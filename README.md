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

### Using the test command

```bash
cd test/
go run . -input <tscn_file> [-output <json_file>] [-replacements <json_file>] [-oldStr <old_string>] [-newStr <new_string>]
```

### Parameters

- `-input`: Required. Path to the input TSCN file
- `-output`: Optional. Path to the output JSON file. If not specified, generates `<input_filename>_tilemap.json`
- `-replacements`: Optional. JSON file containing multiple replacement rules
- `-oldStr`: Optional. Single string to replace in JSON output (applied after replacements file)
- `-newStr`: Optional. Replacement string for oldStr

### Examples

```bash
# Basic usage
cd test/
go run . -input main.tscn

# With custom output file
cd test/
go run . -input main.tscn -output my_tilemap.json

# With multiple replacements from JSON file
cd test/
go run . -input main.tscn -replacements replacements.json

# With both JSON replacements and single replacement
cd test/
go run . -input main.tscn -replacements replacements.json -oldStr "additional_old" -newStr "additional_new"
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

## Testing

Run the test script to verify the tool works correctly:

```bash
cd test/
./test.sh
```

This will process the test TSCN file and generate the corresponding JSON output.