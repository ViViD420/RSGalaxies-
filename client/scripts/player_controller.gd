extends CharacterBody3D

@export_group("Movement")
@export var move_speed := 7.0
@export var ground_acceleration := 30.0
@export var ground_deceleration := 36.0
@export var air_acceleration := 9.0
@export var air_deceleration := 3.5
@export var jump_velocity := 5.0
@export var gravity_multiplier := 1.0
@export var max_fall_speed := 32.0
@export var coyote_time := 0.12
@export var jump_buffer_time := 0.14
@export var turn_speed := 12.0

@export_group("Camera")
@export var mouse_sensitivity := 0.0025
@export var min_camera_pitch_degrees := -65.0
@export var max_camera_pitch_degrees := 35.0

@export_group("Placeholder Locomotion Presentation")
@export var movement_bob_height := 0.035
@export var movement_bob_frequency := 2.0
@export var movement_lean_radians := 0.075

@onready var camera_pivot: Node3D = $CameraPivot
@onready var body_mesh: MeshInstance3D = $BodyMesh

var gravity: float = ProjectSettings.get_setting("physics/3d/default_gravity")
var _coyote_time_left := 0.0
var _jump_buffer_left := 0.0
var _body_mesh_base_y := 0.9
var _locomotion_clock := 0.0


func _ready() -> void:
	_body_mesh_base_y = body_mesh.position.y
	if not DisplayServer.is_touchscreen_available():
		Input.mouse_mode = Input.MOUSE_MODE_CAPTURED


func _notification(what: int) -> void:
	if what == NOTIFICATION_APPLICATION_FOCUS_OUT and Input.mouse_mode == Input.MOUSE_MODE_CAPTURED:
		Input.mouse_mode = Input.MOUSE_MODE_VISIBLE


func _unhandled_input(event: InputEvent) -> void:
	if event is InputEventMouseMotion and Input.mouse_mode == Input.MOUSE_MODE_CAPTURED:
		_rotate_camera(event.relative)
	elif event is InputEventScreenDrag:
		_rotate_camera(event.relative)
	elif event is InputEventMouseButton and event.pressed and event.button_index == MOUSE_BUTTON_LEFT:
		Input.mouse_mode = Input.MOUSE_MODE_CAPTURED


func _rotate_camera(relative: Vector2) -> void:
	camera_pivot.rotate_y(-relative.x * mouse_sensitivity)
	camera_pivot.rotation.x = clampf(
		camera_pivot.rotation.x - relative.y * mouse_sensitivity,
		deg_to_rad(min_camera_pitch_degrees),
		deg_to_rad(max_camera_pitch_degrees)
	)


func _physics_process(delta: float) -> void:
	_update_jump_timers(delta)
	_update_vertical_velocity(delta)

	var input_direction := Input.get_vector("move_left", "move_right", "move_forward", "move_back")
	var forward := -camera_pivot.global_basis.z
	var right := camera_pivot.global_basis.x
	var move_direction := right * input_direction.x + forward * -input_direction.y
	move_direction.y = 0.0
	if move_direction.length_squared() > 1.0:
		move_direction = move_direction.normalized()

	_update_horizontal_velocity(move_direction, delta)
	_update_facing(move_direction, delta)
	_update_locomotion_presentation(move_direction, delta)
	move_and_slide()


func _update_jump_timers(delta: float) -> void:
	if Input.is_action_just_pressed("jump"):
		_jump_buffer_left = jump_buffer_time
	else:
		_jump_buffer_left = maxf(0.0, _jump_buffer_left - delta)

	if is_on_floor():
		_coyote_time_left = coyote_time
		if velocity.y < 0.0:
			velocity.y = 0.0
	else:
		_coyote_time_left = maxf(0.0, _coyote_time_left - delta)

	if _jump_buffer_left > 0.0 and _coyote_time_left > 0.0:
		velocity.y = jump_velocity
		_jump_buffer_left = 0.0
		_coyote_time_left = 0.0


func _update_vertical_velocity(delta: float) -> void:
	if not is_on_floor() or velocity.y > 0.0:
		velocity.y = maxf(
			velocity.y - gravity * gravity_multiplier * delta,
			-max_fall_speed
		)


func _update_horizontal_velocity(move_direction: Vector3, delta: float) -> void:
	var horizontal_velocity := Vector3(velocity.x, 0.0, velocity.z)
	var target_velocity := move_direction * move_speed
	var acceleration := ground_acceleration if is_on_floor() else air_acceleration
	var deceleration := ground_deceleration if is_on_floor() else air_deceleration
	var rate := acceleration if move_direction.length_squared() > 0.0001 else deceleration
	horizontal_velocity = horizontal_velocity.move_toward(target_velocity, rate * delta)
	velocity.x = horizontal_velocity.x
	velocity.z = horizontal_velocity.z


func _update_facing(move_direction: Vector3, delta: float) -> void:
	if move_direction.length_squared() <= 0.0001:
		return

	# Preserve the camera's world yaw while turning the character underneath it.
	var camera_world_rotation := camera_pivot.global_rotation
	var target_yaw := atan2(-move_direction.x, -move_direction.z)
	rotation.y = lerp_angle(rotation.y, target_yaw, clampf(turn_speed * delta, 0.0, 1.0))
	camera_pivot.global_rotation = camera_world_rotation


func _update_locomotion_presentation(move_direction: Vector3, delta: float) -> void:
	var horizontal_speed := Vector2(velocity.x, velocity.z).length()
	var movement_weight := clampf(horizontal_speed / maxf(move_speed, 0.01), 0.0, 1.0)
	_locomotion_clock += delta * lerpf(1.0, 4.0, movement_weight)

	# This deliberately animates only the placeholder mesh, never the collision body.
	# A future rig/AnimationTree can replace this presentation without changing movement.
	var bob := sin(_locomotion_clock * movement_bob_frequency) * movement_bob_height * movement_weight
	var target_height := _body_mesh_base_y + bob
	body_mesh.position.y = lerpf(body_mesh.position.y, target_height, clampf(delta * 14.0, 0.0, 1.0))

	var local_direction := global_basis.inverse() * move_direction
	var target_lean := -local_direction.x * movement_lean_radians * movement_weight
	body_mesh.rotation.z = lerpf(body_mesh.rotation.z, target_lean, clampf(delta * 10.0, 0.0, 1.0))
	var target_pitch := -0.06 if velocity.y > 0.5 else (0.06 if velocity.y < -0.5 else 0.0)
	body_mesh.rotation.x = lerpf(body_mesh.rotation.x, target_pitch, clampf(delta * 8.0, 0.0, 1.0))
