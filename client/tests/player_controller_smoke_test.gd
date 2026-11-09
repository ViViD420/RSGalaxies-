extends SceneTree

const PLAYER_SCENE := preload("res://scenes/player.tscn")
const WORLD_TERRAIN_SCRIPT := preload("res://scripts/world_terrain.gd")


func _initialize() -> void:
	call_deferred("_run_movement_smoke_test")


func _run_movement_smoke_test() -> void:
	for action_name in ["move_left", "move_right", "move_forward", "move_back", "jump"]:
		if not InputMap.has_action(action_name):
			InputMap.add_action(action_name)

	var terrain := WORLD_TERRAIN_SCRIPT.new() as Node3D
	root.add_child(terrain)
	var terrain_surface := terrain.get_node_or_null("TerrainSurface") as MeshInstance3D
	if terrain_surface == null:
		_fail("Procedural terrain did not create its TerrainSurface node.")
		return
	var generated_mesh := terrain_surface.mesh
	if generated_mesh == null:
		_fail("Procedural terrain TerrainSurface has no generated mesh resource.")
		return
	if generated_mesh.get_surface_count() == 0:
		_fail("Procedural terrain mesh has no renderable surfaces.")
		return

	var terrain_body := terrain.get_node_or_null("TerrainCollision") as StaticBody3D
	var terrain_collision := terrain.get_node_or_null("TerrainCollision/TerrainShape") as CollisionShape3D
	if terrain_body == null or terrain_collision == null or terrain_collision.shape == null:
		_fail("Procedural terrain did not create collision geometry.")
		return
	if terrain_body.collision_layer == 0 or terrain_collision.disabled:
		_fail("Procedural terrain collision is disabled or has no collision layer.")
		return
	var collision_mesh := terrain_collision.shape as ConcavePolygonShape3D
	if collision_mesh == null or collision_mesh.get_faces().is_empty():
		_fail("Procedural terrain collision has no triangle faces.")
		return
	print("Terrain collision face vertices: ", collision_mesh.get_faces().size())

	var spawn_height := float(terrain.call("get_height_at", 0.0, 0.0))
	var distant_height := float(terrain.call("get_height_at", 24.0, -18.0))
	var repeated_height := float(terrain.call("get_height_at", 24.0, -18.0))
	if not is_finite(spawn_height) or not is_finite(distant_height):
		_fail("Terrain height sampling returned a non-finite value.")
		return
	if absf(distant_height - repeated_height) > 0.0001:
		_fail("Terrain height sampling is not deterministic.")
		return
	if absf(spawn_height - distant_height) < 0.01:
		_fail("Terrain heightmap did not produce any elevation variation.")
		return

	# Wait for the physics server to register the generated static trimesh.
	await physics_frame
	var ray_from := Vector3(0.0, spawn_height + 10.0, 0.0)
	var ray_to := Vector3(0.0, spawn_height - 10.0, 0.0)
	var ray_query := PhysicsRayQueryParameters3D.create(ray_from, ray_to, 1)
	var terrain_hit := root.world_3d.direct_space_state.intersect_ray(ray_query)
	if terrain_hit.is_empty():
		_fail("Physics ray could not hit generated terrain at the spawn point; spawn_height=%s faces=%s." % [spawn_height, collision_mesh.get_faces().size()])
		return
	if absf(float(terrain_hit.position.y) - spawn_height) > 0.15:
		_fail("Terrain collision ray hit does not match deterministic height sample; hit=%s expected=%s." % [terrain_hit.position.y, spawn_height])
		return

	var player := PLAYER_SCENE.instantiate() as CharacterBody3D
	player.collision_layer = 1
	player.collision_mask = 1
	player.position = Vector3(0.0, spawn_height + 0.1, 0.0)
	root.add_child(player)

	var initial_ground_frames := 0
	while not player.is_on_floor() and initial_ground_frames < 120:
		await physics_frame
		await process_frame
		initial_ground_frames += 1
	if not player.is_on_floor():
		var current_ground_height := float(terrain.call("get_height_at", player.global_position.x, player.global_position.z))
		_fail("Player did not settle on generated terrain collision; position=%s ground_height=%s velocity=%s faces=%s." % [player.global_position, current_ground_height, player.velocity, collision_mesh.get_faces().size()])
		return

	var starting_z := player.global_position.z
	Input.action_press("move_forward")
	await physics_frame
	await process_frame
	var initial_speed := Vector2(player.velocity.x, player.velocity.z).length()
	for frame in range(19):
		await physics_frame
	await process_frame
	Input.action_release("move_forward")

	var travelled_forward := starting_z - player.global_position.z
	var peak_speed := Vector2(player.velocity.x, player.velocity.z).length()
	var configured_move_speed: float = float(player.get("move_speed"))
	if travelled_forward < 0.3:
		_fail("Forward input did not move the character far enough.")
		return
	if initial_speed <= 0.0 or initial_speed >= configured_move_speed:
		_fail("Ground acceleration did not ramp speed from rest.")
		return
	if peak_speed < configured_move_speed * 0.8:
		_fail("Sustained movement did not approach the configured movement speed.")
		return

	for frame in range(8):
		await physics_frame
	await process_frame
	var slowed_speed := Vector2(player.velocity.x, player.velocity.z).length()
	if slowed_speed >= peak_speed:
		_fail("Ground deceleration did not reduce horizontal speed after input release.")
		return

	var landing_frames := 0
	while not player.is_on_floor() and landing_frames < 120:
		await physics_frame
		landing_frames += 1
	if not player.is_on_floor():
		_fail("Player did not settle onto the test floor before jump validation; position=%s velocity=%s." % [player.global_position, player.velocity])
		return

	var jump_binding := InputEventKey.new()
	jump_binding.physical_keycode = KEY_SPACE
	InputMap.action_add_event("jump", jump_binding)
	var jump_press := InputEventKey.new()
	jump_press.physical_keycode = KEY_SPACE
	jump_press.pressed = true
	Input.parse_input_event(jump_press)
	await physics_frame
	await process_frame
	var jump_speed := player.velocity.y
	jump_press.pressed = false
	Input.parse_input_event(jump_press)
	if jump_speed <= 0.0:
		_fail("Jump input did not produce upward velocity while grounded; on_floor=%s position=%s velocity=%s action_pressed=%s." % [player.is_on_floor(), player.global_position, player.velocity, Input.is_action_pressed("jump")])
		return

	var collision_shape := player.get_node("CollisionShape3D") as CollisionShape3D
	if not is_equal_approx(collision_shape.position.y, 0.9):
		_fail("Locomotion presentation unexpectedly changed the collision shape transform.")
		return

	var camera_arm := player.get_node("CameraPivot/SpringArm3D") as SpringArm3D
	var zoom_in := InputEventMouseButton.new()
	zoom_in.button_index = MOUSE_BUTTON_WHEEL_UP
	zoom_in.pressed = true
	player.call("_unhandled_input", zoom_in)
	if not is_equal_approx(camera_arm.spring_length, 4.75):
		_fail("Mouse-wheel camera zoom-in did not adjust the spring arm by one step.")
		return
	for frame in range(20):
		player.call("_unhandled_input", zoom_in)
	if not is_equal_approx(camera_arm.spring_length, float(player.get("camera_min_distance"))):
		_fail("Camera zoom-in exceeded or failed to reach its configured minimum distance.")
		return
	var zoom_out := InputEventMouseButton.new()
	zoom_out.button_index = MOUSE_BUTTON_WHEEL_DOWN
	zoom_out.pressed = true
	for frame in range(30):
		player.call("_unhandled_input", zoom_out)
	if not is_equal_approx(camera_arm.spring_length, float(player.get("camera_max_distance"))):
		_fail("Camera zoom-out exceeded or failed to reach its configured maximum distance.")
		return

	player.call("play_attack_animation")
	var attack_tween := player.get("_attack_tween") as Tween
	if attack_tween == null or not attack_tween.is_running():
		_fail("Successful combat presentation did not start the attack swing animation.")
		return
	await player.get_tree().create_timer(0.3).timeout
	var final_attack_rotation := (player.get_node("BodyMesh") as MeshInstance3D).rotation.y
	if absf(final_attack_rotation) > 0.01 or attack_tween.is_running():
		_fail("Attack swing animation did not return the body mesh to its neutral pose; rotation_y=%s tween_running=%s." % [final_attack_rotation, attack_tween.is_running()])
		return

	print("PASS: terrain mesh/collision, deterministic height, movement acceleration, sustained speed, deceleration, jump, camera zoom bounds, attack swing and collision/presentation separation.")
	for child in root.get_children():
		child.queue_free()
	await process_frame
	quit(0)


func _fail(message: String) -> void:
	push_error("Player movement smoke test failed: " + message)
	for child in root.get_children():
		child.queue_free()
	call_deferred("_quit_after_cleanup", 1)


func _quit_after_cleanup(exit_code: int) -> void:
	await process_frame
	quit(exit_code)
