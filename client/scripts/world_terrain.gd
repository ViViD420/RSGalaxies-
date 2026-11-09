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
var _biome_noise: FastNoiseLite
var _trail_paths: Array[PackedVector2Array] = []


func _ready() -> void:
	_configure_noise()
	_configure_trails()
	_build_terrain()
	_build_environment_dressing()


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

	_biome_noise = FastNoiseLite.new()
	_biome_noise.seed = terrain_seed + 104729
	_biome_noise.noise_type = FastNoiseLite.TYPE_SIMPLEX
	_biome_noise.frequency = 0.019
	_biome_noise.fractal_type = FastNoiseLite.FRACTAL_FBM
	_biome_noise.fractal_octaves = 3
	_biome_noise.fractal_gain = 0.55
	_biome_noise.fractal_lacunarity = 2.0


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
	var base_color := _base_color_for_position(position)
	var trail_distance := _distance_to_trails(Vector2(position.x, position.z))
	var blend := clampf((2.4 - trail_distance) / 0.9, 0.0, 1.0)
	blend = blend * blend * (3.0 - 2.0 * blend)
	var trail_color := Color(0.48, 0.34, 0.20)
	return base_color.lerp(trail_color, blend)


func _add_vertex(surface: SurfaceTool, position: Vector3, uv: Vector2) -> void:
	surface.set_color(_color_for_position(position))
	surface.set_uv(uv)
	surface.add_vertex(position)


func _base_color_for_position(position: Vector3) -> Color:
	var grass := Color(0.17, 0.29, 0.16)
	var moss := Color(0.25, 0.36, 0.20)
	var dry_grass := Color(0.39, 0.37, 0.25)
	var rocky_soil := Color(0.32, 0.33, 0.29)
	var moisture := _biome_noise.get_noise_2d(position.x, position.z)
	if position.y > 1.0:
		return rocky_soil.lerp(moss, clampf(1.0 - (position.y - 1.0) / 2.0, 0.0, 0.35))
	if moisture > 0.28:
		return grass.lerp(moss, clampf((moisture - 0.28) / 0.55, 0.0, 1.0))
	if moisture < -0.28:
		return grass.lerp(dry_grass, clampf((-moisture - 0.28) / 0.55, 0.0, 1.0))
	return grass.lerp(moss, clampf((position.y + 0.55) / 1.3, 0.0, 1.0))


@export_group("Environmental Dressing")
@export_range(0, 500, 1) var foliage_instance_count := 180
@export_range(0, 200, 1) var rock_instance_count := 55
@export var dressing_trail_clearance := 5.0
@export var dressing_spawn_clearance := 14.0


func _build_environment_dressing() -> void:
	if get_node_or_null("TerrainSurface") == null or get_node_or_null("TerrainCollision") == null:
		push_error("WorldTerrain cannot place environmental dressing before terrain visuals and collision exist.")
		return

	var foliage_transforms := _build_dressing_transforms(
		foliage_instance_count, 3571, 0.72, 1.28, 0.42
	)
	var rock_transforms := _build_dressing_transforms(
		rock_instance_count, 7919, 0.62, 1.34, 0.22
	)

	var foliage_mesh := SphereMesh.new()
	foliage_mesh.radius = 0.52
	foliage_mesh.height = 1.02
	foliage_mesh.radial_segments = 6
	foliage_mesh.rings = 3
	var foliage_material := StandardMaterial3D.new()
	foliage_material.albedo_color = Color(0.16, 0.27, 0.13)
	foliage_material.roughness = 1.0
	foliage_mesh.material = foliage_material
	_add_dressing_multimesh("Foliage", foliage_mesh, foliage_transforms, GeometryInstance3D.SHADOW_CASTING_SETTING_OFF)

	var rock_mesh := SphereMesh.new()
	rock_mesh.radius = 0.58
	rock_mesh.height = 0.78
	rock_mesh.radial_segments = 5
	rock_mesh.rings = 2
	var rock_material := StandardMaterial3D.new()
	rock_material.albedo_color = Color(0.34, 0.35, 0.32)
	rock_material.roughness = 1.0
	rock_mesh.material = rock_material
	_add_dressing_multimesh("Rocks", rock_mesh, rock_transforms, GeometryInstance3D.SHADOW_CASTING_SETTING_ON)


func _build_dressing_transforms(
	requested_count: int,
	seed_offset: int,
	minimum_scale: float,
	maximum_scale: float,
	y_offset: float
) -> Array[Transform3D]:
	var transforms: Array[Transform3D] = []
	if requested_count <= 0:
		return transforms

	var rng := RandomNumberGenerator.new()
	rng.seed = int(terrain_seed) + seed_offset
	var half_size := maxf(0.0, world_size * 0.5 - 8.0)
	var maximum_attempts := requested_count * 80
	var attempts := 0
	while transforms.size() < requested_count and attempts < maximum_attempts:
		attempts += 1
		var x := rng.randf_range(-half_size, half_size)
		var z := rng.randf_range(-half_size, half_size)
		if not _is_valid_dressing_site(x, z):
			continue

		var scale_x := rng.randf_range(minimum_scale, maximum_scale)
		var scale_y := rng.randf_range(minimum_scale, maximum_scale)
		var scale_z := rng.randf_range(minimum_scale, maximum_scale)
		var yaw := rng.randf_range(0.0, TAU)
		var basis := Basis(Vector3.UP, yaw).scaled(Vector3(scale_x, scale_y, scale_z))
		var ground_height := get_height_at(x, z)
		transforms.append(Transform3D(basis, Vector3(x, ground_height + y_offset, z)))
	return transforms


func _is_valid_dressing_site(x: float, z: float) -> bool:
	var half_size := world_size * 0.5
	if absf(x) > half_size - 8.0 or absf(z) > half_size - 8.0:
		return false
	if Vector2(x, z).length() < dressing_spawn_clearance:
		return false
	if _distance_to_trails(Vector2(x, z)) < dressing_trail_clearance:
		return false

	# Avoid placing decorative meshes on visibly abrupt terrain transitions.
	var sample_offset := 1.5
	var slope_x := get_height_at(x + sample_offset, z) - get_height_at(x - sample_offset, z)
	var slope_z := get_height_at(x, z + sample_offset) - get_height_at(x, z - sample_offset)
	return Vector2(slope_x, slope_z).length() <= 1.2


func _add_dressing_multimesh(
	node_name: String,
	mesh: Mesh,
	transforms: Array[Transform3D],
	shadow_setting: int
) -> void:
	var multimesh := MultiMesh.new()
	multimesh.transform_format = MultiMesh.TRANSFORM_3D
	multimesh.mesh = mesh
	multimesh.instance_count = transforms.size()
	for index in range(transforms.size()):
		multimesh.set_instance_transform(index, transforms[index])

	var instance := MultiMeshInstance3D.new()
	instance.name = node_name
	instance.multimesh = multimesh
	instance.cast_shadow = shadow_setting
	add_child(instance)
