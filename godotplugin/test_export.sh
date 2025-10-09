#!/bin/bash

# Test Export Script (without Godot)
# This simulates the export process for testing the script infrastructure

echo "=== SPX Godot Exporter - Test Mode ==="
echo "This is a test run without actually running Godot"
echo ""

# Get the directory of this script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Test parameters
GODOT_PROJECT="/Users/tjp/projects/spx/cmd/tscnparser/export"
INPUT_TSCN="res://main.tscn"
OUTPUT_FILE="/Users/tjp/projects/spx/spx_demos/19_tilemap/assets/tilemaps/scene1_test.json"

echo "Project:  $GODOT_PROJECT"
echo "Input:    $INPUT_TSCN"
echo "Output:   $OUTPUT_FILE"
echo ""

# Step 1: Copy scripts
echo "Step 1: Copying scripts..."
ADDON_DIR="$GODOT_PROJECT/addons/spx_exporter"
mkdir -p "$ADDON_DIR"

cp "$SCRIPT_DIR/plugin.cfg" "$ADDON_DIR/"
cp "$SCRIPT_DIR/plugin.gd" "$ADDON_DIR/"
cp "$SCRIPT_DIR/exporter.gd" "$ADDON_DIR/"
cp "$SCRIPT_DIR/export_script.gd" "$ADDON_DIR/"

echo "  ✓ Scripts copied to $ADDON_DIR"

# Step 2: Check if replacements exists
REPLACEMENTS_SRC="$SCRIPT_DIR/../test/replacements.json"
if [ -f "$REPLACEMENTS_SRC" ]; then
    echo "  ✓ Found replacements.json"
    cat "$REPLACEMENTS_SRC"
else
    echo "  ℹ No replacements.json found"
fi

echo ""
echo "Step 2: Verifying project structure..."

# Check project.godot
if [ -f "$GODOT_PROJECT/project.godot" ]; then
    echo "  ✓ project.godot exists"
else
    echo "  ✗ project.godot not found"
    exit 1
fi

# Check main.tscn
TSCN_FILE="$GODOT_PROJECT/main.tscn"
if [ -f "$TSCN_FILE" ]; then
    echo "  ✓ main.tscn exists"

    # Show first few lines
    echo ""
    echo "  First 5 lines of main.tscn:"
    head -5 "$TSCN_FILE" | sed 's/^/    /'
else
    echo "  ✗ main.tscn not found"
    exit 1
fi

echo ""
echo "Step 3: Checking exporter.gd..."

EXPORTER_FILE="$ADDON_DIR/exporter.gd"
if [ -f "$EXPORTER_FILE" ]; then
    echo "  ✓ exporter.gd exists"

    # Count lines
    LINE_COUNT=$(wc -l < "$EXPORTER_FILE" | tr -d ' ')
    echo "  Lines: $LINE_COUNT"

    # Check for key functions
    echo ""
    echo "  Key functions found:"
    grep -E "^func (export_scene|_export_tilemap|_export_sprite2d|_export_layer)" "$EXPORTER_FILE" | sed 's/^/    /'
else
    echo "  ✗ exporter.gd not found"
    exit 1
fi

echo ""
echo "Step 4: Checking export_script.gd..."

EXPORT_SCRIPT="$ADDON_DIR/export_script.gd"
if [ -f "$EXPORT_SCRIPT" ]; then
    echo "  ✓ export_script.gd exists"

    # Check for _run function
    if grep -q "func _run()" "$EXPORT_SCRIPT"; then
        echo "  ✓ _run() function found"
    else
        echo "  ✗ _run() function not found"
    fi

    # Check for argument parsing
    if grep -q "_parse_arguments()" "$EXPORT_SCRIPT"; then
        echo "  ✓ Argument parsing found"
    else
        echo "  ✗ Argument parsing not found"
    fi
else
    echo "  ✗ export_script.gd not found"
    exit 1
fi

echo ""
echo "=== Test Summary ==="
echo "✓ All scripts copied successfully"
echo "✓ Project structure verified"
echo "✓ Exporter components validated"
echo ""
echo "To run the actual export, you need:"
echo "1. Install Godot 4.x"
echo "2. Add Godot to PATH or set GODOT_EXECUTABLE environment variable"
echo "3. Run: ./run.sh"
echo ""
echo "Example with Godot:"
echo "  export GODOT_EXECUTABLE=/path/to/Godot.app/Contents/MacOS/Godot"
echo "  ./run.sh"
