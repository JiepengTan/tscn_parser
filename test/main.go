package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	tscnparser "github.com/JiepengTan/tscn_parser"
)

func main() {
	var inputFile = flag.String("input", "", "Input TSCN file path")
	var outputFile = flag.String("output", "", "Output JSON file path")
	flag.Parse()

	if *inputFile == "" {
		log.Fatal("Please provide input TSCN file with -input flag")
	}

	if *outputFile == "" {
		// Generate output filename based on input
		ext := filepath.Ext(*inputFile)
		*outputFile = (*inputFile)[:len(*inputFile)-len(ext)] + "_tilemap.json"
	}

	// Parse TSCN file
	tileMapData, err := tscnparser.Parse(*inputFile)
	if err != nil {
		log.Fatalf("Error converting TSCN: %v", err)
	}

	// Output to JSON
	jsonData, err := json.MarshalIndent(tileMapData, "", "  ")
	if err != nil {
		log.Fatalf("Error marshaling JSON: %v", err)
	}

	err = os.WriteFile(*outputFile, jsonData, 0644)
	if err != nil {
		log.Fatalf("Error writing output file: %v", err)
	}

	fmt.Printf("Successfully converted %s to %s\n", *inputFile, *outputFile)
}
