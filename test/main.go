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
	var generateGo = flag.Bool("generateGo", false, "Generate Go code file (.go.txt)")
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
	// Generate Go code if requested
	if *generateGo {
		goCodeFile := generateGoFileName(*outputFile)
		goCode := generateGoCode(tileMapData, filepath.Base(*inputFile))
		err = os.WriteFile(goCodeFile, []byte(goCode), 0644)
		if err != nil {
			log.Fatalf("Error writing Go code file: %v", err)
		}
		fmt.Printf("Successfully generated Go code: %s\n", goCodeFile)
	}

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

func generateGoFileName(jsonFile string) string {
	ext := filepath.Ext(jsonFile)
	baseName := jsonFile[:len(jsonFile)-len(ext)]
	return baseName + ".go.txt"
}

func generateGoCode(mapData *tscnparser.MapData, originalFileName string) string {
	var sb strings.Builder

	// Generate package and imports
	sb.WriteString("package main\n\n")
	sb.WriteString("// Auto-generated from " + originalFileName + "\n")
	sb.WriteString("// This file contains the parsed tilemap data as Go code\n\n")

	// Generate variable declaration
	sb.WriteString("var MapData = &MapData{\n")

	// Generate TileMap
	sb.WriteString("\tTileMap: TileMapData{\n")
	sb.WriteString(fmt.Sprintf("\t\tFormat: %d,\n", mapData.TileMap.Format))
	sb.WriteString(fmt.Sprintf("\t\tTileSize: TileSize{Width: %d, Height: %d},\n",
		mapData.TileMap.TileSize.Width, mapData.TileMap.TileSize.Height))

	// Generate TileSet
	sb.WriteString("\t\tTileSet: TileSet{\n")
	sb.WriteString("\t\t\tSources: []TileSource{\n")
	for _, source := range mapData.TileMap.TileSet.Sources {
		sb.WriteString("\t\t\t\t{\n")
		sb.WriteString(fmt.Sprintf("\t\t\t\t\tID: %d,\n", source.ID))
		sb.WriteString(fmt.Sprintf("\t\t\t\t\tTexturePath: \"%s\",\n", source.TexturePath))
		sb.WriteString("\t\t\t\t\tTiles: []TileInfo{\n")
		for _, tile := range source.Tiles {
			sb.WriteString("\t\t\t\t\t\t{\n")
			sb.WriteString(fmt.Sprintf("\t\t\t\t\t\t\tAtlasCoords: Point{X: %d, Y: %d},\n",
				tile.AtlasCoords.X, tile.AtlasCoords.Y))
			sb.WriteString("\t\t\t\t\t\t\tPhysics: PhysicsData{},\n")
			sb.WriteString("\t\t\t\t\t\t},\n")
		}
		sb.WriteString("\t\t\t\t\t},\n")
		sb.WriteString("\t\t\t\t},\n")
	}
	sb.WriteString("\t\t\t},\n")
	sb.WriteString("\t\t},\n")

	// Generate Layers
	sb.WriteString("\t\tLayers: []Layer{\n")
	for _, layer := range mapData.TileMap.Layers {
		sb.WriteString("\t\t\t{\n")
		sb.WriteString(fmt.Sprintf("\t\t\t\tID: %d,\n", layer.ID))
		sb.WriteString(fmt.Sprintf("\t\t\t\tName: \"%s\",\n", layer.Name))
		sb.WriteString("\t\t\t\tTiles: []TileInstance{\n")
		for _, tile := range layer.Tiles {
			sb.WriteString("\t\t\t\t\t{\n")
			sb.WriteString(fmt.Sprintf("\t\t\t\t\t\tTileCoords: Point{X: %d, Y: %d},\n",
				tile.TileCoords.X, tile.TileCoords.Y))
			sb.WriteString(fmt.Sprintf("\t\t\t\t\t\tWorldCoords: WorldPoint{X: %.1f, Y: %.1f},\n",
				tile.WorldCoords.X, tile.WorldCoords.Y))
			sb.WriteString(fmt.Sprintf("\t\t\t\t\t\tSourceID: %d,\n", tile.SourceID))
			sb.WriteString(fmt.Sprintf("\t\t\t\t\t\tAtlasCoords: Point{X: %d, Y: %d},\n",
				tile.AtlasCoords.X, tile.AtlasCoords.Y))
			sb.WriteString("\t\t\t\t\t},\n")
		}
		sb.WriteString("\t\t\t\t},\n")
		sb.WriteString("\t\t\t},\n")
	}
	sb.WriteString("\t\t},\n")
	sb.WriteString("\t},\n")

	// Generate Sprite2Ds
	sb.WriteString("\tSprite2Ds: []Sprite2DNode{\n")
	for _, sprite := range mapData.Sprite2Ds {
		sb.WriteString("\t\t{\n")
		sb.WriteString(fmt.Sprintf("\t\t\tName: \"%s\",\n", sprite.Name))
		sb.WriteString(fmt.Sprintf("\t\t\tParent: \"%s\",\n", sprite.Parent))
		sb.WriteString(fmt.Sprintf("\t\t\tPosition: WorldPoint{X: %.1f, Y: %.1f},\n",
			sprite.Position.X, sprite.Position.Y))
		sb.WriteString(fmt.Sprintf("\t\t\tTexturePath: \"%s\",\n", sprite.TexturePath))
		sb.WriteString(fmt.Sprintf("\t\t\tZIndex: %d,\n", sprite.ZIndex))
		sb.WriteString("\t\t},\n")
	}
	sb.WriteString("\t},\n")

	// Generate Prefabs
	sb.WriteString("\tPrefabs: []PrefabNode{\n")
	for _, prefab := range mapData.Prefabs {
		sb.WriteString("\t\t{\n")
		sb.WriteString(fmt.Sprintf("\t\t\tName: \"%s\",\n", prefab.Name))
		sb.WriteString(fmt.Sprintf("\t\t\tParent: \"%s\",\n", prefab.Parent))
		sb.WriteString(fmt.Sprintf("\t\t\tPosition: WorldPoint{X: %.1f, Y: %.1f},\n",
			prefab.Position.X, prefab.Position.Y))
		sb.WriteString(fmt.Sprintf("\t\t\tPrefabPath: \"%s\",\n", prefab.PrefabPath))

		if len(prefab.Properties) > 0 {
			sb.WriteString("\t\t\tProperties: map[string]interface{}{\n")
			for key, value := range prefab.Properties {
				sb.WriteString(fmt.Sprintf("\t\t\t\t\"%s\": %v,\n", key, formatGoValue(value)))
			}
			sb.WriteString("\t\t\t},\n")
		} else {
			sb.WriteString("\t\t\tProperties: map[string]interface{}{},\n")
		}
		sb.WriteString("\t\t},\n")
	}
	sb.WriteString("\t},\n")

	sb.WriteString("}\n")

	return sb.String()
}

func formatGoValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		return fmt.Sprintf("\"%s\"", v)
	case int:
		return fmt.Sprintf("%d", v)
	case float64:
		return fmt.Sprintf("%.1f", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}
