extends Node3D

@export_group("Terrain Shape")
@export var world_size := 240.0
@export_range(16, 192, 1) var grid_resolution := 96
@export var height_scale := 1.8
@export var detail_scale := 0.22

@export_group("Deterministic Noise")
@export var terrain_seed := 42026
@export var base_frequency := 0.012
@export var detail_frequency := 0.045

var _broad_noise: FastNoiseLite
var _detail_noise: FastNoiseLite
var _trail_paths: Array[PackedVector2Array] = []


func _ready() -> void:
	_configure_noise()
	_configure_trails()
	_build_terrain()


func get_height_at(x: float, z: float) -> float:
	if _broad_noise == null or _detail_noise == null:
		_configure_noise()
	return (
		_broad_noise.get_noise_2d(x, z) * height_scale
		+ _detail_noise.get_noise_2d(x, z) * detail_scale
	)


func _configure_noise() -> void:
	_broad_noise = FastNoiseLite.new()
	_broad_noise.seed = terrain_seed
	_broad_noise.noise_type = FastNoiseLite.TYPE_SIMPLEX
	_broad_noise.frequency = base_frequency
	_broad_noise.fractal_type = FastNoiseLite.FRACTAL_FBM
	_broad_noise.fractal_octaves = 4
	_broad_noise.fractal_gain = 0.48
	_broad_noise.fractal_lacunarity = 2.0

	_detail_noise = FastNoiseLite.new()
	_detail_noise.seed = terrain_seed + 7919
	_detail_noise.noise_type = FastNoiseLite.TYPE_PERLIN
	_detail_noise.frequency = detail_frequency
	_detail_noise.fractal_type = FastNoiseLite.FRACTAL_FBM
	_detail_noise.fractal_octaves = 2
	_detail_noise.fractal_gain = 0.42
	_detail_noise.fractal_lacunarity = 2.0


func _build_terrain() -> void:
	var surface := SurfaceTool.new()
	surface.begin(Mesh.PRIMITIVE_TRIANGLES)
	var half_size := world_size * 0.5
	var step := world_size / float(grid_resolution)
	var uv_scale := 12.0

	for grid_z in range(grid_resolution):
		for grid_x in range(grid_resolution):
			var x0 := -half_size + float(grid_x) * step
			var x1 := x0 + step
			var z0 := -half_size + float(grid_z) * step
			var z1 := z0 + step

			var h00 := get_height_at(x0, z0)
			var h10 := get_height_at(x1, z0)
			var h01 := get_height_at(x0, z1)
			var h11 := get_height_at(x1, z1)

			var v00 := Vector3(x0, h00, z0)
			var v10 := Vector3(x1, h10, z0)
			var v01 := Vector3(x0, h01, z1)
			var v11 := Vector3(x1, h11, z1)

			# Counter-clockwise winding keeps generated surface normals upward.
			_add_vertex(surface, v00, Vector2(float(grid_x), float(grid_z)) / grid_resolution * uv_scale)
			_add_vertex(surface, v01, Vector2(float(grid_x), float(grid_z + 1)) / grid_resolution * uv_scale)
			_add_vertex(surface, v10, Vector2(float(grid_x + 1), float(grid_z)) / grid_resolution * uv_scale)

			_add_vertex(surface, v10, Vector2(float(grid_x + 1), float(grid_z)) / grid_resolution * uv_scale)
			_add_vertex(surface, v01, Vector2(float(grid_x), float(grid_z + 1)) / grid_resolution * uv_scale)
			_add_vertex(surface, v11, Vector2(float(grid_x + 1), float(grid_z + 1)) / grid_resolution * uv_scale)

	surface.index()
	surface.generate_normals()
	var terrain_mesh := surface.commit()
	if terrain_mesh == null:
		push_error("WorldTerrain could not build its terrain mesh.")
		return

	var material := StandardMaterial3D.new()
	material.vertex_color_use_as_albedo = true
	material.roughness = 1.0
	material.cull_mode = BaseMaterial3D.CULL_DISABLED

	var terrain_visual := MeshInstance3D.new()
	terrain_visual.name = "TerrainSurface"
	terrain_visual.mesh = terrain_mesh
	terrain_visual.material_override = material
	terrain_visual.cast_shadow = GeometryInstance3D.SHADOW_CASTING_SETTING_ON
	add_child(terrain_visual)

	var terrain_body := StaticBody3D.new()
	terrain_body.name = "TerrainCollision"
	var terrain_collision := CollisionShape3D.new()
	terrain_collision.name = "TerrainShape"
	var collision_mesh := terrain_mesh.create_trimesh_shape() as ConcavePolygonShape3D
	if collision_mesh == null or collision_mesh.get_faces().is_empty():
		push_error("WorldTerrain could not create non-empty terrain collision faces.")
		terrain_visual.queue_free()
		terrain_body.queue_free()
		return
	# Double-sided collision protects grounding at triangle seams and keeps
	# collision robust if a future terrain mesh winding change is introduced.
	collision_mesh.backface_collision = true
	terrain_collision.shape = collision_mesh
	terrain_body.collision_layer = 1
	terrain_body.collision_mask = 1
	terrain_body.add_child(terrain_collision)
	add_child(terrain_body)


func _configure_trails() -> void:
	# Trail centerlines are stored in zone-local X/Z coordinates and are
	# painted into the terrain's existing vertex-color surface. This avoids
	# coplanar ribbon meshes and keeps the terrain to one visual draw surface.
	_trail_paths = [
		PackedVector2Array([
			Vector2(0.0, -4.0), Vector2(0.0, -18.0), Vector2(-4.0, -34.0),
			Vector2(-15.0, -52.0), Vector2(-32.0, -72.0), Vector2(-48.0, -110.0)
		]),
		PackedVector2Array([
			Vector2(4.0, 0.0), Vector2(18.0, 0.0), Vector2(36.0, -5.0),
			Vector2(55.0, -17.0), Vector2(78.0, -28.0), Vector2(110.0, -38.0)
		]),
		PackedVector2Array([
			Vector2(0.0, 0.0), Vector2(-16.0, 5.0), Vector2(-34.0, 14.0),
			Vector2(-52.0, 29.0), Vector2(-78.0, 48.0), Vector2(-110.0, 65.0)
		]),
		PackedVector2Array([
			Vector2(0.0, 6.0), Vector2(4.0, 24.0), Vector2(18.0, 42.0),
			Vector2(38.0, 61.0), Vector2(62.0, 82.0), Vector2(85.0, 110.0)
		])
	]


func _distance_to_trails(point: Vector2) -> float:
	var nearest_distance := INF
	for path in _trail_paths:
		for index in range(path.size() - 1):
			var start := path[index]
			var segment := path[index + 1] - start
			var length_squared := segment.length_squared()
			if length_squared <= 0.0001:
				continue
			var along := clampf((point - start).dot(segment) / length_squared, 0.0, 1.0)
			var nearest_point := start + segment * along
			nearest_distance = minf(nearest_distance, point.distance_to(nearest_point))
	return nearest_distance


func _color_for_position(position: Vector3) -> Color:
	var base_color := _base_color_for_height(position.y)
	var trail_distance := _distance_to_trails(Vector2(position.x, position.z))
	var blend := clampf((2.4 - trail_distance) / 0.9, 0.0, 1.0)
	blend = blend * blend * (3.0 - 2.0 * blend)
	var trail_color := Color(0.48, 0.34, 0.20)
	return base_color.lerp(trail_color, blend)


func _add_vertex(surface: SurfaceTool, position: Vector3, uv: Vector2) -> void:
	surface.set_color(_color_for_position(position))
	surface.set_uv(uv)
	surface.add_vertex(position)


func _base_color_for_height(height: float) -> Color:
	var grass := Color(0.17, 0.29, 0.16)
	var moss := Color(0.25, 0.36, 0.20)
	var dry_grass := Color(0.39, 0.37, 0.25)
	if height < -0.55:
		return grass.darkened(0.12)
	if height > 0.75:
		return moss.lerp(dry_grass, clampf((height - 0.75) / 1.0, 0.0, 0.65))
	return grass.lerp(moss, clampf((height + 0.55) / 1.3, 0.0, 1.0))
