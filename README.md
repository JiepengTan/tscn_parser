# TSCN Parser

A command-line tool that converts Godot TSCN (Text Scene) files to tilemap JSON format.

## Usage

```bash
go run . -input <tscn_file> [-output <json_file>]
```

### Parameters

- `-input`: Required. Path to the input TSCN file
- `-output`: Optional. Path to the output JSON file. If not specified, generates `<input_filename>_tilemap.json`

### Examples

```bash
# Basic usage
go run . -input test/main.tscn

# With custom output file
go run . -input test/main.tscn -output my_tilemap.json
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
  "sprite2ds": [],
  "prefabs": []
}
```

## Testing

Run the test script to verify the tool works correctly:

```bash
./test.sh
```

This will process the test TSCN file and generate the corresponding JSON output.