# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go library for optimizing PNG images by removing unnecessary metadata chunks while preserving essential information for web display. The library uses a blacklist approach to selectively remove ancillary chunks that don't affect image rendering or color accuracy.

## Build and Development Commands

```bash
# Build the package
go build ./...

# Run all tests
go test -v ./...

# Run a specific test
go test -v -run TestPngMetaWebStrip/Remove_text_chunks

# Generate test coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run linting (requires golangci-lint)
golangci-lint run --timeout=5m

# Generate test data (requires ImageMagick and ExifTool)
make data
# or
go run datacreator/cmd/main.go

# Clean test data
make clean
```

## Architecture

### Core Processing Flow

The library follows a chunk-based PNG processing approach:

1. **Parse PNG Structure**: Reads PNG signature and parses chunks sequentially
2. **Chunk Processing**: Each chunk is evaluated by `processChunk()` function
3. **Selective Removal**: Chunks are either kept or removed based on chunk type
4. **Reconstruction**: Valid chunks are reassembled with proper CRC recalculation

### Key Components

- **webstrip.go**: Main processing logic
  - `PngMetaWebStrip()`: Entry point that orchestrates the cleaning process
  - `processChunk()`: Evaluates each PNG chunk
  - `isEssentialChunk()`: Determines if a chunk should be preserved
  - `calculateCRC()`: Recalculates CRC32 for chunk integrity
  - Chunk type identification and filtering logic

- **datacreator/**: Test data generation utility
  - Creates 18+ different PNG variations with various chunk combinations
  - Uses ImageMagick for basic image operations
  - Uses pngcrush/optipng for adding specific chunks
  - Uses ExifTool for eXIf chunk manipulation

### Chunk Handling Strategy

**Removed (Blacklist)**:
- tEXt/zTXt/iTXt: Text comments and metadata
- tIME: Last modification timestamp
- bKGD: Background color suggestions
- sPLT: Suggested palette entries
- hIST: Histogram data
- eXIf: EXIF metadata (PNG 1.6+)
- Private/ancillary chunks

**Preserved**:
- IHDR: Image header (critical)
- PLTE: Palette data (critical for indexed color)
- IDAT: Image data (critical)
- IEND: Image trailer (critical)
- tRNS: Transparency information
- gAMA: Gamma correction
- cHRM: Chromaticity coordinates
- sRGB: sRGB color space indicator
- iCCP: ICC color profiles
- sBIT: Significant bits (color precision)
- pHYs: Physical pixel dimensions (DPI)

### PNG Chunk Processing

The library directly manipulates PNG chunk structure:
- Reads 8-byte PNG signature validation
- Parses chunk length (4 bytes, big-endian)
- Identifies chunk type (4 ASCII characters)
- Processes chunk data based on type
- Validates/recalculates CRC32 checksums
- Handles both critical and ancillary chunks

## Code Quality Standards

The project enforces strict linting via `.golangci.yml`:
- Cyclomatic complexity limit: 15
- Required formatting: gofmt -s, goimports, gofumpt
- Security scanning with gosec
- Comprehensive static analysis

When modifying PNG chunk parsing code, be aware of:
- All multi-byte integers are big-endian
- CRC must be recalculated after any chunk modification
- Critical chunks (uppercase first letter) cannot be removed
- Chunk ordering matters: some chunks must appear before IDAT

## Testing Approach

Tests verify both chunk removal and image integrity:
- Chunk verification uses pngcheck or custom chunk parser
- Image integrity verified via pixel data CRC validation
- Test data covers edge cases (indexed color, interlaced, various chunk combinations)
- Validates preserved chunks remain bit-identical

## Module Naming Note

The module is named `go-png-meta-web-strip` in go.mod, and the package is `pngmetawebstrip`. The main function is `PngMetaWebStrip()` which returns `([]byte, *Result, error)`.