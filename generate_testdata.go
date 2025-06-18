// Temporary test data generator - run this to create testdata directory
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
)

func main() {
	// Create testdata directory
	if err := os.MkdirAll("testdata", 0755); err != nil {
		log.Fatalf("Failed to create testdata directory: %v", err)
	}

	// Generate minimal test files needed for tests
	createBasicCopy()
	createWithTextChunks()
	createWithTime()
	createWithBackground()  
	createWithGamma()
	createWithChromaticity()
	createWithSRGB()
	createWithPhysicalDims()
	createIndexedColor()
	createWithTransparency()
	createWithSignificantBits()
	createWithMixedChunks()

	fmt.Println("Test data files created successfully!")
}

func createBasicCopy() {
	img := createTestImage()
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		log.Fatalf("Failed to encode PNG: %v", err)
	}
	writeFile("testdata/basic_copy.png", buf.Bytes())
}

func createWithTextChunks() {
	data := createPNGWithChunks([]chunkSpec{
		{"IHDR", createIHDRData()},
		{"tEXt", createTextData("Comment", "Test comment")},
		{"tEXt", createTextData("Copyright", "Test copyright")},
		{"IDAT", createIDATData()},
		{"IEND", []byte{}},
	})
	writeFile("testdata/with_text_chunks.png", data)
}

func createWithTime() {
	data := createPNGWithChunks([]chunkSpec{
		{"IHDR", createIHDRData()},
		{"tIME", []byte{0x07, 0xE8, 0x01, 0x01, 0x00, 0x00, 0x00}}, // 2024-01-01 00:00:00
		{"IDAT", createIDATData()},
		{"IEND", []byte{}},
	})
	writeFile("testdata/with_time.png", data)
}

func createWithBackground() {
	data := createPNGWithChunks([]chunkSpec{
		{"IHDR", createIHDRData()},
		{"bKGD", []byte{0xFF, 0xFF, 0x00, 0x00, 0x00, 0x00}}, // Red background
		{"IDAT", createIDATData()},
		{"IEND", []byte{}},
	})
	writeFile("testdata/with_background.png", data)
}

func createWithGamma() {
	data := createPNGWithChunks([]chunkSpec{
		{"IHDR", createIHDRData()},
		{"gAMA", []byte{0, 0, 0xB1, 0x8F}}, // Gamma 2.2
		{"IDAT", createIDATData()},
		{"IEND", []byte{}},
	})
	writeFile("testdata/with_gamma.png", data)
}

func createWithChromaticity() {
	chrmData := make([]byte, 32)
	// sRGB chromaticity values
	binary.BigEndian.PutUint32(chrmData[0:4], 31270)   // White point x
	binary.BigEndian.PutUint32(chrmData[4:8], 32900)   // White point y
	binary.BigEndian.PutUint32(chrmData[8:12], 64000)  // Red x
	binary.BigEndian.PutUint32(chrmData[12:16], 33000) // Red y
	binary.BigEndian.PutUint32(chrmData[16:20], 30000) // Green x
	binary.BigEndian.PutUint32(chrmData[20:24], 60000) // Green y
	binary.BigEndian.PutUint32(chrmData[24:28], 15000) // Blue x
	binary.BigEndian.PutUint32(chrmData[28:32], 6000)  // Blue y

	data := createPNGWithChunks([]chunkSpec{
		{"IHDR", createIHDRData()},
		{"cHRM", chrmData},
		{"IDAT", createIDATData()},
		{"IEND", []byte{}},
	})
	writeFile("testdata/with_chromaticity.png", data)
}

func createWithSRGB() {
	data := createPNGWithChunks([]chunkSpec{
		{"IHDR", createIHDRData()},
		{"sRGB", []byte{0}}, // Perceptual rendering intent
		{"IDAT", createIDATData()},
		{"IEND", []byte{}},
	})
	writeFile("testdata/with_srgb.png", data)
}

func createWithPhysicalDims() {
	physData := make([]byte, 9)
	binary.BigEndian.PutUint32(physData[0:4], 11811) // X pixels per unit (300 DPI)
	binary.BigEndian.PutUint32(physData[4:8], 11811) // Y pixels per unit (300 DPI)
	physData[8] = 1                                   // Unit is meter

	data := createPNGWithChunks([]chunkSpec{
		{"IHDR", createIHDRData()},
		{"pHYs", physData},
		{"IDAT", createIDATData()},
		{"IEND", []byte{}},
	})
	writeFile("testdata/with_physical_dims.png", data)
}

func createIndexedColor() {
	// Create simple indexed color image
	palette := make([]color.Color, 4)
	palette[0] = color.RGBA{255, 0, 0, 255}   // Red
	palette[1] = color.RGBA{0, 255, 0, 255}   // Green
	palette[2] = color.RGBA{0, 0, 255, 255}   // Blue
	palette[3] = color.RGBA{255, 255, 0, 255} // Yellow

	img := image.NewPaletted(image.Rect(0, 0, 2, 2), palette)
	img.SetColorIndex(0, 0, 0) // Red
	img.SetColorIndex(1, 0, 1) // Green
	img.SetColorIndex(0, 1, 2) // Blue
	img.SetColorIndex(1, 1, 3) // Yellow

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		log.Fatalf("Failed to encode indexed PNG: %v", err)
	}
	writeFile("testdata/indexed_color.png", buf.Bytes())
}

func createWithTransparency() {
	// Create RGBA image with transparency
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{255, 0, 0, 0})   // Transparent red
	img.Set(1, 0, color.RGBA{0, 255, 0, 255}) // Opaque green
	img.Set(0, 1, color.RGBA{0, 0, 255, 128}) // Semi-transparent blue
	img.Set(1, 1, color.RGBA{255, 255, 0, 255}) // Opaque yellow

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		log.Fatalf("Failed to encode transparent PNG: %v", err)
	}
	writeFile("testdata/with_transparency.png", buf.Bytes())
}

func createWithSignificantBits() {
	data := createPNGWithChunks([]chunkSpec{
		{"IHDR", createIHDRData()},
		{"sBIT", []byte{4, 4, 4, 4}}, // 4 significant bits per channel
		{"IDAT", createIDATData()},
		{"IEND", []byte{}},
	})
	writeFile("testdata/with_significant_bits.png", data)
}

func createWithMixedChunks() {
	data := createPNGWithChunks([]chunkSpec{
		{"IHDR", createIHDRData()},
		{"gAMA", []byte{0, 0, 0xB1, 0x8F}},                     // Keep
		{"tEXt", createTextData("Comment", "Mixed chunks")},    // Remove
		{"bKGD", []byte{0xFF, 0xFF, 0x00, 0x00, 0x00, 0x00}}, // Remove
		{"pHYs", createPhysData()},                             // Keep
		{"tIME", []byte{0x07, 0xE8, 0x01, 0x01, 0x00, 0x00, 0x00}}, // Remove
		{"IDAT", createIDATData()},
		{"IEND", []byte{}},
	})
	writeFile("testdata/with_mixed_chunks.png", data)
}

// Helper functions

type chunkSpec struct {
	Type string
	Data []byte
}

func createTestImage() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{255, 0, 0, 255})   // Red
	img.Set(1, 0, color.RGBA{0, 255, 0, 255})   // Green
	img.Set(0, 1, color.RGBA{0, 0, 255, 255})   // Blue
	img.Set(1, 1, color.RGBA{255, 255, 255, 255}) // White
	return img
}

func createPNGWithChunks(chunks []chunkSpec) []byte {
	var buf bytes.Buffer
	
	// PNG signature
	buf.Write([]byte{137, 80, 78, 71, 13, 10, 26, 10})
	
	// Write chunks
	for _, chunk := range chunks {
		writeChunk(&buf, chunk.Type, chunk.Data)
	}
	
	return buf.Bytes()
}

func createIHDRData() []byte {
	// 2x2 RGB image, 8-bit depth
	return []byte{
		0, 0, 0, 2, // Width: 2
		0, 0, 0, 2, // Height: 2
		8,    // Bit depth: 8
		2,    // Color type: RGB
		0,    // Compression: deflate
		0,    // Filter: none
		0,    // Interlace: none
	}
}

func createIDATData() []byte {
	// Minimal compressed RGB data for 2x2 image
	// This is a simplified IDAT - real images would have proper filtering
	return []byte{
		0x78, 0x9C, // zlib header
		0x01, 0x0D, 0x00, 0xF2, 0xFF, // deflate block
		0x00, 0xFF, 0x00, 0x00, 0xFF, 0x00, 0x00, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, // RGBA data
		0x00, 0x00, 0x00, 0x00, 0x00, 0x01, // Adler-32 checksum
	}
}

func createTextData(key, value string) []byte {
	data := append([]byte(key), 0) // null separator
	data = append(data, []byte(value)...)
	return data
}

func createPhysData() []byte {
	physData := make([]byte, 9)
	binary.BigEndian.PutUint32(physData[0:4], 11811) // X pixels per unit
	binary.BigEndian.PutUint32(physData[4:8], 11811) // Y pixels per unit
	physData[8] = 1                                   // Unit is meter
	return physData
}

func writeChunk(buf *bytes.Buffer, chunkType string, data []byte) {
	// Length
	if err := binary.Write(buf, binary.BigEndian, uint32(len(data))); err != nil {
		log.Fatalf("Failed to write chunk length: %v", err)
	}
	
	// Type
	buf.WriteString(chunkType)
	
	// Data
	buf.Write(data)
	
	// CRC
	crcData := append([]byte(chunkType), data...)
	crc := crc32.ChecksumIEEE(crcData)
	if err := binary.Write(buf, binary.BigEndian, crc); err != nil {
		log.Fatalf("Failed to write CRC: %v", err)
	}
}

func writeFile(filename string, data []byte) {
	if err := os.WriteFile(filename, data, 0644); err != nil {
		log.Fatalf("Failed to write %s: %v", filename, err)
	}
	fmt.Printf("Created %s (%d bytes)\n", filename, len(data))
}