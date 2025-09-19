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
	// Generate Go code if requested
	if *generateGo {
		goCodeFile := generateGoFileName(*outputFile)
		// Use empty slice if customLayers is nil
		if customLayers == nil {
			customLayers = []CustomLayer{}
		}
		goCode := generateGoCode(tileMapData, customLayers, filepath.Base(*inputFile))
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

func generateGoCode(mapData *tscnparser.MapData, customLayers []CustomLayer, originalFileName string) string {
	var sb strings.Builder

	// Generate package and imports
	//sb.WriteString("package main\n\n")
	sb.WriteString("// Auto-generated from " + originalFileName + "\n")
	sb.WriteString("// This file contains the parsed tilemap data as Go code\n")
	sb.WriteString("// All type definitions are included, so no external dependencies are needed\n\n")
	sb.WriteString("// Required imports for the LoadFromJson function:\n")
	sb.WriteString("// import (\n")
	sb.WriteString("//     \"encoding/json\"\n")
	sb.WriteString("//     \"fmt\"\n")
	sb.WriteString("//     \"os\"\n")
	sb.WriteString("// )\n\n")

	// Generate type definitions
	sb.WriteString(generateTypeDefinitions())
	sb.WriteString("\n")

	// Generate variable declaration
	sb.WriteString("func Load() *tscnMapData{\n")
	sb.WriteString("return &tscnMapData{\n")

	// Generate TileMap
	sb.WriteString("\tTileMap: tscnTileMapData{\n")
	sb.WriteString(fmt.Sprintf("\t\tFormat: %d,\n", mapData.TileMap.Format))
	sb.WriteString(fmt.Sprintf("\t\tTileSize: tscnTileSize{Width: %d, Height: %d},\n",
		mapData.TileMap.TileSize.Width, mapData.TileMap.TileSize.Height))

	// Generate TileSet
	sb.WriteString("\t\tTileSet: tscnTileSet{\n")
	sb.WriteString("\t\t\tSources: []tscnTileSource{\n")
	for _, source := range mapData.TileMap.TileSet.Sources {
		sb.WriteString("\t\t\t\t{\n")
		sb.WriteString(fmt.Sprintf("\t\t\t\t\tID: %d,\n", source.ID))
		sb.WriteString(fmt.Sprintf("\t\t\t\t\tTexturePath: \"%s\",\n", source.TexturePath))
		sb.WriteString("\t\t\t\t\tTiles: []tscnTileInfo{\n")
		for _, tile := range source.Tiles {
			sb.WriteString("\t\t\t\t\t\t{\n")
			sb.WriteString(fmt.Sprintf("\t\t\t\t\t\t\tAtlasCoords: tscnPoint{X: %d, Y: %d},\n",
				tile.AtlasCoords.X, tile.AtlasCoords.Y))
			sb.WriteString("\t\t\t\t\t\t\tPhysics: tscnPhysicsData{},\n")
			sb.WriteString("\t\t\t\t\t\t},\n")
		}
		sb.WriteString("\t\t\t\t\t},\n")
		sb.WriteString("\t\t\t\t},\n")
	}
	sb.WriteString("\t\t\t},\n")
	sb.WriteString("\t\t},\n")

	// Generate Layers
	sb.WriteString("\t\tLayers: []tscnLayer{\n")
	for _, layer := range customLayers {
		sb.WriteString("\t\t\t{\n")
		sb.WriteString(fmt.Sprintf("\t\t\t\tID: %d,\n", layer.ID))
		sb.WriteString(fmt.Sprintf("\t\t\t\tName: \"%s\",\n", layer.Name))
		sb.WriteString("\t\t\t\tTileData: []int{")
		for i, tileValue := range layer.TileData {
			if i > 0 {
				sb.WriteString(", ")
			}
			if i%20 == 0 {
				sb.WriteString("\n\t\t\t\t\t")
			}
			sb.WriteString(fmt.Sprintf("%d", tileValue))
		}
		sb.WriteString("\n\t\t\t\t},\n")
		sb.WriteString("\t\t\t},\n")
	}
	sb.WriteString("\t\t},\n")
	sb.WriteString("\t},\n")

	// Generate Sprite2Ds
	sb.WriteString("\tSprite2Ds: []tscnSprite2DNode{\n")
	for _, sprite := range mapData.Sprite2Ds {
		sb.WriteString("\t\t{\n")
		sb.WriteString(fmt.Sprintf("\t\t\tName: \"%s\",\n", sprite.Name))
		sb.WriteString(fmt.Sprintf("\t\t\tParent: \"%s\",\n", sprite.Parent))
		sb.WriteString(fmt.Sprintf("\t\t\tPosition: tscnWorldPoint{X: %.1f, Y: %.1f},\n",
			sprite.Position.X, sprite.Position.Y))
		sb.WriteString(fmt.Sprintf("\t\t\tTexturePath: \"%s\",\n", sprite.TexturePath))
		sb.WriteString(fmt.Sprintf("\t\t\tZIndex: %d,\n", sprite.ZIndex))
		sb.WriteString("\t\t},\n")
	}
	sb.WriteString("\t},\n")

	// Generate Prefabs
	sb.WriteString("\tPrefabs: []tscnPrefabNode{\n")
	for _, prefab := range mapData.Prefabs {
		sb.WriteString("\t\t{\n")
		sb.WriteString(fmt.Sprintf("\t\t\tName: \"%s\",\n", prefab.Name))
		sb.WriteString(fmt.Sprintf("\t\t\tParent: \"%s\",\n", prefab.Parent))
		sb.WriteString(fmt.Sprintf("\t\t\tPosition: tscnWorldPoint{X: %.1f, Y: %.1f},\n",
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
	sb.WriteString("}\n")
	sb.WriteString("\n")

	// Add LoadFromJson function
	sb.WriteString(generateLoadFromJsonFunction())
	sb.WriteString("\n")

	// Add runtime parsing utilities
	sb.WriteString(generateRuntimeUtilities())

	return sb.String()
}

func generateLoadFromJsonFunction() string {
	return `// LoadFromJson loads tilemap data from a JSON file
// Usage: mapData := LoadFromJson("path/to/tilemap.json")
func LoadFromJson(dataPath string) *tscnMapData {
	// Read the JSON file
	data, err := os.ReadFile(dataPath)
	if err != nil {
		panic(fmt.Sprintf("Failed to read JSON file %s: %v", dataPath, err))
	}
	
	// Parse JSON into our data structure
	var mapData tscnMapData
	err = json.Unmarshal(data, &mapData)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse JSON file %s: %v", dataPath, err))
	}
	
	return &mapData
}`
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

				// Create layer and add to result
				layer := CustomLayer{
					ID:       currentLayerID,
					Name:     currentLayerName,
					TileData: currentTileData,
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
	TileMap   CustomTileMapData         `json:"tilemap"`
	Sprite2Ds []tscnparser.Sprite2DNode `json:"sprite2ds"`
	Prefabs   []tscnparser.PrefabNode   `json:"prefabs"`
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
		Sprite2Ds: original.Sprite2Ds,
		Prefabs:   original.Prefabs,
	}
}

func generateRuntimeUtilities() string {
	return `
// Runtime utilities for parsing tile data

// TileMapParser provides utilities for parsing compact tile data
// ParseTileData converts compact tile data array to tile instances
// The tile_data format is: [x, source_id, atlas_coords_encoded, x2, source_id2, atlas_coords_encoded2, ...]
// Where atlas_coords_encoded combines atlas X and Y coordinates
func ParseTileData(tileData []int, tileSize tscnTileSize) []tscnTileInstance {
	var tiles []tscnTileInstance
	
	for i := 0; i < len(tileData); i += 3 {
		if i+2 >= len(tileData) {
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
		
		tile := tscnTileInstance{
			TileCoords: tscnPoint{X: int(tileX), Y: int(tileY)},
			WorldCoords: tscnWorldPoint{
				X: float64(tileX * tileSize.Width),
				Y: float64(tileY * tileSize.Height),
			},
			SourceID: sourceID,
			AtlasCoords: tscnPoint{X: int(atlasX), Y: int(atlasY)},
		}
		
		tiles = append(tiles, tile)
	}
	
	return tiles
}

// GetTileAt returns the tile at the specified tile coordinates
func (p *TileMapParser) GetTileAt(tileData []int, targetX, targetY int) (tscnTileInstance, bool) {
	for i := 0; i < len(tileData); i += 3 {
		if i+2 >= len(tileData) {
			break
		}
		
		tilePos := tileData[i]
		sourceID := tileData[i+1]
		atlasEncoded := tileData[i+2]
		
		// Decode tile position
		tileX := tilePos & 0xFFFF
		if tileX >= 0x8000 {
			tileX -= 0x10000
		}
		tileY := (tilePos >> 16) & 0xFFFF
		if tileY >= 0x8000 {
			tileY -= 0x10000
		}
		
		if int(tileX) == targetX && int(tileY) == targetY {
			// Decode atlas coordinates
			atlasX := atlasEncoded & 0xFFFF
			atlasY := (atlasEncoded >> 16) & 0xFFFF
			
			return tscnTileInstance{
				TileCoords: tscnPoint{X: int(tileX), Y: int(tileY)},
				SourceID: sourceID,
				AtlasCoords: tscnPoint{X: int(atlasX), Y: int(atlasY)},
			}, true
		}
	}
	
	return tscnTileInstance{}, false
}
`
}

func generateTypeDefinitions() string {
	return `// Type definitions for tilemap data structures (prefixed with tscn to avoid naming conflicts)

// tscnPoint represents a 2D coordinate
type tscnPoint struct {
	X int ` + "`json:\"x\"`" + `
	Y int ` + "`json:\"y\"`" + `
}

// tscnWorldPoint represents a 2D coordinate in world space (pixels)
type tscnWorldPoint struct {
	X float64 ` + "`json:\"x\"`" + `
	Y float64 ` + "`json:\"y\"`" + `
}

// tscnTileSize represents the dimensions of a tile
type tscnTileSize struct {
	Width  int ` + "`json:\"width\"`" + `
	Height int ` + "`json:\"height\"`" + `
}

// tscnPhysicsData represents physics properties of a tile
type tscnPhysicsData struct {
	CollisionPoints []tscnWorldPoint ` + "`json:\"collision_points,omitempty\"`" + `
}

// tscnTileInfo represents information about a single tile in the tileset
type tscnTileInfo struct {
	AtlasCoords tscnPoint       ` + "`json:\"atlas_coords\"`" + `
	Physics     tscnPhysicsData ` + "`json:\"physics,omitempty\"`" + `
}

// tscnTileSource represents a tileset source
type tscnTileSource struct {
	ID          int            ` + "`json:\"id\"`" + `
	TexturePath string         ` + "`json:\"texture_path\"`" + `
	Tiles       []tscnTileInfo ` + "`json:\"tiles\"`" + `
}

// tscnTileSet represents the complete tileset information
type tscnTileSet struct {
	Sources []tscnTileSource ` + "`json:\"sources\"`" + `
}

// tscnTileInstance represents a placed tile in the map
type tscnTileInstance struct {
	TileCoords  tscnPoint      ` + "`json:\"tile_coords\"`" + `
	WorldCoords tscnWorldPoint ` + "`json:\"world_coords\"`" + `
	SourceID    int            ` + "`json:\"source_id\"`" + `
	AtlasCoords tscnPoint      ` + "`json:\"atlas_coords\"`" + `
}

// tscnLayer represents a tilemap layer with compact tile data format
type tscnLayer struct {
	ID       int    ` + "`json:\"id\"`" + `
	Name     string ` + "`json:\"name\"`" + `
	TileData []int  ` + "`json:\"tile_data\"`" + `
}

// tscnTileMapData represents the complete tilemap data
type tscnTileMapData struct {
	Format   int          ` + "`json:\"format\"`" + `
	TileSize tscnTileSize ` + "`json:\"tile_size\"`" + `
	TileSet  tscnTileSet  ` + "`json:\"tileset\"`" + `
	Layers   []tscnLayer  ` + "`json:\"layers\"`" + `
}

// tscnSprite2DNode represents a Sprite2D node in the scene
type tscnSprite2DNode struct {
	Name        string         ` + "`json:\"name\"`" + `
	Parent      string         ` + "`json:\"parent\"`" + `
	Position    tscnWorldPoint ` + "`json:\"position\"`" + `
	TexturePath string         ` + "`json:\"texture_path\"`" + `
	ZIndex      int            ` + "`json:\"z_index,omitempty\"`" + `
}

// tscnPrefabNode represents an instantiated prefab node in the scene
type tscnPrefabNode struct {
	Name       string                 ` + "`json:\"name\"`" + `
	Parent     string                 ` + "`json:\"parent\"`" + `
	Position   tscnWorldPoint         ` + "`json:\"position\"`" + `
	PrefabPath string                 ` + "`json:\"prefab_path\"`" + `
	Properties map[string]interface{} ` + "`json:\"properties,omitempty\"`" + `
}

// tscnMapData represents the root structure for JSON output
type tscnMapData struct {
	TileMap   tscnTileMapData    ` + "`json:\"tilemap\"`" + `
	Sprite2Ds []tscnSprite2DNode ` + "`json:\"sprite2ds\"`" + `
	Prefabs   []tscnPrefabNode   ` + "`json:\"prefabs\"`" + `
}
`
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
