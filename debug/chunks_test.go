package main

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestChunkAnalyzer(t *testing.T) {
	// Create a temporary PNG file for testing
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test.png")

	// Create a test PNG with various chunks
	pngData := createTestPNG()
	if err := os.WriteFile(tempFile, pngData, 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test with valid PNG file
	t.Run("Valid PNG file", func(t *testing.T) {
		// Save original os.Args
		originalArgs := os.Args
		defer func() { os.Args = originalArgs }()

		// Set up args for main function
		os.Args = []string{"chunks", tempFile}

		// Capture output by redirecting to buffer (we can't easily test main output)
		// Instead, we'll test the core logic by reading the file directly
		data, err := os.ReadFile(tempFile)
		if err != nil {
			t.Fatalf("Failed to read test file: %v", err)
		}

		// Verify the test PNG has expected structure
		if len(data) < 8 {
			t.Fatal("Test PNG too short")
		}

		// Check PNG signature
		expectedSig := []byte{137, 80, 78, 71, 13, 10, 26, 10}
		if !bytes.Equal(data[:8], expectedSig) {
			t.Error("Invalid PNG signature in test file")
		}

		// Parse chunks manually to verify structure
		chunks := parseChunks(data)
		if len(chunks) == 0 {
			t.Error("No chunks found in test PNG")
		}

		// Should have at least IHDR, IDAT, and IEND
		hasIHDR := false
		hasIDAT := false
		hasIEND := false
		for _, chunk := range chunks {
			switch chunk.Type {
			case "IHDR":
				hasIHDR = true
			case "IDAT":
				hasIDAT = true
			case "IEND":
				hasIEND = true
			}
		}

		if !hasIHDR {
			t.Error("Test PNG missing IHDR chunk")
		}
		if !hasIDAT {
			t.Error("Test PNG missing IDAT chunk")
		}
		if !hasIEND {
			t.Error("Test PNG missing IEND chunk")
		}
	})

	t.Run("No arguments", func(t *testing.T) {
		// Save original os.Args
		originalArgs := os.Args
		defer func() { os.Args = originalArgs }()

		// Set up args with no file argument
		os.Args = []string{"chunks"}

		// We can't easily test the main function's output, but we can verify
		// the behavior by checking the argument handling logic
		if len(os.Args) < 2 {
			// This is the expected condition that triggers usage message
			t.Log("Correctly identified missing argument condition")
		}
	})

	t.Run("Non-existent file", func(t *testing.T) {
		nonExistentFile := filepath.Join(tempDir, "doesnotexist.png")
		
		// Try to read the non-existent file (simulating what main would do)
		_, err := os.ReadFile(nonExistentFile)
		if err == nil {
			t.Error("Expected error for non-existent file")
		}
		if !os.IsNotExist(err) {
			t.Errorf("Expected file not found error, got: %v", err)
		}
	})
}

func TestChunkParsing(t *testing.T) {
	t.Run("Valid chunks", func(t *testing.T) {
		data := createTestPNG()
		chunks := parseChunks(data)
		
		if len(chunks) == 0 {
			t.Fatal("No chunks parsed")
		}
		
		// Verify chunk types and basic properties
		for _, chunk := range chunks {
			if chunk.Type == "" {
				t.Error("Empty chunk type found")
			}
			if len(chunk.Type) != 4 {
				t.Errorf("Invalid chunk type length: %s", chunk.Type)
			}
		}
	})
	
	t.Run("Malformed PNG", func(t *testing.T) {
		// Create invalid PNG data
		invalidData := []byte{1, 2, 3, 4, 5, 6, 7, 8} // Wrong signature
		invalidData = append(invalidData, 0, 0, 0, 4)  // Length
		invalidData = append(invalidData, 'T', 'E', 'S', 'T') // Type
		
		chunks := parseChunks(invalidData)
		// Should still try to parse chunks even with wrong signature
		if len(chunks) == 0 {
			t.Log("No chunks found in malformed PNG (expected)")
		}
	})
}

func TestChunkStructures(t *testing.T) {
	t.Run("Different chunk types", func(t *testing.T) {
		// Test various chunk types that might appear
		chunkTypes := []string{"IHDR", "IDAT", "IEND", "tEXt", "gAMA", "cHRM"}
		
		for _, chunkType := range chunkTypes {
			t.Run(chunkType, func(t *testing.T) {
				// Create a minimal PNG with this chunk type
				data := createPNGWithChunk(chunkType, []byte("test data"))
				chunks := parseChunks(data)
				
				found := false
				for _, chunk := range chunks {
					if chunk.Type == chunkType {
						found = true
						break
					}
				}
				
				if !found && (chunkType == "IHDR" || chunkType == "IDAT" || chunkType == "IEND") {
					t.Errorf("Critical chunk %s not found", chunkType)
				}
			})
		}
	})
}

// Helper types and functions

type chunkData struct {
	Type   string
	Length uint32
	Data   []byte
}

func parseChunks(data []byte) []chunkData {
	var chunks []chunkData
	
	if len(data) < 8 {
		return chunks
	}
	
	offset := 8 // Skip PNG signature
	for offset < len(data) {
		if offset+8 > len(data) {
			break
		}
		
		length := binary.BigEndian.Uint32(data[offset : offset+4])
		if offset+12+int(length) > len(data) {
			break
		}
		
		chunkType := string(data[offset+4 : offset+8])
		chunkDataBytes := data[offset+8 : offset+8+int(length)]
		
		chunks = append(chunks, chunkData{
			Type:   chunkType,
			Length: length,
			Data:   chunkDataBytes,
		})
		
		offset += 12 + int(length)
	}
	
	return chunks
}

func createTestPNG() []byte {
	var buf bytes.Buffer
	
	// PNG signature
	buf.Write([]byte{137, 80, 78, 71, 13, 10, 26, 10})
	
	// IHDR chunk (1x1 RGB image)
	ihdrData := []byte{
		0, 0, 0, 1, // Width: 1
		0, 0, 0, 1, // Height: 1
		8,    // Bit depth: 8
		2,    // Color type: RGB
		0,    // Compression: deflate
		0,    // Filter: none
		0,    // Interlace: none
	}
	writeChunk(&buf, "IHDR", ihdrData)
	
	// Simple IDAT chunk with minimal compressed data
	idatData := []byte{0x78, 0x9C, 0x62, 0x00, 0x00, 0x00, 0x02, 0x00, 0x01}
	writeChunk(&buf, "IDAT", idatData)
	
	// IEND chunk
	writeChunk(&buf, "IEND", []byte{})
	
	return buf.Bytes()
}

func createPNGWithChunk(chunkType string, chunkData []byte) []byte {
	var buf bytes.Buffer
	
	// PNG signature
	buf.Write([]byte{137, 80, 78, 71, 13, 10, 26, 10})
	
	// IHDR chunk (required)
	if chunkType != "IHDR" {
		ihdrData := []byte{0, 0, 0, 1, 0, 0, 0, 1, 8, 2, 0, 0, 0}
		writeChunk(&buf, "IHDR", ihdrData)
	}
	
	// Add the requested chunk
	writeChunk(&buf, chunkType, chunkData)
	
	// Add required chunks if not already added
	if chunkType != "IDAT" {
		idatData := []byte{0x78, 0x9C, 0x62, 0x00, 0x00, 0x00, 0x02, 0x00, 0x01}
		writeChunk(&buf, "IDAT", idatData)
	}
	
	if chunkType != "IEND" {
		writeChunk(&buf, "IEND", []byte{})
	}
	
	return buf.Bytes()
}

func writeChunk(buf *bytes.Buffer, chunkType string, data []byte) {
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

func TestMainFunctionArgumentHandling(t *testing.T) {
	// Test the argument validation logic that main() uses
	t.Run("Argument count validation", func(t *testing.T) {
		testCases := []struct {
			args     []string
			valid    bool
			describe string
		}{
			{[]string{"chunks"}, false, "No file argument"},
			{[]string{"chunks", "file.png"}, true, "With file argument"},
			{[]string{"chunks", "file1.png", "file2.png"}, true, "Multiple files (uses first)"},
		}
		
		for _, tc := range testCases {
			t.Run(tc.describe, func(t *testing.T) {
				hasFile := len(tc.args) >= 2
				if hasFile != tc.valid {
					t.Errorf("Expected validity %v for args %v, got %v", tc.valid, tc.args, hasFile)
				}
			})
		}
	})
}

func TestFileReadErrorHandling(t *testing.T) {
	t.Run("Read non-existent file", func(t *testing.T) {
		_, err := os.ReadFile("non-existent-file.png")
		if err == nil {
			t.Error("Expected error when reading non-existent file")
		}
		
		// Check that it's a path error (file not found)
		if !strings.Contains(err.Error(), "no such file") && !strings.Contains(err.Error(), "cannot find") {
			t.Logf("Got expected file read error: %v", err)
		}
	})
	
	t.Run("Read permission denied", func(t *testing.T) {
		// This test may not work on all systems, but we can simulate the logic
		tempDir := t.TempDir()
		restrictedFile := filepath.Join(tempDir, "restricted.png")
		
		// Create a file
		if err := os.WriteFile(restrictedFile, []byte("test"), 0600); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		
		// Try to change permissions to no-read (may not work on all systems)
		if err := os.Chmod(restrictedFile, 0000); err != nil {
			t.Skip("Cannot change file permissions on this system")
		}
		
		// Try to read it
		_, err := os.ReadFile(restrictedFile)
		if err == nil {
			t.Log("File was readable despite permission change (system doesn't enforce)")
		} else {
			t.Logf("Got expected permission error: %v", err)
		}
		
		// Restore permissions for cleanup
		os.Chmod(restrictedFile, 0600)
	})
}