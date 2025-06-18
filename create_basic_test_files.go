// create_basic_test_files.go - Simple test file generator that doesn't require external dependencies
// This is a temporary utility to create basic test files when ImageMagick/ExifTool aren't available

package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"path/filepath"
)

func main() {
	// Create testdata directory
	if err := os.MkdirAll("testdata", 0755); err != nil {
		log.Fatalf("Failed to create testdata directory: %v", err)
	}

	fmt.Println("Creating basic test files...")

	// Create basic test images
	createBasicImages()
	createTextChunkPNG()
	createTimeChunkPNG()
	createGammaPNG()
	createPhysPNG()
	createTransparencyPNG()
	createIndexedColorPNG()
	createMixedChunksPNG()
	createMinimalPNG()

	// Create error case files
	createTruncatedPNG()
	createCorruptedCRCPNG()

	fmt.Println("Basic test files created successfully!")
	fmt.Println("Note: For full test coverage, run datacreator with ImageMagick and ExifTool")
}

func createBasicImages() {
	// Create base image
	img := createTestImage(100, 100)
	saveAsPNG(img, "testdata/base.png")
	saveAsPNG(img, "testdata/basic_copy.png")
}

func createTestImage(width, height int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	// Create gradient
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			blue := uint8((255 * y) / height)
			img.Set(x, y, color.RGBA{0, 0, blue, 255})
		}
	}
	return img
}

func saveAsPNG(img image.Image, filename string) {
	file, err := os.Create(filename)
	if err != nil {
		log.Printf("Warning: Failed to create %s: %v", filename, err)
		return
	}
	defer file.Close()

	if err := png.Encode(file, img); err != nil {
		log.Printf("Warning: Failed to encode %s: %v", filename, err)
	}
}

func createTextChunkPNG() {
	img := createTestImage(100, 100)
	
	var buf bytes.Buffer
	buf.Write([]byte{137, 80, 78, 71, 13, 10, 26, 10}) // PNG signature

	// Encode base image to get IHDR and IDAT
	var imgBuf bytes.Buffer
	if err := png.Encode(&imgBuf, img); err != nil {
		log.Printf("Warning: Failed to encode image for text chunk PNG: %v", err)
		return
	}
	imgData := imgBuf.Bytes()

	// Copy IHDR
	copyChunk(&buf, imgData, "IHDR")

	// Add text chunks
	writeTextChunk(&buf, "Comment", "This is a test comment")
	writeTextChunk(&buf, "Copyright", "Copyright 2024 Test")
	writeTextChunk(&buf, "Description", "Test image with text chunks")

	// Copy IDAT and IEND
	copyChunk(&buf, imgData, "IDAT")
	copyChunk(&buf, imgData, "IEND")

	if err := os.WriteFile("testdata/with_text_chunks.png", buf.Bytes(), 0644); err != nil {
		log.Printf("Warning: Failed to write text chunks PNG: %v", err)
	}
}

func createTimeChunkPNG() {
	img := createTestImage(100, 100)
	
	var buf bytes.Buffer
	buf.Write([]byte{137, 80, 78, 71, 13, 10, 26, 10}) // PNG signature

	var imgBuf bytes.Buffer
	if err := png.Encode(&imgBuf, img); err != nil {
		log.Printf("Warning: Failed to encode image for time chunk PNG: %v", err)
		return
	}
	imgData := imgBuf.Bytes()

	copyChunk(&buf, imgData, "IHDR")

	// Add tIME chunk
	timeData := []byte{
		0x07, 0xE8, // Year: 2024
		0x01, // Month: January
		0x01, // Day: 1
		0x00, // Hour: 0
		0x00, // Minute: 0
		0x00, // Second: 0
	}
	writeChunk(&buf, "tIME", timeData)

	copyChunk(&buf, imgData, "IDAT")
	copyChunk(&buf, imgData, "IEND")

	if err := os.WriteFile("testdata/with_time.png", buf.Bytes(), 0644); err != nil {
		log.Printf("Warning: Failed to write time chunk PNG: %v", err)
	}
}

func createGammaPNG() {
	img := createTestImage(100, 100)
	
	var buf bytes.Buffer
	buf.Write([]byte{137, 80, 78, 71, 13, 10, 26, 10}) // PNG signature

	var imgBuf bytes.Buffer
	if err := png.Encode(&imgBuf, img); err != nil {
		log.Printf("Warning: Failed to encode image for gamma PNG: %v", err)
		return
	}
	imgData := imgBuf.Bytes()

	copyChunk(&buf, imgData, "IHDR")

	// Add gAMA chunk (gamma = 2.2 = 45455 in PNG encoding)
	writeChunk(&buf, "gAMA", []byte{0, 0, 0xB1, 0x8F})

	copyChunk(&buf, imgData, "IDAT")
	copyChunk(&buf, imgData, "IEND")

	if err := os.WriteFile("testdata/with_gamma.png", buf.Bytes(), 0644); err != nil {
		log.Printf("Warning: Failed to write gamma PNG: %v", err)
	}
}

func createPhysPNG() {
	img := createTestImage(100, 100)
	
	var buf bytes.Buffer
	buf.Write([]byte{137, 80, 78, 71, 13, 10, 26, 10}) // PNG signature

	var imgBuf bytes.Buffer
	if err := png.Encode(&imgBuf, img); err != nil {
		log.Printf("Warning: Failed to encode image for phys PNG: %v", err)
		return
	}
	imgData := imgBuf.Bytes()

	copyChunk(&buf, imgData, "IHDR")

	// Add pHYs chunk (300 DPI = 11811 pixels per meter)
	physData := make([]byte, 9)
	binary.BigEndian.PutUint32(physData[0:4], 11811) // X pixels per unit
	binary.BigEndian.PutUint32(physData[4:8], 11811) // Y pixels per unit
	physData[8] = 1                                  // Unit is meter
	writeChunk(&buf, "pHYs", physData)

	copyChunk(&buf, imgData, "IDAT")
	copyChunk(&buf, imgData, "IEND")

	if err := os.WriteFile("testdata/with_physical_dims.png", buf.Bytes(), 0644); err != nil {
		log.Printf("Warning: Failed to write physical dimensions PNG: %v", err)
	}
}

func createTransparencyPNG() {
	// Create image with alpha channel
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			alpha := uint8(255)
			if x < 50 && y < 50 {
				alpha = 0 // Top-left quadrant transparent
			}
			blue := uint8((255 * y) / 100)
			img.Set(x, y, color.RGBA{0, 0, blue, alpha})
		}
	}

	saveAsPNG(img, "testdata/with_transparency.png")
}

func createIndexedColorPNG() {
	// Create paletted image
	palette := make([]color.Color, 256)
	for i := 0; i < 256; i++ {
		palette[i] = color.RGBA{uint8(i), 0, uint8(255 - i), 255}
	}

	img := image.NewPaletted(image.Rect(0, 0, 100, 100), palette)
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.SetColorIndex(x, y, uint8((x+y)/2))
		}
	}

	saveAsPNG(img, "testdata/indexed_color.png")
}

func createMixedChunksPNG() {
	img := createTestImage(100, 100)
	
	var buf bytes.Buffer
	buf.Write([]byte{137, 80, 78, 71, 13, 10, 26, 10}) // PNG signature

	var imgBuf bytes.Buffer
	if err := png.Encode(&imgBuf, img); err != nil {
		log.Printf("Warning: Failed to encode image for mixed chunks PNG: %v", err)
		return
	}
	imgData := imgBuf.Bytes()

	copyChunk(&buf, imgData, "IHDR")

	// Add mix of removable and essential chunks
	writeTextChunk(&buf, "Comment", "Mixed test")
	writeChunk(&buf, "gAMA", []byte{0, 0, 0xB1, 0x8F})
	writeChunk(&buf, "tIME", []byte{0x07, 0xE8, 0x01, 0x01, 0x00, 0x00, 0x00})

	copyChunk(&buf, imgData, "IDAT")
	copyChunk(&buf, imgData, "IEND")

	if err := os.WriteFile("testdata/with_mixed_chunks.png", buf.Bytes(), 0644); err != nil {
		log.Printf("Warning: Failed to write mixed chunks PNG: %v", err)
	}
}

func createMinimalPNG() {
	// Create minimal 1x1 image
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{255, 0, 0, 255})
	saveAsPNG(img, "testdata/edge_case_small.png")
}

func createTruncatedPNG() {
	// Create a basic PNG and truncate it
	img := createTestImage(50, 50)
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		log.Printf("Warning: Failed to create base PNG for truncation: %v", err)
		return
	}

	data := buf.Bytes()
	if len(data) > 100 {
		truncated := data[:len(data)-50] // Remove last 50 bytes
		if err := os.WriteFile("testdata/truncated.png", truncated, 0644); err != nil {
			log.Printf("Warning: Failed to write truncated PNG: %v", err)
		}
	}
}

func createCorruptedCRCPNG() {
	// Create a basic PNG and corrupt its CRC
	img := createTestImage(50, 50)
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		log.Printf("Warning: Failed to create base PNG for CRC corruption: %v", err)
		return
	}

	data := buf.Bytes()
	if len(data) > 50 {
		corrupted := make([]byte, len(data))
		copy(corrupted, data)
		// Corrupt the CRC of IHDR (last byte of first 33 bytes)
		corrupted[32] = corrupted[32] ^ 0xFF
		
		if err := os.WriteFile("testdata/corrupted_crc.png", corrupted, 0644); err != nil {
			log.Printf("Warning: Failed to write corrupted CRC PNG: %v", err)
		}
	}
}

// Helper functions
func writeChunk(buf *bytes.Buffer, chunkType string, data []byte) {
	// Write length
	if err := binary.Write(buf, binary.BigEndian, uint32(len(data))); err != nil {
		log.Printf("Warning: Failed to write chunk length: %v", err)
		return
	}

	// Write chunk type
	buf.WriteString(chunkType)

	// Write data
	buf.Write(data)

	// Calculate and write CRC
	crcData := append([]byte(chunkType), data...)
	crc := crc32.ChecksumIEEE(crcData)
	if err := binary.Write(buf, binary.BigEndian, crc); err != nil {
		log.Printf("Warning: Failed to write CRC: %v", err)
	}
}

func writeTextChunk(buf *bytes.Buffer, key, value string) {
	data := append([]byte(key), 0) // null separator
	data = append(data, []byte(value)...)
	writeChunk(buf, "tEXt", data)
}

func copyChunk(dest *bytes.Buffer, src []byte, chunkType string) {
	offset := 8 // Skip PNG signature
	for offset < len(src) {
		if offset+8 > len(src) {
			break
		}

		length := binary.BigEndian.Uint32(src[offset : offset+4])
		chunk := string(src[offset+4 : offset+8])

		if chunk == chunkType {
			// Copy entire chunk including length and CRC
			chunkSize := 12 + int(length)
			dest.Write(src[offset : offset+chunkSize])
			return
		}

		offset += 12 + int(length)
	}

	if chunkType == "IDAT" || chunkType == "IEND" {
		log.Printf("Warning: Required chunk %s not found", chunkType)
	}
}