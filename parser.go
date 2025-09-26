package tscnparser

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// ExtResource represents an external resource reference
type ExtResource struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Path string `json:"path"`
	UID  string `json:"uid,omitempty"`
}

var (
	tilemapTileSize = TileSize{Width: 16, Height: 16}
)

// TSCNConverter handles conversion from TSCN to TileMap JSON
type TSCNConverter struct {
	tileSize            TileSize
	sources             map[int]*TileSource
	extResources        map[string]*ExtResource
	subResourceTextures map[string]string // Maps SubResource ID to ExtResource ID
	tilePhysicsData     map[string][]Vec2 // Maps SubResource ID to physics points
	decorators          []DecoratorNode   // Collected Decorator nodes
	currentDecorator    *DecoratorNode    // Currently parsing Decorator node
	sprites             []SpriteNode      // Collected Sprite nodes
	currentSprite       *SpriteNode       // Currently parsing Sprite node
}

// NewTSCNConverter creates a new converter instance
func newTSCNConverter() *TSCNConverter {
	return &TSCNConverter{
		tileSize:            TileSize{Width: 16, Height: 16}, // Default tile size
		sources:             make(map[int]*TileSource),
		extResources:        make(map[string]*ExtResource),
		subResourceTextures: make(map[string]string),
		tilePhysicsData:     make(map[string][]Vec2),
		decorators:          []DecoratorNode{},
		sprites:             []SpriteNode{},
	}
}

func parseLayersFromTSCN(filename string) ([]Layer, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var layers []Layer
	lines := strings.Split(string(content), "\n")

	currentLayerID := -1
	currentLayerName := ""
	currentZIndex := 0
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

		// Parse layer z_index: layer_0/z_index = -3
		if strings.Contains(line, "/z_index = ") {
			parts := strings.Split(line, "/z_index = ")
			if len(parts) == 2 {
				if zIndex, err := strconv.Atoi(strings.TrimSpace(parts[1])); err == nil {
					currentZIndex = zIndex
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
				layer := Layer{
					ID:       currentLayerID,
					Name:     currentLayerName,
					ZIndex:   currentZIndex,
					TileData: convertedTileData,
				}
				layers = append(layers, layer)

				// Reset for next layer
				currentLayerID = -1
				currentLayerName = ""
				currentZIndex = 0
				currentTileData = nil
			}
		}
	}
	println("minTile ,max Tile", minYTile, maxYTile, "diff ", maxYTile-minYTile, "total ", tileTotalCount)
	return layers, nil
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

		if tileY < minYTile {
			minYTile = tileY
		}
		if tileY > maxYTile {
			maxYTile = tileY
		}
		// Decode atlas coordinates (usually just X and Y)
		atlasX := atlasEncoded & 0xFFFF
		atlasY := (atlasEncoded >> 16) & 0xFFFF
		tileTotalCount++
		// Append in new format: [source_id, tile_x, tile_y, atlas_x, atlas_y]
		newData = append(newData, sourceID, tileX, -tileY, atlasX, atlasY)
	}

	return newData
}

// ConvertTSCNToTileMap converts a TSCN file to TileMap data structure
func (c *TSCNConverter) ConvertTSCNToTileMap(filename string) (*MapData, error) {
	data, err := c.convertTSCNToTileMap(filename)
	// Parse layer data directly from TSCN file since tscnparser doesn't handle it properly
	data.TileMap.Layers, _ = parseLayersFromTSCN(filename)
	return data, err
}

var (
	minYTile, maxYTile int
	tileTotalCount     int
)

func (c *TSCNConverter) convertTSCNToTileMap(filename string) (*MapData, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// Increase buffer size to handle very long lines
	const maxCapacity = 1024 * 1024 // 1MB buffer
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	var currentSection string
	var currentSubResource string
	var format int
	var layers []Layer
	minYTile = 1000000
	maxYTile = -1000000
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Detect sections
		if strings.HasPrefix(line, "[") {
			if strings.Contains(line, "ext_resource") {
				currentSection = "ext_resource"
				// Parse ExtResource immediately since it's all on one line
				c.parseExtResource(line)
			} else if strings.Contains(line, "sub_resource") {
				currentSection = "sub_resource"
				currentSubResource = c.extractSubResourceID(line)
			} else if strings.Contains(line, "node name=\"TileMap\"") {
				currentSection = "tilemap"
			} else if strings.Contains(line, "type=\"Sprite2D\"") {
				// Finish current nodes if we were processing any
				if currentSection == "decorator" && c.currentDecorator != nil {
					c.decorators = append(c.decorators, *c.currentDecorator)
					c.currentDecorator = nil
				}
				if currentSection == "sprite" && c.currentSprite != nil {
					c.sprites = append(c.sprites, *c.currentSprite)
					c.currentSprite = nil
				}
				currentSection = "decorator"
				// Initialize new Decorator node
				c.currentDecorator = c.parseDecoratorNode(line)
			} else if strings.Contains(line, "instance=ExtResource") {
				// Finish current nodes if we were processing any
				if currentSection == "decorator" && c.currentDecorator != nil {
					c.decorators = append(c.decorators, *c.currentDecorator)
					c.currentDecorator = nil
				}
				if currentSection == "sprite" && c.currentSprite != nil {
					c.sprites = append(c.sprites, *c.currentSprite)
					c.currentSprite = nil
				}
				currentSection = "sprite"
				// Initialize new Sprite node
				c.currentSprite = c.parseSpriteNode(line)
			} else {
				// Finish current nodes if we're leaving their sections
				if currentSection == "decorator" && c.currentDecorator != nil {
					c.decorators = append(c.decorators, *c.currentDecorator)
					c.currentDecorator = nil
				}
				if currentSection == "sprite" && c.currentSprite != nil {
					c.sprites = append(c.sprites, *c.currentSprite)
					c.currentSprite = nil
				}
				currentSection = "other"
			}
			continue
		}

		// Parse content based on current section
		switch currentSection {
		case "ext_resource":
			c.parseExtResource(line)
		case "sub_resource":
			c.parseSubResource(line, currentSubResource)
		case "decorator":
			c.parseDecoratorProperty(line)
		case "sprite":
			c.parseSpriteProperty(line)
		case "tilemap":
			if strings.HasPrefix(line, "format =") {
				format = c.extractIntValue(line)
			} else if strings.HasPrefix(line, "layer_") && strings.Contains(line, "tile_data") {
				layerID := c.extractLayerID(line)
				tileData := c.extractTileData(line)
				layer := c.parseTileData(layerID, tileData)
				layers = append(layers, layer)
			}
		}
	}

	// Handle any remaining Decorator node
	if c.currentDecorator != nil {
		c.decorators = append(c.decorators, *c.currentDecorator)
	}

	// Handle any remaining Sprite node
	if c.currentSprite != nil {
		c.sprites = append(c.sprites, *c.currentSprite)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	// Build tileset from sources
	var tilesetSources []TileSource
	for _, source := range c.sources {
		tilesetSources = append(tilesetSources, *source)
	}
	// sort tilesetSources
	sort.Slice(tilesetSources, func(i, j int) bool {
		return tilesetSources[i].ID < tilesetSources[j].ID
	})
	return &MapData{
		TileMap: TileMapData{
			Format:   format,
			TileSize: tilemapTileSize,
			TileSet: TileSet{
				Sources: tilesetSources,
			},
			Layers: layers,
		},
		Decorators: c.decorators,
		Sprites:    c.sprites,
	}, nil
}

// parseExtResource parses external resource declarations
func (c *TSCNConverter) parseExtResource(line string) {
	// ExtResource format: [ext_resource type="Texture2D" uid="uid://..." path="res://..." id="1_grrf0"]
	// Note: the line detection happens in the main loop, here we process any line in ext_resource section
	extRes := &ExtResource{}

	// Extract type
	if re := regexp.MustCompile(`type="([^"]+)"`); re != nil {
		if matches := re.FindStringSubmatch(line); len(matches) > 1 {
			extRes.Type = matches[1]
		}
	}

	// Extract path
	if re := regexp.MustCompile(`path="([^"]+)"`); re != nil {
		if matches := re.FindStringSubmatch(line); len(matches) > 1 {
			extRes.Path = matches[1]
		}
	}

	// Extract id (the actual resource id, not the uid)
	if re := regexp.MustCompile(`\sid="([^"]+)"`); re != nil {
		matches := re.FindAllStringSubmatch(line, -1)
		if len(matches) > 0 {
			// Get the last match (the actual ID field, not the UID)
			lastMatch := matches[len(matches)-1]
			if len(lastMatch) > 1 {
				extRes.ID = lastMatch[1]
			}
		}
	}

	// Extract uid
	if re := regexp.MustCompile(`uid="([^"]+)"`); re != nil {
		if matches := re.FindStringSubmatch(line); len(matches) > 1 {
			extRes.UID = matches[1]
		}
	}

	if extRes.ID != "" {
		c.extResources[extRes.ID] = extRes
	}
}

// extractSubResourceID extracts the SubResource ID from a line
func (c *TSCNConverter) extractSubResourceID(line string) string {
	re := regexp.MustCompile(`id="([^"]+)"`)
	matches := re.FindStringSubmatch(line)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// parseSubResource parses sub-resource data (TileSetAtlasSource, TileSet)
func (c *TSCNConverter) parseSubResource(line, resourceID string) {
	if strings.HasPrefix(line, "texture = ExtResource(") {
		// Extract ExtResource ID from the line
		extResourceID := c.extractExtResourceID(line)
		if extResourceID != "" && resourceID != "" {
			// Store the mapping for later use
			c.subResourceTextures[resourceID] = extResourceID
		}
	} else if strings.HasPrefix(line, "0:0/0/physics_layer_0/polygon_0/points") {
		// Parse collision polygon to determine tile size and store physics data
		points := c.extractPolygonPoints(line)
		if len(points) >= 4 {
			// Calculate tile size from collision box
			c.tileSize = c.calculateTileSizeFromPoints(points)
			// Store physics data for this resource
			c.tilePhysicsData[resourceID] = points
		}
	} else if strings.HasPrefix(line, "sources/") {
		// Parse tileset sources
		sourceID := c.extractSourceID(line)
		subResourceID := c.extractSubResourceReference(line)
		if sourceID != -1 {
			texturePath := "unknown"
			// Try to resolve texture path using our mappings
			if subResourceID != "" {
				if extResourceID, exists := c.subResourceTextures[subResourceID]; exists {
					if extRes, extExists := c.extResources[extResourceID]; extExists {
						texturePath = extRes.Path
					}
				}
			}

			// Get physics data for this tile source
			var physicsData PhysicsData
			if subResourceID != "" {
				if points, hasPhysics := c.tilePhysicsData[subResourceID]; hasPhysics {
					physicsData = PhysicsData{CollisionPoints: points}
				}
			}

			c.sources[sourceID] = &TileSource{
				ID:          sourceID,
				TexturePath: texturePath,
				Tiles:       []TileInfo{{AtlasCoords: Vec2i{X: 0, Y: 0}, Physics: physicsData}},
			}
		}
	}
}

// extractLayerID extracts layer ID from tile_data line
func (c *TSCNConverter) extractLayerID(line string) int {
	re := regexp.MustCompile(`layer_(\d+)/`)
	matches := re.FindStringSubmatch(line)
	if len(matches) > 1 {
		id, _ := strconv.Atoi(matches[1])
		return id
	}
	return 0
}

// extractTileData extracts PackedInt32Array data from tile_data line
func (c *TSCNConverter) extractTileData(line string) []int {
	// Find PackedInt32Array content
	start := strings.Index(line, "PackedInt32Array(")
	if start == -1 {
		return []int{}
	}
	start += len("PackedInt32Array(")

	end := strings.LastIndex(line, ")")
	if end == -1 {
		return []int{}
	}

	content := line[start:end]
	parts := strings.Split(content, ",")
	var data []int

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if val, err := strconv.Atoi(part); err == nil {
			data = append(data, val)
		}
	}

	return data
}

// parseTileData converts raw tile data to Layer structure
func (c *TSCNConverter) parseTileData(layerID int, data []int) Layer {
	layer := Layer{
		ID:    layerID,
		Name:  fmt.Sprintf("layer_%d", layerID),
		Tiles: []TileInstance{},
	}

	// Process data in groups of 5: [source_id, tile_x, tile_y, atlas_x, atlas_y]
	for i := 0; i < len(data); i += 5 {
		if i+4 >= len(data) {
			break
		}

		sourceID := data[i]
		tileX := data[i+1]
		tileY := data[i+2]
		atlasX := data[i+3]
		atlasY := data[i+4]

		tile := TileInstance{
			TileCoords:  Vec2i{X: tileX, Y: tileY},
			SourceID:    sourceID,
			AtlasCoords: Vec2i{X: atlasX, Y: atlasY},
		}

		layer.Tiles = append(layer.Tiles, tile)
	}

	return layer
}

// extractExtResourceID extracts ExtResource ID from texture assignment
func (c *TSCNConverter) extractExtResourceID(line string) string {
	// Pattern: texture = ExtResource("1_grrf0")
	re := regexp.MustCompile(`ExtResource\("([^"]+)"\)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// extractSubResourceReference extracts SubResource ID from sources/ line
func (c *TSCNConverter) extractSubResourceReference(line string) string {
	// Pattern: sources/0 = SubResource("TileSetAtlasSource_8xjng")
	re := regexp.MustCompile(`SubResource\("([^"]+)"\)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// Helper functions
func (c *TSCNConverter) extractIntValue(line string) int {
	parts := strings.Split(line, "=")
	if len(parts) > 1 {
		val, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
		return val
	}
	return 0
}

func (c *TSCNConverter) extractSourceID(line string) int {
	re := regexp.MustCompile(`sources/(\d+)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) > 1 {
		id, _ := strconv.Atoi(matches[1])
		return id
	}
	return -1
}

func (c *TSCNConverter) extractPolygonPoints(line string) []Vec2 {
	re := regexp.MustCompile(`PackedVector2Array\(([^)]+)\)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) < 2 {
		return []Vec2{}
	}

	content := matches[1]
	parts := strings.Split(content, ",")
	var points []Vec2

	for i := 0; i < len(parts); i += 2 {
		if i+1 >= len(parts) {
			break
		}

		x, err1 := strconv.ParseFloat(strings.TrimSpace(parts[i]), 64)
		y, err2 := strconv.ParseFloat(strings.TrimSpace(parts[i+1]), 64)

		if err1 == nil && err2 == nil {
			points = append(points, Vec2{X: x, Y: y})
		}
	}

	return points
}

func (c *TSCNConverter) calculateTileSizeFromPoints(points []Vec2) TileSize {
	if len(points) < 4 {
		return c.tileSize // Return default
	}

	// Assuming rectangular collision box, find width and height
	minX, maxX := points[0].X, points[0].X
	minY, maxY := points[0].Y, points[0].Y

	for _, point := range points {
		if point.X < minX {
			minX = point.X
		}
		if point.X > maxX {
			maxX = point.X
		}
		if point.Y < minY {
			minY = point.Y
		}
		if point.Y > maxY {
			maxY = point.Y
		}
	}

	return TileSize{
		Width:  int(maxX - minX),
		Height: int(maxY - minY),
	}
}

// parseDecoratorNode creates a new Decorator node from the node declaration line
func (c *TSCNConverter) parseDecoratorNode(line string) *DecoratorNode {
	// Extract node name and parent from line like: [node name="Cloud1" type="Sprite2D" parent="Decorations/Clouds"]
	decorator := &DecoratorNode{
		Path: "unknown", // Default until we find texture property
	}

	// Extract name
	if re := regexp.MustCompile(`name="([^"]+)"`); re != nil {
		if matches := re.FindStringSubmatch(line); len(matches) > 1 {
			decorator.Name = matches[1]
		}
	}

	// Extract parent
	if re := regexp.MustCompile(`parent="([^"]+)"`); re != nil {
		if matches := re.FindStringSubmatch(line); len(matches) > 1 {
			decorator.Parent = matches[1]
		}
	}

	return decorator
}

// parseDecoratorProperty parses properties of the current Decorator node
func (c *TSCNConverter) parseDecoratorProperty(line string) {
	if c.currentDecorator == nil {
		return
	}

	if strings.HasPrefix(line, "position = Vector2(") {
		// Extract position coordinates
		position := c.extractVector2(line)
		position.Y = -position.Y
		c.currentDecorator.Position = position
	} else if strings.HasPrefix(line, "texture = ExtResource(") {
		// Extract texture ExtResource ID and resolve path
		extResourceID := c.extractExtResourceID(line)
		if extRes, exists := c.extResources[extResourceID]; exists {
			c.currentDecorator.Path = extRes.Path
		}
	} else if strings.HasPrefix(line, "z_index = ") {
		// Extract z_index
		c.currentDecorator.ZIndex = c.extractIntValue(line)
	}
}

// extractVector2 extracts Vector2 coordinates from a line like "position = Vector2(-120, -48)"
func (c *TSCNConverter) extractVector2(line string) Vec2 {
	re := regexp.MustCompile(`Vector2\(([^,]+),\s*([^)]+)\)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) < 3 {
		return Vec2{X: 0, Y: 0}
	}

	x, err1 := strconv.ParseFloat(strings.TrimSpace(matches[1]), 64)
	y, err2 := strconv.ParseFloat(strings.TrimSpace(matches[2]), 64)

	if err1 != nil || err2 != nil {
		return Vec2{X: 0, Y: 0}
	}

	return Vec2{X: x, Y: y}
}

// parseSpriteNode creates a new Sprite node from the node declaration line
func (c *TSCNConverter) parseSpriteNode(line string) *SpriteNode {
	// Extract node name, parent, and instance from line like: [node name="Brick" parent="Environment/Platforms/Platform1" instance=ExtResource("6_vt4yb")]
	sprite := &SpriteNode{
		Path:       "unknown", // Default until we resolve ExtResource
		Properties: make(map[string]any),
	}

	// Extract name
	if re := regexp.MustCompile(`name="([^"]+)"`); re != nil {
		if matches := re.FindStringSubmatch(line); len(matches) > 1 {
			sprite.Name = matches[1]
		}
	}

	// Extract parent
	if re := regexp.MustCompile(`parent="([^"]+)"`); re != nil {
		if matches := re.FindStringSubmatch(line); len(matches) > 1 {
			sprite.Parent = matches[1]
		}
	}

	// Extract instance ExtResource and resolve path
	if re := regexp.MustCompile(`instance=ExtResource\("([^"]+)"\)`); re != nil {
		if matches := re.FindStringSubmatch(line); len(matches) > 1 {
			extResourceID := matches[1]
			if extRes, exists := c.extResources[extResourceID]; exists {
				sprite.Path = extRes.Path
			}
		}
	}

	return sprite
}

// parseSpriteProperty parses properties of the current Sprite node
func (c *TSCNConverter) parseSpriteProperty(line string) {
	if c.currentSprite == nil {
		return
	}

	if strings.HasPrefix(line, "position = Vector2(") {
		// Extract position coordinates
		position := c.extractVector2(line)
		position.Y = -position.Y
		c.currentSprite.Position = position
	} else if strings.HasPrefix(line, "gid = ") {
		// Extract gid (common in enemy nodes)
		gidValue := c.extractIntValue(line)
		c.currentSprite.Properties["gid"] = gidValue
	} else if strings.HasPrefix(line, "zoom = Vector2(") {
		// Extract zoom for Camera2D
		zoom := c.extractVector2(line)
		c.currentSprite.Properties["zoom"] = zoom
	} else if strings.Contains(line, " = ") && !strings.HasPrefix(line, "[") {
		// Generic property extraction
		parts := strings.SplitN(line, " = ", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			c.currentSprite.Properties[key] = value
		}
	}
}
