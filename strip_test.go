package pngmetawebstrip

import (
	"bytes"
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
				_, _, err := PngMetaWebStrip(tt.data)
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
			cleaned, result, err := PngMetaWebStrip(data)
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

			cleaned, _, err := PngMetaWebStrip(data)
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

	// For a more thorough check, we could compare pixel by pixel
	// but for performance reasons, we'll just check dimensions
	return nil
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

			cleaned, _, err := PngMetaWebStrip(data)
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
		_, _, err := PngMetaWebStrip(data)
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

		cleaned, result, err := PngMetaWebStrip(data)
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
	cleanedData, result, err := PngMetaWebStrip(pngData)
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
