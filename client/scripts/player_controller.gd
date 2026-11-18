extends CharacterBody3D

enum PresentationState {
	IDLE,
	WALK,
	RUN,
	JUMP,
	FALL,
	LAND,
	ATTACK,
	GATHER,
	CRAFT,
	USE_ITEM,
	HIT,
	DEFEAT,
	RESPAWN,
}

@export_group("Movement")
@export var move_speed := 7.0
@export var walk_speed_multiplier := 0.55
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
@export var camera_zoom_step := 0.75
@export var camera_min_distance := 0.5
@export var camera_max_distance := 9.0
@export var attack_swing_radians := 0.95
@export var attack_swing_duration := 0.18

@export_group("Touch Controls")
@export var touch_dpad_exclusion_width := 260.0
@export var touch_dpad_exclusion_height := 250.0
@export var touch_action_exclusion_width := 145.0
@export var touch_action_exclusion_height := 250.0

@export_group("Placeholder Locomotion Presentation")
@export var movement_bob_height := 0.035
@export var movement_bob_frequency := 2.0
@export var movement_lean_radians := 0.075

@onready var camera_pivot: Node3D = $CameraPivot
@onready var camera_arm: SpringArm3D = $CameraPivot/SpringArm3D
@onready var body_mesh: MeshInstance3D = $BodyMesh

var gravity: float = ProjectSettings.get_setting("physics/3d/default_gravity")
var _coyote_time_left := 0.0
var _jump_buffer_left := 0.0
var _body_mesh_base_y := 0.9
var _locomotion_clock := 0.0
var _attack_animation_active := false
var _attack_animation_elapsed := 0.0
var _attack_neutral_y := 0.0
var _active_camera_touch_index := -1
var _was_on_floor := true
var _presentation_state: int = PresentationState.IDLE
var _action_state: int = PresentationState.IDLE
var _action_timer := 0.0
var _action_duration := 0.0
var _land_timer := 0.0
var _is_defeated := false
var _walk_mode := false
var _autorun_enabled := false


func _ready() -> void:
	_body_mesh_base_y = body_mesh.position.y
	if not DisplayServer.is_touchscreen_available():
		Input.mouse_mode = Input.MOUSE_MODE_CAPTURED


func _notification(what: int) -> void:
	if what == NOTIFICATION_APPLICATION_FOCUS_OUT and Input.mouse_mode == Input.MOUSE_MODE_CAPTURED:
		Input.mouse_mode = Input.MOUSE_MODE_VISIBLE


func set_walk_mode(enabled: bool) -> void:
	_walk_mode = enabled


func is_walk_mode() -> bool:
	return _walk_mode


func toggle_walk_mode() -> bool:
	_walk_mode = not _walk_mode
	return _walk_mode


func set_autorun(enabled: bool) -> void:
	_autorun_enabled = enabled


func is_autorun() -> bool:
	return _autorun_enabled


func toggle_autorun() -> bool:
	_autorun_enabled = not _autorun_enabled
	return _autorun_enabled


func get_effective_move_speed() -> float:
	if _walk_mode:
		return move_speed * clampf(walk_speed_multiplier, 0.25, 1.0)
	return move_speed


func _unhandled_input(event: InputEvent) -> void:
	if event is InputEventMouseMotion and Input.mouse_mode == Input.MOUSE_MODE_CAPTURED:
		_rotate_camera(event.relative)
	elif event is InputEventScreenTouch:
		if event.pressed:
			if _active_camera_touch_index == -1 and _is_camera_touch_allowed(event.position):
				_active_camera_touch_index = event.index
		elif event.index == _active_camera_touch_index:
			_active_camera_touch_index = -1
	elif event is InputEventScreenDrag:
		if event.index == _active_camera_touch_index or (_active_camera_touch_index == -1 and _is_camera_touch_allowed(event.position)):
			_rotate_camera(event.relative)
	elif event is InputEventMouseButton and event.pressed:
		if event.button_index == MOUSE_BUTTON_LEFT:
			Input.mouse_mode = Input.MOUSE_MODE_CAPTURED
		elif event.button_index == MOUSE_BUTTON_WHEEL_UP:
			_adjust_camera_zoom(-camera_zoom_step)
		elif event.button_index == MOUSE_BUTTON_WHEEL_DOWN:
			_adjust_camera_zoom(camera_zoom_step)
	elif event is InputEventKey and event.pressed and not event.echo:
		if event.physical_keycode == KEY_EQUAL or event.physical_keycode == KEY_KP_ADD:
			_adjust_camera_zoom(-camera_zoom_step)
		elif event.physical_keycode == KEY_MINUS or event.physical_keycode == KEY_KP_SUBTRACT:
			_adjust_camera_zoom(camera_zoom_step)
		elif event.physical_keycode == KEY_CAPSLOCK:
			toggle_walk_mode()
		elif event.physical_keycode == KEY_R:
			toggle_autorun()
		elif event.physical_keycode == KEY_HOME:
			reset_camera_behind_character()


func _is_camera_touch_allowed(screen_position: Vector2) -> bool:
	var viewport := get_viewport()
	if viewport == null:
		return true
	var viewport_size := viewport.get_visible_rect().size
	if viewport_size.x <= 0.0 or viewport_size.y <= 0.0:
		return true
	var in_dpad_zone := screen_position.x <= touch_dpad_exclusion_width and screen_position.y >= viewport_size.y - touch_dpad_exclusion_height
	var in_action_zone := screen_position.x >= viewport_size.x - touch_action_exclusion_width and screen_position.y >= viewport_size.y - touch_action_exclusion_height
	return not in_dpad_zone and not in_action_zone


func _adjust_camera_zoom(amount: float) -> void:
	camera_arm.spring_length = clampf(
		camera_arm.spring_length + amount,
		camera_min_distance,
		camera_max_distance
	)
	if body_mesh != null:
		body_mesh.visible = camera_arm.spring_length > 1.0


func reset_camera_behind_character() -> void:
	if camera_pivot == null:
		return
	camera_pivot.rotation.y = 0.0
	camera_pivot.rotation.x = deg_to_rad(-18.0)


func get_presentation_state() -> int:
	return _presentation_state


func get_presentation_state_name() -> String:
	return str(PresentationState.keys()[_presentation_state])


func play_attack_animation() -> void:
	_attack_animation_elapsed = 0.0
	_attack_neutral_y = body_mesh.rotation.y
	_attack_animation_active = true
	_start_action_presentation(PresentationState.ATTACK, attack_swing_duration)


func play_gather_animation() -> void:
	_start_action_presentation(PresentationState.GATHER, 0.35)


func play_craft_animation() -> void:
	_start_action_presentation(PresentationState.CRAFT, 0.4)


func play_use_item_animation() -> void:
	_start_action_presentation(PresentationState.USE_ITEM, 0.25)


func play_hit_reaction() -> void:
	if _is_defeated:
		return
	_start_action_presentation(PresentationState.HIT, 0.16)


func play_defeat_animation() -> void:
	_is_defeated = true
	_autorun_enabled = false
	_attack_animation_active = false
	_action_state = PresentationState.DEFEAT
	_action_timer = 0.0
	_action_duration = 0.0
	_presentation_state = PresentationState.DEFEAT


func play_respawn_animation() -> void:
	_is_defeated = false
	_attack_animation_active = false
	_start_action_presentation(PresentationState.RESPAWN, 0.3)


func _start_action_presentation(state: int, duration: float) -> void:
	if _is_defeated and state != PresentationState.RESPAWN:
		return
	_action_state = state
	_action_duration = maxf(duration, 0.01)
	_action_timer = _action_duration
	_presentation_state = state


func _update_attack_animation(delta: float) -> void:
	if not _attack_animation_active:
		return
	_attack_animation_elapsed += delta
	var duration := maxf(attack_swing_duration, 0.01)
	var progress := clampf(_attack_animation_elapsed / duration, 0.0, 1.0)
	# A smooth single-arc swing keeps animation deterministic and allocation-free.
	body_mesh.rotation.y = _attack_neutral_y + sin(progress * PI) * attack_swing_radians
	if progress >= 1.0:
		body_mesh.rotation.y = _attack_neutral_y
		_attack_animation_active = false


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
	if input_direction.length_squared() > 0.0001 and _autorun_enabled:
		_autorun_enabled = false
	if _autorun_enabled and input_direction.length_squared() <= 0.0001:
		input_direction = Vector2(0.0, -1.0)
	var forward := -camera_pivot.global_basis.z
	var right := camera_pivot.global_basis.x
	var move_direction := right * input_direction.x + forward * -input_direction.y
	move_direction.y = 0.0
	if move_direction.length_squared() > 1.0:
		move_direction = move_direction.normalized()

	_update_horizontal_velocity(move_direction, delta)
	_update_facing(move_direction, delta)
	move_and_slide()
	_update_presentation_state(delta)
	_update_locomotion_presentation(move_direction, delta)
	_update_attack_animation(delta)


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
	var target_velocity := move_direction * get_effective_move_speed()
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


func _update_presentation_state(delta: float) -> void:
	var on_floor := is_on_floor()
	if on_floor and not _was_on_floor and not _is_defeated:
		_land_timer = 0.12
	_was_on_floor = on_floor

	if _land_timer > 0.0:
		_land_timer = maxf(0.0, _land_timer - delta)

	if _is_defeated:
		_presentation_state = PresentationState.DEFEAT
		return

	if _action_timer > 0.0:
		_action_timer = maxf(0.0, _action_timer - delta)
		if _action_timer > 0.0:
			_presentation_state = _action_state
			return
		_action_state = PresentationState.IDLE

	if not on_floor:
		_presentation_state = PresentationState.JUMP if velocity.y >= 0.0 else PresentationState.FALL
		return

	if _land_timer > 0.0:
		_presentation_state = PresentationState.LAND
		return

	var horizontal_speed := Vector2(velocity.x, velocity.z).length()
	if horizontal_speed >= move_speed * 0.65:
		_presentation_state = PresentationState.RUN
	elif horizontal_speed > 0.25:
		_presentation_state = PresentationState.WALK
	else:
		_presentation_state = PresentationState.IDLE


func _update_locomotion_presentation(move_direction: Vector3, delta: float) -> void:
	var horizontal_speed := Vector2(velocity.x, velocity.z).length()
	var movement_weight := clampf(horizontal_speed / maxf(move_speed, 0.01), 0.0, 1.0)
	_locomotion_clock += delta * lerpf(1.0, 4.0, movement_weight)

	# This deliberately animates only the placeholder mesh, never the collision body.
	# A future rig/AnimationTree can replace this presentation without changing movement.
	var bob := sin(_locomotion_clock * movement_bob_frequency) * movement_bob_height * movement_weight
	var target_height := _body_mesh_base_y + bob
	var target_scale := Vector3.ONE
	var local_direction := global_basis.inverse() * move_direction
	var target_lean := -local_direction.x * movement_lean_radians * movement_weight
	var target_pitch := -0.06 if velocity.y > 0.5 else (0.06 if velocity.y < -0.5 else 0.0)

	match _presentation_state:
		PresentationState.LAND:
			target_scale = Vector3(1.06, 0.9, 1.06)
			target_height = _body_mesh_base_y - 0.05
		PresentationState.GATHER, PresentationState.CRAFT:
			var progress := 1.0 - (_action_timer / maxf(_action_duration, 0.01))
			target_pitch = -0.25 * sin(progress * PI)
		PresentationState.USE_ITEM:
			target_scale = Vector3(1.04, 1.04, 1.04)
		PresentationState.HIT:
			target_pitch = 0.18
		PresentationState.DEFEAT:
			target_pitch = -1.35
			target_height = _body_mesh_base_y - 0.55
		PresentationState.RESPAWN:
			target_scale = Vector3(1.05, 1.05, 1.05)

	body_mesh.position.y = lerpf(body_mesh.position.y, target_height, clampf(delta * 14.0, 0.0, 1.0))
	body_mesh.rotation.z = lerpf(body_mesh.rotation.z, target_lean, clampf(delta * 10.0, 0.0, 1.0))
	body_mesh.rotation.x = lerpf(body_mesh.rotation.x, target_pitch, clampf(delta * 12.0, 0.0, 1.0))
	body_mesh.scale = body_mesh.scale.lerp(target_scale, clampf(delta * 14.0, 0.0, 1.0))
