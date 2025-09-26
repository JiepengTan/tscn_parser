#!/bin/bash

# Default parameters
DEFAULT_INPUT_TSCN="/Users/tjp/projects/robot/spx/cmd/tscnparser/export2/main.tscn"
DEFAULT_CP_DESTINATION="/Users/tjp/projects/robot/spx/spx_demos/19_tilemap/assets/tilemaps/scene1.json"

# Use provided parameters or defaults
INPUT_TSCN_PATH="${1:-$DEFAULT_INPUT_TSCN}"
CP_DESTINATION_PATH="${2:-$DEFAULT_CP_DESTINATION}"

# Show usage if requested
if [ "$1" = "-h" ] || [ "$1" = "--help" ]; then
    echo "Usage: $0 [input_tscn_path] [cp_destination_tscn_path]"
    echo "  input_tscn_path: Path to the input tscn file (default: $DEFAULT_INPUT_TSCN)"
    echo "  cp_destination_tscn_path: Path where to copy the generated tilemap json file (default: $DEFAULT_CP_DESTINATION)"
    exit 0
fi

# Get the directory of the script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJ_DIR="$SCRIPT_DIR"
cd $PROJ_DIR
go mod tidy

cp -rf "$INPUT_TSCN_PATH" main.tscn
go run . -input  main.tscn -replacements "replacements.json" -tilesize 16 -offsetx 72 -offsety 392
cp -rf main_tilemap.json "$CP_DESTINATION_PATH"
