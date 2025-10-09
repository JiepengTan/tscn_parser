@tool
extends RefCounted

## SPX Scene Exporter
##
## Exports Godot scenes to SPX tilemap JSON format.
## This class handles the conversion of TileMap, Sprite2D nodes, and prefabs
## to the format expected by the SPX game engine.

## Configuration parameters
var tile_size := Vector2i(16, 16)
var offset := Vector2.ZERO
var path_replacements: Array = []

## Internal cache for loaded prefab information
var _prefab_cache: Dictionary = {}


## Main export function - exports a scene to SPX JSON format
func export_scene(scene_root: Node) -> Dictionary:
	var result := {
		"tilemap": {},
		"decorators": [],
		"sprites": [],
		"prefabs": []
	}

	# Clear cache for fresh export
	_prefab_cache.clear()

	# Process the scene tree
	_process_node(scene_root, result)

	return result


## Recursively process nodes in the scene tree
func _process_node(node: Node, result: Dictionary, scene_root: Node = null) -> void:
	# Set scene_root on first call
	if scene_root == null:
		scene_root = node

	# Process TileMap nodes
	if node is TileMap:
		result.tilemap = _export_tilemap(node)

	# Process Sprite2D nodes (decorators)
	elif node is Sprite2D:
		result.decorators.append(_export_sprite2d(node))

	# Process instantiated scenes (prefabs)
	elif not node.scene_file_path.is_empty() and node != scene_root:
		result.sprites.append(_export_sprite_node(node))

	# Recursively process children
	for child in node.get_children():
		_process_node(child, result, scene_root)


## Export TileMap data
func _export_tilemap(tilemap: TileMap) -> Dictionary:
	var min_tile := Vector2i(1000000, 1000000)
	var max_tile := Vector2i(-1000000, -1000000)
	var tile_count := 0

	var data := {
		"format": 2,
		"tile_size": {"width": tile_size.x, "height": tile_size.y},
		"tileset": _export_tileset(tilemap.tile_set),
		"layers": [],
		"world_tile_size": {"width": 0, "height": 0}
	}

	# Export each layer
	for layer_id in tilemap.get_layers_count():
		var layer_data := _export_layer(tilemap, layer_id, min_tile, max_tile)
		data.layers.append(layer_data.layer)
		min_tile = layer_data.min_tile
		max_tile = layer_data.max_tile
		tile_count += layer_data.tile_count

	# Calculate world tile size
	if tile_count > 0:
		var width := max_tile.x - min_tile.x + 1
		var height := max_tile.y - min_tile.y + 1

		# SPX requires even dimensions
		if width % 2 != 0 or height % 2 != 0:
			push_warning("SPX Exporter: World tile size must be even. Width=%d, Height=%d" % [width, height])

		data.world_tile_size = {"width": width, "height": height}

	return data


## Export tileset information
func _export_tileset(tileset: TileSet) -> Dictionary:
	var sources := []

	if not tileset:
		return {"sources": sources}

	for source_id in tileset.get_source_count():
		var source := tileset.get_source(source_id)

		if source is TileSetAtlasSource:
			var texture_path := ""
			if source.texture:
				texture_path = _replace_path(source.texture.resource_path)

			sources.append({
				"id": source_id,
				"texture_path": texture_path,
				"tiles": _export_tiles(source)
			})

	return {"sources": sources}


## Export individual tiles from atlas source
func _export_tiles(atlas_source: TileSetAtlasSource) -> Array:
	var tiles := []

	for i in atlas_source.get_tiles_count():
		var coords := atlas_source.get_tile_id(i)
		var tile_data := atlas_source.get_tile_data(coords, 0)

		var physics := {}

		# Export physics collision data
		if tile_data and tile_data.get_collision_polygons_count(0) > 0:
			var polygon := tile_data.get_collision_polygon_points(0, 0)
			physics.collision_points = _vector2_array_to_dict_array(polygon)

		tiles.append({
			"atlas_coords": {"x": coords.x, "y": coords.y},
			"physics": physics
		})

	return tiles


## Export a single layer from TileMap
func _export_layer(tilemap: TileMap, layer_id: int, min_tile: Vector2i, max_tile: Vector2i) -> Dictionary:
	var tile_offset := Vector2i(
		int(offset.x / tile_size.x),
		int(offset.y / tile_size.y)
	)

	var layer := {
		"id": layer_id,
		"name": tilemap.get_layer_name(layer_id),
		"z_index": tilemap.get_layer_z_index(layer_id),
		"tile_data": []
	}

	var used_cells := tilemap.get_used_cells(layer_id)
	var tile_count := 0

	for cell_pos in used_cells:
		var source_id := tilemap.get_cell_source_id(layer_id, cell_pos)
		var atlas_coords := tilemap.get_cell_atlas_coords(layer_id, cell_pos)

		# Update bounds
		if cell_pos.x < min_tile.x:
			min_tile.x = cell_pos.x
		if cell_pos.x > max_tile.x:
			max_tile.x = cell_pos.x
		if cell_pos.y < min_tile.y:
			min_tile.y = cell_pos.y
		if cell_pos.y > max_tile.y:
			max_tile.y = cell_pos.y

		tile_count += 1

		# Apply offset and Y-axis inversion
		var adjusted_pos := cell_pos + tile_offset

		# Convert to SPX format: [source_id, tile_x, tile_y, atlas_x, atlas_y]
		layer.tile_data.append(source_id)
		layer.tile_data.append(adjusted_pos.x)
		layer.tile_data.append(-adjusted_pos.y)  # Y-axis inversion
		layer.tile_data.append(atlas_coords.x)
		layer.tile_data.append(atlas_coords.y)

	return {
		"layer": layer,
		"min_tile": min_tile,
		"max_tile": max_tile,
		"tile_count": tile_count
	}


## Export Sprite2D as decorator
func _export_sprite2d(sprite: Sprite2D) -> Dictionary:
	var pos := sprite.global_position + offset

	var decorator := {
		"name": sprite.name,
		"parent": sprite.get_parent().name if sprite.get_parent() else "",
		"path": "",
		"position": {"x": pos.x, "y": -pos.y},  # Y-axis inversion
		"scale": {"x": sprite.scale.x, "y": sprite.scale.y},
		"rotation": sprite.rotation + deg_to_rad(90),  # Adjust rotation for SPX
		"z_index": sprite.z_index
	}

	# Get texture path
	if sprite.texture:
		decorator.path = _replace_path(sprite.texture.resource_path)

	return decorator


## Export instantiated scene as sprite node
func _export_sprite_node(node: Node) -> Dictionary:
	var pos := Vector2.ZERO

	# Try to get position (works for Node2D and derived classes)
	if "global_position" in node:
		pos = node.global_position + offset

	var sprite := {
		"name": node.name,
		"parent": node.get_parent().name if node.get_parent() else "",
		"position": {"x": pos.x, "y": -pos.y},  # Y-axis inversion
		"path": _replace_path(node.scene_file_path),
		"properties": {}
	}

	# Export scale if available
	if "scale" in node:
		sprite.scale = {"x": node.scale.x, "y": node.scale.y}

	# Export rotation if available
	if "rotation" in node:
		sprite.rotation = node.rotation

	# Export custom properties
	for property in node.get_property_list():
		if property.usage & PROPERTY_USAGE_SCRIPT_VARIABLE:
			sprite.properties[property.name] = node.get(property.name)

	return sprite


## Build prefab nodes from sprites with prefab information
func build_prefab_nodes(sprites: Array) -> Array:
	var prefabs := []

	for sprite in sprites:
		var prefab := _sprite_to_prefab(sprite)
		if prefab:
			prefabs.append(prefab)

	return prefabs


## Convert sprite node data to prefab node with enriched information
func _sprite_to_prefab(sprite: Dictionary) -> Dictionary:
	var prefab := sprite.duplicate(true)

	# Load prefab information
	var scene_path: String = sprite.path

	# Restore .tscn extension if it was removed
	if not scene_path.ends_with(".tscn"):
		# Try common prefab locations
		var possible_paths := [
			"res://scenes/" + scene_path + ".tscn",
			"res://" + scene_path + ".tscn",
			scene_path + ".tscn"
		]

		for path in possible_paths:
			if ResourceLoader.exists(path):
				scene_path = path
				break

	var prefab_info := _load_prefab_info(scene_path)

	if prefab_info:
		# Merge prefab information
		for key in prefab_info.keys():
			prefab[key] = prefab_info[key]

	return prefab


## Load information from a prefab scene file
func _load_prefab_info(scene_path: String) -> Dictionary:
	# Check cache first
	if _prefab_cache.has(scene_path):
		return _prefab_cache[scene_path]

	# Load the scene
	if not ResourceLoader.exists(scene_path):
		push_warning("SPX Exporter: Prefab scene not found: " + scene_path)
		return {}

	var scene: PackedScene = load(scene_path)
	if not scene:
		push_warning("SPX Exporter: Failed to load prefab scene: " + scene_path)
		return {}

	var instance := scene.instantiate()
	var info := {}

	# Extract name from root node
	info.name = instance.name

	# Find Sprite2D child for texture and transform info
	var sprite := _find_child_by_type(instance, Sprite2D)
	if sprite:
		if sprite.texture:
			info.texture = _replace_path(sprite.texture.resource_path)
		info.pivot = {"x": sprite.position.x, "y": sprite.position.y}
		info.z_index = sprite.z_index
		info.scale = {"x": sprite.scale.x, "y": sprite.scale.y}
		info.rotation = sprite.rotation

	# Find collision shape
	var collision := _find_child_by_type(instance, CollisionShape2D)
	if collision:
		info.collider_pivot = {"x": collision.position.x, "y": -collision.position.y}  # Y-axis inversion

		if collision.shape:
			info.collider_type = _get_shape_type(collision.shape)
			info.collider_params = _get_shape_params(collision.shape)

		# Determine collider parent relationship
		if collision.get_parent() == instance:
			info.collider_parent = "."
		else:
			info.collider_parent = collision.get_parent().name

	# Clean up
	instance.queue_free()

	# Cache the result
	_prefab_cache[scene_path] = info

	return info


## Determine collision shape type
func _get_shape_type(shape: Shape2D) -> String:
	if shape is RectangleShape2D:
		return "rect"
	elif shape is CircleShape2D:
		return "circle"
	elif shape is CapsuleShape2D:
		return "capsule"
	elif shape is ConvexPolygonShape2D or shape is ConcavePolygonShape2D:
		return "polygon"
	else:
		return "auto"


## Extract shape parameters based on shape type
func _get_shape_params(shape: Shape2D) -> Array:
	if shape is RectangleShape2D:
		return [shape.size.x, shape.size.y]
	elif shape is CircleShape2D:
		return [shape.radius]
	elif shape is CapsuleShape2D:
		return [shape.radius, shape.height]
	elif shape is ConvexPolygonShape2D:
		return _flatten_vector2_array(shape.points)
	elif shape is ConcavePolygonShape2D:
		return _flatten_vector2_array(shape.segments)
	else:
		return []


## Find first child of specific type in node tree
func _find_child_by_type(node: Node, type) -> Node:
	if is_instance_of(node, type):
		return node

	for child in node.get_children():
		var result := _find_child_by_type(child, type)
		if result:
			return result

	return null


## Replace paths according to configuration
func _replace_path(path: String) -> String:
	for replacement in path_replacements:
		if replacement.has("old") and replacement.has("new"):
			path = path.replace(replacement.old, replacement.new)
	return path


## Convert PackedVector2Array to array of dictionaries
func _vector2_array_to_dict_array(arr: PackedVector2Array) -> Array:
	var result := []
	for v in arr:
		result.append({"x": v.x, "y": v.y})
	return result


## Flatten PackedVector2Array to flat array of floats
func _flatten_vector2_array(arr: PackedVector2Array) -> Array:
	var result := []
	for v in arr:
		result.append(v.x)
		result.append(v.y)
	return result
