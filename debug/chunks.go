package main

import (
	"encoding/binary"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run chunks.go <png-file>")
		return
	}

	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	fmt.Printf("File: %s (%d bytes)\n", os.Args[1], len(data))
	fmt.Println("Chunks:")

	offset := 8 // Skip PNG signature
	for offset < len(data) {
		if offset+8 > len(data) {
			break
		}

		length := binary.BigEndian.Uint32(data[offset : offset+4])
		chunkType := string(data[offset+4 : offset+8])

		fmt.Printf("  %s: %d bytes\n", chunkType, length)

		offset += 12 + int(length)
	}
}
