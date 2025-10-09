@tool
extends EditorScript

## SPX Command-Line Export Script
##
## This script allows exporting Godot scenes to SPX JSON format from the command line.
## It's designed to be run in headless mode for automation and CI/CD pipelines.
##
## Usage:
##   godot --headless --script export_script.gd -- \
##     --input res://scenes/main.tscn \
##     --output /path/to/output.json \
##     --tile-size 16 \
##     --offset-x 0 \
##     --offset-y 0

## Configuration
var input_file := "res://main.tscn"
var output_file := ""
var tile_size := 16
var offset_x := 0
var offset_y := 0
var replacements_file := ""

## Path replacements configuration
var default_replacements := [
	{"old": "res://textures/", "new": "textures/"},
	{"old": "res://scenes/", "new": ""},
	{"old": ".tscn", "new": ""}
]


func _run() -> void:
	print("=== SPX Tilemap Exporter - Command Line ===")

	# Parse command line arguments
	_parse_arguments()

	# Validate input file
	if not ResourceLoader.exists(input_file):
		printerr("Error: Input file not found: ", input_file)
		return

	# Generate default output file if not specified
	if output_file.is_empty():
		output_file = input_file.get_basename() + "_tilemap.json"

	print("Input:  ", input_file)
	print("Output: ", output_file)
	print("Tile Size: ", tile_size, "x", tile_size)
	print("Offset: (", offset_x, ", ", offset_y, ")")
	print()

	# Load the scene
	var scene: PackedScene = load(input_file)
	if not scene:
		printerr("Error: Failed to load scene: ", input_file)
		return

	var scene_instance := scene.instantiate()
	if not scene_instance:
		printerr("Error: Failed to instantiate scene")
		return

	# Create exporter
	var exporter_script = load("res://addons/spx_exporter/exporter.gd")
	if not exporter_script:
		# Try relative path from this script
		var script_dir: String = get_script().resource_path.get_base_dir()
		exporter_script = load(script_dir.path_join("exporter.gd"))

	if not exporter_script:
		printerr("Error: Failed to load exporter.gd")
		printerr("Make sure exporter.gd is in the same directory as this script")
		scene_instance.queue_free()
		return

	var exporter = exporter_script.new()

	# Configure exporter
	exporter.tile_size = Vector2i(tile_size, tile_size)
	exporter.offset = Vector2(offset_x, offset_y)

	# Load replacements from file if specified
	if not replacements_file.is_empty():
		var custom_replacements := _load_replacements_file(replacements_file)
		if custom_replacements:
			exporter.path_replacements = custom_replacements
		else:
			exporter.path_replacements = default_replacements
	else:
		exporter.path_replacements = default_replacements

	print("Starting export...")

	# Export the scene
	var result: Dictionary = exporter.export_scene(scene_instance)

	# Convert to JSON
	var json_string := JSON.stringify(result, "\t")

	# Write to file
	var file := FileAccess.open(output_file, FileAccess.WRITE)
	if not file:
		printerr("Error: Failed to open output file: ", output_file)
		printerr("Error code: ", FileAccess.get_open_error())
		scene_instance.queue_free()
		return

	file.store_string(json_string)
	file.close()

	# Clean up
	scene_instance.queue_free()

	print("Export completed successfully!")
	print("Output saved to: ", output_file)
	print()


func _parse_arguments() -> void:
	var args := OS.get_cmdline_args()

	# Find the -- separator that marks script arguments
	var script_args_start := -1
	for i in args.size():
		if args[i] == "--":
			script_args_start = i + 1
			break

	if script_args_start == -1:
		return

	# Parse arguments after --
	var i := script_args_start
	while i < args.size():
		var arg: String = args[i]

		if arg == "--input" and i + 1 < args.size():
			input_file = args[i + 1]
			i += 2
		elif arg == "--output" and i + 1 < args.size():
			output_file = args[i + 1]
			i += 2
		elif arg == "--tile-size" and i + 1 < args.size():
			tile_size = int(args[i + 1])
			i += 2
		elif arg == "--offset-x" and i + 1 < args.size():
			offset_x = int(args[i + 1])
			i += 2
		elif arg == "--offset-y" and i + 1 < args.size():
			offset_y = int(args[i + 1])
			i += 2
		elif arg == "--replacements" and i + 1 < args.size():
			replacements_file = args[i + 1]
			i += 2
		elif arg == "--help" or arg == "-h":
			_print_help()
			i += 1
		else:
			i += 1


func _load_replacements_file(path: String) -> Array:
	if not FileAccess.file_exists(path):
		printerr("Warning: Replacements file not found: ", path)
		return []

	var file := FileAccess.open(path, FileAccess.READ)
	if not file:
		printerr("Warning: Failed to open replacements file: ", path)
		return []

	var content := file.get_as_text()
	file.close()

	var json := JSON.new()
	var error := json.parse(content)

	if error != OK:
		printerr("Warning: Failed to parse replacements JSON: ", json.get_error_message())
		return []

	var data = json.data
	if data is Dictionary and data.has("replacements"):
		return data.replacements
	else:
		printerr("Warning: Invalid replacements file format")
		return []


func _print_help() -> void:
	print("""
SPX Tilemap Exporter - Command Line Tool

Usage:
  godot --headless --script export_script.gd -- [OPTIONS]

Options:
  --input <path>          Input TSCN file path (default: res://main.tscn)
  --output <path>         Output JSON file path (default: <input>_tilemap.json)
  --tile-size <size>      Tile size in pixels (default: 16)
  --offset-x <pixels>     X offset for positioning (default: 0)
  --offset-y <pixels>     Y offset for positioning (default: 0)
  --replacements <path>   JSON file with path replacement rules
  --help, -h              Show this help message

Replacements File Format:
  {
    "replacements": [
      {"old": "res://textures/", "new": "textures/"},
      {"old": "res://scenes/", "new": ""},
      {"old": ".tscn", "new": ""}
    ]
  }

Examples:
  # Basic export
  godot --headless --script export_script.gd -- --input res://main.tscn

  # Export with custom output and tile size
  godot --headless --script export_script.gd -- \\
    --input res://scenes/level1.tscn \\
    --output /output/level1.json \\
    --tile-size 32

  # Export with offset and replacements
  godot --headless --script export_script.gd -- \\
    --input res://main.tscn \\
    --offset-x 100 \\
    --offset-y 50 \\
    --replacements replacements.json
""")
