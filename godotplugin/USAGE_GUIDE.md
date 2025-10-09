# Usage Guide - SPX Godot Plugin

## Quick Start with run.sh

The `run.sh` script automates the entire export process. It copies the exporter scripts to your Godot project and runs the export.

### Basic Usage

```bash
cd cmd/tscnparser/godotplugin
./run.sh
```

This uses all default parameters:
- **Godot Project**: `/Users/tjp/projects/spx/cmd/tscnparser/export`
- **Input Scene**: `res://main.tscn`
- **Output**: `/Users/tjp/projects/spx/spx_demos/19_tilemap/assets/tilemaps/scene1.json`
- **Tile Size**: 16x16
- **Offset**: (80, 384)

### Custom Parameters

```bash
./run.sh [godot_project] [input_scene] [output_path] [tile_size] [offset_x] [offset_y]
```

**Example 1**: Use different Godot project
```bash
./run.sh /path/to/my/godot/project
```

**Example 2**: Custom scene and output
```bash
./run.sh \
  /path/to/godot/project \
  res://scenes/level1.tscn \
  /output/level1.json
```

**Example 3**: Custom tile size and offset
```bash
./run.sh \
  /path/to/godot/project \
  res://main.tscn \
  /output/main.json \
  32 \
  100 \
  200
```

### Environment Variables

**Override Godot executable:**
```bash
export GODOT_EXECUTABLE=/path/to/godot
./run.sh
```

### Help

```bash
./run.sh --help
```

## What run.sh Does

The script performs the following steps:

1. **Validates Godot project**
   - Checks if project directory exists
   - Verifies `project.godot` file is present

2. **Copies exporter scripts**
   - Creates `addons/spx_exporter/` directory in the project
   - Copies `plugin.cfg`, `plugin.gd`, `exporter.gd`, `export_script.gd`
   - Copies `replacements.json` if available

3. **Runs Godot export**
   - Finds Godot executable (checks PATH and common locations)
   - Runs Godot in headless mode
   - Executes `export_script.gd` with configured parameters

4. **Copies output**
   - Moves exported JSON to destination path
   - Cleans up temporary files

5. **Reports results**
   - Shows success/failure status
   - Displays file size and location

## Workflow Comparison

### Old Workflow (Go Parser)

```bash
cd cmd/tscnparser/test
./run.sh /path/to/input.tscn /path/to/output.json
```

**Process:**
1. Copy TSCN to test directory
2. Run Go parser (manual text parsing)
3. Copy JSON to destination

### New Workflow (Godot Plugin)

```bash
cd cmd/tscnparser/godotplugin
./run.sh /path/to/godot/project res://scene.tscn /path/to/output.json
```

**Process:**
1. Copy scripts to project
2. Run Godot exporter (native API access)
3. Copy JSON to destination

## Testing Your Export

### Step 1: Run the export

```bash
cd /Users/tjp/projects/spx/cmd/tscnparser/godotplugin
./run.sh
```

### Step 2: Check the output

```bash
# View the exported JSON
cat /Users/tjp/projects/spx/spx_demos/19_tilemap/assets/tilemaps/scene1.json

# Format and view
cat scene1.json | python3 -m json.tool | less
```

### Step 3: Compare with Go parser (optional)

```bash
# Export using Go parser
cd /Users/tjp/projects/spx/cmd/tscnparser/test
./run.sh

# Compare outputs
diff \
  /Users/tjp/projects/spx/spx_demos/19_tilemap/assets/tilemaps/scene1.json \
  main_tilemap.json
```

## Troubleshooting

### Error: Godot executable not found

**Solution 1:** Install Godot and add to PATH
```bash
# macOS with Homebrew
brew install godot

# Or download from https://godotengine.org/
```

**Solution 2:** Set GODOT_EXECUTABLE environment variable
```bash
export GODOT_EXECUTABLE=/Applications/Godot.app/Contents/MacOS/Godot
./run.sh
```

### Error: Not a valid Godot project

```
Error: Not a valid Godot project (project.godot not found)
```

**Solution:** Check that the first parameter points to a valid Godot project directory
```bash
ls /path/to/godot/project/project.godot  # Should exist
./run.sh /correct/path/to/project
```

### Error: Export failed. Output file not created

**Possible causes:**
1. Input scene doesn't exist
2. Scene path is wrong (must start with `res://`)
3. Scene has errors that prevent loading
4. Godot crashed during export

**Debug:**
```bash
# Check if scene exists
ls /Users/tjp/projects/spx/cmd/tscnparser/export/main.tscn

# Try running export manually to see detailed errors
cd /Users/tjp/projects/spx/cmd/tscnparser/export
godot --headless --quit --script addons/spx_exporter/export_script.gd -- \
  --input res://main.tscn \
  --output test.json
```

### Error: Scene path must start with res://

```
Error: Input file not found: main.tscn
```

**Solution:** Use `res://` prefix for scene paths
```bash
# Wrong
./run.sh /project main.tscn /output.json

# Correct
./run.sh /project res://main.tscn /output.json
```

## Advanced Usage

### Custom Replacements

Create a custom `replacements.json`:

```json
{
  "replacements": [
    {"old": "res://textures/", "new": "assets/textures/"},
    {"old": "res://scenes/", "new": "prefabs/"},
    {"old": ".tscn", "new": ""},
    {"old": ".png", "new": ".webp"}
  ]
}
```

Place it in the Godot project root:
```bash
cp my_replacements.json /path/to/godot/project/replacements.json
./run.sh /path/to/godot/project
```

### Batch Export Multiple Scenes

```bash
#!/bin/bash
# batch_export.sh

SCENES=(
    "res://scenes/level1.tscn"
    "res://scenes/level2.tscn"
    "res://scenes/level3.tscn"
)

for scene in "${SCENES[@]}"; do
    # Extract scene name
    scene_name=$(basename "$scene" .tscn)
    output="/output/${scene_name}.json"

    echo "Exporting $scene..."
    ./run.sh /path/to/project "$scene" "$output"
done
```

### Integration with Makefile

```makefile
# Makefile

GODOT_PROJECT = /Users/tjp/projects/spx/cmd/tscnparser/export
PLUGIN_DIR = cmd/tscnparser/godotplugin
OUTPUT_DIR = /Users/tjp/projects/spx/spx_demos/19_tilemap/assets/tilemaps

.PHONY: export-scene
export-scene:
	cd $(PLUGIN_DIR) && ./run.sh \
		$(GODOT_PROJECT) \
		res://main.tscn \
		$(OUTPUT_DIR)/scene1.json

.PHONY: export-all
export-all:
	$(MAKE) export-scene
	# Add more exports here
```

Usage:
```bash
make export-scene
make export-all
```

### CI/CD Integration

**GitHub Actions example:**

```yaml
name: Export Godot Scenes

on:
  push:
    paths:
      - 'godot_project/**/*.tscn'

jobs:
  export:
    runs-on: macos-latest

    steps:
      - uses: actions/checkout@v3

      - name: Install Godot
        run: |
          brew install godot

      - name: Export Scene
        run: |
          cd cmd/tscnparser/godotplugin
          ./run.sh \
            ${GITHUB_WORKSPACE}/godot_project \
            res://main.tscn \
            ${GITHUB_WORKSPACE}/output/scene.json

      - name: Upload Artifact
        uses: actions/upload-artifact@v3
        with:
          name: exported-scenes
          path: output/*.json
```

## Performance Notes

### First Run
- First run may be slower (Godot initializes, loads resources)
- Typical time: 3-5 seconds

### Subsequent Runs
- Faster due to Godot's resource caching
- Typical time: 1-2 seconds

### Large Scenes
- Scenes with many tiles/sprites may take longer
- Consider splitting very large scenes

## Best Practices

### 1. Version Control

Add to `.gitignore`:
```
# Godot export artifacts
addons/spx_exporter/
exported_tilemap.json
replacements.json
```

### 2. Automated Exports

Set up a git hook to auto-export on scene changes:

```bash
# .git/hooks/pre-commit
#!/bin/bash

# Check if any .tscn files changed
if git diff --cached --name-only | grep -q '\.tscn$'; then
    echo "TSCN files changed, running export..."
    cd cmd/tscnparser/godotplugin
    ./run.sh

    # Add exported file to commit
    git add /path/to/output.json
fi
```

### 3. Testing

Always test exported JSON with SPX runtime:
```bash
# After export
cd /Users/tjp/projects/spx/spx_demos/19_tilemap
spx run
```

### 4. Backup

Keep backups of working exports:
```bash
# Before re-exporting
cp scene1.json scene1.json.bak
./run.sh
# If something broke
cp scene1.json.bak scene1.json
```

## Migration from Go Parser

### 1. Verify Output Compatibility

```bash
# Export with both methods
cd cmd/tscnparser/test
./run.sh  # Go parser

cd ../godotplugin
./run.sh  # Godot plugin

# Compare
diff \
  /output/from/go/parser.json \
  /output/from/godot/plugin.json
```

### 2. Update Build Scripts

Replace Go parser calls with Godot plugin:

```bash
# Old
cd cmd/tscnparser/test && ./run.sh

# New
cd cmd/tscnparser/godotplugin && ./run.sh
```

### 3. Update Documentation

Update any references to the export process in your docs.

## Next Steps

- Read [INSTALLATION.md](INSTALLATION.md) for manual installation options
- Check [README.md](README.md) for detailed technical documentation
- Explore [example_usage.gd](example_usage.gd) for programmatic usage
