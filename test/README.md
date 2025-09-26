# TSCN Parser for spx

## Usage

Use the `run.sh` script to parse TSCN files and generate tilemap JSON files.

### Basic Usage

```bash
# Specify both input and output paths
./run.sh input.tscn /path/to/output.json
```

### Parameters

- `input_tscn_path` (optional): Path to the input TSCN file
  - Default: `main.tscn`
- `cp_destination_tscn_path` (optional): Path where to copy the generated tilemap JSON file  
  - Default: `/Users/tjp/projects/robot/spx/spx_demos/19_tilemap/assets/tilemaps/scene1.json`

### Help

```bash
./run.sh --help
```
