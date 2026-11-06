extends Node3D

const PLAYER_SCENE := preload("res://scenes/player.tscn")

var status_label: Label
var inventory_label: Label
var api_request: HTTPRequest
var movement_request: HTTPRequest
var gather_request: HTTPRequest
var movement_timer: Timer
var player_node: CharacterBody3D
var player_id := ""
var api_stage := "health"
var movement_pending := false
var interaction_pending := false
var interaction_type := "gather"
var api_base_url := "http://127.0.0.1:8080"

func _ready() -> void:
	_setup_input_actions()
	_build_world()
	_build_hud()
	_connect_api()

func _process(_delta: float) -> void:
	if Input.is_action_just_pressed("interact") and not player_id.is_empty() and not interaction_pending:
		_request_interaction()

func _setup_input_actions() -> void:
	_register_key_action("move_forward", KEY_W)
	_register_key_action("move_back", KEY_S)
	_register_key_action("move_left", KEY_A)
	_register_key_action("move_right", KEY_D)
	_register_key_action("jump", KEY_SPACE)
	_register_key_action("interact", KEY_E)

func _register_key_action(action_name: String, key: int) -> void:
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

	_add_tree(Vector3(2.0, 0.0, 0.0))
	_add_stone_deposit(Vector3(-2.0, 0.0, 0.0))
	_add_workbench(Vector3(0.0, 0.0, 4.0))

	player_node = PLAYER_SCENE.instantiate() as CharacterBody3D
	player_node.position = Vector3(0.0, 0.05, 0.0)
	add_child(player_node)

func _add_tree(at: Vector3) -> void:
	var tree := Node3D.new()
	tree.position = at
	var trunk := MeshInstance3D.new()
	var trunk_mesh := CylinderMesh.new()
	trunk_mesh.top_radius = 0.22
	trunk_mesh.bottom_radius = 0.35
	trunk_mesh.height = 2.2
	trunk.mesh = trunk_mesh
	trunk.position.y = 1.1
	var trunk_material := StandardMaterial3D.new()
	trunk_material.albedo_color = Color(0.25, 0.15, 0.08)
	trunk.material_override = trunk_material
	tree.add_child(trunk)

	var canopy := MeshInstance3D.new()
	var canopy_mesh := SphereMesh.new()
	canopy_mesh.radius = 1.25
	canopy_mesh.height = 2.5
	canopy.mesh = canopy_mesh
	canopy.position.y = 2.7
	var canopy_material := StandardMaterial3D.new()
	canopy_material.albedo_color = Color(0.12, 0.32, 0.16)
	canopy.material_override = canopy_material
	tree.add_child(canopy)
	add_child(tree)

func _add_workbench(at: Vector3) -> void:
	var bench := Node3D.new()
	bench.position = at
	var wood_material := StandardMaterial3D.new()
	wood_material.albedo_color = Color(0.30, 0.19, 0.10)
	var top := MeshInstance3D.new()
	var top_mesh := BoxMesh.new()
	top_mesh.size = Vector3(2.4, 0.25, 1.2)
	top.mesh = top_mesh
	top.position.y = 1.05
	top.material_override = wood_material
	bench.add_child(top)
	for x_offset in [-0.95, 0.95]:
		for z_offset in [-0.42, 0.42]:
			var leg := MeshInstance3D.new()
			var leg_mesh := BoxMesh.new()
			leg_mesh.size = Vector3(0.18, 1.0, 0.18)
			leg.mesh = leg_mesh
			leg.position = Vector3(x_offset, 0.5, z_offset)
			leg.material_override = wood_material
			bench.add_child(leg)
	add_child(bench)

func _add_stone_deposit(at: Vector3) -> void:
	var rock := MeshInstance3D.new()
	var rock_mesh := SphereMesh.new()
	rock_mesh.radius = 1.0
	rock_mesh.height = 1.5
	rock.mesh = rock_mesh
	rock.position = at + Vector3(0.0, 0.55, 0.0)
	rock.scale = Vector3(1.3, 0.75, 1.0)
	var material := StandardMaterial3D.new()
	material.albedo_color = Color(0.34, 0.36, 0.38)
	material.roughness = 1.0
	rock.material_override = material
	add_child(rock)

func _build_hud() -> void:
	var layer := CanvasLayer.new()
	var panel := PanelContainer.new()
	panel.position = Vector2(18, 18)
	panel.custom_minimum_size = Vector2(390, 120)
	layer.add_child(panel)

	var column := VBoxContainer.new()
	panel.add_child(column)
	var heading := Label.new()
	heading.text = "RS GALAXIES  /  WORLD PROTOTYPE"
	heading.add_theme_font_size_override("font_size", 18)
	column.add_child(heading)

	var instructions := Label.new()
	instructions.text = "WASD / pad: move  •  E / USE: gather or craft  •  Drag / mouse: look"
	column.add_child(instructions)

	status_label = Label.new()
	status_label.text = "Server: connecting..."
	column.add_child(status_label)

	inventory_label = Label.new()
	inventory_label.text = "Inventory: empty"
	column.add_child(inventory_label)

	if DisplayServer.is_touchscreen_available():
		_add_hold_button(layer, "▲", "move_forward", 95, -230, 165, -160)
		_add_hold_button(layer, "◀", "move_left", 20, -155, 90, -85)
		_add_hold_button(layer, "▼", "move_back", 95, -155, 165, -85)
		_add_hold_button(layer, "▶", "move_right", 170, -155, 240, -85)
		_add_hold_button(layer, "USE", "interact", -125, -230, -25, -130, true)
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
	if OS.has_feature("web"):
		var forwarded_api_origin: Variant = JavaScriptBridge.eval("window.location.origin.replace('-8000.', '-8080.')")
		if forwarded_api_origin is String and not forwarded_api_origin.is_empty():
			api_base_url = forwarded_api_origin
	api_request = HTTPRequest.new()
	add_child(api_request)
	api_request.request_completed.connect(_on_api_request_completed)
	var error := api_request.request(api_base_url + "/healthz")
	if error != OK:
		status_label.text = "Server: unavailable (start the Go world service)"

func _on_api_request_completed(result: int, response_code: int, _headers: PackedStringArray, body: PackedByteArray) -> void:
	if result != HTTPRequest.RESULT_SUCCESS or response_code != 200:
		status_label.text = "Server: unavailable — world service did not respond"
		return
	var payload: Variant = JSON.parse_string(body.get_string_from_utf8())
	if not payload is Dictionary:
		status_label.text = "Server: invalid response"
		return

	if api_stage == "health":
		if payload.get("status") != "ok":
			status_label.text = "Server: unexpected health response"
			return
		api_stage = "create_player"
		var error := api_request.request(
			api_base_url + "/api/v1/players",
			PackedStringArray(["Content-Type: application/json"]),
			HTTPClient.METHOD_POST,
			JSON.stringify({"name": "Wayfarer"})
		)
		if error != OK:
			status_label.text = "Server: could not create development character"
	elif api_stage == "create_player":
		player_id = str(payload.get("id", ""))
		if player_id.is_empty():
			status_label.text = "Server: character creation returned no ID"
			return
		var spawn: Variant = payload.get("position", {})
		if spawn is Dictionary:
			_apply_server_position(spawn, true)
		status_label.text = "Server: connected • character ready"
		_update_inventory(payload.get("inventory", {}))
		_start_movement_sync()
		_start_gathering()

func _start_movement_sync() -> void:
	movement_request = HTTPRequest.new()
	add_child(movement_request)
	movement_request.request_completed.connect(_on_movement_completed)
	movement_timer = Timer.new()
	movement_timer.wait_time = 0.2
	movement_timer.timeout.connect(_sync_player_position)
	add_child(movement_timer)
	movement_timer.start()

func _sync_player_position() -> void:
	if player_id.is_empty() or movement_pending:
		return
	movement_pending = true
	var body := JSON.stringify({
		"x": player_node.global_position.x,
		"z": player_node.global_position.z
	})
	var error := movement_request.request(
		api_base_url + "/api/v1/players/" + player_id + "/move",
		PackedStringArray(["Content-Type: application/json"]),
		HTTPClient.METHOD_POST,
		body
	)
	if error != OK:
		movement_pending = false

func _on_movement_completed(result: int, response_code: int, _headers: PackedStringArray, body: PackedByteArray) -> void:
	movement_pending = false
	if result != HTTPRequest.RESULT_SUCCESS:
		return
	var payload: Variant = JSON.parse_string(body.get_string_from_utf8())
	if not payload is Dictionary:
		return
	var authoritative_position: Variant = payload.get("position", {})
	if authoritative_position is Dictionary:
		_apply_server_position(authoritative_position, response_code != 200)
	if response_code != 200:
		status_label.text = "Server: movement corrected by world rules"

func _start_gathering() -> void:
	gather_request = HTTPRequest.new()
	add_child(gather_request)
	gather_request.request_completed.connect(_on_gather_completed)

func _request_interaction() -> void:
	var tree_distance := Vector2(player_node.global_position.x - 2.0, player_node.global_position.z).length()
	var stone_distance := Vector2(player_node.global_position.x + 2.0, player_node.global_position.z).length()
	var bench_distance := Vector2(player_node.global_position.x, player_node.global_position.z - 4.0).length()
	var error := OK
	if bench_distance <= 3.0 and bench_distance <= tree_distance and bench_distance <= stone_distance:
		interaction_type = "craft"
		status_label.text = "Using frontier workbench..."
		error = gather_request.request(
			api_base_url + "/api/v1/players/" + player_id + "/craft/basic-plank",
			PackedStringArray(),
			HTTPClient.METHOD_POST,
			""
		)
	else:
		var node_id := "tree-001" if tree_distance <= stone_distance else "stone-001"
		var nearest_distance := minf(tree_distance, stone_distance)
		if nearest_distance > 3.0:
			status_label.text = "Move closer to a resource node or workbench"
			return
		interaction_type = "gather"
		error = gather_request.request(
			api_base_url + "/api/v1/players/" + player_id + "/gather/" + node_id,
			PackedStringArray(),
			HTTPClient.METHOD_POST,
			""
		)
	if error == OK:
		interaction_pending = true

func _on_gather_completed(result: int, response_code: int, _headers: PackedStringArray, body: PackedByteArray) -> void:
	interaction_pending = false
	if result != HTTPRequest.RESULT_SUCCESS:
		status_label.text = "Gathering: server connection lost"
		return
	var payload: Variant = JSON.parse_string(body.get_string_from_utf8())
	if not payload is Dictionary:
		return
	if response_code != 200:
		var error_code := str(payload.get("error", "gathering_failed"))
		status_label.text = "Interaction: " + error_code.replace("_", " ")
		return
	var player_data: Dictionary = payload.get("player", {})
	if interaction_type == "craft":
		var crafting: Dictionary = payload.get("crafting", {})
		status_label.text = "Crafted Wooden Plank  •  +" + str(crafting.get("experience", 0)) + " Crafting XP"
	else:
		var gathering: Dictionary = payload.get("gathering", {})
		var item_name := "Softwood Log" if gathering.get("itemId") == "softwood_log" else "Rough Stone"
		status_label.text = "Gathered " + item_name + "  •  +" + str(gathering.get("experience", 0)) + " " + str(gathering.get("skillId", "")) + " XP"
	_update_inventory(player_data.get("inventory", {}))

func _update_inventory(inventory_data: Variant) -> void:
	if not inventory_data is Dictionary:
		return
	inventory_label.text = "Inventory: Logs %d  •  Planks %d  •  Stone %d" % [
		int(inventory_data.get("softwood_log", 0)),
		int(inventory_data.get("wooden_plank", 0)),
		int(inventory_data.get("rough_stone", 0))
	]

func _apply_server_position(position_data: Dictionary, force: bool) -> void:
	var target := Vector3(
		float(position_data.get("x", 0.0)),
		player_node.global_position.y,
		float(position_data.get("z", 0.0))
	)
	if force or player_node.global_position.distance_to(target) > 2.5:
		player_node.global_position = target
	else:
		player_node.global_position = player_node.global_position.lerp(target, 0.08)

func _unhandled_input(event: InputEvent) -> void:
	if event is InputEventKey and event.pressed and event.keycode == KEY_ESCAPE:
		Input.mouse_mode = Input.MOUSE_MODE_VISIBLE
