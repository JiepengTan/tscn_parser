package main

// Point represents a 2D coordinate
type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// WorldPoint represents a 2D coordinate in world space (pixels)
type WorldPoint struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// TileSize represents the dimensions of a tile
type TileSize struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// PhysicsData represents physics properties of a tile
type PhysicsData struct {
	CollisionPoints []WorldPoint `json:"collision_points,omitempty"`
}

// TileInfo represents information about a single tile in the tileset
type TileInfo struct {
	AtlasCoords Point       `json:"atlas_coords"`
	Physics     PhysicsData `json:"physics,omitempty"`
}

// TileSource represents a tileset source
type TileSource struct {
	ID          int        `json:"id"`
	TexturePath string     `json:"texture_path"`
	Tiles       []TileInfo `json:"tiles"`
}

// TileSet represents the complete tileset information
type TileSet struct {
	Sources []TileSource `json:"sources"`
}

// TileInstance represents a placed tile in the map
type TileInstance struct {
	TileCoords  Point      `json:"tile_coords"`
	WorldCoords WorldPoint `json:"world_coords"`
	SourceID    int        `json:"source_id"`
	AtlasCoords Point      `json:"atlas_coords"`
}

// Layer represents a tilemap layer
type Layer struct {
	ID    int            `json:"id"`
	Name  string         `json:"name"`
	Tiles []TileInstance `json:"tiles"`
}

// TileMapData represents the complete tilemap data
type TileMapData struct {
	Format   int      `json:"format"`
	TileSize TileSize `json:"tile_size"`
	TileSet  TileSet  `json:"tileset"`
	Layers   []Layer  `json:"layers"`
}

// Sprite2DNode represents a Sprite2D node in the scene
type Sprite2DNode struct {
	Name        string     `json:"name"`
	Parent      string     `json:"parent"`
	Position    WorldPoint `json:"position"`
	TexturePath string     `json:"texture_path"`
	ZIndex      int        `json:"z_index,omitempty"`
}

// PrefabNode represents an instantiated prefab node in the scene
type PrefabNode struct {
	Name       string                 `json:"name"`
	Parent     string                 `json:"parent"`
	Position   WorldPoint             `json:"position"`
	PrefabPath string                 `json:"prefab_path"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// Root structure for JSON output
type MapData struct {
	TileMap   TileMapData    `json:"tilemap"`
	Sprite2Ds []Sprite2DNode `json:"sprite2ds"`
	Prefabs   []PrefabNode   `json:"prefabs"`
}
