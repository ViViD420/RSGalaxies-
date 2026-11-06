extends Node3D

const PLAYER_SCENE := preload("res://scenes/player.tscn")

var status_label: Label
var api_request: HTTPRequest
var api_base_url := "http://127.0.0.1:8080"

func _ready() -> void:
	_setup_input_actions()
	_build_world()
	_build_hud()
	_connect_api()

func _setup_input_actions() -> void:
	_register_key_action("move_forward", KEY_W)
	_register_key_action("move_back", KEY_S)
	_register_key_action("move_left", KEY_A)
	_register_key_action("move_right", KEY_D)
	_register_key_action("jump", KEY_SPACE)

func _register_key_action(action_name: String, key: Key) -> void:
	if not InputMap.has_action(action_name):
		InputMap.add_action(action_name)
	var event := InputEventKey.new()
	event.physical_keycode = key
	InputMap.action_add_event(action_name, event)

func _build_world() -> void:
	var environment := WorldEnvironment.new()
	var world_environment := Environment.new()
	world_environment.background_mode = Environment.BG_COLOR
	world_environment.background_color = Color(0.035, 0.055, 0.075, 1)
	world_environment.ambient_light_source = Environment.AMBIENT_SOURCE_COLOR
	world_environment.ambient_light_color = Color(0.48, 0.56, 0.68)
	world_environment.ambient_light_energy = 0.65
	environment.environment = world_environment
	add_child(environment)

	var sun := DirectionalLight3D.new()
	sun.rotation_degrees = Vector3(-52.0, -28.0, 0.0)
	sun.light_energy = 1.2
	add_child(sun)

	var floor_body := StaticBody3D.new()
	floor_body.name = "WorldFloor"
	var floor_mesh := MeshInstance3D.new()
	var plane := PlaneMesh.new()
	plane.size = Vector2(240.0, 240.0)
	floor_mesh.mesh = plane
	var floor_material := StandardMaterial3D.new()
	floor_material.albedo_color = Color(0.18, 0.23, 0.20)
	floor_material.roughness = 1.0
	floor_mesh.material_override = floor_material
	floor_body.add_child(floor_mesh)

	var floor_collision := CollisionShape3D.new()
	var floor_shape := BoxShape3D.new()
	floor_shape.size = Vector3(240.0, 1.0, 240.0)
	floor_collision.shape = floor_shape
	floor_collision.position.y = -0.5
	floor_body.add_child(floor_collision)
	add_child(floor_body)

	var player := PLAYER_SCENE.instantiate()
	player.position = Vector3(0.0, 0.05, 0.0)
	add_child(player)

func _build_hud() -> void:
	var layer := CanvasLayer.new()
	var panel := PanelContainer.new()
	panel.position = Vector2(18, 18)
	panel.custom_minimum_size = Vector2(360, 90)
	layer.add_child(panel)

	var column := VBoxContainer.new()
	panel.add_child(column)
	var heading := Label.new()
	heading.text = "RS GALAXIES  /  WORLD PROTOTYPE"
	heading.add_theme_font_size_override("font_size", 18)
	column.add_child(heading)

	var instructions := Label.new()
	instructions.text = "WASD / touch pad: move  •  Space / jump: jump  •  Drag / mouse: look"
	column.add_child(instructions)

	status_label = Label.new()
	status_label.text = "Server: checking local development API..."
	column.add_child(status_label)

	if DisplayServer.is_touchscreen_available():
		_add_hold_button(layer, "▲", "move_forward", 95, -230, 165, -160)
		_add_hold_button(layer, "◀", "move_left", 20, -155, 90, -85)
		_add_hold_button(layer, "▼", "move_back", 95, -155, 165, -85)
		_add_hold_button(layer, "▶", "move_right", 170, -155, 240, -85)
		_add_hold_button(layer, "JUMP", "jump", -125, -120, -25, -25, true)

	add_child(layer)

func _add_hold_button(parent: Control, caption: String, action: String, left: float, top: float, right: float, bottom: float, anchor_right: bool = false) -> void:
	var button := Button.new()
	button.text = caption
	button.custom_minimum_size = Vector2(70, 70)
	if anchor_right:
		button.anchor_left = 1.0
		button.anchor_right = 1.0
		button.anchor_top = 1.0
		button.anchor_bottom = 1.0
	else:
		button.anchor_top = 1.0
		button.anchor_bottom = 1.0
	button.offset_left = left
	button.offset_top = top
	button.offset_right = right
	button.offset_bottom = bottom
	button.button_down.connect(func() -> void: Input.action_press(action))
	button.button_up.connect(func() -> void: Input.action_release(action))
	parent.add_child(button)

func _connect_api() -> void:
	api_request = HTTPRequest.new()
	add_child(api_request)
	api_request.request_completed.connect(_on_api_request_completed)
	var error := api_request.request(api_base_url + "/healthz")
	if error != OK:
		status_label.text = "Server: unavailable (start the Go service on port 8080)"

func _on_api_request_completed(result: int, response_code: int, _headers: PackedStringArray, body: PackedByteArray) -> void:
	if result != HTTPRequest.RESULT_SUCCESS or response_code != 200:
		status_label.text = "Server: offline — run the Go world service on port 8080"
		return
	var payload: Variant = JSON.parse_string(body.get_string_from_utf8())
	if payload is Dictionary and payload.get("status") == "ok":
		status_label.text = "Server: connected to RS Galaxies world service"
	else:
		status_label.text = "Server: response received; health payload was unexpected"

func _unhandled_input(event: InputEvent) -> void:
	if event is InputEventKey and event.pressed and event.keycode == KEY_ESCAPE:
		Input.mouse_mode = Input.MOUSE_MODE_VISIBLE
