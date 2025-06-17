# go-jpeg-meta-web-strip

A Go library for optimizing JPEG images by removing unnecessary metadata while preserving essential information and image quality.

## Features

- **Selective Metadata Removal**: Removes unnecessary metadata using a blacklist approach
- **Preservation of Essential Data**: Keeps important metadata like orientation, ICC profiles, DPI settings
- **Image Integrity**: Ensures pixel data remains unchanged after processing
- **High Performance**: Efficient processing with minimal memory overhead

### Metadata Removed

- EXIF thumbnails
- GPS information
- Camera information (Make, Model, Lens data)
- Maker-specific data
- XMP metadata
- IPTC metadata
- Photoshop IRB data
- Comments

### Metadata Preserved

- Orientation
- ICC color profiles
- DPI/Resolution settings
- Color space information
- Gamma values
- Essential image rendering data

## Installation

```bash
go get github.com/ideamans/go-jpeg-meta-web-strip
```

## Usage

```go
package main

import (
    "fmt"
    "os"
    jpegmetawebstrip "github.com/ideamans/go-jpeg-meta-web-strip"
)

func main() {
    // Read JPEG file
    jpegData, err := os.ReadFile("input.jpg")
    if err != nil {
        panic(err)
    }

    // Remove unnecessary metadata
    cleanedData, result, err := jpegmetawebstrip.jpegmetawebstrip(jpegData)
    if err != nil {
        panic(err)
    }

    // Write cleaned JPEG
    err = os.WriteFile("output.jpg", cleanedData, 0644)
    if err != nil {
        panic(err)
    }

    // Display results
    fmt.Printf("Removed metadata:\n")
    fmt.Printf("  EXIF Thumbnail: %d bytes\n", result.Removed.ExifThumbnail)
    fmt.Printf("  GPS: %d bytes\n", result.Removed.ExifGPS)
    fmt.Printf("  Camera Info: %d bytes\n", result.Removed.CameraInfo)
    fmt.Printf("  XMP: %d bytes\n", result.Removed.XMP)
    fmt.Printf("  IPTC: %d bytes\n", result.Removed.IPTC)
    fmt.Printf("  Comments: %d bytes\n", result.Removed.Comments)
    fmt.Printf("Total removed: %d bytes\n", result.Total)
}
```

## Test Data Generator

The package includes a test data generator that creates various JPEG files with different metadata combinations.

### Usage

```bash
# Generate test data
make data

# Or run directly
go run datacreator/cmd/main.go
```

### Generated Test Images

The following test images are generated in the `testdata` directory:

| Filename                       | Description                      | Metadata                                 |
| ------------------------------ | -------------------------------- | ---------------------------------------- |
| `basic_copy.jpg`               | Basic copy of original           | Minimal metadata                         |
| `with_exif_thumbnail.jpg`      | JPEG with EXIF thumbnail         | 160x120 thumbnail embedded               |
| `with_gps.jpg`                 | JPEG with GPS data               | GPS coordinates                          |
| `with_camera_info.jpg`         | JPEG with camera information     | Make, Model tags                         |
| `with_xmp.jpg`                 | JPEG with XMP metadata           | Creator, creation date, etc.             |
| `with_iptc.jpg`                | JPEG with IPTC metadata          | Caption, keywords, copyright             |
| `with_photoshop_irb.jpg`       | JPEG with Photoshop IRB          | Photoshop-specific data                  |
| `with_comment.jpg`             | JPEG with comment                | Text comment                             |
| `with_orientation.jpg`         | JPEG with orientation            | 90° rotation (preserved)                 |
| `with_icc_profile_srgb.jpg`    | JPEG with sRGB ICC profile       | Color profile (preserved)                |
| `with_icc_profile_p3.jpg`      | JPEG with Display P3 ICC profile | Color profile (preserved)                |
| `with_dpi.jpg`                 | JPEG with DPI settings           | 300 DPI (preserved)                      |
| `with_colorspace.jpg`          | JPEG with specific colorspace    | sRGB colorspace (preserved)              |
| `with_gamma.jpg`               | JPEG with gamma value            | Gamma 2.2 (preserved)                    |
| `with_quality.jpg`             | JPEG with specific quality       | Quality 95                               |
| `with_all_removable.jpg`       | JPEG with all removable metadata | Comprehensive test                       |
| `with_mixed_metadata.jpg`      | JPEG with mixed metadata         | Removable + preservable                  |
| `with_comprehensive_mixed.jpg` | Comprehensive mixed metadata     | Thumbnail, GPS, camera, orientation, DPI |
| `with_thumbnail_and_icc.jpg`   | JPEG with thumbnail and ICC      | Tests selective removal                  |

### Requirements for Test Data Generation

- ImageMagick (`magick` command)
- ExifTool (`exiftool` command) - optional but recommended for comprehensive metadata

## Testing

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific test
go test -v -run Testjpegmetawebstrip

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Test Cases

The package includes comprehensive tests:

1. **Metadata Removal Tests**: Verify specific metadata types are removed
2. **Metadata Preservation Tests**: Ensure essential metadata is preserved
3. **Invalid Data Handling**: Test error handling for invalid inputs
4. **Image Integrity Tests**: Verify pixel data remains unchanged using MD5 checksums
5. **Comprehensive Tests**: Mixed metadata scenarios

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
