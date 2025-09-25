package tscnparser

type Vec2i struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// WorldPoint represents a 2D coordinate in world space (pixels)
type Vec2 struct {
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
	CollisionPoints []Vec2 `json:"collision_points,omitempty"`
}

// TileInfo represents information about a single tile in the tileset
type TileInfo struct {
	AtlasCoords Vec2i       `json:"atlas_coords"`
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
	TileCoords  Vec2i `json:"tile_coords"`
	SourceID    int   `json:"source_id"`
	AtlasCoords Vec2i `json:"atlas_coords"`
}

// Layer represents a tilemap layer
type Layer struct {
	ID       int            `json:"id"`
	Name     string         `json:"name"`
	Tiles    []TileInstance `-`
	ZIndex   int            `json:"z_index"`
	TileData []int          `json:"tile_data"`
}

// TileMapData represents the complete tilemap data
type TileMapData struct {
	Format   int      `json:"format"`
	TileSize TileSize `json:"tile_size"`
	TileSet  TileSet  `json:"tileset"`
	Layers   []Layer  `json:"layers"`
}

// DecoratorNode represents a Sprite2D node in the scene
type DecoratorNode struct {
	Name     string `json:"name"`
	Parent   string `json:"parent"`
	Position Vec2   `json:"position"`
	Path     string `json:"path"`
	ZIndex   int    `json:"z_index,omitempty"`
}

// SpriteNode represents an instantiated prefab node in the scene
type SpriteNode struct {
	Name       string         `json:"name"`
	Parent     string         `json:"parent"`
	Position   Vec2           `json:"position"`
	Path       string         `json:"path"`
	Properties map[string]any `json:"properties,omitempty"`
}

// Root structure for JSON output
type MapData struct {
	TileMap    TileMapData     `json:"tilemap"`
	Decorators []DecoratorNode `json:"decorators"`
	Sprites    []SpriteNode    `json:"sprites"`
}
