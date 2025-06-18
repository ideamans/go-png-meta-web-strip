package pngmetawebstrip

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"image"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestPngMetaWebStrip(t *testing.T) {
	// Test with invalid data
	t.Run("Invalid data", func(t *testing.T) {
		tests := []struct {
			name string
			data []byte
		}{
			{"Empty data", []byte{}},
			{"Too short", []byte{1, 2, 3}},
			{"Invalid signature", []byte{0, 0, 0, 0, 0, 0, 0, 0}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, _, err := Strip(tt.data)
				if err == nil {
					t.Error("Expected error for invalid data")
				}
			})
		}
	})

	// Test with real PNG files
	testFiles := []struct {
		name             string
		file             string
		expectRemoved    bool
		removedChunkType string
	}{
		{"Basic copy", "basic_copy.png", false, ""},
		{"Remove text chunks", "with_text_chunks.png", true, "text"},
		{"Remove time chunk", "with_time.png", true, "time"},
		{"Remove background", "with_background.png", true, "background"},
		{"Remove EXIF", "with_exif.png", true, "exif"},
		{"Preserve gamma", "with_gamma.png", false, ""},
		{"Preserve chromaticity", "with_chromaticity.png", false, ""},
		{"Preserve sRGB", "with_srgb.png", false, ""},
		{"Preserve physical dimensions", "with_physical_dims.png", false, ""},
		{"Preserve palette", "indexed_color.png", false, ""},
		{"Preserve transparency", "with_transparency.png", false, ""},
		{"Preserve significant bits", "with_significant_bits.png", false, ""},
		{"Remove all removable", "with_all_removable.png", true, "multiple"},
		{"Mixed chunks", "with_mixed_chunks.png", true, "mixed"},
	}

	for _, tt := range testFiles {
		t.Run(tt.name, func(t *testing.T) {
			// Skip if file doesn't exist
			path := filepath.Join("testdata", tt.file)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Skipf("Test file %s not found", path)
			}

			// Read test file
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("Failed to read test file: %v", err)
			}

			// Process the PNG
			cleaned, result, err := Strip(data)
			if err != nil {
				t.Fatalf("Failed to process PNG: %v", err)
			}

			// Validate the result is still a valid PNG
			if err := validatePNG(cleaned); err != nil {
				t.Fatalf("Result is not a valid PNG: %v", err)
			}

			// Check if chunks were removed as expected
			if tt.expectRemoved && result.Total == 0 {
				t.Errorf("Expected chunks to be removed but none were")
			} else if !tt.expectRemoved && result.Total > 0 {
				t.Errorf("Expected no chunks to be removed but %d bytes were removed", result.Total)
			}

			// Verify specific chunk types were removed
			if tt.removedChunkType != "" {
				switch tt.removedChunkType {
				case "text":
					if result.Removed.TextChunks == 0 {
						t.Error("Expected text chunks to be removed")
					}
				case "time":
					if result.Removed.TimeChunk == 0 {
						t.Error("Expected time chunk to be removed")
					}
				case "background":
					if result.Removed.Background == 0 {
						t.Error("Expected background chunk to be removed")
					}
				case "exif":
					if result.Removed.ExifData == 0 {
						t.Error("Expected EXIF data to be removed")
					}
				}
			}

			// Verify image integrity (pixel data unchanged)
			if err := verifyImageIntegrity(data, cleaned); err != nil {
				t.Errorf("Image integrity check failed: %v", err)
			}
		})
	}
}

func TestChunkPreservation(t *testing.T) {
	// Test that essential chunks are preserved
	essentialFiles := map[string]string{
		"gAMA": "with_gamma.png",
		"cHRM": "with_chromaticity.png",
		"sRGB": "with_srgb.png",
		"pHYs": "with_physical_dims.png",
		"PLTE": "indexed_color.png",
		"tRNS": "with_transparency.png",
		"sBIT": "with_significant_bits.png",
	}

	for chunkType, filename := range essentialFiles {
		t.Run(fmt.Sprintf("Preserve %s", chunkType), func(t *testing.T) {
			path := filepath.Join("testdata", filename)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Skipf("Test file %s not found", path)
			}

			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("Failed to read test file: %v", err)
			}

			cleaned, _, err := Strip(data)
			if err != nil {
				t.Fatalf("Failed to process PNG: %v", err)
			}

			// Check if the chunk exists in the cleaned data
			if !hasChunk(cleaned, chunkType) && hasChunk(data, chunkType) {
				t.Errorf("Essential chunk %s was removed", chunkType)
			}
		})
	}
}

func TestPngMetaWebStripReader(t *testing.T) {
	path := filepath.Join("testdata", "with_text_chunks.png")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skip("Test file not found")
	}

	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer file.Close()

	cleaned, result, err := PngMetaWebStripReader(file)
	if err != nil {
		t.Fatalf("Failed to process PNG from reader: %v", err)
	}

	if result.Total == 0 {
		t.Error("Expected chunks to be removed")
	}

	if err := validatePNG(cleaned); err != nil {
		t.Fatalf("Result is not a valid PNG: %v", err)
	}
}

func TestPngMetaWebStripWriter(t *testing.T) {
	path := filepath.Join("testdata", "with_text_chunks.png")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skip("Test file not found")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	var buf bytes.Buffer
	result, err := PngMetaWebStripWriter(data, &buf)
	if err != nil {
		t.Fatalf("Failed to process PNG to writer: %v", err)
	}

	if result.Total == 0 {
		t.Error("Expected chunks to be removed")
	}

	if err := validatePNG(buf.Bytes()); err != nil {
		t.Fatalf("Result is not a valid PNG: %v", err)
	}
}

// Helper functions

func validatePNG(data []byte) error {
	_, err := png.Decode(bytes.NewReader(data))
	return err
}

func hasChunk(data []byte, chunkType string) bool {
	if len(data) < 8 {
		return false
	}

	offset := 8 // Skip PNG signature
	for offset < len(data) {
		if offset+8 > len(data) {
			break
		}

		length := int(binary.BigEndian.Uint32(data[offset : offset+4]))
		if offset+12+length > len(data) {
			break
		}

		chunk := string(data[offset+4 : offset+8])
		if chunk == chunkType {
			return true
		}

		offset += 12 + length
	}

	return false
}

func verifyImageIntegrity(original, cleaned []byte) error {
	// Decode original image
	origImg, err := png.Decode(bytes.NewReader(original))
	if err != nil {
		return fmt.Errorf("failed to decode original: %w", err)
	}

	// Decode cleaned image
	cleanImg, err := png.Decode(bytes.NewReader(cleaned))
	if err != nil {
		return fmt.Errorf("failed to decode cleaned: %w", err)
	}

	// Compare bounds
	if !origImg.Bounds().Eq(cleanImg.Bounds()) {
		return fmt.Errorf("image bounds differ")
	}

	// Calculate checksum of original image pixels
	origChecksum, err := calculateImageChecksum(origImg)
	if err != nil {
		return fmt.Errorf("failed to calculate original checksum: %w", err)
	}

	// Calculate checksum of cleaned image pixels
	cleanChecksum, err := calculateImageChecksum(cleanImg)
	if err != nil {
		return fmt.Errorf("failed to calculate cleaned checksum: %w", err)
	}

	// Compare checksums
	if origChecksum != cleanChecksum {
		return fmt.Errorf("image pixel data differs: original=%s, cleaned=%s", origChecksum, cleanChecksum)
	}

	return nil
}

func calculateImageChecksum(img image.Image) (string, error) {
	bounds := img.Bounds()
	hasher := md5.New()

	// Write image dimensions to hasher
	if err := binary.Write(hasher, binary.BigEndian, int32(bounds.Dx())); err != nil {
		return "", err
	}
	if err := binary.Write(hasher, binary.BigEndian, int32(bounds.Dy())); err != nil {
		return "", err
	}

	// Write all pixels to hasher
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			if err := binary.Write(hasher, binary.BigEndian, uint16(r)); err != nil {
				return "", err
			}
			if err := binary.Write(hasher, binary.BigEndian, uint16(g)); err != nil {
				return "", err
			}
			if err := binary.Write(hasher, binary.BigEndian, uint16(b)); err != nil {
				return "", err
			}
			if err := binary.Write(hasher, binary.BigEndian, uint16(a)); err != nil {
				return "", err
			}
		}
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

// Integration test using external tools
func TestExternalValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping external validation in short mode")
	}

	// Check if pngcheck is available
	if _, err := exec.LookPath("pngcheck"); err != nil {
		t.Skip("pngcheck not found, skipping external validation")
	}

	testFiles := []string{
		"with_text_chunks.png",
		"with_mixed_chunks.png",
		"indexed_color.png",
	}

	for _, filename := range testFiles {
		t.Run(filename, func(t *testing.T) {
			path := filepath.Join("testdata", filename)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Skipf("Test file %s not found", path)
			}

			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("Failed to read test file: %v", err)
			}

			cleaned, _, err := Strip(data)
			if err != nil {
				t.Fatalf("Failed to process PNG: %v", err)
			}

			// Write cleaned file to temp
			tmpfile, err := os.CreateTemp("", "cleaned-*.png")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpfile.Name())

			if _, err := tmpfile.Write(cleaned); err != nil {
				t.Fatalf("Failed to write temp file: %v", err)
			}
			tmpfile.Close()

			// Run pngcheck
			cmd := exec.Command("pngcheck", "-v", tmpfile.Name()) // #nosec G204 -- pngcheck is a trusted tool
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Errorf("pngcheck failed: %v\nOutput: %s", err, output)
			}

			// Check that removed chunks are not present
			outputStr := string(output)
			if strings.Contains(outputStr, "tEXt") ||
				strings.Contains(outputStr, "zTXt") ||
				strings.Contains(outputStr, "iTXt") {
				if filename == "with_text_chunks.png" {
					t.Error("Text chunks were not removed")
				}
			}
		})
	}
}

// Benchmark
func BenchmarkPngMetaWebStrip(b *testing.B) {
	path := filepath.Join("testdata", "with_mixed_chunks.png")
	data, err := os.ReadFile(path)
	if err != nil {
		b.Skip("Test file not found")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := Strip(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Helper to generate test report
func TestGenerateReport(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping report generation in short mode")
	}

	files, err := os.ReadDir("testdata")
	if err != nil {
		t.Skip("testdata directory not found")
	}

	fmt.Println("\n=== PNG Metadata Removal Report ===")
	fmt.Printf("%-30s | %-10s | %-10s | %-10s | %-30s\n", "File", "Original", "Cleaned", "Removed", "Removed Chunks")
	fmt.Println(strings.Repeat("-", 100))

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".png") || file.Name() == "base.png" {
			continue
		}

		path := filepath.Join("testdata", file.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		cleaned, result, err := Strip(data)
		if err != nil {
			fmt.Printf("%-30s | ERROR: %v\n", file.Name(), err)
			continue
		}

		removedTypes := []string{}
		if result.Removed.TextChunks > 0 {
			removedTypes = append(removedTypes, "text")
		}
		if result.Removed.TimeChunk > 0 {
			removedTypes = append(removedTypes, "time")
		}
		if result.Removed.Background > 0 {
			removedTypes = append(removedTypes, "background")
		}
		if result.Removed.ExifData > 0 {
			removedTypes = append(removedTypes, "exif")
		}
		if result.Removed.OtherChunks > 0 {
			removedTypes = append(removedTypes, "other")
		}

		fmt.Printf("%-30s | %-10d | %-10d | %-10d | %-30s\n",
			file.Name(),
			len(data),
			len(cleaned),
			result.Total,
			strings.Join(removedTypes, ", "))
	}
}

// Test edge cases and error conditions for better coverage
func TestStripEdgeCases(t *testing.T) {
	t.Run("Chunk length extends beyond data", func(t *testing.T) {
		// PNG signature + chunk with length larger than remaining data
		data := []byte{137, 80, 78, 71, 13, 10, 26, 10} // PNG signature
		data = append(data, 0xFF, 0xFF, 0xFF, 0xFF)      // Huge chunk length
		data = append(data, 'I', 'H', 'D', 'R')          // Chunk type
		data = append(data, []byte{1, 2, 3, 4}...)       // Insufficient data
		
		_, _, err := Strip(data)
		if err == nil {
			t.Error("Expected error for chunk extending beyond data")
		}
		if !strings.Contains(err.Error(), "chunk extends beyond data") {
			t.Errorf("Expected 'chunk extends beyond data' error, got: %v", err)
		}
	})

	t.Run("Incomplete chunk at end", func(t *testing.T) {
		// PNG signature + incomplete chunk header
		data := []byte{137, 80, 78, 71, 13, 10, 26, 10} // PNG signature
		data = append(data, 0, 0, 0, 4)                  // Chunk length
		data = append(data, 'I', 'H')                    // Incomplete chunk type
		
		_, _, err := Strip(data)
		if err == nil {
			t.Error("Expected error for incomplete chunk")
		}
		if !strings.Contains(err.Error(), "incomplete chunk") {
			t.Errorf("Expected 'incomplete chunk' error, got: %v", err)
		}
	})

	t.Run("Invalid CRC", func(t *testing.T) {
		// Create a valid PNG chunk structure but with wrong CRC
		data := []byte{137, 80, 78, 71, 13, 10, 26, 10} // PNG signature
		data = append(data, 0, 0, 0, 4)                  // Chunk length = 4
		data = append(data, 'I', 'H', 'D', 'R')          // Chunk type
		data = append(data, 1, 2, 3, 4)                  // Chunk data
		data = append(data, 0xFF, 0xFF, 0xFF, 0xFF)      // Wrong CRC
		
		_, _, err := Strip(data)
		if err == nil {
			t.Error("Expected error for invalid CRC")
		}
		if !strings.Contains(err.Error(), "invalid CRC") {
			t.Errorf("Expected 'invalid CRC' error, got: %v", err)
		}
	})
}

func TestShouldKeepChunkFunction(t *testing.T) {
	tests := []struct {
		chunkType string
		expected  bool
	}{
		// Essential chunks that should be kept
		{"IHDR", true},
		{"PLTE", true},
		{"IDAT", true},
		{"IEND", true},
		{"tRNS", true},
		{"gAMA", true},
		{"cHRM", true},
		{"sRGB", true},
		{"iCCP", true},
		{"sBIT", true},
		{"pHYs", true},
		
		// Chunks that should be removed
		{"tEXt", false},
		{"zTXt", false},
		{"iTXt", false},
		{"tIME", false},
		{"bKGD", false},
		{"eXIf", false},
		{"sPLT", false},
		{"hIST", false},
		{"unknown", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Chunk_%s", tt.chunkType), func(t *testing.T) {
			result := shouldKeepChunk(tt.chunkType)
			if result != tt.expected {
				t.Errorf("shouldKeepChunk(%s) = %v, expected %v", tt.chunkType, result, tt.expected)
			}
		})
	}
}

func TestTrackRemovedChunkFunction(t *testing.T) {
	tests := []struct {
		chunkType string
		size      int
		checkFunc func(*Result) bool
	}{
		{"tEXt", 100, func(r *Result) bool { return r.Removed.TextChunks == 100 }},
		{"zTXt", 200, func(r *Result) bool { return r.Removed.TextChunks == 200 }},
		{"iTXt", 300, func(r *Result) bool { return r.Removed.TextChunks == 300 }},
		{"tIME", 50, func(r *Result) bool { return r.Removed.TimeChunk == 50 }},
		{"bKGD", 75, func(r *Result) bool { return r.Removed.Background == 75 }},
		{"eXIf", 150, func(r *Result) bool { return r.Removed.ExifData == 150 }},
		{"sPLT", 80, func(r *Result) bool { return r.Removed.OtherChunks == 80 }},
		{"hIST", 90, func(r *Result) bool { return r.Removed.OtherChunks == 90 }},
		{"unkn", 120, func(r *Result) bool { return r.Removed.OtherChunks == 120 }},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Track_%s", tt.chunkType), func(t *testing.T) {
			result := &Result{}
			trackRemovedChunk(result, tt.chunkType, tt.size)
			
			if result.Total != tt.size {
				t.Errorf("Expected total %d, got %d", tt.size, result.Total)
			}
			
			if !tt.checkFunc(result) {
				t.Errorf("Chunk tracking failed for %s", tt.chunkType)
			}
		})
	}
	
	// Test cumulative tracking
	t.Run("Cumulative tracking", func(t *testing.T) {
		result := &Result{}
		trackRemovedChunk(result, "tEXt", 100)
		trackRemovedChunk(result, "tIME", 50)
		trackRemovedChunk(result, "bKGD", 75)
		trackRemovedChunk(result, "eXIf", 150)
		trackRemovedChunk(result, "sPLT", 80)
		
		expectedTotal := 100 + 50 + 75 + 150 + 80
		if result.Total != expectedTotal {
			t.Errorf("Expected total %d, got %d", expectedTotal, result.Total)
		}
		
		if result.Removed.TextChunks != 100 {
			t.Errorf("Expected TextChunks %d, got %d", 100, result.Removed.TextChunks)
		}
		if result.Removed.TimeChunk != 50 {
			t.Errorf("Expected TimeChunk %d, got %d", 50, result.Removed.TimeChunk)
		}
		if result.Removed.Background != 75 {
			t.Errorf("Expected Background %d, got %d", 75, result.Removed.Background)
		}
		if result.Removed.ExifData != 150 {
			t.Errorf("Expected ExifData %d, got %d", 150, result.Removed.ExifData)
		}
		if result.Removed.OtherChunks != 80 {
			t.Errorf("Expected OtherChunks %d, got %d", 80, result.Removed.OtherChunks)
		}
	})
}

func TestPngMetaWebStripReaderErrorHandling(t *testing.T) {
	t.Run("Read error", func(t *testing.T) {
		// Create a failing reader
		failingReader := &failingReader{}
		
		_, _, err := PngMetaWebStripReader(failingReader)
		if err == nil {
			t.Error("Expected error from failing reader")
		}
		if !strings.Contains(err.Error(), "failed to read data") {
			t.Errorf("Expected 'failed to read data' error, got: %v", err)
		}
	})
}

func TestPngMetaWebStripWriterErrorHandling(t *testing.T) {
	t.Run("Write error", func(t *testing.T) {
		// Create a simple valid PNG
		img := image.NewRGBA(image.Rect(0, 0, 1, 1))
		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err != nil {
			t.Fatalf("Failed to create test image: %v", err)
		}
		
		// Create a failing writer
		failingWriter := &failingWriter{}
		
		_, err := PngMetaWebStripWriter(buf.Bytes(), failingWriter)
		if err == nil {
			t.Error("Expected error from failing writer")
		}
		if !strings.Contains(err.Error(), "failed to write data") {
			t.Errorf("Expected 'failed to write data' error, got: %v", err)
		}
	})
	
	t.Run("Strip error propagation", func(t *testing.T) {
		// Use invalid PNG data to trigger Strip error
		invalidData := []byte{1, 2, 3, 4}
		var buf bytes.Buffer
		
		_, err := PngMetaWebStripWriter(invalidData, &buf)
		if err == nil {
			t.Error("Expected error from invalid PNG data")
		}
	})
}

func TestComplexPNGProcessing(t *testing.T) {
	t.Run("PNG with multiple chunk types", func(t *testing.T) {
		// Create a PNG with various chunks using manual construction
		data := createTestPNGWithChunks([]chunkInfo{
			{"IHDR", []byte{0, 0, 0, 1, 0, 0, 0, 1, 8, 2, 0, 0, 0}}, // 1x1 RGB
			{"gAMA", []byte{0, 0, 0xB1, 0x8F}},                       // Gamma (keep)
			{"tEXt", []byte("Comment\x00Test comment")},              // Text (remove)
			{"bKGD", []byte{0xFF, 0xFF, 0x00, 0x00, 0x00, 0x00}},    // Background (remove)
			{"tIME", []byte{0x07, 0xE8, 0x01, 0x01, 0x00, 0x00, 0x00}}, // Time (remove)
			{"IDAT", []byte{0x78, 0x9C, 0x62, 0x00, 0x00, 0x00, 0x02, 0x00, 0x01}}, // Minimal image data
			{"IEND", []byte{}}, // End
		})
		
		cleaned, result, err := Strip(data)
		if err != nil {
			t.Fatalf("Failed to process complex PNG: %v", err)
		}
		
		// Should have removed text, background, and time chunks
		if result.Removed.TextChunks == 0 {
			t.Error("Expected text chunks to be removed")
		}
		if result.Removed.Background == 0 {
			t.Error("Expected background chunk to be removed")
		}
		if result.Removed.TimeChunk == 0 {
			t.Error("Expected time chunk to be removed")
		}
		
		// Should still be valid PNG
		if err := validatePNG(cleaned); err != nil {
			t.Errorf("Cleaned PNG is invalid: %v", err)
		}
		
		// Should still contain gamma chunk
		if !hasChunk(cleaned, "gAMA") {
			t.Error("Expected gamma chunk to be preserved")
		}
		
		// Should not contain removed chunks
		if hasChunk(cleaned, "tEXt") {
			t.Error("Text chunk should have been removed")
		}
		if hasChunk(cleaned, "bKGD") {
			t.Error("Background chunk should have been removed")
		}
		if hasChunk(cleaned, "tIME") {
			t.Error("Time chunk should have been removed")
		}
	})
}

// Test checksum calculation specifically
func TestImageChecksumVerification(t *testing.T) {
	// Load a test image with metadata
	path := filepath.Join("testdata", "with_text_chunks.png")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skip("Test file not found")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	// Process the image
	cleaned, _, err := Strip(data)
	if err != nil {
		t.Fatalf("Failed to process PNG: %v", err)
	}

	// Decode both images
	origImg, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Failed to decode original: %v", err)
	}

	cleanImg, err := png.Decode(bytes.NewReader(cleaned))
	if err != nil {
		t.Fatalf("Failed to decode cleaned: %v", err)
	}

	// Calculate checksums
	origChecksum, err := calculateImageChecksum(origImg)
	if err != nil {
		t.Fatalf("Failed to calculate original checksum: %v", err)
	}

	cleanChecksum, err := calculateImageChecksum(cleanImg)
	if err != nil {
		t.Fatalf("Failed to calculate cleaned checksum: %v", err)
	}

	// Checksums should be identical
	if origChecksum != cleanChecksum {
		t.Errorf("Checksums differ: original=%s, cleaned=%s", origChecksum, cleanChecksum)
	}

	t.Logf("Checksum verification successful: %s", origChecksum)
}

// Helper types and functions for additional tests

type failingReader struct{}

func (f *failingReader) Read([]byte) (int, error) {
	return 0, fmt.Errorf("simulated read error")
}

type failingWriter struct{}

func (f *failingWriter) Write([]byte) (int, error) {
	return 0, fmt.Errorf("simulated write error")
}

type chunkInfo struct {
	Type string
	Data []byte
}

func createTestPNGWithChunks(chunks []chunkInfo) []byte {
	var buf bytes.Buffer
	
	// Write PNG signature
	buf.Write([]byte{137, 80, 78, 71, 13, 10, 26, 10})
	
	// Write all chunks
	for _, chunk := range chunks {
		writeTestChunk(&buf, chunk.Type, chunk.Data)
	}
	
	return buf.Bytes()
}

func writeTestChunk(buf *bytes.Buffer, chunkType string, data []byte) {
	// Write length
	if err := binary.Write(buf, binary.BigEndian, uint32(len(data))); err != nil {
		panic(fmt.Sprintf("Failed to write chunk length: %v", err))
	}
	
	// Write chunk type
	buf.WriteString(chunkType)
	
	// Write data
	buf.Write(data)
	
	// Calculate and write CRC
	crcData := append([]byte(chunkType), data...)
	crc := crc32.ChecksumIEEE(crcData)
	if err := binary.Write(buf, binary.BigEndian, crc); err != nil {
		panic(fmt.Sprintf("Failed to write CRC: %v", err))
	}
}

// Test to ensure README examples work
func TestReadmeExample(t *testing.T) {
	// Create a simple test PNG
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("Failed to create test image: %v", err)
	}

	pngData := buf.Bytes()

	// Remove unnecessary metadata
	cleanedData, result, err := Strip(pngData)
	if err != nil {
		t.Fatalf("Failed to process PNG: %v", err)
	}

	// The example should work without errors
	if len(cleanedData) == 0 {
		t.Error("Cleaned data is empty")
	}

	if result == nil {
		t.Error("Result is nil")
	}

	// Basic PNG should not have metadata to remove
	if result != nil && result.Total != 0 {
		t.Logf("Removed %d bytes (unexpected for basic PNG)", result.Total)
	}
}
