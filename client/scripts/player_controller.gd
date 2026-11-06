extends CharacterBody3D

@export var move_speed := 7.0
@export var jump_velocity := 5.0
@export var mouse_sensitivity := 0.0025

@onready var camera_pivot: Node3D = $CameraPivot

var gravity: float = ProjectSettings.get_setting("physics/3d/default_gravity")

func _ready() -> void:
	Input.mouse_mode = Input.MOUSE_MODE_CAPTURED

func _unhandled_input(event: InputEvent) -> void:
	if event is InputEventMouseMotion and Input.mouse_mode == Input.MOUSE_MODE_CAPTURED:
		camera_pivot.rotate_y(-event.relative.x * mouse_sensitivity)
		camera_pivot.rotation.x = clampf(
			camera_pivot.rotation.x - event.relative.y * mouse_sensitivity,
			deg_to_rad(-65.0),
			deg_to_rad(35.0)
		)
	elif event is InputEventMouseButton and event.pressed:
		Input.mouse_mode = Input.MOUSE_MODE_CAPTURED

func _physics_process(delta: float) -> void:
	if not is_on_floor():
		velocity.y -= gravity * delta

	if Input.is_key_pressed(KEY_SPACE) and is_on_floor():
		velocity.y = jump_velocity

	var input_direction := Vector2.ZERO
	if Input.is_key_pressed(KEY_A):
		input_direction.x -= 1.0
	if Input.is_key_pressed(KEY_D):
		input_direction.x += 1.0
	if Input.is_key_pressed(KEY_W):
		input_direction.y -= 1.0
	if Input.is_key_pressed(KEY_S):
		input_direction.y += 1.0
	input_direction = input_direction.normalized()

	var forward := -camera_pivot.global_basis.z
	var right := camera_pivot.global_basis.x
	var direction := right * input_direction.x + forward * -input_direction.y
	direction.y = 0.0
	direction = direction.normalized()

	velocity.x = direction.x * move_speed
	velocity.z = direction.z * move_speed


	move_and_slide()
