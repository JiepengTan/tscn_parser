@tool
extends EditorPlugin

## SPX Tilemap Exporter Plugin
##
## This plugin adds the ability to export Godot scenes to SPX tilemap JSON format.
## It provides menu integration for easy access from the editor.

var exporter_instance: RefCounted
var export_dialog: FileDialog

func _enter_tree() -> void:
	# Add menu item to the Project menu
	add_tool_menu_item("Export Scene to SPX JSON", _on_export_to_spx_pressed)

	# Create file dialog for output selection
	export_dialog = FileDialog.new()
	export_dialog.file_mode = FileDialog.FILE_MODE_SAVE_FILE
	export_dialog.access = FileDialog.ACCESS_FILESYSTEM
	export_dialog.add_filter("*.json", "JSON Files")
	export_dialog.file_selected.connect(_on_export_file_selected)
	get_editor_interface().get_base_control().add_child(export_dialog)

	print("SPX Tilemap Exporter: Plugin loaded successfully")


func _exit_tree() -> void:
	# Clean up menu item
	remove_tool_menu_item("Export Scene to SPX JSON")

	# Clean up dialog
	if export_dialog:
		export_dialog.queue_free()
		export_dialog = null

	print("SPX Tilemap Exporter: Plugin unloaded")


func _on_export_to_spx_pressed() -> void:
	# Get the currently edited scene
	var current_scene := get_editor_interface().get_edited_scene_root()

	if not current_scene:
		push_error("SPX Exporter: No scene is currently open. Please open a scene first.")
		printerr("No scene opened. Please open a scene before exporting.")
		return

	# Get the scene file path to suggest output path
	var scene_path: String = current_scene.scene_file_path
	if scene_path.is_empty():
		push_error("SPX Exporter: Current scene has not been saved. Please save the scene first.")
		printerr("Scene not saved. Please save the scene before exporting.")
		return

	# Set default output path
	var default_output_path := scene_path.get_basename() + "_tilemap.json"
	export_dialog.current_path = default_output_path
	export_dialog.current_file = default_output_path.get_file()

	# Show file dialog
	export_dialog.popup_centered(Vector2i(800, 600))


func _on_export_file_selected(path: String) -> void:
	var current_scene := get_editor_interface().get_edited_scene_root()

	if not current_scene:
		push_error("SPX Exporter: No scene available for export")
		return

	# Perform the export
	_export_scene_to_file(current_scene, path)


func _export_scene_to_file(scene_root: Node, output_path: String) -> void:
	print("SPX Exporter: Starting export...")
	print("  Scene: ", scene_root.scene_file_path)
	print("  Output: ", output_path)

	# Load the exporter script
	var exporter_script = load("res://addons/spx_exporter/exporter.gd")
	if not exporter_script:
		push_error("SPX Exporter: Failed to load exporter.gd")
		printerr("Failed to load exporter script")
		return

	# Create exporter instance
	exporter_instance = exporter_script.new()

	# Configure exporter (these can be exposed as plugin settings later)
	exporter_instance.tile_size = Vector2i(16, 16)
	exporter_instance.offset = Vector2.ZERO
	exporter_instance.path_replacements = [
		{"old": "res://textures/", "new": "textures/"},
		{"old": "res://scenes/", "new": ""},
		{"old": ".tscn", "new": ""}
	]

	# Export the scene
	var result: Dictionary = exporter_instance.export_scene(scene_root)

	# Convert to JSON
	var json_string := JSON.stringify(result, "\t")

	# Write to file
	var file := FileAccess.open(output_path, FileAccess.WRITE)
	if not file:
		push_error("SPX Exporter: Failed to open output file: " + output_path)
		printerr("Failed to open file for writing: ", output_path)
		printerr("Error: ", FileAccess.get_open_error())
		return

	file.store_string(json_string)
	file.close()

	print("SPX Exporter: Export completed successfully!")
	print("  Output file: ", output_path)

	# Show success notification in editor
	get_editor_interface().get_base_control().notification(NOTIFICATION_WM_ABOUT)
