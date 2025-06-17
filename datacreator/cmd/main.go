package main

import (
	"fmt"
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
