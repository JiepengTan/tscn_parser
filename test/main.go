package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	tscnparser "github.com/JiepengTan/tscn_parser"
)

type Replacement struct {
	Old string `json:"old"`
	New string `json:"new"`
}

type ReplacementConfig struct {
	Replacements []Replacement `json:"replacements"`
}

func main() {
	var inputFile = flag.String("input", "", "Input TSCN file path")
	var outputFile = flag.String("output", "", "Output JSON file path")
	var replacementsFile = flag.String("replacements", "", "JSON file containing replacement rules")
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

	// Output to JSON with custom layers if available
	var jsonData []byte
	jsonData, err = json.MarshalIndent(tileMapData, "", "  ")
	if err != nil {
		log.Fatalf("Error marshaling JSON: %v", err)
	}

	// Apply string replacements
	jsonStr := string(jsonData)

	// Apply replacements from JSON file first (if provided)
	if *replacementsFile != "" {
		jsonStr = applyReplacementsFromFile(jsonStr, *replacementsFile)
	}

	err = os.WriteFile(*outputFile, []byte(jsonStr), 0644)
	if err != nil {
		log.Fatalf("Error writing output file: %v", err)
	}
	tileMapData = &tscnparser.MapData{}
	json.Unmarshal([]byte(jsonStr), tileMapData)

	fmt.Printf("Successfully converted %s to %s\n", *inputFile, *outputFile)
}

func applyReplacementsFromFile(jsonStr, replacementsFile string) string {
	// Read replacements file
	data, err := os.ReadFile(replacementsFile)
	if err != nil {
		log.Printf("Warning: Could not read replacements file %s: %v", replacementsFile, err)
		return jsonStr
	}

	// Parse JSON
	var config ReplacementConfig
	if err := json.Unmarshal(data, &config); err != nil {
		log.Printf("Warning: Could not parse replacements file %s: %v", replacementsFile, err)
		return jsonStr
	}

	// Apply all replacements
	result := jsonStr
	for _, replacement := range config.Replacements {
		if replacement.Old != "" {
			result = strings.ReplaceAll(result, replacement.Old, replacement.New)
		}
	}

	return result
}
