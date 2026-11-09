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


func _ready() -> void:
	_configure_noise()
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
	terrain_collision.shape = terrain_mesh.create_trimesh_shape()
	if terrain_collision.shape == null:
		push_error("WorldTerrain could not create terrain collision.")
		terrain_visual.queue_free()
		terrain_body.queue_free()
		return
	terrain_body.add_child(terrain_collision)
	add_child(terrain_body)


func _add_vertex(surface: SurfaceTool, position: Vector3, uv: Vector2) -> void:
	surface.set_color(_color_for_height(position.y))
	surface.set_uv(uv)
	surface.add_vertex(position)


func _color_for_height(height: float) -> Color:
	var grass := Color(0.17, 0.29, 0.16)
	var moss := Color(0.25, 0.36, 0.20)
	var dry_grass := Color(0.39, 0.37, 0.25)
	if height < -0.55:
		return grass.darkened(0.12)
	if height > 0.75:
		return moss.lerp(dry_grass, clampf((height - 0.75) / 1.0, 0.0, 0.65))
	return grass.lerp(moss, clampf((height + 0.55) / 1.3, 0.0, 1.0))
