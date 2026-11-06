extends Node3D

const PLAYER_SCENE := preload("res://scenes/player.tscn")

var status_label: Label
var api_request: HTTPRequest
var api_base_url := "http://127.0.0.1:8080"

func _ready() -> void:
	_build_world()
	_build_hud()
	_connect_api()

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
	instructions.text = "WASD move  •  Space jump  •  Mouse look  •  Esc release cursor"
	column.add_child(instructions)

	status_label = Label.new()
	status_label.text = "Server: checking local development API..."
	column.add_child(status_label)
	add_child(layer)

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
