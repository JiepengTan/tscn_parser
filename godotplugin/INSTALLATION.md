# Installation Guide - SPX Tilemap Exporter

## Quick Start

### Method 1: Install as Godot Addon (Recommended)

1. **Copy the plugin to your Godot project:**

   ```bash
   # Navigate to your Godot project directory
   cd /path/to/your/godot/project

   # Create addons directory if it doesn't exist
   mkdir -p addons/spx_exporter

   # Copy all plugin files
   cp /path/to/spx/cmd/tscnparser/godotplugin/* addons/spx_exporter/
   ```

2. **Enable the plugin in Godot:**

   - Open your project in Godot Editor
   - Go to `Project → Project Settings → Plugins`
   - Find "SPX Tilemap Exporter" in the list
   - Check the "Enable" checkbox

3. **Verify installation:**

   - You should see a new menu item: `Tools → Export Scene to SPX JSON`
   - If you don't see it, try restarting the Godot Editor

### Method 2: Use as Standalone Scripts (No Installation)

If you don't want to install as a plugin, you can use the scripts directly:

1. **Copy scripts to your project:**

   ```bash
   # Copy to your project root or any folder
   cp /path/to/spx/cmd/tscnparser/godotplugin/exporter.gd /your/project/
   cp /path/to/spx/cmd/tscnparser/godotplugin/export_script.gd /your/project/
   ```

2. **Run from command line:**

   ```bash
   godot --headless --script export_script.gd -- \
     --input res://main.tscn \
     --output output.json
   ```

## Usage

### A. Using the Editor Plugin

1. **Open a scene** in Godot Editor (must be saved first)

2. **Export the scene:**
   - Click `Tools → Export Scene to SPX JSON`
   - Choose output location in the file dialog
   - Click "Save"

3. **Find your exported JSON:**
   - By default: `<scene_name>_tilemap.json` in the same directory as your scene

### B. Using Command Line

```bash
# Basic usage
godot --headless --script addons/spx_exporter/export_script.gd -- \
  --input res://scenes/main.tscn

# With custom output path
godot --headless --script addons/spx_exporter/export_script.gd -- \
  --input res://scenes/level1.tscn \
  --output /output/level1.json

# With custom tile size and offset
godot --headless --script addons/spx_exporter/export_script.gd -- \
  --input res://main.tscn \
  --tile-size 32 \
  --offset-x 100 \
  --offset-y 50

# Show help
godot --headless --script addons/spx_exporter/export_script.gd -- --help
```

### C. Using Programmatically (GDScript)

```gdscript
# In any GDScript file
var exporter = load("res://addons/spx_exporter/exporter.gd").new()

# Configure
exporter.tile_size = Vector2i(16, 16)
exporter.offset = Vector2.ZERO
exporter.path_replacements = [
    {"old": "res://textures/", "new": "textures/"}
]

# Export current scene
var scene = get_tree().get_edited_scene_root()
var result = exporter.export_scene(scene)

# Save to file
var file = FileAccess.open("output.json", FileAccess.WRITE)
file.store_string(JSON.stringify(result, "\t"))
file.close()
```

## Configuration

### Path Replacements

Configure how resource paths are transformed in the exported JSON.

**Create `replacements.json` in your project root:**

```json
{
  "replacements": [
    {"old": "res://textures/", "new": "textures/"},
    {"old": "res://scenes/", "new": ""},
    {"old": ".tscn", "new": ""},
    {"old": ".png", "new": ".webp"}
  ]
}
```

**Use with command line:**

```bash
godot --headless --script addons/spx_exporter/export_script.gd -- \
  --input res://main.tscn \
  --replacements replacements.json
```

**Use programmatically:**

```gdscript
var exporter = load("res://addons/spx_exporter/exporter.gd").new()
exporter.path_replacements = [
    {"old": "res://textures/", "new": "textures/"},
    {"old": ".tscn", "new": ""}
]
```

## Integration with Existing Workflow

### For Go-based tscnparser Users

If you're currently using the Go-based parser (`cmd/tscnparser/test`):

1. **Install this plugin** as described above

2. **Compare outputs:**

   ```bash
   # Export using Go parser
   cd cmd/tscnparser/test
   go run . -input main.tscn -output go_output.json

   # Export using Godot plugin
   godot --headless --script addons/spx_exporter/export_script.gd -- \
     --input res://main.tscn \
     --output godot_output.json

   # Compare results
   diff go_output.json godot_output.json
   ```

3. **Gradually migrate:**
   - Start using the plugin for new scenes
   - Keep Go parser for existing scenes during transition
   - Validate that SPX runtime works with plugin output

### CI/CD Integration

**GitHub Actions Example:**

```yaml
name: Export Godot Scenes

on: [push]

jobs:
  export:
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v2

      - name: Install Godot
        run: |
          wget https://downloads.tuxfamily.org/godotengine/4.2/Godot_v4.2-stable_linux.x86_64.zip
          unzip Godot_v4.2-stable_linux.x86_64.zip
          chmod +x Godot_v4.2-stable_linux.x86_64

      - name: Export Scene
        run: |
          ./Godot_v4.2-stable_linux.x86_64 --headless \
            --script addons/spx_exporter/export_script.gd -- \
            --input res://scenes/main.tscn \
            --output output/main.json

      - name: Upload Artifacts
        uses: actions/upload-artifact@v2
        with:
          name: exported-scenes
          path: output/*.json
```

## Troubleshooting

### Plugin Not Showing in Editor

**Symptoms:** Can't find "SPX Tilemap Exporter" in Plugins menu

**Solutions:**
- Check that files are in `addons/spx_exporter/` directory
- Verify `plugin.cfg` exists and has correct format
- Restart Godot Editor
- Check console for error messages

### Export Produces Empty JSON

**Symptoms:** Output file is empty or has no data

**Solutions:**
- Verify scene contains TileMap or Sprite2D nodes
- Check that scene has been saved (not untitled)
- Look at Godot console for error messages
- Ensure scene can be loaded without errors

### Command Line Export Fails

**Symptoms:** `godot --headless` command doesn't work

**Solutions:**
- Make sure Godot is in your PATH or use full path to godot executable
- Verify script path is correct (relative to project root)
- Check that `--` separator is present before script arguments
- Use `--help` flag to see available options

### Paths Not Replaced Correctly

**Symptoms:** Output JSON has wrong paths (e.g., `res://` prefix still present)

**Solutions:**
- Check `path_replacements` configuration
- Verify replacements.json format if using file
- Make sure replacements are applied in correct order
- Test with simple replacement first

### Missing Collision Data

**Symptoms:** Prefabs exported without `collider_type` or `collider_params`

**Solutions:**
- Verify prefab scene has CollisionShape2D node
- Check that CollisionShape2D has a valid Shape2D assigned
- Ensure prefab scene can be loaded (path is correct)
- Check console for warnings about missing prefabs

## Updating the Plugin

To update to a newer version:

1. **Disable the plugin:**
   - `Project → Project Settings → Plugins`
   - Uncheck "SPX Tilemap Exporter"

2. **Replace files:**
   ```bash
   rm -rf addons/spx_exporter/*
   cp /path/to/new/version/* addons/spx_exporter/
   ```

3. **Re-enable the plugin:**
   - Check "SPX Tilemap Exporter" again in Plugins menu

## Uninstallation

1. **Disable the plugin:**
   - `Project → Project Settings → Plugins`
   - Uncheck "SPX Tilemap Exporter"

2. **Remove files:**
   ```bash
   rm -rf addons/spx_exporter
   ```

3. **Restart Godot Editor**

## System Requirements

- **Godot Version:** 4.0 or later (tested with 4.2)
- **Platform:** Windows, macOS, Linux
- **For command-line usage:** Godot executable in PATH or specified by full path

## Getting Help

If you encounter issues:

1. Check the console output for error messages
2. Review this troubleshooting guide
3. Consult the main README.md for detailed documentation
4. Check example_usage.gd for code samples
5. Report issues to the SPX project repository

## Next Steps

- Read [README.md](README.md) for detailed documentation
- Check [example_usage.gd](example_usage.gd) for code examples
- Explore the plugin settings and configuration options
- Integrate into your build pipeline
