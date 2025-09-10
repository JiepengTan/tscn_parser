#!/bin/bash

# Get the directory of the script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJ_DIR="$SCRIPT_DIR"
cd $PROJ_DIR
go run . -input main.tscn -replacements "replacements.json" -generateGo true
