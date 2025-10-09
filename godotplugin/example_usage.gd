@tool
extends EditorScript

## Example Usage Script for SPX Exporter
##
## This script demonstrates various ways to use the SPX exporter programmatically.
## You can run this from the Godot editor: File → Run Script → Select this file

func _run() -> void:
	print("=== SPX Exporter - Example Usage ===\n")

	# Example 1: Basic single scene export
	example_1_basic_export()

	print()

	# Example 2: Export with custom configuration
	example_2_custom_config()

	print()

	# Example 3: Batch export multiple scenes
	example_3_batch_export()

	print()

	# Example 4: Export with custom path replacements
	example_4_custom_replacements()

	print("\n=== Examples completed ===")


## Example 1: Basic single scene export with default settings
func example_1_basic_export() -> void:
	print("--- Example 1: Basic Export ---")

	var scene_path := "res://main.tscn"
	var output_path := "res://main_tilemap.json"

	# Check if scene exists
	if not ResourceLoader.exists(scene_path):
		print("Scene not found: ", scene_path)
		print("Skipping this example")
		return

	# Load and instantiate the scene
	var scene: PackedScene = load(scene_path)
	var scene_instance := scene.instantiate()

	# Create exporter with default settings
	var exporter = load("res://addons/spx_exporter/exporter.gd").new()

	# Export the scene
	var result := exporter.export_scene(scene_instance)

	# Save to JSON
	_save_json(result, output_path)

	# Clean up
	scene_instance.queue_free()

	print("Exported: ", scene_path, " → ", output_path)


## Example 2: Export with custom tile size and offset
func example_2_custom_config() -> void:
	print("--- Example 2: Custom Configuration ---")

	var scene_path := "res://main.tscn"
	var output_path := "res://main_custom.json"

	if not ResourceLoader.exists(scene_path):
		print("Scene not found: ", scene_path)
		print("Skipping this example")
		return

	var scene: PackedScene = load(scene_path)
	var scene_instance := scene.instantiate()

	# Create exporter with custom settings
	var exporter = load("res://addons/spx_exporter/exporter.gd").new()

	# Configure custom tile size
	exporter.tile_size = Vector2i(32, 32)

	# Configure offset
	exporter.offset = Vector2(100, 50)

	# Configure path replacements
	exporter.path_replacements = [
		{"old": "res://", "new": ""},
		{"old": ".png", "new": ".webp"}  # Example: change texture format
	]

	# Export
	var result := exporter.export_scene(scene_instance)
	_save_json(result, output_path)

	scene_instance.queue_free()

	print("Exported with custom config:")
	print("  Tile size: 32x32")
	print("  Offset: (100, 50)")
	print("  Output: ", output_path)


## Example 3: Batch export multiple scenes
func example_3_batch_export() -> void:
	print("--- Example 3: Batch Export ---")

	# List of scenes to export
	var scenes := [
		"res://scenes/level1.tscn",
		"res://scenes/level2.tscn",
		"res://scenes/level3.tscn"
	]

	var exporter = load("res://addons/spx_exporter/exporter.gd").new()

	# Configure exporter once
	exporter.tile_size = Vector2i(16, 16)
	exporter.path_replacements = [
		{"old": "res://textures/", "new": "textures/"},
		{"old": "res://scenes/", "new": ""},
		{"old": ".tscn", "new": ""}
	]

	var exported_count := 0

	for scene_path in scenes:
		if not ResourceLoader.exists(scene_path):
			print("  Skipping (not found): ", scene_path)
			continue

		var output_path := scene_path.get_basename() + "_tilemap.json"

		var scene: PackedScene = load(scene_path)
		var scene_instance := scene.instantiate()

		var result := exporter.export_scene(scene_instance)
		_save_json(result, output_path)

		scene_instance.queue_free()

		print("  Exported: ", scene_path.get_file(), " → ", output_path.get_file())
		exported_count += 1

	print("Batch export completed: ", exported_count, " scenes")


## Example 4: Export with custom path replacements loaded from file
func example_4_custom_replacements() -> void:
	print("--- Example 4: Custom Replacements from File ---")

	var scene_path := "res://main.tscn"
	var output_path := "res://main_replaced.json"
	var replacements_file := "res://replacements.json"

	if not ResourceLoader.exists(scene_path):
		print("Scene not found: ", scene_path)
		print("Skipping this example")
		return

	var scene: PackedScene = load(scene_path)
	var scene_instance := scene.instantiate()

	var exporter = load("res://addons/spx_exporter/exporter.gd").new()

	# Load replacements from file if it exists
	if FileAccess.file_exists(replacements_file):
		var replacements := _load_replacements(replacements_file)
		if replacements:
			exporter.path_replacements = replacements
			print("  Loaded replacements from: ", replacements_file)
	else:
		print("  Replacements file not found, using defaults")
		exporter.path_replacements = [
			{"old": "res://textures/", "new": "textures/"},
			{"old": "res://scenes/", "new": ""},
			{"old": ".tscn", "new": ""}
		]

	# Export
	var result := exporter.export_scene(scene_instance)
	_save_json(result, output_path)

	scene_instance.queue_free()

	print("  Output: ", output_path)


## Helper: Save dictionary as JSON file
func _save_json(data: Dictionary, path: String) -> void:
	var json_string := JSON.stringify(data, "\t")
	var file := FileAccess.open(path, FileAccess.WRITE)

	if not file:
		printerr("Failed to write file: ", path)
		return

	file.store_string(json_string)
	file.close()


## Helper: Load replacements from JSON file
func _load_replacements(path: String) -> Array:
	var file := FileAccess.open(path, FileAccess.READ)
	if not file:
		return []

	var content := file.get_as_text()
	file.close()

	var json := JSON.new()
	var error := json.parse(content)

	if error != OK:
		printerr("Failed to parse JSON: ", json.get_error_message())
		return []

	var data = json.data
	if data is Dictionary and data.has("replacements"):
		return data.replacements

	return []
