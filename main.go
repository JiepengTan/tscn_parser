package tscnparser

import "errors"

func Parse(inputFile string) (*MapData, error) {

	if inputFile == "" {
		return nil, errors.New("input file is empty")
	}

	// Parse TSCN file
	converter := newTSCNConverter()
	return converter.ConvertTSCNToTileMap(inputFile)
}
