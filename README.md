# TSCN Parser

A Go library and command-line tool that converts Godot TSCN (Text Scene) files to tilemap JSON format.

## Library Usage

The package can be imported and used as a library:

```go
import tscnparser "github.com/JiepengTan/tscn_parser"

// Parse a TSCN file
mapData, err := tscnparser.Parse("path/to/scene.tscn")
if err != nil {
    log.Fatal(err)
}
```

## Command Line Usage

### Quick Start: Using the Standalone Script

The easiest way to convert TSCN files is using the `convert2spx.sh` script:

```bash
# Basic usage with default parameters
./convert2spx.sh

# Custom input and output paths
./convert2spx.sh <input_tscn_file> <output_json_file>

# Show help
./convert2spx.sh --help
```

**Features:**
- Automatically downloads and uses the latest version of tscnparser
- No manual setup required (only Go needs to be installed)
- Includes built-in replacement rules for SPX format
- Processes with optimized settings (tilesize: 16, offset: 80x384)

### Using the Command Line Tool Directly

You can also install and use the command-line tool directly:

```bash
# Install the latest version
go install github.com/JiepengTan/tscn_parser/cmd/tscnparser@latest

# Or run directly without installation
go run github.com/JiepengTan/tscn_parser/cmd/tscnparser@latest -input <tscn_file> [options]
```

#### Basic Command Usage

```bash
tscnparser -input <tscn_file> [-output <json_file>] [-replacements <json_file>] [-generateGo] [-oldStr <old_string>] [-newStr <new_string>]
```

### Parameters

- `-input`: Required. Path to the input TSCN file
- `-output`: Optional. Path to the output JSON file. If not specified, generates `<input_filename>_tilemap.json`
- `-replacements`: Optional. JSON file containing multiple replacement rules
- `-tilesize`: Optional. Tile size in pixels (default: 16)
- `-offsetx`: Optional. X offset for tile coordinates (default: 0)
- `-offsety`: Optional. Y offset for tile coordinates (default: 0)
- `--prefabs`: Optional. Directory containing prefab .tscn files for scene instantiation
- `-generateGo`: Optional. Generate Go code file (.go.txt) for direct integration
- `-oldStr`: Optional. Single string to replace in JSON output (applied after replacements file)
- `-newStr`: Optional. Replacement string for oldStr

### Examples

```bash
# Basic usage
tscnparser -input main.tscn

# With custom output file
tscnparser -input main.tscn -output my_tilemap.json

# With multiple replacements from JSON file
tscnparser -input main.tscn -replacements replacements.json

# With tile size and offset parameters
tscnparser -input main.tscn -tilesize 16 -offsetx 80 -offsety 384

# With prefabs directory
tscnparser -input main.tscn --prefabs "./scenes"

# With both JSON replacements and single replacement
tscnparser -input main.tscn -replacements replacements.json -oldStr "additional_old" -newStr "additional_new"

# Generate Go code for direct integration
tscnparser -input main.tscn -generateGo

# Generate both JSON and Go code with replacements
tscnparser -input main.tscn -replacements replacements.json -generateGo
```

### Replacement Configuration File

Create a JSON file (e.g., `replacements.json`) with multiple replacement rules:

```json
{
  "replacements": [
    {"old": "res://assets/sprites/", "new": "textures/"},
    {"old": "res://assets/scenes/", "new": "prefabs/"},
    {"old": ".png", "new": ".webp"},
    {"old": ".tscn", "new": ".scene"}
  ]
}
```

### Go Code Generation

When using the `-generateGo` flag, the tool generates a `.go.txt` file containing Go code that defines the tilemap data as a variable. The file includes all necessary type definitions, making it completely self-contained with no external dependencies. This allows for direct integration into other Go projects without needing to import or define the data structures separately.

Example generated Go code structure:
```go
package main

// Auto-generated from main.tscn
// This file contains the parsed tilemap data as Go code
// All type definitions are included, so no external dependencies are needed

// Type definitions for tilemap data structures (prefixed with tscn to avoid naming conflicts)
type tscnPoint struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type tscnWorldPoint struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// ... (all other type definitions with tscn prefix included)

var MapData = &tscnMapData{
	TileMap: tscnTileMapData{
		Format: 2,
		TileSize: tscnTileSize{Width: 16, Height: 16},
		TileSet: tscnTileSet{
			Sources: []tscnTileSource{
				{
					ID: 0,
					TexturePath: "textures/GroundBlock.png",
					Tiles: []tscnTileInfo{
						{
							AtlasCoords: tscnPoint{X: 0, Y: 0},
							Physics: tscnPhysicsData{},
						},
					},
				},
			},
		},
		// ... (complete data structure)
	},
	Sprite2Ds: []tscnSprite2DNode{},
	Prefabs: []tscnPrefabNode{},
}
```

## Output Format

The tool generates a JSON file with the following structure:

```json
{
  "tilemap": {
    "format": 2,
    "tile_size": {"width": 16, "height": 16},
    "tileset": {
      "sources": [
        {
          "id": 0,
          "texture_path": "res://assets/sprites/GroundBlock.png",
          "tiles": [
            {
              "atlas_coords": {"x": 0, "y": 0},
              "physics": {}
            }
          ]
        }
      ]
    },
    "layers": [
      {
        "id": 0,
        "name": "layer_0",
        "tiles": [
          {
            "tile_coords": {"x": 5, "y": 13},
            "world_coords": {"x": 80, "y": 208},
            "source_id": 0,
            "atlas_coords": {"x": 0, "y": 0}
          }
        ]
      }
    ]
  },
  "sprite2ds": [
    {
      "name": "Cloud1",
      "parent": "Decorations/Clouds",
      "position": {"x": -120, "y": -48},
      "texture_path": "res://assets/sprites/Cloud1.png",
      "z_index": 0
    }
  ],
  "prefabs": [
    {
      "name": "Brick",
      "parent": "Environment/Platforms/Platform1",
      "position": {"x": 200, "y": 100},
      "prefab_path": "res://assets/scenes/brick.tscn",
      "properties": {"gid": 123}
    }
  ]
}
```

The output includes:
- **tilemap**: Core tilemap data with tile layers and tilesets
- **sprite2ds**: Individual Sprite2D nodes found in the scene
- **prefabs**: Instantiated scene prefabs with their properties

## Installation

### Option 1: Use the Standalone Script (Recommended)

Download the `convert2spx.sh` script from the repository. No installation needed - the script will automatically download and use the latest version of the tool.

```bash
# Download the script
curl -O https://raw.githubusercontent.com/JiepengTan/tscn_parser/main/convert2spx.sh
chmod +x convert2spx.sh

# Run it
./convert2spx.sh <input_tscn_file> <output_json_file>
```

### Option 2: Install the Command-Line Tool

```bash
go install github.com/JiepengTan/tscn_parser/cmd/tscnparser@latest
```

### Option 3: Run Directly Without Installation

```bash
go run github.com/JiepengTan/tscn_parser/cmd/tscnparser@latest -input <tscn_file>
```

## Requirements

- Go 1.21 or higher (for running the tool or script)
- For library usage: Import as a Go module