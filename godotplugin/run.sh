#!/bin/bash

# SPX Godot Plugin Export Script
# This script copies the exporter scripts to a Godot project and runs the export

# Default parameters
DEFAULT_GODOT_PROJECT="/Users/tjp/projects/spx/cmd/tscnparser/export"
DEFAULT_INPUT_TSCN="res://main.tscn"
DEFAULT_CP_DESTINATION="/Users/tjp/projects/spx/spx_demos/19_tilemap/assets/tilemaps/scene1.json"
DEFAULT_TILE_SIZE=16
DEFAULT_OFFSET_X=80
DEFAULT_OFFSET_Y=384

# Use provided parameters or defaults
GODOT_PROJECT_PATH="${1:-$DEFAULT_GODOT_PROJECT}"
INPUT_TSCN_PATH="${2:-$DEFAULT_INPUT_TSCN}"
CP_DESTINATION_PATH="${3:-$DEFAULT_CP_DESTINATION}"
TILE_SIZE="${4:-$DEFAULT_TILE_SIZE}"
OFFSET_X="${5:-$DEFAULT_OFFSET_X}"
OFFSET_Y="${6:-$DEFAULT_OFFSET_Y}"

# Show usage if requested
if [ "$1" = "-h" ] || [ "$1" = "--help" ]; then
    echo "Usage: $0 [godot_project_path] [input_tscn_path] [cp_destination_path] [tile_size] [offset_x] [offset_y]"
    echo ""
    echo "Parameters:"
    echo "  godot_project_path:    Path to Godot project directory (default: $DEFAULT_GODOT_PROJECT)"
    echo "  input_tscn_path:       Scene path relative to project (default: $DEFAULT_INPUT_TSCN)"
    echo "  cp_destination_path:   Where to copy the output JSON (default: $DEFAULT_CP_DESTINATION)"
    echo "  tile_size:             Tile size in pixels (default: $DEFAULT_TILE_SIZE)"
    echo "  offset_x:              X offset in pixels (default: $DEFAULT_OFFSET_X)"
    echo "  offset_y:              Y offset in pixels (default: $DEFAULT_OFFSET_Y)"
    echo ""
    echo "Example:"
    echo "  $0 /path/to/godot/project res://main.tscn /output/scene1.json 16 80 384"
    exit 0
fi

# Get the directory of this script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Validate Godot project path
if [ ! -d "$GODOT_PROJECT_PATH" ]; then
    echo "Error: Godot project directory not found: $GODOT_PROJECT_PATH"
    exit 1
fi

if [ ! -f "$GODOT_PROJECT_PATH/project.godot" ]; then
    echo "Error: Not a valid Godot project (project.godot not found): $GODOT_PROJECT_PATH"
    exit 1
fi

echo "=== SPX Godot Exporter ==="
echo "Project:     $GODOT_PROJECT_PATH"
echo "Input:       $INPUT_TSCN_PATH"
echo "Output:      $CP_DESTINATION_PATH"
echo "Tile Size:   ${TILE_SIZE}x${TILE_SIZE}"
echo "Offset:      ($OFFSET_X, $OFFSET_Y)"
echo ""

# Create addons directory in Godot project if it doesn't exist
ADDON_DIR="$GODOT_PROJECT_PATH/addons/spx_exporter"
mkdir -p "$ADDON_DIR"

echo "Step 1: Copying exporter scripts to Godot project..."

# Copy necessary files to the addon directory
cp "$SCRIPT_DIR/plugin.cfg" "$ADDON_DIR/"
cp "$SCRIPT_DIR/plugin.gd" "$ADDON_DIR/"
cp "$SCRIPT_DIR/exporter.gd" "$ADDON_DIR/"
cp "$SCRIPT_DIR/export_script.gd" "$ADDON_DIR/"
cp "$SCRIPT_DIR/cli_export.gd" "$ADDON_DIR/"

echo "  ✓ Scripts copied to $ADDON_DIR"

# Copy replacements.json from test directory if it exists
REPLACEMENTS_SRC="$SCRIPT_DIR/../test/replacements.json"
REPLACEMENTS_DEST="$GODOT_PROJECT_PATH/replacements.json"

if [ -f "$REPLACEMENTS_SRC" ]; then
    cp "$REPLACEMENTS_SRC" "$REPLACEMENTS_DEST"
    echo "  ✓ Replacements config copied"
    USE_REPLACEMENTS="--replacements replacements.json"
else
    echo "  ℹ No replacements.json found, using defaults"
    USE_REPLACEMENTS=""
fi

echo ""
echo "Step 2: Running Godot export..."

# Find Godot executable
GODOT_BIN=""

# Check for gdspx (SPX's custom Godot build) first
if [ -f "$HOME/go/bin/gdspx2.1.18" ]; then
    GODOT_BIN="$HOME/go/bin/gdspx2.1.18"
elif command -v godot &> /dev/null; then
    GODOT_BIN="godot"
elif command -v /Applications/Godot.app/Contents/MacOS/Godot &> /dev/null; then
    GODOT_BIN="/Applications/Godot.app/Contents/MacOS/Godot"
elif [ -f "/Applications/Godot.app/Contents/MacOS/Godot" ]; then
    GODOT_BIN="/Applications/Godot.app/Contents/MacOS/Godot"
else
    echo "Error: Godot executable not found. Please ensure Godot is installed and in PATH."
    echo "Or set GODOT_EXECUTABLE environment variable to point to Godot executable."
    exit 1
fi

# Allow override via environment variable
if [ ! -z "$GODOT_EXECUTABLE" ]; then
    GODOT_BIN="$GODOT_EXECUTABLE"
fi

echo "  Using Godot: $GODOT_BIN"

# Change to Godot project directory (required for res:// paths to work)
cd "$GODOT_PROJECT_PATH"

# Temporary output file in project directory
TEMP_OUTPUT="exported_tilemap.json"

# Run Godot in headless mode to export the scene
# Note: Using environment variables to pass parameters since -s doesn't support -- arguments
export SPX_INPUT="$INPUT_TSCN_PATH"
export SPX_OUTPUT="$TEMP_OUTPUT"
export SPX_TILE_SIZE="$TILE_SIZE"
export SPX_OFFSET_X="$OFFSET_X"
export SPX_OFFSET_Y="$OFFSET_Y"
if [ ! -z "$USE_REPLACEMENTS" ]; then
    export SPX_REPLACEMENTS="replacements.json"
fi

"$GODOT_BIN" --headless -s addons/spx_exporter/cli_export.gd

# Check if export was successful
if [ ! -f "$TEMP_OUTPUT" ]; then
    echo ""
    echo "Error: Export failed. Output file not created."
    echo "Check the output above for error messages."
    exit 1
fi

echo ""
echo "Step 3: Copying output to destination..."

# Copy the exported file to the destination
cp "$TEMP_OUTPUT" "$CP_DESTINATION_PATH"

if [ $? -eq 0 ]; then
    echo "  ✓ Output copied to $CP_DESTINATION_PATH"
else
    echo "  Error: Failed to copy output to destination"
    exit 1
fi

# Clean up temporary file
rm -f "$TEMP_OUTPUT"

echo ""
echo "=== Export Completed Successfully ==="
echo "Output: $CP_DESTINATION_PATH"
echo ""

# Optional: Show file size
if command -v wc &> /dev/null; then
    FILE_SIZE=$(wc -c < "$CP_DESTINATION_PATH" | tr -d ' ')
    echo "File size: $FILE_SIZE bytes"
fi
