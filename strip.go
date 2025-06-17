package pngmetawebstrip

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
)

// Result contains information about removed chunks
type Result struct {
	Removed struct {
		TextChunks  int // tEXt, zTXt, iTXt
		TimeChunk   int // tIME
		Background  int // bKGD
		ExifData    int // eXIf
		OtherChunks int // All other removed chunks
	}
	Total int // Total bytes removed
}

// Essential chunks that must be preserved
var essentialChunks = map[string]bool{
	// Core
	"IHDR": true,
	"PLTE": true,
	"IDAT": true,
	"IEND": true,
	// Transparency
	"tRNS": true,
	// Color space
	"gAMA": true,
	"cHRM": true,
	"sRGB": true,
	"iCCP": true,
	"sBIT": true,
	// Physical dimensions
	"pHYs": true,
}

// PngMetaWebStrip removes unnecessary metadata chunks from PNG data
func PngMetaWebStrip(data []byte) ([]byte, *Result, error) {
	if len(data) < 8 {
		return nil, nil, fmt.Errorf("data too short to be a PNG")
	}

	// Verify PNG signature
	pngSignature := []byte{137, 80, 78, 71, 13, 10, 26, 10}
	if !bytes.Equal(data[:8], pngSignature) {
		return nil, nil, fmt.Errorf("invalid PNG signature")
	}

	result := &Result{}
	output := bytes.NewBuffer(nil)
	
	// Write PNG signature
	output.Write(pngSignature)
	
	// Process chunks
	offset := 8
	for offset < len(data) {
		if offset+8 > len(data) {
			return nil, nil, fmt.Errorf("incomplete chunk at offset %d", offset)
		}

		// Read chunk length
		length := binary.BigEndian.Uint32(data[offset : offset+4])
		
		if offset+12+int(length) > len(data) {
			return nil, nil, fmt.Errorf("chunk extends beyond data at offset %d", offset)
		}

		// Read chunk type
		chunkType := string(data[offset+4 : offset+8])
		
		// Calculate full chunk size (length + type + data + CRC)
		fullChunkSize := 12 + int(length)
		
		// Verify CRC
		chunkData := data[offset+4 : offset+8+int(length)]
		crc := binary.BigEndian.Uint32(data[offset+8+int(length) : offset+12+int(length)])
		calculatedCRC := crc32.ChecksumIEEE(chunkData)
		
		if crc != calculatedCRC {
			return nil, nil, fmt.Errorf("invalid CRC for chunk %s at offset %d", chunkType, offset)
		}

		// Decide whether to keep the chunk
		if shouldKeepChunk(chunkType) {
			// Write the entire chunk
			output.Write(data[offset : offset+fullChunkSize])
		} else {
			// Track removed chunk
			trackRemovedChunk(result, chunkType, fullChunkSize)
		}

		offset += fullChunkSize
	}

	return output.Bytes(), result, nil
}

// shouldKeepChunk determines if a chunk should be preserved
func shouldKeepChunk(chunkType string) bool {
	return essentialChunks[chunkType]
}

// trackRemovedChunk updates the result statistics
func trackRemovedChunk(result *Result, chunkType string, size int) {
	result.Total += size
	
	switch chunkType {
	case "tEXt", "zTXt", "iTXt":
		result.Removed.TextChunks += size
	case "tIME":
		result.Removed.TimeChunk += size
	case "bKGD":
		result.Removed.Background += size
	case "eXIf":
		result.Removed.ExifData += size
	default:
		result.Removed.OtherChunks += size
	}
}

// PngMetaWebStripReader processes PNG data from a reader
func PngMetaWebStripReader(r io.Reader) ([]byte, *Result, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read data: %w", err)
	}
	
	return PngMetaWebStrip(data)
}

// PngMetaWebStripWriter processes PNG data and writes to a writer
func PngMetaWebStripWriter(data []byte, w io.Writer) (*Result, error) {
	cleaned, result, err := PngMetaWebStrip(data)
	if err != nil {
		return nil, err
	}
	
	_, err = w.Write(cleaned)
	if err != nil {
		return nil, fmt.Errorf("failed to write data: %w", err)
	}
	
	return result, nil
}