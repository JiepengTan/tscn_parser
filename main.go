package tscnparser

import (
	"errors"
)

func SetTileSize(size int) {
	tilemapTileSize = TileSize{size, size}
}
func SetOffset(x, y int) {
	tilemapOffset = Vec2{float64(x), float64(y)}
}
func SetPrefabsDir(dir string) {
	prefabsDirectory = dir
}
func Parse(inputFile string) (*MapData, error) {

	if inputFile == "" {
		return nil, errors.New("input file is empty")
	}

	// Parse TSCN file
	converter := newTSCNConverter()
	return converter.ConvertTSCNToTileMap(inputFile)
}
