#!/bin/bash

# Standalone TSCN Parser Script
# This script uses 'go run' to directly run tscnparser from GitHub
# Version: latest

set -e

# Default parameters
DEFAULT_INPUT_TSCN="/Users/tjp/projects/spx/cmd/tscnparser/export/main.tscn"
DEFAULT_CP_DESTINATION="/Users/tjp/projects/spx/spx_demos/19_tilemap/assets/tilemaps/scene1.json"

# Use provided parameters or defaults
INPUT_TSCN_PATH="${1:-$DEFAULT_INPUT_TSCN}"
CP_DESTINATION_PATH="${2:-$DEFAULT_CP_DESTINATION}"

# Show usage if requested
if [ "$1" = "-h" ] || [ "$1" = "--help" ]; then
    echo "Usage: $0 [input_tscn_path] [cp_destination_path]"
    echo "  input_tscn_path: Path to the input tscn file (default: $DEFAULT_INPUT_TSCN)"
    echo "  cp_destination_path: Path where to copy the generated tilemap json file (default: $DEFAULT_CP_DESTINATION)"
    echo ""
    echo "Requirements: Go must be installed (this script uses 'go run')"
    exit 0
fi

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed. Please install Go from https://go.dev/dl/"
    exit 1
fi

# Get the directory of the script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEMP_DIR="$SCRIPT_DIR/_temp"
mkdir -p "$TEMP_DIR"

# Create replacements.json file
REPLACEMENTS_FILE="$TEMP_DIR/replacements.json"
cat > "$REPLACEMENTS_FILE" << 'EOF'
{
  "replacements": [
    {"old": "res://textures/", "new": "textures/"},
    {"old": "res://scenes/", "new": ""},
    {"old": ".tscn", "new": ""}
  ]
}
EOF

# Check if input file exists
if [ ! -f "$INPUT_TSCN_PATH" ]; then
    echo "Error: Input TSCN file not found: $INPUT_TSCN_PATH"
    exit 1
fi

# Copy input file to temp directory
TEMP_INPUT="$TEMP_DIR/main.tscn"
cp "$INPUT_TSCN_PATH" "$TEMP_INPUT"

# Get the prefabs directory (assuming it's relative to the input file's parent)
INPUT_DIR=$(dirname "$INPUT_TSCN_PATH")
PREFABS_DIR="$INPUT_DIR/scenes"

echo "Processing TSCN file using tscnparser (latest version)..."
echo "Input: $INPUT_TSCN_PATH"
echo "Output will be copied to: $CP_DESTINATION_PATH"

# Change to temp directory and run the parser directly from GitHub
cd "$TEMP_DIR"
go run github.com/JiepengTan/tscn_parser/cmd/tscnparser@latest \
    -input "main.tscn" \
    -replacements "replacements.json" \
    -tilesize 16 \
    -offsetx 80 \
    -offsety 384 \
    --prefabs "$PREFABS_DIR"

# Check if output was generated
OUTPUT_FILE="$TEMP_DIR/main_tilemap.json"
if [ ! -f "$OUTPUT_FILE" ]; then
    echo "Error: Output file was not generated"
    exit 1
fi

# Copy the output to destination
mkdir -p "$(dirname "$CP_DESTINATION_PATH")"
cp "$OUTPUT_FILE" "$CP_DESTINATION_PATH"

echo "Success! Output copied to: $CP_DESTINATION_PATH"
