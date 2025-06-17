# go-png-meta-web-strip

A Go library for optimizing PNG images by removing unnecessary metadata while preserving essential information for web display.

## Features

- **Selective Chunk Removal**: Removes unnecessary PNG chunks using a blacklist approach
- **Preservation of Essential Data**: Keeps important chunks like gamma correction, color profiles, and DPI settings
- **Image Integrity**: Ensures pixel data remains unchanged after processing
- **High Performance**: Efficient chunk-based processing with minimal memory overhead

### Chunks Removed

- tEXt/zTXt/iTXt: Text metadata and comments
- tIME: Last modification time
- bKGD: Background color
- sPLT: Suggested palette
- hIST: Histogram
- eXIf: EXIF metadata (PNG 1.6+)
- Private/ancillary chunks

### Chunks Preserved

- IHDR: Image header (required)
- PLTE: Palette (required for indexed color)
- IDAT: Image data (required)
- IEND: Image trailer (required)
- tRNS: Transparency information
- gAMA: Gamma correction
- cHRM: Chromaticity
- sRGB: sRGB color space
- iCCP: ICC color profiles
- sBIT: Significant bits (color precision)
- pHYs: Physical pixel dimensions (DPI)

## Installation

```bash
go get github.com/ideamans/go-png-meta-web-strip
```

## Usage

```go
package main

import (
    "fmt"
    "os"
    pngmetawebstrip "github.com/ideamans/go-png-meta-web-strip"
)

func main() {
    // Read PNG file
    pngData, err := os.ReadFile("input.png")
    if err != nil {
        panic(err)
    }

    // Remove unnecessary metadata
    cleanedData, result, err := pngmetawebstrip.PngMetaWebStrip(pngData)
    if err != nil {
        panic(err)
    }

    // Write cleaned PNG
    err = os.WriteFile("output.png", cleanedData, 0644)
    if err != nil {
        panic(err)
    }

    // Display results
    fmt.Printf("Removed chunks:\n")
    fmt.Printf("  Text chunks: %d bytes\n", result.Removed.TextChunks)
    fmt.Printf("  Time chunk: %d bytes\n", result.Removed.TimeChunk)
    fmt.Printf("  Background: %d bytes\n", result.Removed.Background)
    fmt.Printf("  EXIF data: %d bytes\n", result.Removed.ExifData)
    fmt.Printf("  Other chunks: %d bytes\n", result.Removed.OtherChunks)
    fmt.Printf("Total removed: %d bytes\n", result.Total)
}
```

## Test Data Generator

The package includes a test data generator that creates various PNG files with different chunk combinations.

### Usage

```bash
# Generate test data
make data

# Or run directly
go run datacreator/cmd/main.go
```

### Generated Test Images

The following test images are generated in the `testdata` directory:

| Filename                       | Description                      | Chunks/Metadata                          |
| ------------------------------ | -------------------------------- | ---------------------------------------- |
| `basic_copy.png`               | Basic copy of original           | Minimal chunks                           |
| `with_text_chunks.png`         | PNG with tEXt/zTXt/iTXt chunks  | Comments, keywords, metadata             |
| `with_time.png`                | PNG with tIME chunk              | Last modification time                   |
| `with_background.png`          | PNG with bKGD chunk              | Background color                         |
| `with_exif.png`                | PNG with eXIf chunk              | EXIF metadata (PNG 1.6+)                 |
| `with_histogram.png`           | PNG with hIST chunk              | Histogram data                           |
| `with_suggested_palette.png`   | PNG with sPLT chunk              | Suggested palette                        |
| `with_significant_bits.png`    | PNG with sBIT chunk              | Significant bits info (preserved)        |
| `with_gamma.png`               | PNG with gAMA chunk              | Gamma 2.2 (preserved)                    |
| `with_chromaticity.png`        | PNG with cHRM chunk              | Chromaticity (preserved)                 |
| `with_srgb.png`                | PNG with sRGB chunk              | sRGB indicator (preserved)               |
| `with_icc_profile.png`         | PNG with iCCP chunk              | ICC color profile (preserved)            |
| `with_physical_dims.png`       | PNG with pHYs chunk              | 300 DPI (preserved)                      |
| `indexed_color.png`            | PNG with PLTE chunk              | Palette (preserved)                      |
| `with_transparency.png`        | PNG with tRNS chunk              | Transparency (preserved)                 |
| `with_all_removable.png`       | PNG with all removable chunks   | Comprehensive test                       |
| `with_mixed_chunks.png`        | PNG with mixed chunks            | Removable + preservable                  |
| `with_comprehensive_mixed.png` | Comprehensive mixed chunks       | Text, time, background, gamma, DPI       |
| `with_text_and_icc.png`        | PNG with text and ICC            | Tests selective removal                  |

### Requirements for Test Data Generation

- ImageMagick (`magick` command)
- pngcrush or optipng - for adding specific PNG chunks
- ExifTool (`exiftool` command) - optional for eXIf chunk manipulation

## Testing

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific test
go test -v -run TestPngMetaWebStrip

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Test Cases

The package includes comprehensive tests:

1. **Chunk Removal Tests**: Verify specific chunk types are removed
2. **Chunk Preservation Tests**: Ensure essential chunks are preserved
3. **Invalid Data Handling**: Test error handling for invalid inputs
4. **Image Integrity Tests**: Verify pixel data remains unchanged using CRC checksums
5. **Comprehensive Tests**: Mixed chunk scenarios
6. **Palette Preservation**: Ensure PLTE chunks remain intact for indexed images

## Requirements

- Go 1.22 or higher
- Dependencies are managed via Go modules

## License

MIT License

Copyright (c) 2024 IdeaMans Inc.

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
