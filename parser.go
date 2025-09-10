package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
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

// TSCNConverter handles conversion from TSCN to TileMap JSON
type TSCNConverter struct {
	tileSize            TileSize
	sources             map[int]*TileSource
	extResources        map[string]*ExtResource
	subResourceTextures map[string]string // Maps SubResource ID to ExtResource ID
	sprite2Ds           []Sprite2DNode    // Collected Sprite2D nodes
	currentSprite2D     *Sprite2DNode     // Currently parsing Sprite2D node
	prefabs             []PrefabNode      // Collected Prefab nodes
	currentPrefab       *PrefabNode       // Currently parsing Prefab node
}

// NewTSCNConverter creates a new converter instance
func NewTSCNConverter() *TSCNConverter {
	return &TSCNConverter{
		tileSize:            TileSize{Width: 16, Height: 16}, // Default tile size
		sources:             make(map[int]*TileSource),
		extResources:        make(map[string]*ExtResource),
		subResourceTextures: make(map[string]string),
		sprite2Ds:           []Sprite2DNode{},
		prefabs:             []PrefabNode{},
	}
}

// ConvertTSCNToTileMap converts a TSCN file to TileMap data structure
func (c *TSCNConverter) ConvertTSCNToTileMap(filename string) (*MapData, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var currentSection string
	var currentSubResource string
	var format int
	var layers []Layer

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
				if currentSection == "sprite2d" && c.currentSprite2D != nil {
					c.sprite2Ds = append(c.sprite2Ds, *c.currentSprite2D)
					c.currentSprite2D = nil
				}
				if currentSection == "prefab" && c.currentPrefab != nil {
					c.prefabs = append(c.prefabs, *c.currentPrefab)
					c.currentPrefab = nil
				}
				currentSection = "sprite2d"
				// Initialize new Sprite2D node
				c.currentSprite2D = c.parseSprite2DNode(line)
			} else if strings.Contains(line, "instance=ExtResource") {
				// Finish current nodes if we were processing any
				if currentSection == "sprite2d" && c.currentSprite2D != nil {
					c.sprite2Ds = append(c.sprite2Ds, *c.currentSprite2D)
					c.currentSprite2D = nil
				}
				if currentSection == "prefab" && c.currentPrefab != nil {
					c.prefabs = append(c.prefabs, *c.currentPrefab)
					c.currentPrefab = nil
				}
				currentSection = "prefab"
				// Initialize new Prefab node
				c.currentPrefab = c.parsePrefabNode(line)
			} else {
				// Finish current nodes if we're leaving their sections
				if currentSection == "sprite2d" && c.currentSprite2D != nil {
					c.sprite2Ds = append(c.sprite2Ds, *c.currentSprite2D)
					c.currentSprite2D = nil
				}
				if currentSection == "prefab" && c.currentPrefab != nil {
					c.prefabs = append(c.prefabs, *c.currentPrefab)
					c.currentPrefab = nil
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
		case "sprite2d":
			c.parseSprite2DProperty(line)
		case "prefab":
			c.parsePrefabProperty(line)
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

	// Handle any remaining Sprite2D node
	if c.currentSprite2D != nil {
		c.sprite2Ds = append(c.sprite2Ds, *c.currentSprite2D)
	}

	// Handle any remaining Prefab node
	if c.currentPrefab != nil {
		c.prefabs = append(c.prefabs, *c.currentPrefab)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	// Build tileset from sources
	var tilesetSources []TileSource
	for _, source := range c.sources {
		tilesetSources = append(tilesetSources, *source)
	}

	return &MapData{
		TileMap: TileMapData{
			Format:   format,
			TileSize: c.tileSize,
			TileSet: TileSet{
				Sources: tilesetSources,
			},
			Layers: layers,
		},
		Sprite2Ds: c.sprite2Ds,
		Prefabs:   c.prefabs,
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
		// Parse collision polygon to determine tile size
		points := c.extractPolygonPoints(line)
		if len(points) >= 4 {
			// Calculate tile size from collision box
			c.tileSize = c.calculateTileSizeFromPoints(points)
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

			c.sources[sourceID] = &TileSource{
				ID:          sourceID,
				TexturePath: texturePath,
				Tiles:       []TileInfo{{AtlasCoords: Point{X: 0, Y: 0}}},
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

	// Process data in groups of 3: [position, source_id, atlas_coords]
	for i := 0; i < len(data); i += 3 {
		if i+2 >= len(data) {
			break
		}

		encodedPos := data[i]
		sourceID := data[i+1]
		atlasCoords := data[i+2]

		// Decode tile position
		tileX, tileY := c.decodeTilePosition(encodedPos)

		// Calculate world coordinates
		worldX := float64(tileX * c.tileSize.Width)
		worldY := float64(tileY * c.tileSize.Height)

		// Decode atlas coordinates (simplified - assuming single atlas coord)
		atlasX, atlasY := c.decodeAtlasCoords(atlasCoords)

		tile := TileInstance{
			TileCoords:  Point{X: tileX, Y: tileY},
			WorldCoords: WorldPoint{X: worldX, Y: worldY},
			SourceID:    sourceID,
			AtlasCoords: Point{X: atlasX, Y: atlasY},
		}

		layer.Tiles = append(layer.Tiles, tile)
	}

	return layer
}

// decodeTilePosition decodes the encoded tile position
func (c *TSCNConverter) decodeTilePosition(encoded int) (int, int) {
	x := (encoded >> 16) & 0xFFFF
	y := encoded & 0xFFFF

	// Convert from unsigned to signed if necessary
	if x >= 32768 {
		x -= 65536
	}
	if y >= 32768 {
		y -= 65536
	}

	return x, y
}

// decodeAtlasCoords decodes atlas coordinates (simplified)
func (c *TSCNConverter) decodeAtlasCoords(_ int) (int, int) {
	// For now, assuming simple case where encoded value represents atlas coords
	// In full implementation, this would need proper decoding based on Godot's format
	return 0, 0
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

func (c *TSCNConverter) extractPolygonPoints(line string) []WorldPoint {
	re := regexp.MustCompile(`PackedVector2Array\(([^)]+)\)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) < 2 {
		return []WorldPoint{}
	}

	content := matches[1]
	parts := strings.Split(content, ",")
	var points []WorldPoint

	for i := 0; i < len(parts); i += 2 {
		if i+1 >= len(parts) {
			break
		}

		x, err1 := strconv.ParseFloat(strings.TrimSpace(parts[i]), 64)
		y, err2 := strconv.ParseFloat(strings.TrimSpace(parts[i+1]), 64)

		if err1 == nil && err2 == nil {
			points = append(points, WorldPoint{X: x, Y: y})
		}
	}

	return points
}

func (c *TSCNConverter) calculateTileSizeFromPoints(points []WorldPoint) TileSize {
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

// parseSprite2DNode creates a new Sprite2D node from the node declaration line
func (c *TSCNConverter) parseSprite2DNode(line string) *Sprite2DNode {
	// Extract node name and parent from line like: [node name="Cloud1" type="Sprite2D" parent="Decorations/Clouds"]
	sprite := &Sprite2DNode{
		TexturePath: "unknown", // Default until we find texture property
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

	return sprite
}

// parseSprite2DProperty parses properties of the current Sprite2D node
func (c *TSCNConverter) parseSprite2DProperty(line string) {
	if c.currentSprite2D == nil {
		return
	}

	if strings.HasPrefix(line, "position = Vector2(") {
		// Extract position coordinates
		position := c.extractVector2(line)
		c.currentSprite2D.Position = position
	} else if strings.HasPrefix(line, "texture = ExtResource(") {
		// Extract texture ExtResource ID and resolve path
		extResourceID := c.extractExtResourceID(line)
		if extRes, exists := c.extResources[extResourceID]; exists {
			c.currentSprite2D.TexturePath = extRes.Path
		}
	} else if strings.HasPrefix(line, "z_index = ") {
		// Extract z_index
		c.currentSprite2D.ZIndex = c.extractIntValue(line)
	}
}

// extractVector2 extracts Vector2 coordinates from a line like "position = Vector2(-120, -48)"
func (c *TSCNConverter) extractVector2(line string) WorldPoint {
	re := regexp.MustCompile(`Vector2\(([^,]+),\s*([^)]+)\)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) < 3 {
		return WorldPoint{X: 0, Y: 0}
	}

	x, err1 := strconv.ParseFloat(strings.TrimSpace(matches[1]), 64)
	y, err2 := strconv.ParseFloat(strings.TrimSpace(matches[2]), 64)

	if err1 != nil || err2 != nil {
		return WorldPoint{X: 0, Y: 0}
	}

	return WorldPoint{X: x, Y: y}
}

// parsePrefabNode creates a new Prefab node from the node declaration line
func (c *TSCNConverter) parsePrefabNode(line string) *PrefabNode {
	// Extract node name, parent, and instance from line like: [node name="Brick" parent="Environment/Platforms/Platform1" instance=ExtResource("6_vt4yb")]
	prefab := &PrefabNode{
		PrefabPath: "unknown", // Default until we resolve ExtResource
		Properties: make(map[string]interface{}),
	}

	// Extract name
	if re := regexp.MustCompile(`name="([^"]+)"`); re != nil {
		if matches := re.FindStringSubmatch(line); len(matches) > 1 {
			prefab.Name = matches[1]
		}
	}

	// Extract parent
	if re := regexp.MustCompile(`parent="([^"]+)"`); re != nil {
		if matches := re.FindStringSubmatch(line); len(matches) > 1 {
			prefab.Parent = matches[1]
		}
	}

	// Extract instance ExtResource and resolve path
	if re := regexp.MustCompile(`instance=ExtResource\("([^"]+)"\)`); re != nil {
		if matches := re.FindStringSubmatch(line); len(matches) > 1 {
			extResourceID := matches[1]
			if extRes, exists := c.extResources[extResourceID]; exists {
				prefab.PrefabPath = extRes.Path
			}
		}
	}

	return prefab
}

// parsePrefabProperty parses properties of the current Prefab node
func (c *TSCNConverter) parsePrefabProperty(line string) {
	if c.currentPrefab == nil {
		return
	}

	if strings.HasPrefix(line, "position = Vector2(") {
		// Extract position coordinates
		position := c.extractVector2(line)
		c.currentPrefab.Position = position
	} else if strings.HasPrefix(line, "gid = ") {
		// Extract gid (common in enemy nodes)
		gidValue := c.extractIntValue(line)
		c.currentPrefab.Properties["gid"] = gidValue
	} else if strings.HasPrefix(line, "zoom = Vector2(") {
		// Extract zoom for Camera2D
		zoom := c.extractVector2(line)
		c.currentPrefab.Properties["zoom"] = zoom
	} else if strings.Contains(line, " = ") && !strings.HasPrefix(line, "[") {
		// Generic property extraction
		parts := strings.SplitN(line, " = ", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			c.currentPrefab.Properties[key] = value
		}
	}
}
