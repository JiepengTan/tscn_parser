package tscnparser

// ConvertToTilemap merges sprites with their prefab data into decorators
// and removes sprites and prefabs from the MapData
func ConvertToTilemap(data *MapData) {
	if data == nil {
		return
	}

	// Create a map of prefab paths to prefab data for quick lookup
	prefabMap := make(map[string]*PrefabNode)
	for i := range data.Prefabs {
		prefab := &data.Prefabs[i]
		if prefab.Path != "" {
			prefabMap[prefab.Path] = prefab
		}
	}

	// Process each sprite and convert to decorator
	for _, sprite := range data.Sprites {
		decorator := DecoratorNode{
			Name:     sprite.Name,
			Parent:   sprite.Parent,
			Path:     sprite.Path,
			Position: sprite.Position,
			Scale:    sprite.Scale,
			Ratation: sprite.Ratation,
		}
		// TODO 1. adjust rotation for spx
		// TODO 2. adjust pivot for spx
		decorator.Ratation = sprite.Ratation*180/3.1415926535 + 90 // Adjust rotation for spx
		// If there's a matching prefab, merge its data
		if prefab, exists := prefabMap[sprite.Path]; exists {
			// Use name from prefab if it has one
			if prefab.Name != "" {
				decorator.Name = prefab.Name
			}

			// IMPORTANT: Use the texture path from prefab for rendering
			decorator.Path = prefab.Texture
			decorator.TextureSize = prefab.TextureSize
			// Use pivot from prefab if available
			decorator.Pivot = prefab.Pivot

			// Use z-index from prefab if available
			decorator.ZIndex = prefab.ZIndex

			// Copy collision data from prefab
			decorator.ColliderType = prefab.ColliderType
			decorator.ColliderPivot = prefab.ColliderPivot
			decorator.ColliderParams = prefab.ColliderParams

			decorator.Scale.X *= prefab.Scale.X
			decorator.Scale.Y *= prefab.Scale.Y

			// convert to local collider pivot if needed
			// Note: We use position, scale, and rotation from sprite (instance)
			// not from prefab (definition)
		}

		// Add any sprite-specific properties that might affect rendering
		if gid, ok := sprite.Properties["gid"].(int); ok && gid != 0 {
			// Skip sprites with gid (these might be special markers)
			continue
		}

		// Add the decorator to the list
		data.Decorators = append(data.Decorators, decorator)
	}

	// Clear sprites and prefabs as they've been converted to decorators
	//data.Sprites = nil
	//data.Prefabs = nil
	for idx := range data.Decorators {
		item := &data.Decorators[idx]
		//item.Position.InvertY()
		item.ColliderPivot = item.ColliderPivot.InvertY()
		item.Pivot = item.Pivot.InvertY()
		item.Parent = ""
		// convert pivot  &&  collider pivot to spx's pivot and collider pivot
		item.Pivot = item.Pivot.Add(item.TextureSize.Mulf(0.5))
		item.ColliderPivot = item.ColliderPivot.Sub(item.Pivot)
	}
}
