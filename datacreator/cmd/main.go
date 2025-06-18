package main

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	testDataDir = "testdata"
	baseImage   = "base.png"
)

type TestCase struct {
	Filename    string
	Description string
	Commands    [][]string
}

func main() {
	// Create testdata directory
	if err := os.MkdirAll(testDataDir, 0755); err != nil {
		log.Fatalf("Failed to create testdata directory: %v", err)
	}

	// Create base image using ImageMagick
	fmt.Println("Creating base image...")
	createBaseImage()

	// Define test cases
	testCases := []TestCase{
		{
			Filename:    "basic_copy.png",
			Description: "Basic copy of original",
			Commands: [][]string{
				{"magick", "convert", filepath.Join(testDataDir, baseImage),
					"-define", "png:exclude-chunk=all",
					"-define", "png:include-chunk=none",
					filepath.Join(testDataDir, "basic_copy.png")},
			},
		},
		{
			Filename:    "with_text_chunks.png",
			Description: "PNG with tEXt/zTXt/iTXt chunks",
			Commands: [][]string{
				{"magick", "convert", filepath.Join(testDataDir, baseImage),
					"-set", "comment", "This is a test comment",
					"-set", "copyright", "Copyright 2024 Test",
					"-set", "description", "Test image with text chunks",
					filepath.Join(testDataDir, "with_text_chunks.png")},
			},
		},
		{
			Filename:    "with_time.png",
			Description: "PNG with tIME chunk",
			Commands: [][]string{
				{"magick", "convert", filepath.Join(testDataDir, baseImage),
					"-define", "png:include-chunk=tIME",
					filepath.Join(testDataDir, "with_time.png")},
			},
		},
		{
			Filename:    "with_background.png",
			Description: "PNG with bKGD chunk",
			Commands: [][]string{
				{"magick", "convert", filepath.Join(testDataDir, baseImage),
					"-background", "#FF0000",
					"-define", "png:include-chunk=bKGD",
					filepath.Join(testDataDir, "with_background.png")},
			},
		},
		{
			Filename:    "with_exif.png",
			Description: "PNG with eXIf chunk",
			Commands: [][]string{
				{"magick", "convert", filepath.Join(testDataDir, baseImage), filepath.Join(testDataDir, "with_exif_temp.png")},
				{"exiftool", "-overwrite_original",
					"-Make=TestCamera",
					"-Model=TestModel",
					"-GPSLatitude=40.7128",
					"-GPSLongitude=-74.0060",
					filepath.Join(testDataDir, "with_exif_temp.png")},
				{"mv", filepath.Join(testDataDir, "with_exif_temp.png"), filepath.Join(testDataDir, "with_exif.png")},
			},
		},
		{
			Filename:    "with_histogram.png",
			Description: "PNG with hIST chunk",
			Commands: [][]string{
				{"magick", "convert", filepath.Join(testDataDir, baseImage),
					"-define", "png:include-chunk=hIST",
					filepath.Join(testDataDir, "with_histogram.png")},
			},
		},
		{
			Filename:    "with_gamma.png",
			Description: "PNG with gAMA chunk",
			Commands: [][]string{
				{"magick", "convert", filepath.Join(testDataDir, baseImage),
					"-gamma", "2.2",
					"-define", "png:include-chunk=gAMA",
					filepath.Join(testDataDir, "with_gamma.png")},
			},
		},
		{
			Filename:    "with_chromaticity.png",
			Description: "PNG with cHRM chunk",
			Commands: [][]string{
				{"magick", "convert", filepath.Join(testDataDir, baseImage),
					"-define", "png:include-chunk=cHRM",
					filepath.Join(testDataDir, "with_chromaticity.png")},
			},
		},
		{
			Filename:    "with_srgb.png",
			Description: "PNG with sRGB chunk",
			Commands: [][]string{
				{"magick", "convert", filepath.Join(testDataDir, baseImage),
					"-colorspace", "sRGB",
					"-define", "png:include-chunk=sRGB",
					filepath.Join(testDataDir, "with_srgb.png")},
			},
		},
		{
			Filename:    "with_icc_profile.png",
			Description: "PNG with iCCP chunk",
			Commands: [][]string{
				{"magick", "convert", filepath.Join(testDataDir, baseImage),
					"-profile", "sRGB",
					filepath.Join(testDataDir, "with_icc_profile.png")},
			},
		},
		{
			Filename:    "with_physical_dims.png",
			Description: "PNG with pHYs chunk",
			Commands: [][]string{
				{"magick", "convert", filepath.Join(testDataDir, baseImage),
					"-density", "300",
					"-units", "PixelsPerInch",
					filepath.Join(testDataDir, "with_physical_dims.png")},
			},
		},
		{
			Filename:    "indexed_color.png",
			Description: "PNG with PLTE chunk",
			Commands: [][]string{
				{"magick", "convert", filepath.Join(testDataDir, baseImage),
					"-colors", "256",
					"-type", "Palette",
					filepath.Join(testDataDir, "indexed_color.png")},
			},
		},
		{
			Filename:    "with_transparency.png",
			Description: "PNG with tRNS chunk",
			Commands: [][]string{
				{"magick", "convert", filepath.Join(testDataDir, baseImage),
					"-transparent", "white",
					filepath.Join(testDataDir, "with_transparency.png")},
			},
		},
		{
			Filename:    "with_significant_bits.png",
			Description: "PNG with sBIT chunk",
			Commands: [][]string{
				{"magick", "convert", filepath.Join(testDataDir, baseImage),
					"-depth", "4",
					"-define", "png:include-chunk=sBIT",
					filepath.Join(testDataDir, "with_significant_bits.png")},
			},
		},
		{
			Filename:    "with_suggested_palette.png",
			Description: "PNG with sPLT chunk",
			Commands: [][]string{
				{"magick", "convert", filepath.Join(testDataDir, baseImage),
					"-define", "png:include-chunk=sPLT",
					filepath.Join(testDataDir, "with_suggested_palette.png")},
			},
		},
		{
			Filename:    "with_all_removable.png",
			Description: "PNG with all removable chunks",
			Commands: [][]string{
				{"magick", "convert", filepath.Join(testDataDir, baseImage),
					"-set", "comment", "Test comment",
					"-set", "copyright", "Test copyright",
					"-background", "#FF0000",
					"-define", "png:include-chunk=tIME,bKGD,hIST,sPLT",
					filepath.Join(testDataDir, "with_all_removable_temp.png")},
				{"exiftool", "-overwrite_original",
					"-Make=TestCamera",
					"-Model=TestModel",
					filepath.Join(testDataDir, "with_all_removable_temp.png")},
				{"mv", filepath.Join(testDataDir, "with_all_removable_temp.png"), filepath.Join(testDataDir, "with_all_removable.png")},
			},
		},
		{
			Filename:    "with_mixed_chunks.png",
			Description: "PNG with mixed chunks",
			Commands: [][]string{
				{"magick", "convert", filepath.Join(testDataDir, baseImage),
					"-set", "comment", "Mixed test",
					"-gamma", "2.2",
					"-density", "300",
					"-define", "png:include-chunk=tIME,gAMA,pHYs",
					filepath.Join(testDataDir, "with_mixed_chunks.png")},
			},
		},
		{
			Filename:    "with_comprehensive_mixed.png",
			Description: "Comprehensive mixed chunks",
			Commands: [][]string{
				{"magick", "convert", filepath.Join(testDataDir, baseImage),
					"-set", "comment", "Comprehensive test",
					"-background", "#00FF00",
					"-gamma", "2.2",
					"-density", "300",
					"-transparent", "white",
					"-define", "png:include-chunk=tIME,bKGD,gAMA,pHYs,tRNS",
					filepath.Join(testDataDir, "with_comprehensive_mixed.png")},
			},
		},
		{
			Filename:    "with_text_and_icc.png",
			Description: "PNG with text and ICC",
			Commands: [][]string{
				{"magick", "convert", filepath.Join(testDataDir, baseImage),
					"-set", "comment", "ICC test",
					"-profile", "sRGB",
					filepath.Join(testDataDir, "with_text_and_icc.png")},
			},
		},
		// Additional test cases for improved coverage
		{
			Filename:    "ztxt_chunks.png",
			Description: "PNG with zTXt compressed text chunks",
			Commands: [][]string{
				{"magick", "convert", filepath.Join(testDataDir, baseImage),
					"-set", "comment", "Compressed text comment for zTXt testing",
					"-set", "description", "This is a longer description that should be compressed in zTXt format",
					"-define", "png:format=png8",
					filepath.Join(testDataDir, "ztxt_chunks.png")},
			},
		},
		{
			Filename:    "multiple_idat.png",
			Description: "PNG with multiple IDAT chunks",
			Commands: [][]string{
				{"magick", "convert", filepath.Join(testDataDir, baseImage),
					"-compress", "ZIP",
					"-quality", "95",
					filepath.Join(testDataDir, "multiple_idat.png")},
			},
		},
		{
			Filename:    "minimal_ihdr_only.png",
			Description: "Minimal PNG with only IHDR, IDAT, IEND",
			Commands: [][]string{
				{"magick", "convert", filepath.Join(testDataDir, baseImage),
					"-strip",
					"-define", "png:exclude-chunk=all",
					"-define", "png:include-chunk=none",
					filepath.Join(testDataDir, "minimal_ihdr_only.png")},
			},
		},
		{
			Filename:    "large_text_chunk.png",
			Description: "PNG with very large text chunk",
			Commands: [][]string{
				{"magick", "convert", filepath.Join(testDataDir, baseImage),
					"-set", "comment", generateLargeText(),
					filepath.Join(testDataDir, "large_text_chunk.png")},
			},
		},
		{
			Filename:    "interlaced.png",
			Description: "Interlaced PNG",
			Commands: [][]string{
				{"magick", "convert", filepath.Join(testDataDir, baseImage),
					"-interlace", "PNG",
					filepath.Join(testDataDir, "interlaced.png")},
			},
		},
		{
			Filename:    "grayscale_with_transparency.png",
			Description: "Grayscale PNG with transparency",
			Commands: [][]string{
				{"magick", "convert", filepath.Join(testDataDir, baseImage),
					"-colorspace", "Gray",
					"-transparent", "white",
					filepath.Join(testDataDir, "grayscale_with_transparency.png")},
			},
		},
		{
			Filename:    "all_essential_chunks.png",
			Description: "PNG with all essential chunks",
			Commands: [][]string{
				{"magick", "convert", filepath.Join(testDataDir, baseImage),
					"-gamma", "2.2",
					"-colorspace", "sRGB",
					"-density", "300",
					"-transparent", "white",
					"-define", "png:include-chunk=gAMA,sRGB,pHYs,tRNS",
					filepath.Join(testDataDir, "all_essential_chunks.png")},
			},
		},
		{
			Filename:    "private_chunks.png",
			Description: "PNG with private/unknown chunks (will be simulated)",
			Commands: [][]string{
				{"magick", "convert", filepath.Join(testDataDir, baseImage),
					"-set", "comment", "Base for private chunks",
					filepath.Join(testDataDir, "private_chunks_temp.png")},
			},
		},
		{
			Filename:    "edge_case_small.png",
			Description: "Very small PNG for edge case testing",
			Commands: [][]string{
				{"magick", "convert",
					"-size", "1x1",
					"xc:red",
					filepath.Join(testDataDir, "edge_case_small.png")},
			},
		},
		{
			Filename:    "high_bit_depth.png",
			Description: "PNG with 16-bit color depth",
			Commands: [][]string{
				{"magick", "convert", filepath.Join(testDataDir, baseImage),
					"-depth", "16",
					filepath.Join(testDataDir, "high_bit_depth.png")},
			},
		},
		{
			Filename:    "comprehensive_ancillary.png",
			Description: "PNG with all types of ancillary chunks",
			Commands: [][]string{
				{"magick", "convert", filepath.Join(testDataDir, baseImage),
					"-set", "comment", "Comprehensive test",
					"-set", "copyright", "Test Copyright",
					"-set", "description", "All ancillary chunks test",
					"-background", "#808080",
					"-gamma", "2.2",
					"-density", "300",
					"-define", "png:include-chunk=tIME,bKGD,gAMA,pHYs,hIST,sPLT",
					filepath.Join(testDataDir, "comprehensive_ancillary_temp.png")},
				{"exiftool", "-overwrite_original",
					"-Make=TestCamera",
					"-Model=TestModel",
					"-DateTime=2024:01:01 12:00:00",
					filepath.Join(testDataDir, "comprehensive_ancillary_temp.png")},
				{"mv", filepath.Join(testDataDir, "comprehensive_ancillary_temp.png"), filepath.Join(testDataDir, "comprehensive_ancillary.png")},
			},
		},
	}

	// Execute test cases
	for _, tc := range testCases {
		fmt.Printf("Creating %s: %s\n", tc.Filename, tc.Description)
		for _, cmd := range tc.Commands {
			if err := runCommand(cmd); err != nil {
				log.Printf("Warning: Failed to create %s: %v", tc.Filename, err)
				// Continue with other test cases
			}
		}
	}

	// Post-process special cases
	fmt.Println("Creating special test cases...")
	createPrivateChunkPNG()
	createCorruptedCRCPNG()
	createTruncatedPNG()

	fmt.Println("Test data generation complete!")
}

func createBaseImage() {
	// Create a simple 100x100 RGB image with a gradient
	cmd := []string{
		"magick", "convert",
		"-size", "100x100",
		"gradient:blue-white",
		"-define", "png:exclude-chunk=all",
		"-define", "png:include-chunk=none",
		filepath.Join(testDataDir, baseImage),
	}
	if err := runCommand(cmd); err != nil {
		log.Fatalf("Failed to create base image: %v", err)
	}
}

func runCommand(args []string) error {
	cmd := exec.Command(args[0], args[1:]...) // #nosec G204 -- args are hardcoded in this file
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func generateLargeText() string {
	// Generate a large text string to test large chunk handling
	text := "This is a test of large text chunk handling. "
	result := ""
	for i := 0; i < 100; i++ {
		result += fmt.Sprintf("%s[%d] ", text, i)
	}
	return result
}

func createPrivateChunkPNG() {
	// Create a PNG with a private chunk for testing unknown chunk handling
	sourceFile := filepath.Join(testDataDir, "private_chunks_temp.png")
	targetFile := filepath.Join(testDataDir, "private_chunks.png")
	
	// Read the source file
	data, err := os.ReadFile(sourceFile)
	if err != nil {
		log.Printf("Warning: Failed to read source file for private chunks: %v", err)
		return
	}
	
	// Insert a private chunk (prIV) after IHDR
	output := make([]byte, 0, len(data)+100)
	output = append(output, data[:33]...) // PNG signature + IHDR
	
	// Add private chunk
	privateData := []byte("This is a private chunk for testing")
	output = append(output, writeChunkToBytes("prIV", privateData)...)
	
	// Add rest of the file
	output = append(output, data[33:]...)
	
	if err := os.WriteFile(targetFile, output, 0644); err != nil {
		log.Printf("Warning: Failed to write private chunks PNG: %v", err)
	}
	
	// Clean up temp file
	os.Remove(sourceFile)
}

func createCorruptedCRCPNG() {
	// Create a PNG with corrupted CRC for error testing
	sourceFile := filepath.Join(testDataDir, "basic_copy.png")
	targetFile := filepath.Join(testDataDir, "corrupted_crc.png")
	
	data, err := os.ReadFile(sourceFile)
	if err != nil {
		log.Printf("Warning: Failed to read source file for corrupted CRC: %v", err)
		return
	}
	
	// Corrupt the CRC of the first chunk after IHDR (should be IDAT)
	if len(data) > 50 {
		corrupted := make([]byte, len(data))
		copy(corrupted, data)
		// Corrupt last byte of IHDR CRC
		corrupted[32] = corrupted[32] ^ 0xFF
		
		if err := os.WriteFile(targetFile, corrupted, 0644); err != nil {
			log.Printf("Warning: Failed to write corrupted CRC PNG: %v", err)
		}
	}
}

func createTruncatedPNG() {
	// Create a truncated PNG for error testing
	sourceFile := filepath.Join(testDataDir, "basic_copy.png")
	targetFile := filepath.Join(testDataDir, "truncated.png")
	
	data, err := os.ReadFile(sourceFile)
	if err != nil {
		log.Printf("Warning: Failed to read source file for truncated PNG: %v", err)
		return
	}
	
	// Truncate the file in the middle of a chunk
	if len(data) > 100 {
		truncated := data[:len(data)-50] // Remove last 50 bytes
		
		if err := os.WriteFile(targetFile, truncated, 0644); err != nil {
			log.Printf("Warning: Failed to write truncated PNG: %v", err)
		}
	}
}

func writeChunkToBytes(chunkType string, data []byte) []byte {
	result := make([]byte, 0, 12+len(data))
	
	// Length
	length := make([]byte, 4)
	binary.BigEndian.PutUint32(length, uint32(len(data)))
	result = append(result, length...)
	
	// Type
	result = append(result, []byte(chunkType)...)
	
	// Data
	result = append(result, data...)
	
	// CRC
	crcData := append([]byte(chunkType), data...)
	crc := crc32.ChecksumIEEE(crcData)
	crcBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(crcBytes, crc)
	result = append(result, crcBytes...)
	
	return result
}
