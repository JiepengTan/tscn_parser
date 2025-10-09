# Godot Plugin Solution for TSCN Parsing

## Overview

This directory contains a Godot addon-based solution to replace manual TSCN parsing with native Godot engine capabilities. Instead of parsing TSCN files with regular expressions and string manipulation, this approach uses Godot's built-in APIs to access scene data directly.

## Problem with Current Implementation

The current Go-based parser (`cmd/tscnparser`) has several limitations:

1. **High Maintenance Cost** - Manual parsing of TSCN format requires updates when Godot changes
2. **Error-Prone** - Complex regex patterns and string processing are fragile
3. **Limited Functionality** - Cannot leverage Godot's runtime features
4. **Code Duplication** - Parsing prefab files requires duplicating the same parsing logic

## Solution: Godot EditorPlugin Addon

### Core Concept

Use Godot's EditorPlugin API to create an addon that runs inside the Godot editor, directly accessing the scene tree and resources without any parsing.

### Architecture

```
cmd/tscnparser/godotplugin/
├── README.md                    # This file
├── plugin.cfg                   # Addon configuration
├── plugin.gd                    # EditorPlugin entry point
├── exporter.gd                  # Core export logic
├── export_script.gd            # Command-line export script
└── example_usage.gd            # Example usage script
```

## Implementation Details

### 1. Plugin Structure

**plugin.cfg** - Defines the addon metadata for Godot to recognize it

**plugin.gd** - Main EditorPlugin class that:
- Adds menu items to the editor
- Handles user interaction
- Triggers the export process
- Saves output JSON files

**exporter.gd** - Core exporting logic that:
- Traverses the scene tree using Godot's API
- Extracts TileMap data (layers, tiles, tileset)
- Collects Sprite2D nodes (decorators)
- Processes instantiated scenes (prefabs)
- Extracts collision shapes and physics data
- Converts to SPX JSON format

### 2. Key Features

#### Direct Scene Access
```gdscript
# No parsing needed - direct API access
var tilemap = get_node("TileMap")
var used_cells = tilemap.get_used_cells(layer_id)
var source_id = tilemap.get_cell_source_id(layer_id, cell_pos)
```

#### Automatic Resource Resolution
```gdscript
# Godot handles resource loading automatically
var texture_path = sprite.texture.resource_path
var scene_data = load(prefab_path).instantiate()
```

#### Physics Data Extraction
```gdscript
# Access collision shapes directly
var collision = find_collision_shape(node)
var shape_type = _get_shape_type(collision.shape)
var params = _get_shape_params(collision.shape)
```

### 3. Usage Methods

#### Method A: Editor Integration (Recommended for Artists)

1. Install the addon in Godot project:
   ```
   project/
   └── addons/
       └── spx_exporter/
           ├── plugin.cfg
           ├── plugin.gd
           └── exporter.gd
   ```

2. Enable addon in Godot: `Project → Project Settings → Plugins → SPX Exporter`

3. Export scene: `Tools → Export to SPX JSON` (or from custom dock)

#### Method B: Command-Line (Recommended for CI/CD)

```bash
# Export a scene from command line
godot --headless --script export_script.gd -- \
  --input res://scenes/main.tscn \
  --output /path/to/output.json \
  --tile-size 16 \
  --offset-x 0 \
  --offset-y 0
```

#### Method C: Batch Processing

```gdscript
# Batch export multiple scenes
var scenes = ["res://level1.tscn", "res://level2.tscn"]
for scene_path in scenes:
    exporter.export_scene_file(scene_path, scene_path.get_basename() + ".json")
```

## Data Export Format

The addon exports the same JSON format as the current Go parser:

```json
{
  "tilemap": {
    "format": 2,
    "tile_size": {"width": 16, "height": 16},
    "tileset": {
      "sources": [
        {
          "id": 0,
          "texture_path": "textures/ground.png",
          "tiles": [
            {
              "atlas_coords": {"x": 0, "y": 0},
              "physics": {
                "collision_points": [...]
              }
            }
          ]
        }
      ]
    },
    "layers": [
      {
        "id": 0,
        "name": "layer_0",
        "z_index": 0,
        "tile_data": [0, 5, 13, 0, 0]  // [source_id, tile_x, tile_y, atlas_x, atlas_y]
      }
    ],
    "world_tile_size": {"width": 64, "height": 48}
  },
  "decorators": [
    {
      "name": "Tree",
      "path": "textures/tree.png",
      "position": {"x": 100, "y": -50},
      "scale": {"x": 1, "y": 1},
      "rotation": 0,
      "z_index": 0,
      "pivot": {"x": 0, "y": 0},
      "collider_type": "circle",
      "collider_pivot": {"x": 0, "y": 0},
      "collider_params": [10]
    }
  ],
  "sprites": [...],
  "prefabs": [
    {
      "name": "Enemy",
      "path": "scenes/enemy.tscn",
      "texture": "textures/enemy.png",
      "position": {"x": 200, "y": -100},
      "pivot": {"x": 8, "y": 8},
      "collider_type": "rect",
      "collider_params": [16, 16],
      "properties": {"gid": 123}
    }
  ]
}
```

## Configuration Options

The exporter supports the following configuration:

```gdscript
var exporter = Exporter.new()
exporter.tile_size = Vector2i(16, 16)
exporter.offset = Vector2(0, 0)
exporter.path_replacements = [
    {"old": "res://textures/", "new": "textures/"},
    {"old": "res://scenes/", "new": ""},
    {"old": ".tscn", "new": ""}
]
```

## Advantages Over Manual Parsing

| Aspect | Manual Parsing (Go) | Godot Addon (GDScript) |
|--------|-------------------|----------------------|
| **Parsing Cost** | High - regex, string manipulation | Zero - native API access |
| **Maintenance** | Manual updates for format changes | Automatic compatibility |
| **Data Access** | Limited to text format | Full scene tree access |
| **Complexity** | ~1200 lines of parsing code | ~400 lines of clean logic |
| **Error Handling** | Fragile regex patterns | Type-safe API calls |
| **Extensibility** | Requires parser updates | Easy to add new features |
| **Integration** | External tool | Native editor workflow |
| **Testing** | Manual file comparison | Visual preview in editor |

## Implementation Phases

### Phase 1: Basic Addon Setup ✓
- Create plugin structure
- Implement EditorPlugin entry point
- Add basic menu integration

### Phase 2: Core Export Logic ✓
- Implement TileMap export
- Implement Sprite2D export
- Implement Prefab export with collision data

### Phase 3: Advanced Features
- Add custom editor dock UI
- Support batch export
- Add export validation and preview

### Phase 4: Integration & Migration
- Test compatibility with existing Go parser output
- Add command-line support for CI/CD
- Document migration guide
- Gradual replacement of Go parser

## Migration Strategy

### Step 1: Parallel Implementation
- Keep Go parser as fallback
- Develop and test Godot addon
- Compare outputs for consistency

### Step 2: Validation
- Export same scenes with both methods
- Verify JSON output matches
- Test with SPX runtime

### Step 3: Gradual Adoption
- Start using addon for new scenes
- Convert existing workflow gradually
- Maintain Go parser for edge cases

### Step 4: Full Migration
- Make addon the primary export method
- Archive Go parser as legacy code
- Update documentation and tools

## Technical Notes

### Godot Version Compatibility

The addon is designed for **Godot 4.x**. Key API differences from Godot 3.x:

- `@tool` annotation instead of `tool` keyword
- `TileMap` API changes (layers, cell access)
- `FileAccess` instead of `File`
- Vector2i for tile coordinates

### Coordinate System Conversion

SPX uses Y-down coordinates, while Godot uses Y-up:

```gdscript
# Y-axis inversion
spx_y = -godot_y

# Applied to:
# - Sprite positions
# - Prefab positions
# - Collider pivots
```

### Rotation Adjustment

Based on the converter.go file, rotation needs adjustment:

```gdscript
# Add 90 degrees to match SPX coordinate system
decorator.rotation = sprite.rotation + deg_to_rad(90)
```

### Path Replacements

Convert Godot resource paths to SPX paths:

```gdscript
"res://textures/ground.png" → "textures/ground.png"
"res://scenes/enemy.tscn" → "enemy"
```

## Performance Considerations

1. **Large Scenes** - For scenes with many tiles/sprites, consider:
   - Progress reporting for long exports
   - Chunked processing for very large tilemaps
   - Async export for non-blocking UI

2. **Prefab Loading** - Instantiating prefabs can be slow:
   - Implement caching for repeated prefabs
   - Only load prefab data once per unique scene
   - Consider preloading common prefabs

3. **Memory Usage** - For batch exports:
   - Free instantiated nodes after extraction
   - Clear caches between exports
   - Monitor memory for leak prevention

## Troubleshooting

### Common Issues

**Issue**: Addon not showing in editor
- Solution: Check plugin.cfg syntax, ensure addon is in `addons/` folder

**Issue**: Export produces empty JSON
- Solution: Verify scene has TileMap/Sprite2D nodes, check console for errors

**Issue**: Paths not resolved correctly
- Solution: Configure path_replacements to match your project structure

**Issue**: Collision data missing
- Solution: Ensure prefab scenes have CollisionShape2D nodes with valid shapes

## Future Enhancements

1. **Export Options Dialog** - UI for configuring export settings
2. **Batch Export Tool** - Export multiple scenes at once
3. **Preview Mode** - Visual preview of exported data
4. **Validation Tool** - Check exported data against SPX requirements
5. **Auto-Export** - Automatic export on scene save
6. **Hot Reload** - Signal SPX runtime to reload on export

## Contributing

When extending the addon:

1. Follow GDScript style guide
2. Add type hints for better IDE support
3. Document new configuration options
4. Update this README with new features
5. Test with various scene structures

## License

This addon follows the same license as the SPX project.

## References

- [Godot EditorPlugin Documentation](https://docs.godotengine.org/en/stable/classes/class_editorplugin.html)
- [TileMap API Reference](https://docs.godotengine.org/en/stable/classes/class_tilemap.html)
- [GDScript Style Guide](https://docs.godotengine.org/en/stable/tutorials/scripting/gdscript/gdscript_styleguide.html)
