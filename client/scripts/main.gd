extends Node3D

const PLAYER_SCENE := preload("res://scenes/player.tscn")

var status_label: Label
var presence_label: Label
var inventory_label: Label
var auth_panel: PanelContainer
var auth_email_input: LineEdit
var auth_password_input: LineEdit
var auth_status_label: Label
var character_status_label: Label
var logout_button: Button
var character_button: Button
var equipment_button: Button
var rest_button: Button
var bag_button: Button
var skills_button: Button
var skills_panel: PanelContainer
var skills_list: VBoxContainer
var skills_status_label: Label
var skills_request: HTTPRequest
var skills_pending := false
var inventory_panel: PanelContainer
var inventory_use_select: OptionButton
var inventory_search_input: LineEdit
var inventory_filter_select: OptionButton
var inventory_sort_select: OptionButton
var inventory_items_column: VBoxContainer
var inventory_search_text := ""
var inventory_filter := "all"
var inventory_sort := "name"
var inventory_bank_select: OptionButton
var inventory_quantity_input: LineEdit
var inventory_status_label: Label
var equipment_panel: PanelContainer
var equipment_item_select: OptionButton
var equipment_slot_select: OptionButton
var equipment_status_label: Label
var character_panel: PanelContainer
var character_select: OptionButton
var character_name_input: LineEdit
var market_panel: PanelContainer
var market_item_select: OptionButton
var market_quantity_input: LineEdit
var market_status_label: Label
var auth_token := ""
var current_inventory: Dictionary = {}
var current_bank: Dictionary = {}
var current_coins := 0
var current_equipment: Dictionary = {}
var current_attack_bonus := 0
var current_defense_bonus := 0
var current_health := 100
var chat_panel: PanelContainer
var chat_history: RichTextLabel
var chat_input: LineEdit
var chat_status_label: Label
var chat_poll_request: HTTPRequest
var chat_send_request: HTTPRequest
var chat_poll_timer: Timer
var chat_last_id := 0
var chat_poll_pending := false
var chat_send_pending := false
var chat_active_character_id := ""
var chat_lines: Array[String] = []
var presence_heartbeat_request: HTTPRequest
var presence_roster_request: HTTPRequest
var presence_heartbeat_timer: Timer
var presence_roster_timer: Timer
var presence_heartbeat_pending := false
var presence_roster_pending := false
var remote_player_nodes: Dictionary = {}
var remote_player_targets: Dictionary = {}
const game_max_health := 100
var market_catalog: Array = []
var enemy_catalog: Array = []
var gathering_catalog: Array = []
var api_request: HTTPRequest
var movement_request: HTTPRequest
var gather_request: HTTPRequest
var movement_timer: Timer
var player_node: CharacterBody3D
var equipment_visual_root: Node3D
var player_id := ""
var api_stage := "health"
var auth_mode := "login"
var movement_pending := false
var interaction_pending := false
var interaction_type := "gather"
var api_base_url := "http://127.0.0.1:8080"

func _new_api_request(timeout_seconds: float = 8.0) -> HTTPRequest:
	var request := HTTPRequest.new()
	request.timeout = timeout_seconds
	# Authenticated API requests must not forward bearer credentials through redirects.
	request.max_redirects = 0
	return request


func _ready() -> void:
	_setup_input_actions()
	_build_world()
	_build_hud()
	_connect_api()

func _process(delta: float) -> void:
	_update_remote_player_visuals(delta)
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

	_add_workbench(Vector3(0.0, 0.0, 4.0))
	_add_bank_chest(Vector3(0.0, 0.0, -4.0))
	_add_merchant_stall(Vector3(4.0, 0.0, 0.0))
	_add_forge(Vector3(0.0, 0.0, 8.0))

	player_node = PLAYER_SCENE.instantiate() as CharacterBody3D
	player_node.position = Vector3(0.0, 0.05, 0.0)
	equipment_visual_root = Node3D.new()
	equipment_visual_root.name = "EquipmentVisuals"
	player_node.add_child(equipment_visual_root)
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





func _build_gathering_node_visuals() -> void:
	for node in gathering_catalog:
		if not node is Dictionary:
			continue
		var position_data: Variant = node.get("position", {})
		if not position_data is Dictionary:
			continue
		var at := Vector3(float(position_data.get("x", 0.0)), 0.0, float(position_data.get("z", 0.0)))
		match str(node.get("resourceId", "")):
			"softwood_log":
				_add_tree(at)
			"rough_stone":
				_add_stone_deposit(at)
			"copper_ore":
				_add_copper_vein(at)
			_:
				_add_generic_gathering_node(at, str(node.get("name", "Resource Node")))

func _add_generic_gathering_node(at: Vector3, node_name: String) -> void:
	var root := Node3D.new()
	root.position = at
	var rock := MeshInstance3D.new()
	var mesh := SphereMesh.new()
	mesh.radius = 0.6
	mesh.height = 0.9
	rock.mesh = mesh
	rock.position.y = 0.45
	var material := StandardMaterial3D.new()
	material.albedo_color = Color(0.42, 0.40, 0.30)
	rock.material_override = material
	root.add_child(rock)
	var label := Label3D.new()
	label.text = node_name
	label.position.y = 1.15
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	root.add_child(label)
	add_child(root)

func _build_enemy_spawns() -> void:
	for enemy in enemy_catalog:
		if not enemy is Dictionary:
			continue
		var position_data: Variant = enemy.get("position", {})
		if not position_data is Dictionary:
			continue
		_add_enemy_visual(
			Vector3(float(position_data.get("x", 0.0)), 0.0, float(position_data.get("z", 0.0))),
			str(enemy.get("name", "Frontier Creature")),
			int(enemy.get("maxHealth", 10))
		)

func _add_enemy_visual(at: Vector3, enemy_name: String, max_health: int) -> void:
	var creature := Node3D.new()
	creature.position = at
	var hide_material := StandardMaterial3D.new()
	hide_material.albedo_color = Color(0.48, 0.30, 0.18) if max_health < 20 else Color(0.32, 0.25, 0.18)
	var body := MeshInstance3D.new()
	var body_mesh := CapsuleMesh.new()
	body_mesh.radius = 0.42 if max_health < 20 else 0.7
	body_mesh.height = 0.9 if max_health < 20 else 1.5
	body.mesh = body_mesh
	body.position.y = body_mesh.height * 0.5
	body.material_override = hide_material
	creature.add_child(body)
	var head := MeshInstance3D.new()
	var head_mesh := SphereMesh.new()
	head_mesh.radius = 0.28 if max_health < 20 else 0.42
	head_mesh.height = head_mesh.radius * 2.0
	head.mesh = head_mesh
	head.position = Vector3(0.0, body_mesh.height * 0.75, -0.35)
	head.material_override = hide_material
	creature.add_child(head)
	var label := Label3D.new()
	label.text = enemy_name
	label.position.y = body_mesh.height + 0.45
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 32
	creature.add_child(label)
	add_child(creature)

func _add_copper_vein(at: Vector3) -> void:
	var deposit := Node3D.new()
	deposit.position = at
	var rock_material := StandardMaterial3D.new()
	rock_material.albedo_color = Color(0.30, 0.32, 0.30)
	var copper_material := StandardMaterial3D.new()
	copper_material.albedo_color = Color(0.70, 0.34, 0.16)
	for index in range(4):
		var rock := MeshInstance3D.new()
		var mesh := SphereMesh.new()
		mesh.radius = 0.45 + float(index % 2) * 0.12
		mesh.height = mesh.radius * 1.6
		rock.mesh = mesh
		rock.position = Vector3(float(index % 2) * 0.55 - 0.25, 0.32, float(index / 2) * 0.5 - 0.25)
		rock.material_override = rock_material
		deposit.add_child(rock)
	var ore := MeshInstance3D.new()
	var ore_mesh := SphereMesh.new()
	ore_mesh.radius = 0.23
	ore_mesh.height = 0.5
	ore.mesh = ore_mesh
	ore.position = Vector3(0.0, 0.55, -0.2)
	ore.material_override = copper_material
	deposit.add_child(ore)
	add_child(deposit)

func _add_forge(at: Vector3) -> void:
	var forge := Node3D.new()
	forge.position = at
	var stone := StandardMaterial3D.new()
	stone.albedo_color = Color(0.22, 0.24, 0.25)
	var ember := StandardMaterial3D.new()
	ember.albedo_color = Color(0.92, 0.30, 0.08)
	var base := MeshInstance3D.new()
	var base_mesh := BoxMesh.new()
	base_mesh.size = Vector3(2.0, 1.1, 1.6)
	base.mesh = base_mesh
	base.position.y = 0.55
	base.material_override = stone
	forge.add_child(base)
	var chimney := MeshInstance3D.new()
	var chimney_mesh := BoxMesh.new()
	chimney_mesh.size = Vector3(0.7, 1.8, 0.7)
	chimney.mesh = chimney_mesh
	chimney.position = Vector3(0.45, 1.8, -0.2)
	chimney.material_override = stone
	forge.add_child(chimney)
	var fire := MeshInstance3D.new()
	var fire_mesh := SphereMesh.new()
	fire_mesh.radius = 0.3
	fire_mesh.height = 0.55
	fire.mesh = fire_mesh
	fire.position = Vector3(0.0, 0.7, 0.84)
	fire.material_override = ember
	forge.add_child(fire)
	add_child(forge)

func _add_merchant_stall(at: Vector3) -> void:
	var stall := Node3D.new()
	stall.position = at
	var wood := StandardMaterial3D.new()
	wood.albedo_color = Color(0.30, 0.19, 0.10)
	var fabric := StandardMaterial3D.new()
	var leather_material := StandardMaterial3D.new()
	leather_material.albedo_color = Color(0.62, 0.45, 0.30)
	fabric.albedo_color = Color(0.58, 0.32, 0.16)
	var counter := MeshInstance3D.new()
	var counter_mesh := BoxMesh.new()
	counter_mesh.size = Vector3(2.2, 1.0, 0.9)
	counter.mesh = counter_mesh
	counter.position.y = 0.5
	counter.material_override = wood
	stall.add_child(counter)
	var canopy := MeshInstance3D.new()
	var canopy_mesh := BoxMesh.new()
	canopy_mesh.size = Vector3(2.8, 0.18, 1.8)
	canopy.mesh = canopy_mesh
	canopy.position.y = 2.4
	canopy.material_override = fabric
	stall.add_child(canopy)
	for x_offset in [-1.2, 1.2]:
		for z_offset in [-0.7, 0.7]:
			var post := MeshInstance3D.new()
			var post_mesh := CylinderMesh.new()
			post_mesh.top_radius = 0.07
			post_mesh.bottom_radius = 0.07
			post_mesh.height = 2.3
			post.mesh = post_mesh
			post.position = Vector3(x_offset, 1.15, z_offset)
			post.material_override = wood
			stall.add_child(post)
	var vendor_material := StandardMaterial3D.new()
	vendor_material.albedo_color = Color(0.22, 0.38, 0.30)
	var vendor_body := MeshInstance3D.new()
	var vendor_body_mesh := CapsuleMesh.new()
	vendor_body_mesh.radius = 0.30
	vendor_body_mesh.height = 1.25
	vendor_body.mesh = vendor_body_mesh
	vendor_body.position = Vector3(0.0, 0.62, 1.15)
	vendor_body.material_override = vendor_material
	stall.add_child(vendor_body)
	var vendor_head := MeshInstance3D.new()
	var vendor_head_mesh := SphereMesh.new()
	vendor_head_mesh.radius = 0.23
	vendor_head_mesh.height = 0.46
	vendor_head.mesh = vendor_head_mesh
	vendor_head.position = Vector3(0.0, 1.42, 1.15)
	vendor_head.material_override = leather_material
	stall.add_child(vendor_head)
	var vendor_label := Label3D.new()
	vendor_label.text = "Frontier Trader"
	vendor_label.position = Vector3(0.0, 2.0, 1.15)
	vendor_label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	vendor_label.font_size = 28
	stall.add_child(vendor_label)
	add_child(stall)

func _add_bank_chest(at: Vector3) -> void:
	var chest := Node3D.new()
	chest.position = at
	var wood := StandardMaterial3D.new()
	wood.albedo_color = Color(0.28, 0.16, 0.08)
	var lid_material := StandardMaterial3D.new()
	lid_material.albedo_color = Color(0.48, 0.31, 0.12)
	var body := MeshInstance3D.new()
	var body_mesh := BoxMesh.new()
	body_mesh.size = Vector3(1.6, 0.9, 1.0)
	body.mesh = body_mesh
	body.position.y = 0.48
	body.material_override = wood
	chest.add_child(body)
	var lid := MeshInstance3D.new()
	var lid_mesh := BoxMesh.new()
	lid_mesh.size = Vector3(1.7, 0.22, 1.08)
	lid.mesh = lid_mesh
	lid.position.y = 1.05
	lid.material_override = lid_material
	chest.add_child(lid)
	var clasp := MeshInstance3D.new()
	var clasp_mesh := BoxMesh.new()
	clasp_mesh.size = Vector3(0.16, 0.28, 0.08)
	clasp.mesh = clasp_mesh
	clasp.position = Vector3(0.0, 0.78, -0.54)
	var metal := StandardMaterial3D.new()
	metal.albedo_color = Color(0.72, 0.61, 0.31)
	clasp.material_override = metal
	chest.add_child(clasp)
	add_child(chest)

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
	presence_label = Label.new()
	presence_label.text = "Nearby players: offline"
	column.add_child(presence_label)

	inventory_label = Label.new()
	inventory_label.text = "Inventory: empty"
	column.add_child(inventory_label)

	_build_chat_panel(layer)
	_build_auth_panel(layer)
	skills_button = Button.new()
	skills_button.text = "Skills"
	skills_button.anchor_left = 1.0
	skills_button.anchor_right = 1.0
	skills_button.offset_left = -718
	skills_button.offset_top = 18
	skills_button.offset_right = -608
	skills_button.offset_bottom = 62
	skills_button.visible = false
	skills_button.pressed.connect(_toggle_skills_panel)
	layer.add_child(skills_button)
	bag_button = Button.new()
	bag_button.text = "Bag"
	bag_button.anchor_left = 1.0
	bag_button.anchor_right = 1.0
	bag_button.offset_left = -598
	bag_button.offset_top = 18
	bag_button.offset_right = -488
	bag_button.offset_bottom = 62
	bag_button.visible = false
	bag_button.pressed.connect(_show_inventory_panel)
	layer.add_child(bag_button)
	rest_button = Button.new()
	rest_button.text = "Rest"
	rest_button.anchor_left = 1.0
	rest_button.anchor_right = 1.0
	rest_button.offset_left = -478
	rest_button.offset_top = 18
	rest_button.offset_right = -368
	rest_button.offset_bottom = 62
	rest_button.visible = false
	rest_button.pressed.connect(_rest_character)
	layer.add_child(rest_button)
	equipment_button = Button.new()
	equipment_button.text = "Gear"
	equipment_button.anchor_left = 1.0
	equipment_button.anchor_right = 1.0
	equipment_button.offset_left = -358
	equipment_button.offset_top = 18
	equipment_button.offset_right = -248
	equipment_button.offset_bottom = 62
	equipment_button.visible = false
	equipment_button.pressed.connect(_show_equipment_panel)
	layer.add_child(equipment_button)
	character_button = Button.new()
	character_button.text = "Characters"
	character_button.anchor_left = 1.0
	character_button.anchor_right = 1.0
	character_button.offset_left = -238
	character_button.offset_top = 18
	character_button.offset_right = -128
	character_button.offset_bottom = 62
	character_button.visible = false
	character_button.pressed.connect(_request_character_list)
	layer.add_child(character_button)
	logout_button = Button.new()
	logout_button.text = "Sign out"
	logout_button.anchor_left = 1.0
	logout_button.anchor_right = 1.0
	logout_button.offset_left = -118
	logout_button.offset_top = 18
	logout_button.offset_right = -18
	logout_button.offset_bottom = 62
	logout_button.visible = false
	logout_button.pressed.connect(_logout)
	layer.add_child(logout_button)

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
	api_request = _new_api_request()
	add_child(api_request)
	api_request.request_completed.connect(_on_api_request_completed)
	var error := api_request.request(api_base_url + "/healthz")
	if error != OK:
		status_label.text = "Server: unavailable (start the Go world service)"

func _on_api_request_completed(result: int, response_code: int, _headers: PackedStringArray, body: PackedByteArray) -> void:
	if api_stage == "logout":
		if result == HTTPRequest.RESULT_SUCCESS and response_code == 204:
			auth_token = ""
			_stop_presence()
			player_id = ""
			if movement_timer != null:
				movement_timer.stop()
			movement_pending = false
			interaction_pending = false
			auth_panel.visible = true
			logout_button.visible = false
			character_button.visible = false
			equipment_button.visible = false
			rest_button.visible = false
			bag_button.visible = false
			skills_button.visible = false
			if character_panel != null:
				character_panel.visible = false
					_stop_chat()
			_stop_skills_request()
			if skills_panel != null:
				skills_panel.visible = false
			if market_panel != null:
				market_panel.visible = false
			if inventory_panel != null:
				inventory_panel.visible = false
			status_label.text = "Server: signed out"
			auth_status_label.text = "You have been signed out."
		else:
			status_label.text = "Sign out failed"
		return
	if result != HTTPRequest.RESULT_SUCCESS:
		status_label.text = "Server: connection failed"
		if auth_panel != null and auth_panel.visible:
			auth_status_label.text = "Could not reach the world service."
		return
	var payload: Variant = JSON.parse_string(body.get_string_from_utf8())
	if api_stage == "health":
		if response_code != 200 or not payload is Dictionary or payload.get("status") != "ok":
			status_label.text = "Server: unexpected health response"
			return
		api_stage = "market_catalog"
		var catalog_error := api_request.request(api_base_url + "/api/v1/market/items")
		if catalog_error != OK:
			status_label.text = "Server: ready • market catalog unavailable"
			auth_panel.visible = true
		return
	if not payload is Dictionary:
		if auth_panel != null and auth_panel.visible:
			auth_status_label.text = "Account service returned an invalid response."
		return
	if response_code < 200 or response_code >= 300:
		var error_code := str(payload.get("error", "request_failed"))
		_set_account_status("Account: " + error_code.replace("_", " "))
		return
	if api_stage == "market_catalog":
		var catalog_items: Variant = payload.get("items", [])
		if catalog_items is Array:
			market_catalog = catalog_items
		api_stage = "enemy_catalog"
		var enemy_error := api_request.request(api_base_url + "/api/v1/combat/enemies")
		if enemy_error != OK:
			auth_panel.visible = true
			status_label.text = "Server: ready • enemy catalog unavailable"
		return
	if api_stage == "enemy_catalog":
		var enemies: Variant = payload.get("enemies", [])
		if enemies is Array:
			enemy_catalog = enemies
			_build_enemy_spawns()
		api_stage = "node_catalog"
		var node_error := api_request.request(api_base_url + "/api/v1/world/nodes")
		if node_error != OK:
			status_label.text = "Server: ready • gathering catalog unavailable"
		return
	if api_stage == "node_catalog":
		var nodes: Variant = payload.get("nodes", [])
		if nodes is Array:
			gathering_catalog = nodes
		auth_panel.visible = true
		status_label.text = "Server: ready • sign in to continue"
		return
	match api_stage:
		"login", "register":
			auth_token = str(payload.get("accessToken", ""))
			if auth_token.is_empty():
				auth_status_label.text = "Account service did not return a session."
				return
			auth_status_label.text = "Signed in • loading characters..."
			_request_character_list()
		"list_characters":
			var characters: Variant = payload.get("characters", [])
			if characters is Array and not characters.is_empty():
				_show_character_selection(characters)
			else:
				_create_account_character("Wayfarer")
		"create_character":
			_activate_character(payload)
func _start_movement_sync() -> void:
	if movement_request != null:
		movement_request.queue_free()
	if movement_timer != null:
		movement_timer.queue_free()
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
		api_base_url + "/api/v1/account/characters/" + player_id + "/move",
		PackedStringArray(["Content-Type: application/json", "Authorization: Bearer " + auth_token]),
		HTTPClient.METHOD_POST,
		body
	)
	if error != OK:
		movement_pending = false

func _on_movement_completed(result: int, response_code: int, _headers: PackedStringArray, body: PackedByteArray) -> void:
	movement_pending = false
	if result != HTTPRequest.RESULT_SUCCESS:
		return
	if response_code == 401:
		_expire_session()
		return
	var payload: Variant = JSON.parse_string(body.get_string_from_utf8())
	if not payload is Dictionary:
		return
	if response_code == 409 and str(payload.get("error", "")) == "player_defeated":
		if movement_timer != null:
			movement_timer.stop()
		status_label.text = "Defeated • use Rest to respawn at the frontier"
		return
	var authoritative_position: Variant = payload.get("position", {})
	if authoritative_position is Dictionary:
		_apply_server_position(authoritative_position, response_code != 200)
	if response_code != 200:
		status_label.text = "Server: movement corrected by world rules"

func _start_gathering() -> void:
	if gather_request != null:
		gather_request.queue_free()
	gather_request = HTTPRequest.new()
	add_child(gather_request)
	gather_request.request_completed.connect(_on_gather_completed)

func _request_interaction() -> void:
	var bench_distance := Vector2(player_node.global_position.x, player_node.global_position.z - 4.0).length()
	var bank_distance := Vector2(player_node.global_position.x, player_node.global_position.z + 4.0).length()
	var merchant_distance := Vector2(player_node.global_position.x - 4.0, player_node.global_position.z).length()
	var forge_distance := Vector2(player_node.global_position.x, player_node.global_position.z - 8.0).length()
	var nearest_node_id := ""
	var nearest_node_distance := 1.0e20
	for node in gathering_catalog:
		if node is Dictionary:
			var node_position: Variant = node.get("position", {})
			if node_position is Dictionary:
				var distance := Vector2(
					player_node.global_position.x - float(node_position.get("x", 0.0)),
					player_node.global_position.z - float(node_position.get("z", 0.0))
				).length()
				if distance < nearest_node_distance:
					nearest_node_distance = distance
					nearest_node_id = str(node.get("id", ""))
	var nearest_enemy_id := ""
	var enemy_distance := 1.0e20
	for enemy in enemy_catalog:
		if enemy is Dictionary:
			var enemy_position: Variant = enemy.get("position", {})
			if enemy_position is Dictionary:
				var distance := Vector2(
					player_node.global_position.x - float(enemy_position.get("x", 0.0)),
					player_node.global_position.z - float(enemy_position.get("z", 0.0))
				).length()
				if distance < enemy_distance:
					enemy_distance = distance
					nearest_enemy_id = str(enemy.get("id", ""))
	var error := OK
	if merchant_distance <= 3.0 and merchant_distance <= bank_distance and merchant_distance <= bench_distance and merchant_distance <= nearest_node_distance and merchant_distance <= forge_distance and merchant_distance <= enemy_distance:
		_show_market_panel()
		return
	elif bank_distance <= 3.0 and bank_distance <= bench_distance and bank_distance <= nearest_node_distance and bank_distance <= forge_distance and bank_distance <= enemy_distance:
		var selected_item := ""
		interaction_type = "bank_deposit"
		for item in market_catalog:
			if item is Dictionary:
				var item_id := str(item.get("id", ""))
				if int(current_inventory.get(item_id, 0)) > 0:
					selected_item = item_id
					break
		if selected_item.is_empty():
			interaction_type = "bank_withdraw"
			for item in market_catalog:
				if item is Dictionary:
					var item_id := str(item.get("id", ""))
					if int(current_bank.get(item_id, 0)) > 0:
						selected_item = item_id
						break
		if selected_item.is_empty():
			status_label.text = "Bank: no items to move"
			return
		var bank_action := "deposit" if interaction_type == "bank_deposit" else "withdraw"
		status_label.text = "Bank: " + bank_action + "ing " + selected_item.replace("_", " ") + "..."
		error = gather_request.request(
			api_base_url + "/api/v1/account/characters/" + player_id + "/bank/" + bank_action,
			_api_headers(),
			HTTPClient.METHOD_POST,
			JSON.stringify({"itemId": selected_item, "quantity": 1})
		)
	elif forge_distance <= 3.0 and forge_distance <= bench_distance and forge_distance <= nearest_node_distance and forge_distance <= enemy_distance:
		var recipe_id := "copper-bar"
		if int(current_inventory.get("copper_bar", 0)) >= 3 and int(current_inventory.get("wooden_plank", 0)) >= 1:
			recipe_id = "copper-sword"
		interaction_type = "craft"
		status_label.text = "Using frontier forge..."
		error = gather_request.request(
			api_base_url + "/api/v1/account/characters/" + player_id + "/craft/" + recipe_id,
			_api_headers(),
			HTTPClient.METHOD_POST,
			""
		)
	elif bench_distance <= 3.0 and bench_distance <= nearest_node_distance and bench_distance <= enemy_distance:
		interaction_type = "craft"
		status_label.text = "Using frontier workbench..."
		error = gather_request.request(
			api_base_url + "/api/v1/account/characters/" + player_id + "/craft/basic-plank",
			_api_headers(),
			HTTPClient.METHOD_POST,
			""
		)
	elif not nearest_enemy_id.is_empty() and enemy_distance <= 3.0 and enemy_distance <= nearest_node_distance:
		interaction_type = "combat"
		status_label.text = "Attacking nearby creature..."
		error = gather_request.request(
			api_base_url + "/api/v1/account/characters/" + player_id + "/combat/attack/" + nearest_enemy_id,
			_api_headers(),
			HTTPClient.METHOD_POST,
			""
		)
	else:
		if nearest_node_id.is_empty() or nearest_node_distance > 3.0:
			status_label.text = "Move closer to a resource, station or trader"
			return
		interaction_type = "gather"
		error = gather_request.request(
			api_base_url + "/api/v1/account/characters/" + player_id + "/gather/" + nearest_node_id,
			_api_headers(),
			HTTPClient.METHOD_POST,
			""
		)
	if error == OK:
		interaction_pending = true

func _on_gather_completed(result: int, response_code: int, _headers: PackedStringArray, body: PackedByteArray) -> void:
	interaction_pending = false
	if result != HTTPRequest.RESULT_SUCCESS:
		status_label.text = "Interaction: server connection lost"
		return
	if response_code == 401:
		_expire_session()
		return
	var payload: Variant = JSON.parse_string(body.get_string_from_utf8())
	if not payload is Dictionary:
		return
	if response_code != 200:
		var error_code := str(payload.get("error", "interaction_failed"))
		var message := "Interaction: " + error_code.replace("_", " ")
		status_label.text = message
		if interaction_type == "market_buy" or interaction_type == "market_sell":
			market_status_label.text = message
		if interaction_type == "gear_equip" or interaction_type == "gear_unequip":
			equipment_status_label.text = message
		if interaction_type == "bank_deposit" or interaction_type == "bank_withdraw":
			if inventory_status_label != null:
				inventory_status_label.text = message
		if interaction_type == "combat":
			if error_code == "combat_cooldown":
				status_label.text = "Combat: wait for your next attack"
			elif error_code == "enemy_respawning":
				status_label.text = "Combat: this creature is recovering"
			elif error_code == "player_defeated":
				status_label.text = "Defeated • use Rest to respawn"
		return
	var player_data: Dictionary = payload.get("character", payload.get("player", {}))
	if interaction_type == "combat":
		var combat: Dictionary = payload.get("combat", {})
		_update_inventory(player_data.get("inventory", {}), player_data.get("bank", {}), player_data.get("coins", current_coins), player_data.get("equipment", {}), player_data.get("health", current_health))
		var combat_message := "Hit for " + str(combat.get("damageDealt", 0)) + " • Enemy HP " + str(combat.get("enemyHealth", 0)) + " • Your HP " + str(current_health)
		if bool(combat.get("enemyDefeated", false)):
			combat_message = "Enemy defeated • +" + str(combat.get("experience", 0)) + " Combat XP"
			if not str(combat.get("lootItemId", "")).is_empty():
				combat_message += " • Loot: " + str(combat.get("lootItemId", "")).replace("_", " ")
		if bool(combat.get("playerDefeated", false)):
			combat_message = "You were defeated • use Rest to respawn"
			if movement_timer != null:
				movement_timer.stop()
		status_label.text = "Combat: " + combat_message
		_refresh_skills_panel()
		return
	if interaction_type == "use_item":
		var item_id := str(payload.get("itemId", "minor_healing_potion"))
		_update_inventory(player_data.get("inventory", {}), player_data.get("bank", {}), player_data.get("coins", current_coins), player_data.get("equipment", {}), payload.get("health", player_data.get("health", current_health)))
		status_label.text = "Used " + item_id.replace("_", " ") + " • +" + str(payload.get("healed", 0)) + " health"
		return
	if interaction_type == "rest":
		_update_inventory(player_data.get("inventory", {}), player_data.get("bank", {}), player_data.get("coins", current_coins), player_data.get("equipment", {}), payload.get("health", player_data.get("health", current_health)))
		var position_data: Variant = player_data.get("position", {})
		if position_data is Dictionary:
			_apply_server_position(position_data, true)
		status_label.text = "Rested • Health restored to " + str(current_health)
		if movement_timer == null or movement_timer.is_stopped():
			_start_movement_sync()
		return
	if interaction_type == "market_buy" or interaction_type == "market_sell":
		var trade: Dictionary = payload.get("trade", {})
		var trade_action := "Bought" if interaction_type == "market_buy" else "Sold"
		_update_inventory(payload.get("inventory", player_data.get("inventory", {})), player_data.get("bank", {}), payload.get("coins", player_data.get("coins", current_coins)), player_data.get("equipment", {}), player_data.get("health", current_health))
		var trade_message := trade_action + " " + str(trade.get("quantity", 1)) + " " + str(trade.get("itemId", "item")).replace("_", " ") + " for " + str(trade.get("coins", 0)) + " coins"
		status_label.text = "Trader: " + trade_message
		market_status_label.text = trade_message + " • Balance: " + str(current_coins)
		return
	if interaction_type == "gear_equip" or interaction_type == "gear_unequip":
		var bonuses: Dictionary = payload.get("bonuses", {})
		_update_inventory(player_data.get("inventory", {}), player_data.get("bank", {}), player_data.get("coins", current_coins), payload.get("equipment", player_data.get("equipment", {})), player_data.get("health", current_health))
		var equipment_message := "Equipment updated • ATK +" + str(bonuses.get("attack", current_attack_bonus)) + " DEF +" + str(bonuses.get("defense", current_defense_bonus))
		status_label.text = equipment_message
		equipment_status_label.text = equipment_message
		_show_equipment_panel(equipment_message)
		return
	if interaction_type == "bank_deposit" or interaction_type == "bank_withdraw":
		var bank_action := "Deposited" if interaction_type == "bank_deposit" else "Withdrew"
		var item_bank: Variant = payload.get("bank", player_data.get("bank", {}))
		_update_inventory(payload.get("inventory", player_data.get("inventory", {})), item_bank, player_data.get("coins", current_coins), player_data.get("equipment", {}), player_data.get("health", current_health))
		status_label.text = "Bank: " + bank_action + " stack updated"
		if inventory_panel != null and inventory_panel.visible:
			_show_inventory_panel()
		return
	if interaction_type == "craft":
		var crafting: Dictionary = payload.get("crafting", {})
		var output_name := "crafted item"
		var outputs: Array = crafting.get("outputs", [])
		if not outputs.is_empty():
			var output_id := str(outputs[0].get("itemId", ""))
			for item in market_catalog:
				if item is Dictionary and str(item.get("id", "")) == output_id:
					output_name = str(item.get("name", output_id))
		status_label.text = "Crafted " + output_name + " • +" + str(crafting.get("experience", 0)) + " " + str(crafting.get("skillId", "crafting")) + " XP"
	else:
		var gathering: Dictionary = payload.get("gathering", {})
		var item_name := str(gathering.get("itemId", "resource")).replace("_", " ")
		for item in market_catalog:
			if item is Dictionary and str(item.get("id", "")) == str(gathering.get("itemId", "")):
				item_name = str(item.get("name", item_name))
		status_label.text = "Gathered " + item_name + " • +" + str(gathering.get("experience", 0)) + " " + str(gathering.get("skillId", "")) + " XP"
	_update_inventory(player_data.get("inventory", {}), player_data.get("bank", {}), player_data.get("coins", current_coins), player_data.get("equipment", {}), player_data.get("health", current_health))
	_refresh_skills_panel()


func _on_inventory_search_changed(value: String) -> void:
	inventory_search_text = value.strip_edges().to_lower()
	_refresh_inventory_item_list()


func _on_inventory_filter_selected(index: int) -> void:
	inventory_filter = str(inventory_filter_select.get_item_metadata(index))
	_refresh_inventory_item_list()


func _on_inventory_sort_selected(index: int) -> void:
	inventory_sort = str(inventory_sort_select.get_item_metadata(index))
	_refresh_inventory_item_list()


func _compare_inventory_items(left: Dictionary, right: Dictionary) -> bool:
	var left_id := str(left.get("id", ""))
	var right_id := str(right.get("id", ""))
	if inventory_sort == "bag":
		var left_bag_count := int(current_inventory.get(left_id, 0))
		var right_bag_count := int(current_inventory.get(right_id, 0))
		if left_bag_count != right_bag_count:
			return left_bag_count > right_bag_count
	elif inventory_sort == "bank":
		var left_bank_count := int(current_bank.get(left_id, 0))
		var right_bank_count := int(current_bank.get(right_id, 0))
		if left_bank_count != right_bank_count:
			return left_bank_count > right_bank_count
	return str(left.get("name", left_id)).naturalnocasecmp_to(str(right.get("name", right_id))) < 0


func _refresh_inventory_item_list() -> void:
	if inventory_items_column == null or inventory_bank_select == null:
		return
	for child in inventory_items_column.get_children():
		child.queue_free()
	inventory_bank_select.clear()
	var visible_items: Array[Dictionary] = []
	for item in market_catalog:
		if not item is Dictionary:
			continue
		var item_id := str(item.get("id", ""))
		var carried := int(current_inventory.get(item_id, 0))
		var banked := int(current_bank.get(item_id, 0))
		if carried == 0 and banked == 0:
			continue
		var item_name := str(item.get("name", item_id))
		if not inventory_search_text.is_empty() and not item_name.to_lower().contains(inventory_search_text) and not item_id.to_lower().contains(inventory_search_text):
			continue
		var equipment := str(item.get("equipmentSlot", "")) != ""
		var consumable := bool(item.get("consumable", false))
		if inventory_filter == "resources" and (equipment or consumable):
			continue
		if inventory_filter == "equipment" and not equipment:
			continue
		if inventory_filter == "consumables" and not consumable:
			continue
		visible_items.append(item)
	visible_items.sort_custom(_compare_inventory_items)
	for item in visible_items:
		var item_id := str(item.get("id", ""))
		var carried := int(current_inventory.get(item_id, 0))
		var banked := int(current_bank.get(item_id, 0))
		var item_label := Label.new()
		item_label.text = str(item.get("name", item_id)) + "  •  Bag " + str(carried) + "  •  Bank " + str(banked)
		inventory_items_column.add_child(item_label)
		inventory_bank_select.add_item(str(item.get("name", item_id)) + " (Bag " + str(carried) + " / Bank " + str(banked) + ")")
		inventory_bank_select.set_item_metadata(inventory_bank_select.item_count - 1, item_id)
	if visible_items.is_empty():
		var empty_label := Label.new()
		empty_label.text = "No items match this filter."
		inventory_items_column.add_child(empty_label)


func _update_inventory(inventory_data: Variant, bank_data: Variant = {}, coins_data: Variant = null, equipment_data: Variant = {}, health_data: Variant = null) -> void:
	if not inventory_data is Dictionary:
		return
	current_inventory = inventory_data.duplicate()
	current_bank = bank_data.duplicate() if bank_data is Dictionary else {}
	if coins_data != null:
		current_coins = int(coins_data)
	if health_data != null:
		current_health = int(health_data)
	current_equipment = equipment_data.duplicate() if equipment_data is Dictionary else {}
	current_attack_bonus = 0
	current_defense_bonus = 0
	for item_id in current_equipment.values():
		for item in market_catalog:
			if item is Dictionary and str(item.get("id", "")) == str(item_id):
				current_attack_bonus += int(item.get("attackBonus", 0))
				current_defense_bonus += int(item.get("defenseBonus", 0))
	_refresh_equipment_visuals()
	inventory_label.text = "HP %d/%d • Coins %d • ATK %d DEF %d • Bag L:%d P:%d S:%d Cu:%d/%d • Bank L:%d P:%d S:%d Cu:%d/%d" % [
			current_health, game_max_health, current_coins, current_attack_bonus, current_defense_bonus,
		int(current_inventory.get("softwood_log", 0)),
		int(current_inventory.get("wooden_plank", 0)),
		int(current_inventory.get("rough_stone", 0)),
		int(current_inventory.get("copper_ore", 0)),
		int(current_inventory.get("copper_bar", 0)),
		int(current_bank.get("softwood_log", 0)),
		int(current_bank.get("wooden_plank", 0)),
		int(current_bank.get("rough_stone", 0)),
		int(current_bank.get("copper_ore", 0)),
		int(current_bank.get("copper_bar", 0))
	]
	if inventory_panel != null and inventory_panel.visible:
		_show_inventory_panel()

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
func _build_auth_panel(parent: CanvasLayer) -> void:
	auth_panel = PanelContainer.new()
	auth_panel.position = Vector2(18, 155)
	auth_panel.custom_minimum_size = Vector2(360, 220)
	parent.add_child(auth_panel)
	var column := VBoxContainer.new()
	auth_panel.add_child(column)
	var title := Label.new()
	title.text = "ACCOUNT SIGN IN"
	title.add_theme_font_size_override("font_size", 18)
	column.add_child(title)
	auth_email_input = LineEdit.new()
	auth_email_input.placeholder_text = "Email address"
	auth_email_input.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	column.add_child(auth_email_input)
	auth_password_input = LineEdit.new()
	auth_password_input.placeholder_text = "Password (12+ characters)"
	auth_password_input.secret = true
	auth_password_input.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	column.add_child(auth_password_input)
	var actions := HBoxContainer.new()
	column.add_child(actions)
	var login_button := Button.new()
	login_button.text = "Login"
	login_button.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	login_button.pressed.connect(func() -> void: _request_auth("login"))
	actions.add_child(login_button)
	var register_button := Button.new()
	register_button.text = "Create account"
	register_button.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	register_button.pressed.connect(func() -> void: _request_auth("register"))
	actions.add_child(register_button)
	auth_status_label = Label.new()
	auth_status_label.text = "Sign in or create a new account."
	auth_status_label.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	column.add_child(auth_status_label)

func _request_auth(mode: String) -> void:
	if auth_email_input.text.strip_edges().is_empty() or auth_password_input.text.length() < 12:
		auth_status_label.text = "Enter a valid email and a password of at least 12 characters."
		return
	auth_mode = mode
	api_stage = mode
	auth_status_label.text = "Connecting to account service..."
	var body := JSON.stringify({"email": auth_email_input.text.strip_edges(), "password": auth_password_input.text})
	var error := api_request.request(
		api_base_url + "/api/v1/auth/" + mode,
		PackedStringArray(["Content-Type: application/json"]),
		HTTPClient.METHOD_POST,
		body
	)
	if error != OK:
		auth_status_label.text = "Could not send account request."

func _api_headers() -> PackedStringArray:
	return PackedStringArray(["Authorization: Bearer " + auth_token])

func _logout() -> void:
	if auth_token.is_empty():
		return
	api_stage = "logout"
	var error := api_request.request(
		api_base_url + "/api/v1/auth/logout",
		_api_headers(),
		HTTPClient.METHOD_POST,
		""
	)
	if error != OK:
		status_label.text = "Sign out failed; try again."

func _activate_character(character_data: Dictionary) -> void:
	_stop_presence()
	_stop_skills_request()
	if movement_timer != null:
		movement_timer.stop()
	player_id = str(character_data.get("id", ""))
	if player_id.is_empty():
		status_label.text = "Account: character response had no ID"
		return
	var spawn: Variant = character_data.get("position", {})
	if spawn is Dictionary:
		_apply_server_position(spawn, true)
	status_label.text = "Server: connected • account character ready"
	_update_inventory(character_data.get("inventory", {}), character_data.get("bank", {}), character_data.get("coins", current_coins), character_data.get("equipment", {}), character_data.get("health", current_health))
	auth_panel.visible = false
	if character_panel != null:
		character_panel.visible = false
	if equipment_panel != null:
		equipment_panel.visible = false
	if skills_panel != null:
		skills_panel.visible = false
	if inventory_panel != null:
		inventory_panel.visible = false
	if market_panel != null:
		market_panel.visible = false
	logout_button.visible = true
	character_button.visible = true
	equipment_button.visible = true
	rest_button.visible = true
	bag_button.visible = true
	skills_button.visible = true
	_start_movement_sync()
	_start_gathering()
	_start_chat()
	_start_presence()

func _request_character_list() -> void:
	api_stage = "list_characters"
	var error := api_request.request(
		api_base_url + "/api/v1/account/characters",
		_api_headers(),
		HTTPClient.METHOD_GET
	)
	if error != OK:
		_set_account_status("Could not load account characters.")

func _create_account_character(character_name: String) -> void:
	character_name = character_name.strip_edges()
	if character_name.length() < 3 or character_name.length() > 20:
		_set_account_status("Character names must be 3–20 characters.")
		return
	api_stage = "create_character"
	var error := api_request.request(
		api_base_url + "/api/v1/account/characters",
		PackedStringArray(["Content-Type: application/json", "Authorization: Bearer " + auth_token]),
		HTTPClient.METHOD_POST,
		JSON.stringify({"name": character_name})
	)
	if error != OK:
		_set_account_status("Could not create a character.")



func _expire_session() -> void:
	_stop_chat()
	_stop_skills_request()
	auth_token = ""
	player_id = ""
	movement_pending = false
	interaction_pending = false
	if movement_timer != null:
		movement_timer.stop()
	auth_panel.visible = true
	logout_button.visible = false
	character_button.visible = false
	rest_button.visible = false
	if character_panel != null:
		character_panel.visible = false
	if market_panel != null:
		market_panel.visible = false
	if equipment_panel != null:
		equipment_panel.visible = false
	if skills_panel != null:
		skills_panel.visible = false
	if inventory_panel != null:
		inventory_panel.visible = false
	bag_button.visible = false
	skills_button.visible = false
	status_label.text = "Server: session expired"
	auth_status_label.text = "Your session expired. Please sign in again."

func _show_character_selection(characters: Array) -> void:
	_stop_chat()
	_stop_skills_request()
	if market_panel != null:
		market_panel.visible = false
	if skills_panel != null:
		skills_panel.visible = false
	if inventory_panel != null:
		inventory_panel.visible = false
	if movement_timer != null:
		movement_timer.stop()
	movement_pending = false
	player_id = ""
	interaction_pending = false
	if character_panel != null:
		character_panel.queue_free()
	character_panel = PanelContainer.new()
	character_panel.position = Vector2(18, 155)
	character_panel.custom_minimum_size = Vector2(360, 230)
	auth_panel.visible = false
	logout_button.visible = true
	character_button.visible = false
	equipment_button.visible = false
	rest_button.visible = false
	bag_button.visible = false
	skills_button.visible = false
	var column := VBoxContainer.new()
	character_panel.add_child(column)
	var heading := Label.new()
	heading.text = "SELECT A CHARACTER"
	heading.add_theme_font_size_override("font_size", 18)
	column.add_child(heading)
	character_select = OptionButton.new()
	character_select.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	for character in characters:
		if character is Dictionary:
			character_select.add_item(str(character.get("name", "Unnamed")))
			character_select.set_item_metadata(character_select.item_count - 1, character)
	column.add_child(character_select)
	var enter_button := Button.new()
	enter_button.text = "Enter world"
	enter_button.pressed.connect(func() -> void:
		var selected: Variant = character_select.get_selected_metadata()
		if selected is Dictionary:
			_activate_character(selected)
	)
	column.add_child(enter_button)
	var separator := HSeparator.new()
	column.add_child(separator)
	var new_character_label := Label.new()
	new_character_label.text = "Create another character"
	column.add_child(new_character_label)
	character_name_input = LineEdit.new()
	character_name_input.placeholder_text = "Character name (3–20 characters)"
	column.add_child(character_name_input)
	var create_button := Button.new()
	create_button.text = "Create character"
	create_button.pressed.connect(func() -> void: _create_account_character(character_name_input.text))
	column.add_child(create_button)
	character_status_label = Label.new()
	character_status_label.text = "Choose an existing character or create another."
	character_status_label.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	column.add_child(character_status_label)
	auth_panel.get_parent().add_child(character_panel)


func _show_market_panel() -> void:
	if skills_panel != null:
		skills_panel.visible = false
	if inventory_panel != null:
		inventory_panel.visible = false
	if equipment_panel != null:
		equipment_panel.visible = false
	if market_panel != null:
		market_panel.visible = true
		market_status_label.text = "Coins: " + str(current_coins)
		return
	market_panel = PanelContainer.new()
	market_panel.position = Vector2(420, 155)
	market_panel.custom_minimum_size = Vector2(320, 220)
	var column := VBoxContainer.new()
	market_panel.add_child(column)
	var heading := Label.new()
	heading.text = "FRONTIER TRADER"
	heading.add_theme_font_size_override("font_size", 18)
	column.add_child(heading)
	market_item_select = OptionButton.new()
	market_item_select.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	var market_items := market_catalog
	if market_items.is_empty():
		market_items = [
			{"id":"softwood_log","name":"Softwood Log","buyPrice":5,"sellPrice":2},
			{"id":"rough_stone","name":"Rough Stone","buyPrice":3,"sellPrice":1},
			{"id":"wooden_plank","name":"Wooden Plank","buyPrice":12,"sellPrice":5}
		]
	for item in market_items:
		if item is Dictionary:
			var item_name := str(item.get("name", item.get("id", "Item")))
			var prices := " (buy %d / sell %d)" % [int(item.get("buyPrice", 0)), int(item.get("sellPrice", 0))]
			market_item_select.add_item(item_name + prices)
			market_item_select.set_item_metadata(market_item_select.item_count-1,str(item.get("id", "")))
	column.add_child(market_item_select)
	market_quantity_input = LineEdit.new()
	market_quantity_input.text = "1"
	market_quantity_input.placeholder_text = "Quantity (1–100)"
	market_quantity_input.virtual_keyboard_type = LineEdit.KEYBOARD_TYPE_NUMBER
	column.add_child(market_quantity_input)
	var actions := HBoxContainer.new()
	column.add_child(actions)
	var buy_button := Button.new()
	buy_button.text = "Buy"
	buy_button.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	buy_button.pressed.connect(func() -> void: _submit_market_trade(true))
	actions.add_child(buy_button)
	var sell_button := Button.new()
	sell_button.text = "Sell"
	sell_button.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	sell_button.pressed.connect(func() -> void: _submit_market_trade(false))
	actions.add_child(sell_button)
	var close_button := Button.new()
	close_button.text = "Close"
	close_button.pressed.connect(func() -> void: market_panel.visible = false)
	column.add_child(close_button)
	market_status_label = Label.new()
	market_status_label.text = "Coins: " + str(current_coins)
	market_status_label.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	column.add_child(market_status_label)
	auth_panel.get_parent().add_child(market_panel)

func _submit_market_trade(buy: bool) -> void:
	var quantity := int(market_quantity_input.text)
	if quantity < 1 or quantity > 100:
		market_status_label.text = "Quantity must be between 1 and 100."
		return
	var item_id := str(market_item_select.get_selected_metadata())
	if item_id.is_empty():
		market_status_label.text = "Choose an item."
		return
	interaction_type = "market_buy" if buy else "market_sell"
	var action := "buy" if buy else "sell"
	var error := gather_request.request(
		api_base_url + "/api/v1/account/characters/" + player_id + "/market/" + action + "/" + item_id,
		PackedStringArray(["Content-Type: application/json", "Authorization: Bearer " + auth_token]),
		HTTPClient.METHOD_POST,
		JSON.stringify({"quantity": quantity})
	)
	if error == OK:
		interaction_pending = true
		market_status_label.text = "Processing trade..."
	else:
		market_status_label.text = "Could not send market request."





func _show_equipment_panel(message: String = "") -> void:
	if skills_panel != null:
		skills_panel.visible = false
	if market_panel != null:
		market_panel.visible = false
	if inventory_panel != null:
		inventory_panel.visible = false
	if equipment_panel != null:
		equipment_panel.queue_free()
	equipment_panel = PanelContainer.new()
	equipment_panel.position = Vector2(420, 390)
	equipment_panel.custom_minimum_size = Vector2(320, 260)
	var column := VBoxContainer.new()
	equipment_panel.add_child(column)
	var heading := Label.new()
	heading.text = "EQUIPMENT"
	heading.add_theme_font_size_override("font_size", 18)
	column.add_child(heading)
	var current_label := Label.new()
	current_label.text = "Equipped: " + str(current_equipment)
	current_label.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	column.add_child(current_label)
	equipment_item_select = OptionButton.new()
	equipment_item_select.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	for item in market_catalog:
		if item is Dictionary:
			var item_id := str(item.get("id", ""))
			if str(item.get("equipmentSlot", "")) != "" and int(current_inventory.get(item_id, 0)) > 0:
				equipment_item_select.add_item(str(item.get("name", item_id)))
				equipment_item_select.set_item_metadata(equipment_item_select.item_count - 1, item_id)
	column.add_child(equipment_item_select)
	var equip_button := Button.new()
	equip_button.text = "Equip selected item"
	equip_button.disabled = equipment_item_select.item_count == 0
	equip_button.pressed.connect(_submit_equip_item)
	column.add_child(equip_button)
	equipment_slot_select = OptionButton.new()
	for slot in ["main_hand", "head", "body", "feet"]:
		equipment_slot_select.add_item(slot.capitalize())
		equipment_slot_select.set_item_metadata(equipment_slot_select.item_count - 1, slot)
	column.add_child(equipment_slot_select)
	var unequip_button := Button.new()
	unequip_button.text = "Unequip selected slot"
	unequip_button.pressed.connect(_submit_unequip_item)
	column.add_child(unequip_button)
	var close_button := Button.new()
	close_button.text = "Close"
	close_button.pressed.connect(func() -> void: equipment_panel.visible = false)
	column.add_child(close_button)
	equipment_status_label = Label.new()
	equipment_status_label.text = message if not message.is_empty() else "Buy starter gear from the trader to equip it."
	equipment_status_label.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	column.add_child(equipment_status_label)
	auth_panel.get_parent().add_child(equipment_panel)

func _submit_equip_item() -> void:
	var item_id := str(equipment_item_select.get_selected_metadata())
	if item_id.is_empty():
		equipment_status_label.text = "Choose an item from your inventory."
		return
	_submit_equipment_request("equip", item_id)

func _submit_unequip_item() -> void:
	var slot := str(equipment_slot_select.get_selected_metadata())
	if slot.is_empty():
		equipment_status_label.text = "Choose an equipment slot."
		return
	_submit_equipment_request("unequip", slot)

func _submit_equipment_request(action: String, item_or_slot: String) -> void:
	interaction_type = "gear_equip" if action == "equip" else "gear_unequip"
	var route := "equipment/" + action + "/" + item_or_slot
	var error := gather_request.request(
		api_base_url + "/api/v1/account/characters/" + player_id + "/" + route,
		_api_headers(),
		HTTPClient.METHOD_POST,
		""
	)
	if error == OK:
		interaction_pending = true
		equipment_status_label.text = "Updating equipment..."
	else:
		equipment_status_label.text = "Could not send equipment request."


func _rest_character() -> void:
	if player_id.is_empty() or auth_token.is_empty() or interaction_pending:
		return
	interaction_type = "rest"
	var error := gather_request.request(
		api_base_url + "/api/v1/account/characters/" + player_id + "/rest",
		_api_headers(),
		HTTPClient.METHOD_POST,
		""
	)
	if error == OK:
		interaction_pending = true
		status_label.text = "Resting at the frontier spawn..."
	else:
		status_label.text = "Could not send rest request."


func _start_presence() -> void:
	_stop_presence()
	if player_id.is_empty() or auth_token.is_empty():
		return
	presence_heartbeat_request = HTTPRequest.new()
	add_child(presence_heartbeat_request)
	presence_heartbeat_request.request_completed.connect(_on_presence_heartbeat_completed)
	presence_roster_request = HTTPRequest.new()
	add_child(presence_roster_request)
	presence_roster_request.request_completed.connect(_on_presence_roster_completed)
	presence_heartbeat_timer = Timer.new()
	presence_heartbeat_timer.wait_time = 5.0
	presence_heartbeat_timer.timeout.connect(_send_presence_heartbeat)
	add_child(presence_heartbeat_timer)
	presence_roster_timer = Timer.new()
	presence_roster_timer.wait_time = 3.0
	presence_roster_timer.timeout.connect(_poll_zone_presence)
	add_child(presence_roster_timer)
	presence_heartbeat_timer.start()
	presence_roster_timer.start()
	_send_presence_heartbeat()
	_poll_zone_presence()


func _stop_presence() -> void:
	if presence_heartbeat_timer != null:
		presence_heartbeat_timer.stop()
		presence_heartbeat_timer.queue_free()
		presence_heartbeat_timer = null
	if presence_roster_timer != null:
		presence_roster_timer.stop()
		presence_roster_timer.queue_free()
		presence_roster_timer = null
	if presence_heartbeat_request != null:
		if presence_heartbeat_request.get_http_client_status() != HTTPClient.STATUS_DISCONNECTED:
			presence_heartbeat_request.cancel_request()
		presence_heartbeat_request.queue_free()
		presence_heartbeat_request = null
	if presence_roster_request != null:
		if presence_roster_request.get_http_client_status() != HTTPClient.STATUS_DISCONNECTED:
			presence_roster_request.cancel_request()
		presence_roster_request.queue_free()
		presence_roster_request = null
	presence_heartbeat_pending = false
	presence_roster_pending = false
	for remote in remote_player_nodes.values():
		if is_instance_valid(remote):
			remote.queue_free()
	remote_player_nodes.clear()
	remote_player_targets.clear()
	if presence_label != null:
		presence_label.text = "Nearby players: offline"


func _send_presence_heartbeat() -> void:
	if player_id.is_empty() or presence_heartbeat_pending or presence_heartbeat_request == null:
		return
	var url := api_base_url + "/api/v1/account/characters/" + player_id + "/presence/heartbeat"
	var error := presence_heartbeat_request.request(url, _api_headers(), HTTPClient.METHOD_POST, "")
	if error == OK:
		presence_heartbeat_pending = true


func _on_presence_heartbeat_completed(result: int, response_code: int, _headers: PackedStringArray, _body: PackedByteArray) -> void:
	presence_heartbeat_pending = false
	if player_id.is_empty():
		return
	if response_code == 401:
		_expire_session()
		return
	if result != HTTPRequest.RESULT_SUCCESS or response_code != 204:
		if response_code == 404:
			_stop_presence()
		return


func _poll_zone_presence() -> void:
	if player_id.is_empty() or presence_roster_pending or presence_roster_request == null:
		return
	var url := api_base_url + "/api/v1/account/characters/" + player_id + "/presence"
	var error := presence_roster_request.request(url, _api_headers())
	if error == OK:
		presence_roster_pending = true


func _on_presence_roster_completed(result: int, response_code: int, _headers: PackedStringArray, body: PackedByteArray) -> void:
	presence_roster_pending = false
	if player_id.is_empty():
		return
	if response_code == 401:
		_expire_session()
		return
	if result != HTTPRequest.RESULT_SUCCESS or response_code != 200:
		return
	var payload: Variant = JSON.parse_string(body.get_string_from_utf8())
	if not payload is Dictionary:
		return
	var players: Variant = payload.get("players", [])
	if not players is Array:
		return
	var seen: Dictionary = {}
	for remote_data in players:
		if not remote_data is Dictionary:
			continue
		var character_id := str(remote_data.get("characterId", ""))
		if character_id.is_empty() or character_id == player_id:
			continue
		var zone_id := str(remote_data.get("zoneId", ""))
		if zone_id.is_empty():
			continue
		seen[character_id] = true
		var target := Vector3(float(remote_data.get("x", 0.0)), 0.05, float(remote_data.get("z", 0.0)))
		if not remote_player_nodes.has(character_id):
			_create_remote_player_visual(character_id, str(remote_data.get("name", "Wayfarer")), target)
		remote_player_targets[character_id] = target
	for character_id in remote_player_nodes.keys():
		if not seen.has(character_id):
			var stale: Node3D = remote_player_nodes[character_id]
			if is_instance_valid(stale):
				stale.queue_free()
			remote_player_nodes.erase(character_id)
			remote_player_targets.erase(character_id)
	if status_label != null:
		if presence_label != null:
		presence_label.text = "Nearby players: " + str(remote_player_nodes.size())


func _create_remote_player_visual(character_id: String, display_name: String, at: Vector3) -> void:
	var remote := Node3D.new()
	remote.name = "Remote_" + character_id
	remote.position = at
	var body := MeshInstance3D.new()
	var mesh := CapsuleMesh.new()
	mesh.radius = 0.32
	mesh.height = 1.35
	body.mesh = mesh
	body.position.y = 0.68
	var material := StandardMaterial3D.new()
	material.albedo_color = Color(0.18, 0.55, 0.78)
	material.roughness = 0.8
	body.material_override = material
	remote.add_child(body)
	var label := Label3D.new()
	label.text = display_name
	label.position.y = 1.65
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 28
	remote.add_child(label)
	add_child(remote)
	remote_player_nodes[character_id] = remote


func _update_remote_player_visuals(delta: float) -> void:
	for character_id in remote_player_nodes.keys():
		var remote: Node3D = remote_player_nodes[character_id]
		if not is_instance_valid(remote):
			continue
		var target: Vector3 = remote_player_targets.get(character_id, remote.position)
		remote.position = remote.position.lerp(target, clampf(delta * 5.0, 0.0, 1.0))


func _build_chat_panel(layer: CanvasLayer) -> void:
	chat_panel = PanelContainer.new()
	chat_panel.position = Vector2(18, 430)
	chat_panel.custom_minimum_size = Vector2(390, 250)
	chat_panel.visible = false
	var column := VBoxContainer.new()
	chat_panel.add_child(column)
	var heading := Label.new()
	heading.text = "LOCAL ZONE CHAT"
	heading.add_theme_font_size_override("font_size", 16)
	column.add_child(heading)
	chat_history = RichTextLabel.new()
	chat_history.bbcode_enabled = false
	chat_history.scroll_following = true
	chat_history.custom_minimum_size = Vector2(360, 150)
	chat_history.text = "Messages from players in your current zone appear here."
	column.add_child(chat_history)
	var input_row := HBoxContainer.new()
	column.add_child(input_row)
	chat_input = LineEdit.new()
	chat_input.placeholder_text = "Say something to the zone..."
	chat_input.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	chat_input.max_length = 240
	chat_input.text_submitted.connect(func(_value: String) -> void: _send_local_chat())
	input_row.add_child(chat_input)
	var send_button := Button.new()
	send_button.text = "Send"
	send_button.pressed.connect(_send_local_chat)
	input_row.add_child(send_button)
	chat_status_label = Label.new()
	chat_status_label.text = "Zone chat"
	column.add_child(chat_status_label)
	layer.add_child(chat_panel)

func _start_chat() -> void:
	_stop_chat()
	if player_id.is_empty() or auth_token.is_empty():
		return
	chat_active_character_id = player_id
	chat_last_id = 0
	chat_poll_pending = false
	chat_send_pending = false
	chat_lines.clear()
	_append_chat_line("Local zone chat • messages are visible to players in this zone.")
	chat_poll_request = HTTPRequest.new()
	add_child(chat_poll_request)
	chat_poll_request.request_completed.connect(_on_chat_poll_completed)
	chat_send_request = HTTPRequest.new()
	add_child(chat_send_request)
	chat_send_request.request_completed.connect(_on_chat_send_completed)
	chat_poll_timer = Timer.new()
	chat_poll_timer.wait_time = 2.0
	chat_poll_timer.timeout.connect(_poll_local_chat)
	add_child(chat_poll_timer)
	chat_panel.visible = true
	chat_poll_timer.start()
	_poll_local_chat()

func _stop_chat() -> void:
	chat_active_character_id = ""
	if chat_poll_timer != null:
		chat_poll_timer.stop()
		chat_poll_timer.queue_free()
		chat_poll_timer = null
	if chat_poll_request != null:
		if chat_poll_request.get_http_client_status() != HTTPClient.STATUS_DISCONNECTED:
			chat_poll_request.cancel_request()
		chat_poll_request.queue_free()
		chat_poll_request = null
	if chat_send_request != null:
		if chat_send_request.get_http_client_status() != HTTPClient.STATUS_DISCONNECTED:
			chat_send_request.cancel_request()
		chat_send_request.queue_free()
		chat_send_request = null
	chat_poll_pending = false
	chat_send_pending = false
	if chat_panel != null:
		chat_panel.visible = false

func _poll_local_chat() -> void:
	if chat_active_character_id.is_empty() or player_id != chat_active_character_id or chat_poll_pending or chat_poll_request == null:
		return
	var url := api_base_url + "/api/v1/account/characters/" + chat_active_character_id + "/chat/local?after=" + str(chat_last_id) + "&limit=50"
	var error := chat_poll_request.request(url, _api_headers())
	if error == OK:
		chat_poll_pending = true
	else:
		chat_status_label.text = "Chat connection unavailable"

func _send_local_chat() -> void:
	if chat_active_character_id.is_empty() or player_id != chat_active_character_id or chat_send_pending or chat_send_request == null:
		return
	var message := chat_input.text.strip_edges()
	if message.is_empty():
		return
	chat_send_pending = true
	var error := chat_send_request.request(
		api_base_url + "/api/v1/account/characters/" + chat_active_character_id + "/chat/local",
		_api_headers(),
		HTTPClient.METHOD_POST,
		JSON.stringify({"message": message})
	)
	if error == OK:
		chat_status_label.text = "Sending..."
	else:
		chat_send_pending = false
		chat_status_label.text = "Could not send message"

func _on_chat_poll_completed(result: int, response_code: int, _headers: PackedStringArray, body: PackedByteArray) -> void:
	chat_poll_pending = false
	if chat_active_character_id.is_empty() or player_id != chat_active_character_id:
		return
	if result != HTTPRequest.RESULT_SUCCESS:
		chat_status_label.text = "Chat connection unavailable"
		return
	if response_code == 401:
		_expire_session()
		return
	if response_code != 200:
		chat_status_label.text = "Chat: " + str(response_code)
		return
	var payload: Variant = JSON.parse_string(body.get_string_from_utf8())
	if not payload is Dictionary:
		return
	var messages: Variant = payload.get("messages", [])
	if messages is Array:
		for message in messages:
			if message is Dictionary:
				var message_id := int(message.get("id", 0))
				if message_id <= chat_last_id:
					continue
				chat_last_id = message_id
				_append_chat_line(str(message.get("characterName", "Unknown")) + ": " + str(message.get("message", "")))
	chat_status_label.text = "Zone chat • cursor " + str(chat_last_id)

func _on_chat_send_completed(result: int, response_code: int, _headers: PackedStringArray, body: PackedByteArray) -> void:
	chat_send_pending = false
	if chat_active_character_id.is_empty() or player_id != chat_active_character_id:
		return
	if result != HTTPRequest.RESULT_SUCCESS:
		chat_status_label.text = "Chat connection unavailable"
		return
	if response_code == 401:
		_expire_session()
		return
	var payload: Variant = JSON.parse_string(body.get_string_from_utf8())
	if response_code == 201 and payload is Dictionary:
		var message: Dictionary = payload.get("message", {})
		var message_id := int(message.get("id", 0))
		if message_id > chat_last_id:
			chat_last_id = message_id
			_append_chat_line(str(message.get("characterName", "You")) + ": " + str(message.get("message", "")))
		chat_input.clear()
		chat_status_label.text = "Message sent"
	elif response_code == 429:
		chat_status_label.text = "Chat cooldown • wait a moment"
	else:
		chat_status_label.text = "Chat: " + str(payload.get("error", "send_failed") if payload is Dictionary else "send_failed")


func _show_inventory_panel() -> void:
	if skills_panel != null:
		skills_panel.visible = false
	if market_panel != null:
		market_panel.visible = false
	if equipment_panel != null:
		equipment_panel.visible = false
	if inventory_panel != null:
		inventory_panel.queue_free()
	inventory_panel = PanelContainer.new()
	inventory_panel.position = Vector2(18, 155)
	inventory_panel.custom_minimum_size = Vector2(360, 245)
	var column := VBoxContainer.new()
	inventory_panel.add_child(column)
	var heading := Label.new()
	heading.text = "INVENTORY & BANK"
	heading.add_theme_font_size_override("font_size", 18)
	column.add_child(heading)
	var summary := Label.new()
	summary.text = "Health %d/%d  •  Coins %d" % [current_health, game_max_health, current_coins]
	column.add_child(summary)
	var scroll := ScrollContainer.new()
	scroll.custom_minimum_size = Vector2(330, 130)
	scroll.size_flags_vertical = Control.SIZE_EXPAND_FILL
	column.add_child(scroll)
	var items_column := VBoxContainer.new()
	items_column.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	scroll.add_child(items_column)
	var filter_row := HBoxContainer.new()
	column.add_child(filter_row)
	inventory_filter_select = OptionButton.new()
	for filter in [{"label":"All items","id":"all"},{"label":"Resources","id":"resources"},{"label":"Equipment","id":"equipment"},{"label":"Consumables","id":"consumables"}]:
		inventory_filter_select.add_item(str(filter.label))
		inventory_filter_select.set_item_metadata(inventory_filter_select.item_count - 1, str(filter.id))
		if str(filter.id) == inventory_filter:
			inventory_filter_select.select(inventory_filter_select.item_count - 1)
	inventory_filter_select.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	inventory_filter_select.item_selected.connect(_on_inventory_filter_selected)
	filter_row.add_child(inventory_filter_select)
	inventory_sort_select = OptionButton.new()
	for sort_option in [{"label":"Name","id":"name"},{"label":"Bag count","id":"bag"},{"label":"Bank count","id":"bank"}]:
		inventory_sort_select.add_item(str(sort_option.label))
		inventory_sort_select.set_item_metadata(inventory_sort_select.item_count - 1, str(sort_option.id))
		if str(sort_option.id) == inventory_sort:
			inventory_sort_select.select(inventory_sort_select.item_count - 1)
	inventory_sort_select.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	inventory_sort_select.item_selected.connect(_on_inventory_sort_selected)
	filter_row.add_child(inventory_sort_select)
	inventory_search_input = LineEdit.new()
	inventory_search_input.placeholder_text = "Search items..."
	inventory_search_input.text = inventory_search_text
	inventory_search_input.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	inventory_search_input.text_changed.connect(_on_inventory_search_changed)
	column.add_child(inventory_search_input)
	var scroll := ScrollContainer.new()
	scroll.custom_minimum_size = Vector2(330, 130)
	scroll.size_flags_vertical = Control.SIZE_EXPAND_FILL
	column.add_child(scroll)
	inventory_items_column = VBoxContainer.new()
	inventory_items_column.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	scroll.add_child(inventory_items_column)
	inventory_bank_select = OptionButton.new()
	inventory_bank_select.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	column.add_child(inventory_bank_select)
	_refresh_inventory_item_list()
	inventory_quantity_input = LineEdit.new()
	inventory_quantity_input.text = "1"
	inventory_quantity_input.placeholder_text = "Transfer quantity (1–100)"
	inventory_quantity_input.virtual_keyboard_type = LineEdit.KEYBOARD_TYPE_NUMBER
	column.add_child(inventory_quantity_input)
	var bank_actions := HBoxContainer.new()
	var deposit_button := Button.new()
	deposit_button.text = "Deposit"
	deposit_button.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	deposit_button.disabled = inventory_bank_select.item_count == 0
	deposit_button.pressed.connect(func() -> void: _submit_bank_transfer(true))
	bank_actions.add_child(deposit_button)
	var withdraw_button := Button.new()
	withdraw_button.text = "Withdraw"
	withdraw_button.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	withdraw_button.disabled = inventory_bank_select.item_count == 0
	withdraw_button.pressed.connect(func() -> void: _submit_bank_transfer(false))
	bank_actions.add_child(withdraw_button)
	column.add_child(bank_actions)
	inventory_status_label = Label.new()
	inventory_status_label.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	inventory_status_label.text = "Select an item and quantity to move stacks between bag and bank."
	column.add_child(inventory_status_label)
	inventory_use_select = OptionButton.new()
	inventory_use_select.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	for item in market_catalog:
		if item is Dictionary:
			var item_id := str(item.get("id", ""))
			if bool(item.get("consumable", false)) and int(current_inventory.get(item_id, 0)) > 0:
				inventory_use_select.add_item(str(item.get("name", item_id)))
				inventory_use_select.set_item_metadata(inventory_use_select.item_count - 1, item_id)
	column.add_child(inventory_use_select)
	var use_button := Button.new()
	use_button.text = "Use selected consumable"
	use_button.disabled = inventory_use_select.item_count == 0
	use_button.pressed.connect(_use_selected_inventory_item)
	column.add_child(use_button)
	var close_button := Button.new()
	close_button.text = "Close"
	close_button.pressed.connect(func() -> void: inventory_panel.visible = false)
	column.add_child(close_button)
	auth_panel.get_parent().add_child(inventory_panel)


func _refresh_equipment_visuals() -> void:
	if equipment_visual_root == null:
		return
	for child in equipment_visual_root.get_children():
		child.queue_free()
	var leather := StandardMaterial3D.new()
	leather.albedo_color = Color(0.30, 0.17, 0.09)
	var metal := StandardMaterial3D.new()
	metal.albedo_color = Color(0.62, 0.66, 0.70)
	var cloth := StandardMaterial3D.new()
	cloth.albedo_color = Color(0.35, 0.25, 0.16)
	if current_equipment.has("main_hand"):
		var sword := Node3D.new()
		sword.position = Vector3(0.48, 0.95, -0.05)
		sword.rotation_degrees.z = -22.0
		var blade := MeshInstance3D.new()
		var blade_mesh := BoxMesh.new()
		blade_mesh.size = Vector3(0.10, 0.85, 0.07)
		blade.mesh = blade_mesh
		blade.position.y = 0.42
		blade.material_override = metal
		sword.add_child(blade)
		var grip := MeshInstance3D.new()
		var grip_mesh := BoxMesh.new()
		grip_mesh.size = Vector3(0.12, 0.24, 0.12)
		grip.mesh = grip_mesh
		grip.position.y = -0.12
		grip.material_override = leather
		sword.add_child(grip)
		equipment_visual_root.add_child(sword)
	if current_equipment.has("head"):
		var cap := MeshInstance3D.new()
		var cap_mesh := SphereMesh.new()
		cap_mesh.radius = 0.36
		cap_mesh.height = 0.34
		cap.mesh = cap_mesh
		cap.position.y = 1.82
		cap.scale = Vector3(1.0, 0.55, 1.0)
		cap.material_override = leather
		equipment_visual_root.add_child(cap)
	if current_equipment.has("body"):
		var tunic := MeshInstance3D.new()
		var tunic_mesh := CylinderMesh.new()
		tunic_mesh.top_radius = 0.39
		tunic_mesh.bottom_radius = 0.48
		tunic_mesh.height = 0.92
		tunic.mesh = tunic_mesh
		tunic.position.y = 0.92
		tunic.material_override = cloth
		equipment_visual_root.add_child(tunic)
	if current_equipment.has("feet"):
		for side in [-1.0, 1.0]:
			var boot := MeshInstance3D.new()
			var boot_mesh := BoxMesh.new()
			boot_mesh.size = Vector3(0.32, 0.24, 0.48)
			boot.mesh = boot_mesh
			boot.position = Vector3(side * 0.22, 0.12, 0.08)
			boot.material_override = leather
			equipment_visual_root.add_child(boot)


func _append_chat_line(line: String) -> void:
	chat_lines.append(line)
	while chat_lines.size() > 100:
		chat_lines.pop_front()
	if chat_history == null:
		return
	chat_history.clear()
	for entry in chat_lines:
		chat_history.append_text(entry + "\n")


func _use_selected_inventory_item() -> void:
	var item_id := str(inventory_use_select.get_selected_metadata())
	if item_id.is_empty():
		return
	interaction_type = "use_item"
	var error := gather_request.request(
		api_base_url + "/api/v1/account/characters/" + player_id + "/inventory/use/" + item_id,
		_api_headers(),
		HTTPClient.METHOD_POST,
		""
	)
	if error == OK:
		interaction_pending = true
		status_label.text = "Using " + item_id.replace("_", " ") + "..."
	else:
		status_label.text = "Could not use item."


func _toggle_skills_panel() -> void:
	if skills_panel != null and skills_panel.visible:
		skills_panel.visible = false
		return
	if market_panel != null:
		market_panel.visible = false
	if inventory_panel != null:
		inventory_panel.visible = false
	if equipment_panel != null:
		equipment_panel.visible = false
	if skills_panel == null:
		skills_panel = PanelContainer.new()
		skills_panel.position = Vector2(420, 155)
		skills_panel.custom_minimum_size = Vector2(330, 380)
		var column := VBoxContainer.new()
		skills_panel.add_child(column)
		var heading := Label.new()
		heading.text = "SKILLS"
		heading.add_theme_font_size_override("font_size", 18)
		column.add_child(heading)
		var scroll := ScrollContainer.new()
		scroll.custom_minimum_size = Vector2(300, 280)
		scroll.size_flags_vertical = Control.SIZE_EXPAND_FILL
		column.add_child(scroll)
		skills_list = VBoxContainer.new()
		skills_list.size_flags_horizontal = Control.SIZE_EXPAND_FILL
		scroll.add_child(skills_list)
		skills_status_label = Label.new()
		column.add_child(skills_status_label)
		var close_button := Button.new()
		close_button.text = "Close"
		close_button.pressed.connect(func() -> void: skills_panel.visible = false)
		column.add_child(close_button)
		auth_panel.get_parent().add_child(skills_panel)
	skills_panel.visible = true
	if skills_request == null:
		skills_request = HTTPRequest.new()
		add_child(skills_request)
		skills_request.request_completed.connect(_on_skills_completed)
	_refresh_skills_panel()

func _refresh_skills_panel() -> void:
	if skills_panel == null or not skills_panel.visible or skills_pending or skills_request == null:
		return
	var error := skills_request.request(
		api_base_url + "/api/v1/account/characters/" + player_id + "/skills",
		_api_headers()
	)
	if error == OK:
		skills_pending = true
		skills_status_label.text = "Loading skills..."
	else:
		skills_status_label.text = "Could not request skills."

func _on_skills_completed(result: int, response_code: int, _headers: PackedStringArray, body: PackedByteArray) -> void:
	skills_pending = false
	if player_id.is_empty():
		return
	if response_code == 401:
		_expire_session()
		return
	if result != HTTPRequest.RESULT_SUCCESS or response_code != 200:
		skills_status_label.text = "Skills unavailable"
		return
	var payload: Variant = JSON.parse_string(body.get_string_from_utf8())
	if not payload is Dictionary:
		return
	var skills: Variant = payload.get("skills", {})
	if not skills is Dictionary:
		return
	for child in skills_list.get_children():
		child.queue_free()
	var skill_ids: Array = skills.keys()
	skill_ids.sort()
	for skill_id in skill_ids:
		var progress: Variant = skills[skill_id]
		if not progress is Dictionary:
			continue
		var level := clampi(int(progress.get("level", 1)), 1, 99)
		var experience := maxi(int(progress.get("experience", 0)), 0)
		var skill_column := VBoxContainer.new()
		skill_column.size_flags_horizontal = Control.SIZE_EXPAND_FILL
		var skill_label := Label.new()
		skill_label.text = str(skill_id).capitalize() + "  •  Level " + str(level)
		skill_column.add_child(skill_label)
		var progress_bar := ProgressBar.new()
		progress_bar.custom_minimum_size = Vector2(0, 12)
		progress_bar.size_flags_horizontal = Control.SIZE_EXPAND_FILL
		progress_bar.show_percentage = false
		if level >= 99:
			progress_bar.min_value = 0
			progress_bar.max_value = 200000000
			progress_bar.value = clampi(experience, 0, 200000000)
		else:
			var level_start := _experience_for_level(level)
			var next_level := _experience_for_level(level + 1)
			progress_bar.min_value = 0
			progress_bar.max_value = maxi(next_level - level_start, 1)
			progress_bar.value = clampi(experience - level_start, 0, int(progress_bar.max_value))
		skill_column.add_child(progress_bar)
		var detail := Label.new()
		if level >= 99:
			detail.text = str(experience) + " XP • " + str(maxi(200000000 - experience, 0)) + " to XP cap"
		else:
			var remaining := maxi(_experience_for_level(level + 1) - experience, 0)
			detail.text = str(experience) + " XP • " + str(remaining) + " to level " + str(level + 1)
		detail.add_theme_font_size_override("font_size", 11)
		skill_column.add_child(detail)
		skills_list.add_child(skill_column)
	skills_status_label.text = str(skills.size()) + " skills • progression updates after server actions"


func _experience_for_level(level: int) -> int:
	if level <= 1:
		return 0
	var total := 0
	for current_level in range(1, mini(level, 99)):
		total += floori(float(current_level) + 300.0 * pow(2.0, float(current_level) / 7.0))
	return floori(float(total) / 4.0)


func _stop_skills_request() -> void:
	if skills_request != null:
		if skills_request.get_http_client_status() != HTTPClient.STATUS_DISCONNECTED:
			skills_request.cancel_request()
		skills_request.queue_free()
		skills_request = null
	skills_pending = false


func _submit_bank_transfer(deposit: bool) -> void:
	if interaction_pending:
		return
	var item_id := str(inventory_bank_select.get_selected_metadata())
	var quantity := int(inventory_quantity_input.text)
	if item_id.is_empty():
		inventory_status_label.text = "Select an item first."
		return
	if quantity < 1 or quantity > 100:
		inventory_status_label.text = "Quantity must be between 1 and 100."
		return
	interaction_type = "bank_deposit" if deposit else "bank_withdraw"
	var action := "deposit" if deposit else "withdraw"
	var error := gather_request.request(
		api_base_url + "/api/v1/account/characters/" + player_id + "/bank/" + action,
		_api_headers(),
		HTTPClient.METHOD_POST,
		JSON.stringify({"itemId": item_id, "quantity": quantity})
	)
	if error == OK:
		interaction_pending = true
		inventory_status_label.text = "Transferring " + str(quantity) + " × " + item_id.replace("_", " ") + "..."
	else:
		inventory_status_label.text = "Could not send bank request."
