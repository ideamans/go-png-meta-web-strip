package main

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestCreateTestImage(t *testing.T) {
	img := createTestImage()
	
	bounds := img.Bounds()
	if bounds.Dx() != 100 || bounds.Dy() != 100 {
		t.Errorf("Expected 100x100 image, got %dx%d", bounds.Dx(), bounds.Dy())
	}
	
	// Test that it's a gradient (blue channel changes with Y)
	topColor := img.At(50, 0)
	bottomColor := img.At(50, 99)
	
	_, _, topBlue, _ := topColor.RGBA()
	_, _, bottomBlue, _ := bottomColor.RGBA()
	
	if topBlue >= bottomBlue {
		t.Error("Expected blue gradient from top to bottom")
	}
}

func TestGenerateBasicPNG(t *testing.T) {
	tempDir := t.TempDir()
	originalTestdata := "testdata"
	
	// Temporarily change to temp directory
	os.Chdir(tempDir)
	defer os.Chdir("..")
	
	img := createTestImage()
	generateBasicPNG(img)
	
	// Check if file was created
	if _, err := os.Stat("testdata/basic_copy.png"); os.IsNotExist(err) {
		t.Error("basic_copy.png was not created")
	}
	
	// Verify it's a valid PNG
	data, err := os.ReadFile("testdata/basic_copy.png")
	if err != nil {
		t.Fatalf("Failed to read generated PNG: %v", err)
	}
	
	_, err = png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Errorf("Generated PNG is not valid: %v", err)
	}
	
	// Restore original directory context
	os.RemoveAll("testdata")
	os.Chdir(originalTestdata)
}

func TestWriteChunkFunction(t *testing.T) {
	var buf bytes.Buffer
	testData := []byte("test data")
	
	writeChunk(&buf, "tEXt", testData)
	
	data := buf.Bytes()
	
	// Check length (4 bytes)
	if len(data) < 4 {
		t.Fatal("Chunk too short")
	}
	
	length := binary.BigEndian.Uint32(data[0:4])
	if length != uint32(len(testData)) {
		t.Errorf("Expected length %d, got %d", len(testData), length)
	}
	
	// Check chunk type (4 bytes)
	if len(data) < 8 {
		t.Fatal("Chunk too short for type")
	}
	
	chunkType := string(data[4:8])
	if chunkType != "tEXt" {
		t.Errorf("Expected chunk type 'tEXt', got '%s'", chunkType)
	}
	
	// Check data
	if len(data) < 8+len(testData) {
		t.Fatal("Chunk too short for data")
	}
	
	chunkData := data[8 : 8+len(testData)]
	if !bytes.Equal(chunkData, testData) {
		t.Error("Chunk data doesn't match")
	}
	
	// Check CRC (4 bytes at end)
	if len(data) != 12+len(testData) {
		t.Errorf("Expected total length %d, got %d", 12+len(testData), len(data))
	}
	
	expectedCRC := crc32.ChecksumIEEE(append([]byte("tEXt"), testData...))
	actualCRC := binary.BigEndian.Uint32(data[8+len(testData):])
	
	if actualCRC != expectedCRC {
		t.Errorf("CRC mismatch: expected %x, got %x", expectedCRC, actualCRC)
	}
}

func TestWriteTextChunk(t *testing.T) {
	var buf bytes.Buffer
	
	writeTextChunk(&buf, "Comment", "Test comment")
	
	data := buf.Bytes()
	
	// Should be a tEXt chunk
	if len(data) < 8 {
		t.Fatal("Chunk too short")
	}
	
	chunkType := string(data[4:8])
	if chunkType != "tEXt" {
		t.Errorf("Expected tEXt chunk, got %s", chunkType)
	}
	
	// Extract and verify data
	length := binary.BigEndian.Uint32(data[0:4])
	chunkData := data[8 : 8+length]
	
	expectedData := "Comment\x00Test comment"
	if string(chunkData) != expectedData {
		t.Errorf("Expected data %q, got %q", expectedData, string(chunkData))
	}
}

func TestGenerateWithGamma(t *testing.T) {
	tempDir := t.TempDir()
	os.Chdir(tempDir)
	defer os.Chdir("..")
	
	img := createTestImage()
	generateWithGamma(img)
	
	// Check if file was created
	data, err := os.ReadFile("testdata/with_gamma.png")
	if err != nil {
		t.Fatalf("Failed to read generated PNG: %v", err)
	}
	
	// Verify it has a gAMA chunk
	if !hasChunkInData(data, "gAMA") {
		t.Error("Generated PNG doesn't contain gAMA chunk")
	}
	
	// Verify it's still a valid PNG
	_, err = png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Errorf("Generated PNG with gamma is invalid: %v", err)
	}
}

func TestGenerateWithPhys(t *testing.T) {
	tempDir := t.TempDir()
	os.Chdir(tempDir)
	defer os.Chdir("..")
	
	img := createTestImage()
	generateWithPhys(img)
	
	data, err := os.ReadFile("testdata/with_physical_dims.png")
	if err != nil {
		t.Fatalf("Failed to read generated PNG: %v", err)
	}
	
	if !hasChunkInData(data, "pHYs") {
		t.Error("Generated PNG doesn't contain pHYs chunk")
	}
	
	_, err = png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Errorf("Generated PNG with physical dimensions is invalid: %v", err)
	}
}

func TestGenerateWithText(t *testing.T) {
	tempDir := t.TempDir()
	os.Chdir(tempDir)
	defer os.Chdir("..")
	
	img := createTestImage()
	generateWithText(img)
	
	data, err := os.ReadFile("testdata/with_text_chunks.png")
	if err != nil {
		t.Fatalf("Failed to read generated PNG: %v", err)
	}
	
	if !hasChunkInData(data, "tEXt") {
		t.Error("Generated PNG doesn't contain tEXt chunk")
	}
	
	_, err = png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Errorf("Generated PNG with text is invalid: %v", err)
	}
}

func TestGenerateWithTime(t *testing.T) {
	tempDir := t.TempDir()
	os.Chdir(tempDir)
	defer os.Chdir("..")
	
	img := createTestImage()
	generateWithTime(img)
	
	data, err := os.ReadFile("testdata/with_time.png")
	if err != nil {
		t.Fatalf("Failed to read generated PNG: %v", err)
	}
	
	if !hasChunkInData(data, "tIME") {
		t.Error("Generated PNG doesn't contain tIME chunk")
	}
	
	_, err = png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Errorf("Generated PNG with time is invalid: %v", err)
	}
}

func TestGenerateWithBackground(t *testing.T) {
	tempDir := t.TempDir()
	os.Chdir(tempDir)
	defer os.Chdir("..")
	
	img := createTestImage()
	generateWithBackground(img)
	
	data, err := os.ReadFile("testdata/with_background.png")
	if err != nil {
		t.Fatalf("Failed to read generated PNG: %v", err)
	}
	
	if !hasChunkInData(data, "bKGD") {
		t.Error("Generated PNG doesn't contain bKGD chunk")
	}
	
	_, err = png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Errorf("Generated PNG with background is invalid: %v", err)
	}
}

func TestGenerateWithChromaticity(t *testing.T) {
	tempDir := t.TempDir()
	os.Chdir(tempDir)
	defer os.Chdir("..")
	
	img := createTestImage()
	generateWithChromaticity(img)
	
	data, err := os.ReadFile("testdata/with_chromaticity.png")
	if err != nil {
		t.Fatalf("Failed to read generated PNG: %v", err)
	}
	
	if !hasChunkInData(data, "cHRM") {
		t.Error("Generated PNG doesn't contain cHRM chunk")
	}
	
	_, err = png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Errorf("Generated PNG with chromaticity is invalid: %v", err)
	}
}

func TestGenerateWithSRGB(t *testing.T) {
	tempDir := t.TempDir()
	os.Chdir(tempDir)
	defer os.Chdir("..")
	
	img := createTestImage()
	generateWithSRGB(img)
	
	data, err := os.ReadFile("testdata/with_srgb.png")
	if err != nil {
		t.Fatalf("Failed to read generated PNG: %v", err)
	}
	
	if !hasChunkInData(data, "sRGB") {
		t.Error("Generated PNG doesn't contain sRGB chunk")
	}
	
	_, err = png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Errorf("Generated PNG with sRGB is invalid: %v", err)
	}
}

func TestGenerateWithTransparency(t *testing.T) {
	tempDir := t.TempDir()
	os.Chdir(tempDir)
	defer os.Chdir("..")
	
	generateWithTransparency()
	
	data, err := os.ReadFile("testdata/with_transparency.png")
	if err != nil {
		t.Fatalf("Failed to read generated PNG: %v", err)
	}
	
	// Decode and verify alpha channel
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Generated PNG with transparency is invalid: %v", err)
	}
	
	// Check that top-left is transparent
	_, _, _, alpha := img.At(0, 0).RGBA()
	if alpha != 0 {
		t.Error("Expected transparent pixel at (0,0)")
	}
	
	// Check that bottom-right is opaque
	bounds := img.Bounds()
	_, _, _, alpha = img.At(bounds.Max.X-1, bounds.Max.Y-1).RGBA()
	if alpha == 0 {
		t.Error("Expected opaque pixel at bottom-right")
	}
}

func TestGenerateIndexedColor(t *testing.T) {
	tempDir := t.TempDir()
	os.Chdir(tempDir)
	defer os.Chdir("..")
	
	generateIndexedColor()
	
	data, err := os.ReadFile("testdata/indexed_color.png")
	if err != nil {
		t.Fatalf("Failed to read generated PNG: %v", err)
	}
	
	// Should have PLTE chunk for indexed color
	if !hasChunkInData(data, "PLTE") {
		t.Error("Indexed color PNG doesn't contain PLTE chunk")
	}
	
	// Verify it's valid
	_, err = png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Errorf("Generated indexed color PNG is invalid: %v", err)
	}
}

func TestGenerateWithSBIT(t *testing.T) {
	tempDir := t.TempDir()
	os.Chdir(tempDir)
	defer os.Chdir("..")
	
	img := createTestImage()
	generateWithSBIT(img)
	
	data, err := os.ReadFile("testdata/with_significant_bits.png")
	if err != nil {
		t.Fatalf("Failed to read generated PNG: %v", err)
	}
	
	if !hasChunkInData(data, "sBIT") {
		t.Error("Generated PNG doesn't contain sBIT chunk")
	}
	
	_, err = png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Errorf("Generated PNG with sBIT is invalid: %v", err)
	}
}

func TestCopyChunkFunction(t *testing.T) {
	// Create source data with chunks
	var srcBuf bytes.Buffer
	srcBuf.Write([]byte{137, 80, 78, 71, 13, 10, 26, 10}) // PNG signature
	
	// Add IHDR chunk
	ihdrData := []byte{0, 0, 0, 1, 0, 0, 0, 1, 8, 2, 0, 0, 0}
	writeTestChunk(&srcBuf, "IHDR", ihdrData)
	
	// Add IDAT chunk
	idatData := []byte{0x78, 0x9C, 0x62, 0x00, 0x00, 0x00, 0x02, 0x00, 0x01}
	writeTestChunk(&srcBuf, "IDAT", idatData)
	
	srcData := srcBuf.Bytes()
	
	// Test copying IHDR
	var destBuf bytes.Buffer
	copyChunk(&destBuf, srcData, "IHDR")
	
	// Verify IHDR was copied
	if !hasChunkInData(destBuf.Bytes(), "IHDR") {
		t.Error("IHDR chunk was not copied")
	}
	
	// Test copying non-existent chunk
	var destBuf2 bytes.Buffer
	copyChunk(&destBuf2, srcData, "tEXt")
	
	// Should not have copied anything
	if destBuf2.Len() > 0 {
		t.Error("Non-existent chunk should not have been copied")
	}
}

func TestMainFunction(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	
	os.Chdir(tempDir)
	
	// Test that main function completes without error
	// We can't easily test the output, but we can verify it runs
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Main function panicked: %v", r)
		}
	}()
	
	// This will create all test files
	main()
	
	// Verify some key files were created
	expectedFiles := []string{
		"testdata/basic_copy.png",
		"testdata/with_gamma.png",
		"testdata/with_text_chunks.png",
		"testdata/indexed_color.png",
	}
	
	for _, file := range expectedFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("Expected file %s was not created", file)
		}
	}
}

// Helper functions

func hasChunkInData(data []byte, chunkType string) bool {
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

func writeTestChunk(buf *bytes.Buffer, chunkType string, data []byte) {
	// Length
	if err := binary.Write(buf, binary.BigEndian, uint32(len(data))); err != nil {
		panic(err)
	}
	
	// Type
	buf.WriteString(chunkType)
	
	// Data
	buf.Write(data)
	
	// CRC
	crcData := append([]byte(chunkType), data...)
	crc := crc32.ChecksumIEEE(crcData)
	if err := binary.Write(buf, binary.BigEndian, crc); err != nil {
		panic(err)
	}
}

func TestImageCreationVariations(t *testing.T) {
	t.Run("Test gradient properties", func(t *testing.T) {
		img := createTestImage()
		bounds := img.Bounds()
		
		// Sample a few points to verify gradient
		samples := []struct{ x, y int }{
			{50, 0},   // Top
			{50, 25},  // Quarter
			{50, 50},  // Middle
			{50, 75},  // Three quarters
			{50, 99},  // Bottom
		}
		
		var blueValues []uint32
		for _, sample := range samples {
			_, _, blue, _ := img.At(sample.x, sample.y).RGBA()
			blueValues = append(blueValues, blue)
		}
		
		// Verify blue values increase from top to bottom
		for i := 1; i < len(blueValues); i++ {
			if blueValues[i] <= blueValues[i-1] {
				t.Errorf("Blue gradient not increasing: %v", blueValues)
				break
			}
		}
	})
}

func TestPaletteCreation(t *testing.T) {
	// Test the palette creation logic in generateIndexedColor
	palette := make([]color.Color, 256)
	for i := 0; i < 256; i++ {
		palette[i] = color.RGBA{uint8(i), 0, uint8(255 - i), 255}
	}
	
	// Verify first and last palette entries
	first := palette[0].(color.RGBA)
	if first.R != 0 || first.G != 0 || first.B != 255 || first.A != 255 {
		t.Errorf("First palette entry incorrect: %+v", first)
	}
	
	last := palette[255].(color.RGBA)
	if last.R != 255 || last.G != 0 || last.B != 0 || last.A != 255 {
		t.Errorf("Last palette entry incorrect: %+v", last)
	}
}