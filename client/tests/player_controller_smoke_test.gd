extends SceneTree

const PLAYER_SCENE := preload("res://scenes/player.tscn")


func _initialize() -> void:
	call_deferred("_run_movement_smoke_test")


func _run_movement_smoke_test() -> void:
	for action_name in ["move_left", "move_right", "move_forward", "move_back", "jump"]:
		if not InputMap.has_action(action_name):
			InputMap.add_action(action_name)

	var ground := StaticBody3D.new()
	var ground_collision := CollisionShape3D.new()
	var ground_shape := BoxShape3D.new()
	ground_shape.size = Vector3(100.0, 1.0, 100.0)
	ground_collision.shape = ground_shape
	ground.add_child(ground_collision)
	ground.position.y = -0.5
	root.add_child(ground)

	var player := PLAYER_SCENE.instantiate() as CharacterBody3D
	player.position = Vector3(0.0, 0.05, 0.0)
	root.add_child(player)

	for frame in range(5):
		await physics_frame

	var starting_z := player.global_position.z
	Input.action_press("move_forward")
	await physics_frame
	var initial_speed := Vector2(player.velocity.x, player.velocity.z).length()
	for frame in range(19):
		await physics_frame
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
	var slowed_speed := Vector2(player.velocity.x, player.velocity.z).length()
	if slowed_speed >= peak_speed:
		_fail("Ground deceleration did not reduce horizontal speed after input release.")
		return

	for frame in range(5):
		await physics_frame
	Input.action_press("jump")
	await physics_frame
	Input.action_release("jump")
	if player.velocity.y <= 0.0:
		_fail("Buffered jump input did not produce upward velocity while grounded.")
		return

	var collision_shape := player.get_node("CollisionShape3D") as CollisionShape3D
	if not is_equal_approx(collision_shape.position.y, 0.9):
		_fail("Locomotion presentation unexpectedly changed the collision shape transform.")
		return

	print("PASS: movement acceleration, sustained speed, deceleration, jump and collision/presentation separation.")
	quit(0)


func _fail(message: String) -> void:
	push_error("Player movement smoke test failed: " + message)
	quit(1)
