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