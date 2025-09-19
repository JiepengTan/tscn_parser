#!/bin/bash

# Get the directory of the script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJ_DIR="$SCRIPT_DIR"
cd $PROJ_DIR
go run . -input main.tscn -replacements "replacements.json"

# cp -rf main_tilemap.go.txt  /Users/tjp/projects/robot/spx/spx_demos/19_tilemap/Scene1.spx
cp -rf main_tilemap.json  /Users/tjp/projects/robot/spx/spx_demos/19_tilemap/assets/tilemaps/scene1.json
echo done