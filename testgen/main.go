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
	// Create test directory
	if err := os.MkdirAll("testdata", 0755); err != nil {
		log.Fatalf("Failed to create testdata directory: %v", err)
	}

	// Create base image
	img := createTestImage()
	
	// Generate test files
	generateBasicPNG(img)
	generateWithGamma(img)
	generateWithPhys(img)
	generateWithText(img)
	generateWithTime(img)
	generateWithBackground(img)
	generateWithChromaticity(img)
	generateWithSRGB(img)
	generateWithTransparency()
	generateIndexedColor()
	generateWithSBIT(img)
	
	fmt.Println("Test data generation complete!")
}

func createTestImage() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	// Create gradient
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			blue := uint8((255 * y) / 100)
			img.Set(x, y, color.RGBA{0, 0, blue, 255})
		}
	}
	return img
}

func generateBasicPNG(img image.Image) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		log.Fatalf("Failed to encode PNG: %v", err)
	}
	if err := os.WriteFile("testdata/basic_copy.png", buf.Bytes(), 0644); err != nil {
		log.Fatalf("Failed to write file: %v", err)
	}
}

func generateWithGamma(img image.Image) {
	var buf bytes.Buffer
	
	// Write PNG signature
	buf.Write([]byte{137, 80, 78, 71, 13, 10, 26, 10})
	
	// Encode image to get IHDR and IDAT
	var imgBuf bytes.Buffer
	if err := png.Encode(&imgBuf, img); err != nil {
		log.Fatalf("Failed to encode PNG: %v", err)
	}
	imgData := imgBuf.Bytes()
	
	// Copy IHDR
	copyChunk(&buf, imgData, "IHDR")
	
	// Add gAMA chunk (gamma = 2.2 = 45455 in PNG encoding)
	writeChunk(&buf, "gAMA", []byte{0, 0, 0xB1, 0x8F})
	
	// Copy IDAT and IEND
	copyChunk(&buf, imgData, "IDAT")
	copyChunk(&buf, imgData, "IEND")
	
	if err := os.WriteFile("testdata/with_gamma.png", buf.Bytes(), 0644); err != nil {
		log.Fatalf("Failed to write file: %v", err)
	}
}

func generateWithPhys(img image.Image) {
	var buf bytes.Buffer
	
	// Write PNG signature
	buf.Write([]byte{137, 80, 78, 71, 13, 10, 26, 10})
	
	// Encode image to get chunks
	var imgBuf bytes.Buffer
	if err := png.Encode(&imgBuf, img); err != nil {
		log.Fatalf("Failed to encode PNG: %v", err)
	}
	imgData := imgBuf.Bytes()
	
	// Copy IHDR
	copyChunk(&buf, imgData, "IHDR")
	
	// Add pHYs chunk (300 DPI = 11811 pixels per meter)
	physData := make([]byte, 9)
	binary.BigEndian.PutUint32(physData[0:4], 11811) // X pixels per unit
	binary.BigEndian.PutUint32(physData[4:8], 11811) // Y pixels per unit
	physData[8] = 1 // Unit is meter
	writeChunk(&buf, "pHYs", physData)
	
	// Copy IDAT and IEND
	copyChunk(&buf, imgData, "IDAT")
	copyChunk(&buf, imgData, "IEND")
	
	if err := os.WriteFile("testdata/with_physical_dims.png", buf.Bytes(), 0644); err != nil {
		log.Fatalf("Failed to write file: %v", err)
	}
}

func generateWithText(img image.Image) {
	var buf bytes.Buffer
	
	// Write PNG signature
	buf.Write([]byte{137, 80, 78, 71, 13, 10, 26, 10})
	
	// Encode image
	var imgBuf bytes.Buffer
	if err := png.Encode(&imgBuf, img); err != nil {
		log.Fatalf("Failed to encode PNG: %v", err)
	}
	imgData := imgBuf.Bytes()
	
	// Copy IHDR
	copyChunk(&buf, imgData, "IHDR")
	
	// Add tEXt chunks
	writeTextChunk(&buf, "Comment", "This is a test comment")
	writeTextChunk(&buf, "Copyright", "Copyright 2024 Test")
	writeTextChunk(&buf, "Description", "Test image with text chunks")
	
	// Copy IDAT and IEND
	copyChunk(&buf, imgData, "IDAT")
	copyChunk(&buf, imgData, "IEND")
	
	if err := os.WriteFile("testdata/with_text_chunks.png", buf.Bytes(), 0644); err != nil {
		log.Fatalf("Failed to write file: %v", err)
	}
}

func generateWithTime(img image.Image) {
	var buf bytes.Buffer
	
	// Write PNG signature
	buf.Write([]byte{137, 80, 78, 71, 13, 10, 26, 10})
	
	// Encode image
	var imgBuf bytes.Buffer
	if err := png.Encode(&imgBuf, img); err != nil {
		log.Fatalf("Failed to encode PNG: %v", err)
	}
	imgData := imgBuf.Bytes()
	
	// Copy IHDR
	copyChunk(&buf, imgData, "IHDR")
	
	// Add tIME chunk
	timeData := []byte{
		0x07, 0xE8, // Year: 2024
		0x01,       // Month: January
		0x01,       // Day: 1
		0x00,       // Hour: 0
		0x00,       // Minute: 0
		0x00,       // Second: 0
	}
	writeChunk(&buf, "tIME", timeData)
	
	// Copy IDAT and IEND
	copyChunk(&buf, imgData, "IDAT")
	copyChunk(&buf, imgData, "IEND")
	
	if err := os.WriteFile("testdata/with_time.png", buf.Bytes(), 0644); err != nil {
		log.Fatalf("Failed to write file: %v", err)
	}
}

func generateWithBackground(img image.Image) {
	var buf bytes.Buffer
	
	// Write PNG signature
	buf.Write([]byte{137, 80, 78, 71, 13, 10, 26, 10})
	
	// Encode image
	var imgBuf bytes.Buffer
	if err := png.Encode(&imgBuf, img); err != nil {
		log.Fatalf("Failed to encode PNG: %v", err)
	}
	imgData := imgBuf.Bytes()
	
	// Copy IHDR
	copyChunk(&buf, imgData, "IHDR")
	
	// Add bKGD chunk (red background for RGB)
	bkgdData := []byte{0xFF, 0xFF, 0x00, 0x00, 0x00, 0x00} // Red
	writeChunk(&buf, "bKGD", bkgdData)
	
	// Copy IDAT and IEND
	copyChunk(&buf, imgData, "IDAT")
	copyChunk(&buf, imgData, "IEND")
	
	if err := os.WriteFile("testdata/with_background.png", buf.Bytes(), 0644); err != nil {
		log.Fatalf("Failed to write file: %v", err)
	}
}

func generateWithChromaticity(img image.Image) {
	var buf bytes.Buffer
	
	// Write PNG signature
	buf.Write([]byte{137, 80, 78, 71, 13, 10, 26, 10})
	
	// Encode image
	var imgBuf bytes.Buffer
	if err := png.Encode(&imgBuf, img); err != nil {
		log.Fatalf("Failed to encode PNG: %v", err)
	}
	imgData := imgBuf.Bytes()
	
	// Copy IHDR
	copyChunk(&buf, imgData, "IHDR")
	
	// Add cHRM chunk (sRGB chromaticity)
	chrmData := make([]byte, 32)
	// White point
	binary.BigEndian.PutUint32(chrmData[0:4], 31270)  // x
	binary.BigEndian.PutUint32(chrmData[4:8], 32900)  // y
	// Red
	binary.BigEndian.PutUint32(chrmData[8:12], 64000)  // x
	binary.BigEndian.PutUint32(chrmData[12:16], 33000) // y
	// Green
	binary.BigEndian.PutUint32(chrmData[16:20], 30000) // x
	binary.BigEndian.PutUint32(chrmData[20:24], 60000) // y
	// Blue
	binary.BigEndian.PutUint32(chrmData[24:28], 15000) // x
	binary.BigEndian.PutUint32(chrmData[28:32], 6000)  // y
	
	writeChunk(&buf, "cHRM", chrmData)
	
	// Copy IDAT and IEND
	copyChunk(&buf, imgData, "IDAT")
	copyChunk(&buf, imgData, "IEND")
	
	if err := os.WriteFile("testdata/with_chromaticity.png", buf.Bytes(), 0644); err != nil {
		log.Fatalf("Failed to write file: %v", err)
	}
}

func generateWithSRGB(img image.Image) {
	var buf bytes.Buffer
	
	// Write PNG signature
	buf.Write([]byte{137, 80, 78, 71, 13, 10, 26, 10})
	
	// Encode image
	var imgBuf bytes.Buffer
	if err := png.Encode(&imgBuf, img); err != nil {
		log.Fatalf("Failed to encode PNG: %v", err)
	}
	imgData := imgBuf.Bytes()
	
	// Copy IHDR
	copyChunk(&buf, imgData, "IHDR")
	
	// Add sRGB chunk (0 = Perceptual)
	writeChunk(&buf, "sRGB", []byte{0})
	
	// Copy IDAT and IEND
	copyChunk(&buf, imgData, "IDAT")
	copyChunk(&buf, imgData, "IEND")
	
	if err := os.WriteFile("testdata/with_srgb.png", buf.Bytes(), 0644); err != nil {
		log.Fatalf("Failed to write file: %v", err)
	}
}

func generateWithTransparency() {
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
	
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		log.Fatalf("Failed to encode PNG: %v", err)
	}
	if err := os.WriteFile("testdata/with_transparency.png", buf.Bytes(), 0644); err != nil {
		log.Fatalf("Failed to write file: %v", err)
	}
}

func generateIndexedColor() {
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
	
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		log.Fatalf("Failed to encode PNG: %v", err)
	}
	if err := os.WriteFile("testdata/indexed_color.png", buf.Bytes(), 0644); err != nil {
		log.Fatalf("Failed to write file: %v", err)
	}
}

func generateWithSBIT(img image.Image) {
	var buf bytes.Buffer
	
	// Write PNG signature
	buf.Write([]byte{137, 80, 78, 71, 13, 10, 26, 10})
	
	// Encode image
	var imgBuf bytes.Buffer
	if err := png.Encode(&imgBuf, img); err != nil {
		log.Fatalf("Failed to encode PNG: %v", err)
	}
	imgData := imgBuf.Bytes()
	
	// Copy IHDR
	copyChunk(&buf, imgData, "IHDR")
	
	// Add sBIT chunk (4 bits per channel for RGBA)
	writeChunk(&buf, "sBIT", []byte{4, 4, 4, 4})
	
	// Copy IDAT and IEND
	copyChunk(&buf, imgData, "IDAT")
	copyChunk(&buf, imgData, "IEND")
	
	if err := os.WriteFile("testdata/with_significant_bits.png", buf.Bytes(), 0644); err != nil {
		log.Fatalf("Failed to write file: %v", err)
	}
}

// Helper functions

func writeChunk(buf *bytes.Buffer, chunkType string, data []byte) {
	// Write length
	if err := binary.Write(buf, binary.BigEndian, uint32(len(data))); err != nil {
		log.Fatalf("Failed to write chunk length: %v", err)
	}
	
	// Write chunk type
	buf.WriteString(chunkType)
	
	// Write data
	buf.Write(data)
	
	// Calculate and write CRC
	crcData := append([]byte(chunkType), data...)
	crc := crc32.ChecksumIEEE(crcData)
	if err := binary.Write(buf, binary.BigEndian, crc); err != nil {
		log.Fatalf("Failed to write CRC: %v", err)
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
		
		length := binary.BigEndian.Uint32(src[offset:offset+4])
		chunk := string(src[offset+4:offset+8])
		
		if chunk == chunkType {
			// Copy entire chunk including length and CRC
			chunkSize := 12 + int(length)
			dest.Write(src[offset:offset+chunkSize])
			return
		}
		
		offset += 12 + int(length)
	}
	
	if chunkType == "IDAT" || chunkType == "IEND" {
		log.Fatalf("Required chunk %s not found", chunkType)
	}
}