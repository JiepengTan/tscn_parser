package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
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

	// Parse layer data directly from TSCN file since tscnparser doesn't handle it properly
	customLayers, err := parseLayersFromTSCN(*inputFile)
	if err != nil {
		log.Printf("Warning: Could not parse layers from TSCN: %v", err)
	} else {
		// Convert CustomLayer to tscnparser.Layer format for the generation code
		convertedLayers := make([]tscnparser.Layer, len(customLayers))
		for i, customLayer := range customLayers {
			// Create a compatible layer - we'll use our custom data in the generation
			convertedLayers[i] = tscnparser.Layer{
				ID:   customLayer.ID,
				Name: customLayer.Name,
				// Note: tscnparser.Layer doesn't have TileData, so we'll handle this in generateGoCode
			}
		}
		tileMapData.TileMap.Layers = convertedLayers

		fmt.Printf("Successfully parsed %d layers from TSCN file\n", len(customLayers))
	}

	// Output to JSON with custom layers if available
	var jsonData []byte
	if customLayers != nil && len(customLayers) > 0 {
		// Create a custom structure that includes our tile data
		customOutput := createCustomJSONOutput(tileMapData, customLayers)
		jsonData, err = json.MarshalIndent(customOutput, "", "  ")
	} else {
		jsonData, err = json.MarshalIndent(tileMapData, "", "  ")
	}
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

// Define our own layer structure with TileData field
type CustomLayer struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	TileData []int  `json:"tile_data"`
}

func parseLayersFromTSCN(filename string) ([]CustomLayer, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var layers []CustomLayer
	lines := strings.Split(string(content), "\n")

	currentLayerID := -1
	currentLayerName := ""
	var currentTileData []int

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Parse layer name: layer_0/name = "1"
		if strings.Contains(line, "/name = ") {
			parts := strings.Split(line, "/name = ")
			if len(parts) == 2 {
				layerIDStr := strings.Replace(parts[0], "layer_", "", 1)
				if layerID, err := strconv.Atoi(layerIDStr); err == nil {
					currentLayerID = layerID
					currentLayerName = strings.Trim(parts[1], "\"")
				}
			}
		}

		// Parse tile data: layer_0/tile_data = PackedInt32Array(...)
		if strings.Contains(line, "/tile_data = PackedInt32Array(") && currentLayerID >= 0 {
			// Extract the data inside PackedInt32Array(...)
			start := strings.Index(line, "PackedInt32Array(") + len("PackedInt32Array(")
			end := strings.LastIndex(line, ")")
			if end > start {
				dataStr := line[start:end]

				// Parse the comma-separated integers
				if dataStr != "" {
					parts := strings.Split(dataStr, ",")
					for _, part := range parts {
						part = strings.TrimSpace(part)
						if value, err := strconv.Atoi(part); err == nil {
							currentTileData = append(currentTileData, value)
						}
					}
				}

				// Convert from old format [encoded_position, source_id, atlas_coords] to new format [source_id, tile_x, tile_y, atlas_x, atlas_y]
				convertedTileData := convertTileDataFormat(currentTileData)

				// Create layer and add to result
				layer := CustomLayer{
					ID:       currentLayerID,
					Name:     currentLayerName,
					TileData: convertedTileData,
				}
				layers = append(layers, layer)

				// Reset for next layer
				currentLayerID = -1
				currentLayerName = ""
				currentTileData = nil
			}
		}
	}

	return layers, nil
}

// Custom JSON output structure that includes tile data
type CustomJSONOutput struct {
	TileMap    CustomTileMapData          `json:"tilemap"`
	Decorators []tscnparser.DecoratorNode `json:"decorators"`
	Sprites    []tscnparser.SpriteNode    `json:"sprites"`
}

type CustomTileMapData struct {
	Format   int                 `json:"format"`
	TileSize tscnparser.TileSize `json:"tile_size"`
	TileSet  tscnparser.TileSet  `json:"tileset"`
	Layers   []CustomLayer       `json:"layers"`
}

func createCustomJSONOutput(original *tscnparser.MapData, customLayers []CustomLayer) *CustomJSONOutput {
	return &CustomJSONOutput{
		TileMap: CustomTileMapData{
			Format:   original.TileMap.Format,
			TileSize: original.TileMap.TileSize,
			TileSet:  original.TileMap.TileSet,
			Layers:   customLayers,
		},
		Decorators: original.Decorators,
		Sprites:    original.Sprites,
	}
}

// convertTileDataFormat converts tile data from old format to new format
// Old format: [tilePos, source_id, atlas_coords_encoded] (3 elements per tile)
// New format: [source_id, tile_x, tile_y, atlas_x, atlas_y] (5 elements per tile)
// This function uses the original parsing logic from internal/tilemap/tilemap.go before commit f81157b
func convertTileDataFormat(tileData []int) []int {
	var newData []int
	lenght := len(tileData)

	// Original parsing logic from internal/tilemap/tilemap.go
	for i := 0; i < lenght; i += 3 {
		if i+2 >= lenght {
			break
		}
		tilePos := tileData[i]
		sourceID := tileData[i+1]
		atlasEncoded := tileData[i+2]

		// Decode tile position (Godot uses a specific encoding)
		tileX := tilePos & 0xFFFF
		if tileX >= 0x8000 {
			tileX -= 0x10000 // Handle negative coordinates
		}
		tileY := (tilePos >> 16) & 0xFFFF
		if tileY >= 0x8000 {
			tileY -= 0x10000 // Handle negative coordinates
		}

		// Decode atlas coordinates (usually just X and Y)
		atlasX := atlasEncoded & 0xFFFF
		atlasY := (atlasEncoded >> 16) & 0xFFFF

		// Append in new format: [source_id, tile_x, tile_y, atlas_x, atlas_y]
		newData = append(newData, sourceID, tileX, -tileY, atlasX, atlasY)
	}

	return newData
}
