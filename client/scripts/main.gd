extends Node3D

const PLAYER_SCENE := preload("res://scenes/player.tscn")
const WORLD_TERRAIN_SCRIPT := preload("res://scripts/world_terrain.gd")
const CORE_SKILL_IDS := ["blades", "cooking", "crafting", "defense", "farming", "foraging", "magic", "melee", "mining", "pistols", "polearms", "ranged", "rifles", "scavenging", "smithing", "strength", "unarmed", "woodcutting"]

var status_label: Label
var interaction_hint_label: Label
var hud_info_panel: PanelContainer
var action_toolbar: GridContainer
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
var remote_player_names: Dictionary = {}
var enemy_nodes: Dictionary = {}
var enemy_targets: Dictionary = {}
var enemy_labels: Dictionary = {}
var enemy_available: Dictionary = {}
var enemy_snapshot_request: HTTPRequest
var enemy_snapshot_timer: Timer
var enemy_snapshot_pending := false
var boss_hud_panel: PanelContainer
var boss_hud_name_label: Label
var boss_hud_phase_label: Label
var boss_hud_health_bar: ProgressBar
var boss_hud_ability_label: Label
var boss_hud_status_label: Label
var boss_last_phase_id := ""
var boss_phase_feedback_text := ""
var boss_phase_feedback_until_msec := 0
var trade_button: Button
var trade_panel: PanelContainer
var trade_target_select: OptionButton
var trade_item_select: OptionButton
var trade_quantity_input: LineEdit
var trade_coins_input: LineEdit
var trade_offer_items: Dictionary = {}
var trade_offer_items_label: Label
var trade_editor_trade_id := ""
var trade_editor_revision := -1
var trade_status_label: Label
var trade_invite_button: Button
var trade_accept_button: Button
var trade_update_button: Button
var trade_confirm_button: Button
var trade_cancel_button: Button
var trade_request: HTTPRequest
var trade_refresh_request: HTTPRequest
var trade_refresh_pending := false
var trade_poll_timer: Timer
var trade_pending := false
var trade_action := "poll"
var current_trade: Dictionary = {}
var rewards_button: Button
var rewards_panel: PanelContainer
var rewards_claims_column: VBoxContainer
var rewards_status_label: Label
var rewards_request: HTTPRequest
var rewards_pending := false
var rewards_action := "list"
const game_max_health := 100
var market_catalog: Array = []
var enemy_catalog: Array = []
var gathering_catalog: Array = []
var station_catalog: Array = []
var api_request: HTTPRequest
var movement_request: HTTPRequest
var gather_request: HTTPRequest
var movement_timer: Timer
var player_node: CharacterBody3D
var world_terrain: Node3D
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
	_update_enemy_visuals(delta)
	_update_interaction_hint()
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

func _terrain_height_at(x: float, z: float) -> float:
	if world_terrain == null or not is_instance_valid(world_terrain):
		return 0.0
	return float(world_terrain.call("get_height_at", x, z))


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

	world_terrain = WORLD_TERRAIN_SCRIPT.new() as Node3D
	world_terrain.name = "WorldTerrain"
	add_child(world_terrain)

	_add_bank_chest(Vector3(0.0, _terrain_height_at(0.0, -4.0), -4.0))
	_add_merchant_stall(Vector3(4.0, _terrain_height_at(4.0, 0.0), 0.0))

	player_node = PLAYER_SCENE.instantiate() as CharacterBody3D
	player_node.position = Vector3(0.0, _terrain_height_at(0.0, 0.0) + 0.05, 0.0)
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





func _build_crafting_station_visuals() -> void:
	for station in station_catalog:
		if not station is Dictionary:
			continue
		var position_data: Variant = station.get("position", {})
		if not position_data is Dictionary:
			continue
		var at := Vector3(float(position_data.get("x", 0.0)), 0.0, float(position_data.get("z", 0.0)))
		at.y = _terrain_height_at(at.x, at.z)
		match str(station.get("id", "")):
			"frontier-workbench":
				_add_workbench(at)
			"frontier-forge":
				_add_forge(at)
			_:
				_add_generic_gathering_node(at, str(station.get("name", "Crafting Station")))


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
			"wild_herb":
				_add_herb_patch(at)
			_:
				_add_generic_gathering_node(at, str(node.get("name", "Resource Node")))

func _add_herb_patch(at: Vector3) -> void:
	var patch := Node3D.new()
	patch.position = at
	var stem_material := StandardMaterial3D.new()
	stem_material.albedo_color = Color(0.18, 0.38, 0.16)
	var leaf_material := StandardMaterial3D.new()
	leaf_material.albedo_color = Color(0.28, 0.58, 0.22)
	var stem := MeshInstance3D.new()
	var stem_mesh := CylinderMesh.new()
	stem_mesh.top_radius = 0.035
	stem_mesh.bottom_radius = 0.06
	stem_mesh.height = 0.55
	stem.mesh = stem_mesh
	stem.position.y = 0.27
	stem.material_override = stem_material
	patch.add_child(stem)
	for index in range(5):
		var leaf := MeshInstance3D.new()
		var leaf_mesh := SphereMesh.new()
		leaf_mesh.radius = 0.22
		leaf_mesh.height = 0.18
		leaf.mesh = leaf_mesh
		leaf.scale = Vector3(0.85, 0.5, 1.5)
		var angle := float(index) * TAU / 5.0
		leaf.position = Vector3(cos(angle) * 0.18, 0.18 + float(index % 2) * 0.08, sin(angle) * 0.18)
		leaf.rotation.y = angle
		leaf.material_override = leaf_material
		patch.add_child(leaf)
	var label := Label3D.new()
	label.text = "Wild Herbs"
	label.position.y = 0.8
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	patch.add_child(label)
	add_child(patch)


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
		var enemy_id := str(enemy.get("id", ""))
		if enemy_id.is_empty():
			continue
		var at := Vector3(float(position_data.get("x", 0.0)), 0.0, float(position_data.get("z", 0.0)))
		_add_enemy_visual(enemy_id, at, str(enemy.get("name", "Frontier Creature")), int(enemy.get("maxHealth", 10)))

func _add_enemy_visual(enemy_id: String, at: Vector3, enemy_name: String, max_health: int) -> void:
	var creature := Node3D.new()
	creature.name = "Enemy_" + enemy_id
	creature.position = at
	var is_boss := enemy_id == "stoneback-colossus"
	var hide_material := StandardMaterial3D.new()
	hide_material.albedo_color = Color(0.27, 0.29, 0.30) if is_boss else (Color(0.48, 0.30, 0.18) if max_health < 20 else Color(0.32, 0.25, 0.18))
	hide_material.roughness = 0.92
	var body := MeshInstance3D.new()
	var body_mesh := CapsuleMesh.new()
	body_mesh.radius = 1.25 if is_boss else (0.42 if max_health < 20 else 0.7)
	body_mesh.height = 2.5 if is_boss else (0.9 if max_health < 20 else 1.5)
	body.mesh = body_mesh
	body.position.y = body_mesh.height * 0.5
	body.material_override = hide_material
	creature.add_child(body)
	var head := MeshInstance3D.new()
	var head_mesh := SphereMesh.new()
	head_mesh.radius = 0.72 if is_boss else (0.28 if max_health < 20 else 0.42)
	head_mesh.height = head_mesh.radius * 2.0
	head.mesh = head_mesh
	head.position = Vector3(0.0, body_mesh.height * 0.78, -0.8 if is_boss else -0.35)
	head.material_override = hide_material
	creature.add_child(head)
	if is_boss:
		var plate_material := StandardMaterial3D.new()
		plate_material.albedo_color = Color(0.42, 0.44, 0.45)
		plate_material.roughness = 0.88
		for index in range(3):
			var plate := MeshInstance3D.new()
			var plate_mesh := BoxMesh.new()
			plate_mesh.size = Vector3(1.15, 0.42, 0.72)
			plate.mesh = plate_mesh
			plate.position = Vector3(0.0, 1.55 + float(index) * 0.42, 0.25 + float(index) * 0.32)
			plate.rotation.x = -0.25
			plate.material_override = plate_material
			creature.add_child(plate)
		for side in [-1.0, 1.0]:
			var fist := MeshInstance3D.new()
			var fist_mesh := SphereMesh.new()
			fist_mesh.radius = 0.58
			fist_mesh.height = 1.05
			fist.mesh = fist_mesh
			fist.position = Vector3(side * 1.35, 1.0, 0.05)
			fist.material_override = hide_material
			creature.add_child(fist)
	var label := Label3D.new()
	label.text = enemy_name + "\n" + str(max_health) + "/" + str(max_health) + " • Idle"
	label.position.y = body_mesh.height + 0.45
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 32 if is_boss else 24
	creature.add_child(label)
	add_child(creature)
	enemy_nodes[enemy_id] = creature
	enemy_targets[enemy_id] = at
	enemy_labels[enemy_id] = label
	enemy_available[enemy_id] = true

func _start_enemy_snapshot_polling() -> void:
	_stop_enemy_snapshot_polling()
	if player_id.is_empty() or auth_token.is_empty():
		return
	enemy_snapshot_request = _new_api_request(5.0)
	add_child(enemy_snapshot_request)
	enemy_snapshot_request.request_completed.connect(_on_enemy_snapshots_completed)
	enemy_snapshot_timer = Timer.new()
	enemy_snapshot_timer.wait_time = 0.5
	enemy_snapshot_timer.timeout.connect(_poll_enemy_snapshots)
	add_child(enemy_snapshot_timer)
	enemy_snapshot_timer.start()
	_poll_enemy_snapshots()

func _stop_enemy_snapshot_polling() -> void:
	if enemy_snapshot_timer != null:
		enemy_snapshot_timer.stop()
		enemy_snapshot_timer.queue_free()
		enemy_snapshot_timer = null
	if enemy_snapshot_request != null:
		if enemy_snapshot_request.get_http_client_status() != HTTPClient.STATUS_DISCONNECTED:
			enemy_snapshot_request.cancel_request()
		enemy_snapshot_request.queue_free()
		enemy_snapshot_request = null
	enemy_snapshot_pending = false

func _poll_enemy_snapshots() -> void:
	if player_id.is_empty() or auth_token.is_empty() or enemy_snapshot_pending or enemy_snapshot_request == null:
		return
	var url := api_base_url + "/api/v1/account/characters/" + player_id + "/world/enemies"
	var error := enemy_snapshot_request.request(url, _api_headers())
	if error == OK:
		enemy_snapshot_pending = true

func _on_enemy_snapshots_completed(result: int, response_code: int, _headers: PackedStringArray, body: PackedByteArray) -> void:
	enemy_snapshot_pending = false
	if player_id.is_empty():
		return
	if response_code == 401:
		_expire_session()
		return
	if response_code == 404:
		_stop_enemy_snapshot_polling()
		return
	if result != HTTPRequest.RESULT_SUCCESS or response_code != 200:
		return
	var payload: Variant = JSON.parse_string(body.get_string_from_utf8())
	if not payload is Dictionary:
		return
	var snapshots: Variant = payload.get("enemies", [])
	if not snapshots is Array:
		return
	var boss_seen := false
	for snapshot in snapshots:
		if not snapshot is Dictionary:
			continue
		var enemy_id := str(snapshot.get("enemyId", ""))
		if not enemy_nodes.has(enemy_id):
			continue
		var target := Vector3(float(snapshot.get("x", 0.0)), 0.0, float(snapshot.get("z", 0.0)))
		target.y = _terrain_height_at(target.x, target.z)
		enemy_targets[enemy_id] = target
		var health := int(snapshot.get("health", 0))
		var maximum := int(snapshot.get("maxHealth", 1))
		enemy_available[enemy_id] = health > 0
		var behavior := str(snapshot.get("behavior", "idle"))
		var creature: Node3D = enemy_nodes[enemy_id]
		for visual_child in creature.get_children():
			if visual_child is MeshInstance3D:
				visual_child.visible = health > 0
		var label: Label3D = enemy_labels[enemy_id]
		var display_name := str(snapshot.get("name", enemy_id))
		var boss_id := str(snapshot.get("bossId", ""))
		var boss_phase := str(snapshot.get("bossPhaseId", ""))
		var boss_ability := str(snapshot.get("bossActiveAbilityId", ""))
		var boss_phase_changed := not boss_phase.is_empty() and not boss_last_phase_id.is_empty() and boss_phase != boss_last_phase_id
		if boss_phase_changed and health > 0:
			boss_phase_feedback_text = "PHASE SHIFT — " + boss_phase.replace("-", " ").to_upper()
			boss_phase_feedback_until_msec = Time.get_ticks_msec() + 2500
		if not boss_id.is_empty():
			# Keep the encounter panel alive through defeat. The zero-health
			# snapshot is authoritative and should not disappear before players
			# can read the completion/reward feedback.
			boss_seen = true
			if boss_hud_panel != null:
				boss_hud_panel.visible = true
				boss_hud_name_label.text = display_name.to_upper()
				boss_hud_health_bar.max_value = maxi(maximum, 1)
				boss_hud_health_bar.value = clampi(health, 0, maxi(maximum, 1))
				if health <= 0:
					boss_hud_phase_label.text = "ENCOUNTER COMPLETE"
					boss_hud_ability_label.text = "DEFEATED"
					boss_hud_status_label.text = "Rewards are ready to claim from the Rewards panel."
					boss_last_phase_id = ""
					boss_phase_feedback_text = ""
					boss_phase_feedback_until_msec = 0
				elif not boss_ability.is_empty():
					boss_hud_phase_label.text = boss_phase.replace("-", " ").to_upper() if not boss_phase.is_empty() else "ENCOUNTER"
					boss_hud_ability_label.text = "⚠ " + boss_ability.replace("-", " ").to_upper() + " — MOVE NOW"
					boss_hud_status_label.text = "Telegraph active • leave the boss's attack range."
				else:
					boss_hud_phase_label.text = boss_phase.replace("-", " ").to_upper() if not boss_phase.is_empty() else "ENCOUNTER"
					boss_hud_ability_label.text = "No attack telegraphed"
					if Time.get_ticks_msec() < boss_phase_feedback_until_msec:
						boss_hud_status_label.text = boss_phase_feedback_text
					else:
						boss_hud_status_label.text = "Watch the boss for its next tell."
				boss_last_phase_id = boss_phase if health > 0 else ""
		if health <= 0:
			label.text = display_name + "\nRespawning..."
		elif not boss_id.is_empty():
			var boss_status := ""
			if not boss_phase.is_empty():
				boss_status += " • " + boss_phase.replace("-", " ").capitalize()
			if not boss_ability.is_empty():
				boss_status += " • ⚠ " + boss_ability.replace("-", " ").to_upper()
			label.text = display_name + "\n" + str(health) + "/" + str(maximum) + boss_status
		elif behavior == "telegraphing":
			label.text = display_name + "\n" + str(health) + "/" + str(maximum) + " • ⚠ ATTACK"
		else:
			label.text = display_name + "\n" + str(health) + "/" + str(maximum) + " • " + behavior.capitalize()
	if not boss_seen and boss_hud_panel != null:
		boss_hud_panel.visible = false

func _update_enemy_visuals(delta: float) -> void:
	var blend := 1.0 - exp(-delta * 10.0)
	for enemy_id in enemy_nodes.keys():
		var creature: Node3D = enemy_nodes[enemy_id]
		if not is_instance_valid(creature):
			continue
		var target: Vector3 = enemy_targets.get(enemy_id, creature.position)
		if creature.position.distance_to(target) > 20.0:
			creature.position = target
		else:
			creature.position = creature.position.lerp(target, blend)

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
	hud_info_panel = panel
	layer.add_child(panel)

	var column := VBoxContainer.new()
	panel.add_child(column)
	var heading := Label.new()
	heading.text = "RS GALAXIES  /  WORLD PROTOTYPE"
	heading.add_theme_font_size_override("font_size", 18)
	column.add_child(heading)

	var instructions := Label.new()
	instructions.text = "WASD / pad: move  •  E / USE: gather or craft  •  Drag / mouse: look"
	instructions.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	column.add_child(instructions)

	status_label = Label.new()
	status_label.text = "Server: connecting..."
	status_label.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	column.add_child(status_label)
	interaction_hint_label = Label.new()
	interaction_hint_label.name = "InteractionHint"
	interaction_hint_label.text = "Interact: sign in to begin"
	interaction_hint_label.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	interaction_hint_label.add_theme_font_size_override("font_size", 14)
	interaction_hint_label.add_theme_color_override("font_color", Color(0.78, 0.90, 0.70))
	column.add_child(interaction_hint_label)
	presence_label = Label.new()
	presence_label.text = "Nearby players: offline"
	presence_label.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	column.add_child(presence_label)

	inventory_label = Label.new()
	inventory_label.text = "Inventory: empty"
	inventory_label.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	column.add_child(inventory_label)

	_build_chat_panel(layer)
	_build_auth_panel(layer)
	action_toolbar = GridContainer.new()
	action_toolbar.name = "ActionToolbar"
	action_toolbar.columns = 2
	action_toolbar.add_theme_constant_override("h_separation", 6)
	action_toolbar.add_theme_constant_override("v_separation", 6)
	action_toolbar.anchor_left = 1.0
	action_toolbar.anchor_right = 1.0
	action_toolbar.anchor_top = 0.0
	action_toolbar.anchor_bottom = 0.0
	action_toolbar.offset_top = 18
	action_toolbar.offset_bottom = 132
	layer.add_child(action_toolbar)

	skills_button = _create_action_button(action_toolbar, "Skills", _toggle_skills_panel)
	skills_button.visible = false
	bag_button = _create_action_button(action_toolbar, "Bag", _show_inventory_panel)
	bag_button.visible = false
	rest_button = _create_action_button(action_toolbar, "Rest", _rest_character)
	rest_button.visible = false
	equipment_button = _create_action_button(action_toolbar, "Gear", _show_equipment_panel)
	equipment_button.visible = false
	character_button = _create_action_button(action_toolbar, "Chars", _request_character_list)
	character_button.visible = false
	logout_button = _create_action_button(action_toolbar, "Sign out", _logout)
	logout_button.visible = false

	trade_button = Button.new()
	trade_button.text = "Trade"
	trade_button.anchor_left = 0.0
	trade_button.anchor_right = 0.0
	trade_button.offset_left = 18
	trade_button.offset_top = 145
	trade_button.offset_right = 118
	trade_button.offset_bottom = 189
	trade_button.visible = false
	trade_button.pressed.connect(_toggle_trade_panel)
	layer.add_child(trade_button)
	rewards_button = Button.new()
	rewards_button.text = "Rewards"
	rewards_button.anchor_left = 0.0
	rewards_button.anchor_right = 0.0
	rewards_button.offset_left = 18
	rewards_button.offset_top = 195
	rewards_button.offset_right = 118
	rewards_button.offset_bottom = 239
	rewards_button.visible = false
	rewards_button.pressed.connect(_show_rewards_panel)
	layer.add_child(rewards_button)

	if DisplayServer.is_touchscreen_available():
		_add_hold_button(layer, "▲", "move_forward", 95, -230, 165, -160)
		_add_hold_button(layer, "◀", "move_left", 20, -155, 90, -85)
		_add_hold_button(layer, "▼", "move_back", 95, -155, 165, -85)
		_add_hold_button(layer, "▶", "move_right", 170, -155, 240, -85)
		_add_hold_button(layer, "USE", "interact", -125, -230, -25, -130, true)
		_add_hold_button(layer, "JUMP", "jump", -125, -120, -25, -25, true)

	_build_boss_hud(layer)
	add_child(layer)
	get_viewport().size_changed.connect(_update_hud_layout)
	_update_hud_layout()


func _create_action_button(parent: GridContainer, caption: String, action: Callable) -> Button:
	var button := Button.new()
	button.text = caption
	button.custom_minimum_size = Vector2(0.0, 34.0)
	button.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	button.add_theme_font_size_override("font_size", 12)
	button.pressed.connect(action)
	parent.add_child(button)
	return button


func _update_hud_layout() -> void:
	if hud_info_panel == null or action_toolbar == null:
		return
	var viewport_width: float = get_viewport().get_visible_rect().size.x
	var info_width := minf(390.0, maxf(140.0, viewport_width * 0.42))
	var toolbar_width := minf(224.0, maxf(164.0, viewport_width * 0.46))
	hud_info_panel.custom_minimum_size.x = info_width
	action_toolbar.offset_left = -(toolbar_width + 18.0)
	action_toolbar.offset_right = -18.0


func _build_boss_hud(layer: CanvasLayer) -> void:
	boss_hud_panel = PanelContainer.new()
	# Use viewport-relative margins so the encounter panel fits narrow
	# mobile screens as well as desktop windows.
	boss_hud_panel.anchor_left = 0.05
	boss_hud_panel.anchor_right = 0.95
	boss_hud_panel.offset_left = 0
	boss_hud_panel.offset_right = 0
	boss_hud_panel.offset_top = 78
	boss_hud_panel.offset_bottom = 194
	boss_hud_panel.visible = false
	var column := VBoxContainer.new()
	boss_hud_panel.add_child(column)
	boss_hud_name_label = Label.new()
	boss_hud_name_label.text = "BOSS ENCOUNTER"
	boss_hud_name_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	boss_hud_name_label.add_theme_font_size_override("font_size", 18)
	column.add_child(boss_hud_name_label)
	boss_hud_health_bar = ProgressBar.new()
	boss_hud_health_bar.custom_minimum_size = Vector2(0, 18)
	boss_hud_health_bar.show_percentage = false
	column.add_child(boss_hud_health_bar)
	boss_hud_phase_label = Label.new()
	boss_hud_phase_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	column.add_child(boss_hud_phase_label)
	boss_hud_ability_label = Label.new()
	boss_hud_ability_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	boss_hud_ability_label.add_theme_font_size_override("font_size", 15)
	column.add_child(boss_hud_ability_label)
	boss_hud_status_label = Label.new()
	boss_hud_status_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	boss_hud_status_label.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	boss_hud_status_label.add_theme_font_size_override("font_size", 12)
	column.add_child(boss_hud_status_label)
	layer.add_child(boss_hud_panel)

func _add_hold_button(parent: Node, caption: String, action: String, left: float, top: float, right: float, bottom: float, anchor_right: bool = false) -> void:
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
		api_stage = "world_catalog"
		var catalog_error := api_request.request(api_base_url + "/api/v1/world/catalog")
		if catalog_error != OK:
			status_label.text = "Server: ready • world catalog unavailable"
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
	if api_stage == "world_catalog":
		var catalog_items: Variant = payload.get("items", [])
		var enemies: Variant = payload.get("enemies", [])
		var nodes: Variant = payload.get("nodes", [])
		var stations: Variant = payload.get("stations", [])
		if catalog_items is Array:
			market_catalog = catalog_items
		if enemies is Array:
			enemy_catalog = enemies
			_build_enemy_spawns()
		if nodes is Array:
			gathering_catalog = nodes
			_build_gathering_node_visuals()
		if stations is Array:
			station_catalog = stations
			_build_crafting_station_visuals()
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

func _update_interaction_hint() -> void:
	if interaction_hint_label == null:
		return
	if player_id.is_empty() or player_node == null or not is_instance_valid(player_node):
		interaction_hint_label.visible = false
		return

	interaction_hint_label.visible = true
	if interaction_pending:
		interaction_hint_label.text = "Working on your interaction..."
		return

	var player_position: Vector3 = player_node.global_position
	var best_distance := INF
	var best_action := ""
	var best_name := ""

	# Keep the hint's candidate ordering aligned with the E/USE interaction
	# resolver: trader, bank, station, enemy, then gathering node.
	var merchant_distance := Vector2(player_position.x - 4.0, player_position.z).length()
	if merchant_distance < best_distance:
		best_distance = merchant_distance
		best_action = "Trade with"
		best_name = "Frontier Trader"

	var bank_distance := Vector2(player_position.x, player_position.z + 4.0).length()
	if bank_distance < best_distance:
		best_distance = bank_distance
		best_action = "Open"
		best_name = "Frontier Bank"

	for station in station_catalog:
		if not station is Dictionary:
			continue
		var station_position: Variant = station.get("position", {})
		if not station_position is Dictionary:
			continue
		var distance := Vector2(
			player_position.x - float(station_position.get("x", 0.0)),
			player_position.z - float(station_position.get("z", 0.0))
		).length()
		if distance < best_distance:
			best_distance = distance
			best_action = "Craft at"
			best_name = str(station.get("name", "Crafting Station"))

	for enemy in enemy_catalog:
		if not enemy is Dictionary:
			continue
		var enemy_id := str(enemy.get("id", ""))
		if not bool(enemy_available.get(enemy_id, true)):
			continue
		var enemy_position: Variant = enemy.get("position", {})
		if not enemy_position is Dictionary:
			continue
		var fallback_position := Vector3(
			float(enemy_position.get("x", 0.0)),
			0.0,
			float(enemy_position.get("z", 0.0))
		)
		var target_position: Vector3 = enemy_targets.get(enemy_id, fallback_position)
		var enemy_distance := Vector2(
			player_position.x - target_position.x,
			player_position.z - target_position.z
		).length()
		if enemy_distance < best_distance:
			best_distance = enemy_distance
			best_action = "Attack"
			best_name = str(enemy.get("name", "Creature"))

	for node in gathering_catalog:
		if not node is Dictionary:
			continue
		var node_position: Variant = node.get("position", {})
		if not node_position is Dictionary:
			continue
		var node_distance := Vector2(
			player_position.x - float(node_position.get("x", 0.0)),
			player_position.z - float(node_position.get("z", 0.0))
		).length()
		if node_distance < best_distance:
			best_distance = node_distance
			best_action = "Gather"
			best_name = str(node.get("name", "Resource"))

	if best_action.is_empty():
		interaction_hint_label.text = "Explore • no nearby interaction"
		return

	if best_distance <= 3.0:
		var input_name := "USE" if DisplayServer.is_touchscreen_available() else "E"
		interaction_hint_label.text = input_name + " • " + best_action + " " + best_name
	else:
		interaction_hint_label.text = "Move closer • " + best_name + " (" + str(snappedf(best_distance, 0.1)) + "m)"


func _request_interaction() -> void:
	var bank_distance := Vector2(player_node.global_position.x, player_node.global_position.z + 4.0).length()
	var merchant_distance := Vector2(player_node.global_position.x - 4.0, player_node.global_position.z).length()
	var nearest_station_id := ""
	var nearest_station_distance := 1.0e20
	for station in station_catalog:
		if station is Dictionary:
			var station_position: Variant = station.get("position", {})
			if station_position is Dictionary:
				var station_distance := Vector2(player_node.global_position.x - float(station_position.get("x", 0.0)), player_node.global_position.z - float(station_position.get("z", 0.0))).length()
				if station_distance < nearest_station_distance:
					nearest_station_distance = station_distance
					nearest_station_id = str(station.get("id", ""))
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
			var enemy_id := str(enemy.get("id", ""))
			if not bool(enemy_available.get(enemy_id, true)):
				continue
			var enemy_position: Variant = enemy.get("position", {})
			if enemy_position is Dictionary:
				var target_position: Vector3 = enemy_targets.get(enemy_id, Vector3(float(enemy_position.get("x", 0.0)), 0.0, float(enemy_position.get("z", 0.0))))
				var distance := Vector2(
					player_node.global_position.x - target_position.x,
					player_node.global_position.z - target_position.z
				).length()
				if distance < enemy_distance:
					enemy_distance = distance
					nearest_enemy_id = enemy_id
	var error := OK
	if merchant_distance <= 3.0 and merchant_distance <= bank_distance and merchant_distance <= nearest_station_distance and merchant_distance <= nearest_node_distance and merchant_distance <= enemy_distance:
		_show_market_panel()
		return
	elif bank_distance <= 3.0 and bank_distance <= nearest_station_distance and bank_distance <= nearest_node_distance and bank_distance <= enemy_distance:
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
	elif nearest_station_id == "frontier-forge" and nearest_station_distance <= 3.0 and nearest_station_distance <= nearest_node_distance and nearest_station_distance <= enemy_distance:
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
	elif nearest_station_id == "frontier-workbench" and nearest_station_distance <= 3.0 and nearest_station_distance <= nearest_node_distance and nearest_station_distance <= enemy_distance:
		var recipe_id := "basic-plank"
		if int(current_inventory.get("herbal_bundle", 0)) >= 1 and int(current_inventory.get("wild_herb", 0)) >= 1:
			recipe_id = "herbal-tea"
		elif int(current_inventory.get("wild_herb", 0)) >= 3 and int(current_inventory.get("wooden_plank", 0)) >= 1:
			recipe_id = "herbal-bundle"
		interaction_type = "craft"
		status_label.text = "Using frontier workbench..."
		error = gather_request.request(
			api_base_url + "/api/v1/account/characters/" + player_id + "/craft/" + recipe_id,
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
		if player_node != null and is_instance_valid(player_node):
			player_node.call("play_attack_animation")
		_update_inventory(player_data.get("inventory", {}), player_data.get("bank", {}), player_data.get("coins", current_coins), player_data.get("equipment", {}), player_data.get("health", current_health))
		var combat_message := "Hit for " + str(combat.get("damageDealt", 0)) + " • Enemy HP " + str(combat.get("enemyHealth", 0)) + " • Your HP " + str(current_health)
		if bool(combat.get("enemyDefeated", false)):
			combat_message = "Enemy defeated • +" + str(combat.get("experience", 0)) + " " + str(combat.get("skillId", "combat")).capitalize() + " XP"
			if not str(combat.get("parentSkillId", "")).is_empty():
				combat_message += " • +" + str(combat.get("parentExperience", 0)) + " " + str(combat.get("parentSkillId", "")).capitalize() + " XP"
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
	if trade_item_select != null:
		_refresh_trade_items()

func _apply_server_position(position_data: Dictionary, force: bool) -> void:
	var target_x := float(position_data.get("x", 0.0))
	var target_z := float(position_data.get("z", 0.0))
	var vertical_offset := 0.05
	if not force:
		var current_ground_height := _terrain_height_at(player_node.global_position.x, player_node.global_position.z)
		vertical_offset = maxf(0.05, player_node.global_position.y - current_ground_height)
	var target := Vector3(target_x, _terrain_height_at(target_x, target_z) + vertical_offset, target_z)
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
	_cancel_rewards_request()
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
	if rewards_panel != null:
		rewards_panel.visible = false
	logout_button.visible = true
	trade_button.visible = true
	rewards_button.visible = true
	character_button.visible = true
	equipment_button.visible = true
	rest_button.visible = true
	bag_button.visible = true
	skills_button.visible = true
	_start_movement_sync()
	_start_gathering()
	_start_chat()
	_start_presence()

func _set_account_status(message: String) -> void:
	if auth_status_label != null:
		auth_status_label.text = message
	if status_label != null:
		status_label.text = message


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
	_cancel_rewards_request()
	_stop_presence()
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
	_cancel_rewards_request()
	_stop_chat()
	_stop_skills_request()
	if market_panel != null:
		market_panel.visible = false
	if skills_panel != null:
		skills_panel.visible = false
	if inventory_panel != null:
		inventory_panel.visible = false
	if rewards_panel != null:
		rewards_panel.visible = false
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
	rewards_button.visible = false
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
	_start_enemy_snapshot_polling()
	_start_trade_polling()
	_send_presence_heartbeat()
	_poll_zone_presence()



func _stop_presence() -> void:
	_stop_enemy_snapshot_polling()
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
	remote_player_names.clear()
	_stop_trade_polling()
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
		var target := Vector3(float(remote_data.get("x", 0.0)), 0.0, float(remote_data.get("z", 0.0)))
		target.y = _terrain_height_at(target.x, target.z) + 0.05
		if not remote_player_nodes.has(character_id):
			_create_remote_player_visual(character_id, str(remote_data.get("name", "Wayfarer")), target)
		remote_player_targets[character_id] = target
		remote_player_names[character_id] = str(remote_data.get("name", "Wayfarer"))
	for character_id in remote_player_nodes.keys():
		if not seen.has(character_id):
			var stale: Node3D = remote_player_nodes[character_id]
			if is_instance_valid(stale):
				stale.queue_free()
			remote_player_nodes.erase(character_id)
			remote_player_targets.erase(character_id)
			remote_player_names.erase(character_id)
	if presence_label != null:
		presence_label.text = "Nearby players: " + str(remote_player_nodes.size())
	if trade_target_select != null:
		_refresh_trade_targets()


func _create_remote_player_visual(character_id: String, display_name: String, at: Vector3) -> void:
	remote_player_names[character_id] = display_name
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
	var blend := 1.0 - exp(-delta * 8.0)
	for character_id in remote_player_nodes.keys():
		var remote: Node3D = remote_player_nodes[character_id]
		if not is_instance_valid(remote):
			continue
		var target: Vector3 = remote_player_targets.get(character_id, remote.position)
		var displacement := target - remote.position
		displacement.y = 0.0
		if displacement.length_squared() > 0.0025:
			var desired_yaw := atan2(displacement.x, displacement.z)
			remote.rotation.y = lerp_angle(remote.rotation.y, desired_yaw, blend)
		if displacement.length() > 20.0:
			remote.position = target
		else:
			remote.position = remote.position.lerp(target, blend)


func _start_trade_polling() -> void:
	if player_id.is_empty() or auth_token.is_empty():
		return
	if trade_button != null:
		trade_button.visible = true
	if trade_request == null:
		trade_request = _new_api_request()
		add_child(trade_request)
		trade_request.request_completed.connect(_on_trade_request_completed)
	if trade_poll_timer == null:
		trade_poll_timer = Timer.new()
		trade_poll_timer.wait_time = 3.0
		trade_poll_timer.timeout.connect(_poll_active_trade)
		add_child(trade_poll_timer)
	trade_poll_timer.start()
	_poll_active_trade()


func _stop_trade_polling() -> void:
	if trade_poll_timer != null:
		trade_poll_timer.stop()
		trade_poll_timer.queue_free()
		trade_poll_timer = null
	if trade_request != null:
		if trade_request.get_http_client_status() != HTTPClient.STATUS_DISCONNECTED:
			trade_request.cancel_request()
		trade_request.queue_free()
		trade_request = null
	trade_pending = false
	if trade_refresh_request != null:
		if trade_refresh_request.get_http_client_status() != HTTPClient.STATUS_DISCONNECTED:
			trade_refresh_request.cancel_request()
		trade_refresh_request.queue_free()
		trade_refresh_request = null
	trade_refresh_pending = false
	current_trade.clear()
	if trade_panel != null:
		trade_panel.visible = false
	if trade_button != null:
		trade_button.text = "Trade"
		trade_button.visible = false
	if trade_status_label != null:
		trade_status_label.text = "No active trade"


func _poll_active_trade() -> void:
	if player_id.is_empty() or trade_pending or trade_request == null:
		return
	trade_action = "poll"
	var url := api_base_url + "/api/v1/account/characters/" + player_id + "/trades/active"
	var error := trade_request.request(url, _api_headers())
	if error == OK:
		trade_pending = true


func _toggle_trade_panel() -> void:
	if trade_panel != null and trade_panel.visible:
		trade_panel.visible = false
		return
	if inventory_panel != null:
		inventory_panel.visible = false
	if skills_panel != null:
		skills_panel.visible = false
	if equipment_panel != null:
		equipment_panel.visible = false
	if market_panel != null:
		market_panel.visible = false
	if trade_panel == null:
		_build_trade_panel()
	trade_panel.visible = true
	_refresh_trade_targets()
	_refresh_trade_items()
	_render_trade_snapshot()


func _build_trade_panel() -> void:
	trade_panel = PanelContainer.new()
	trade_panel.anchor_left = 0.5
	trade_panel.anchor_right = 0.5
	trade_panel.offset_top = 145
	trade_panel.offset_bottom = 535
	trade_panel.custom_minimum_size = Vector2(410, 390)
	var column := VBoxContainer.new()
	trade_panel.add_child(column)
	var heading := Label.new()
	heading.text = "PLAYER TRADE"
	heading.add_theme_font_size_override("font_size", 18)
	column.add_child(heading)
	var target_label := Label.new()
	target_label.text = "Nearby player"
	column.add_child(target_label)
	trade_target_select = OptionButton.new()
	trade_target_select.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	column.add_child(trade_target_select)
	var item_label := Label.new()
	item_label.text = "Build an offer with up to 20 item stacks"
	column.add_child(item_label)
	trade_item_select = OptionButton.new()
	trade_item_select.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	column.add_child(trade_item_select)
	var quantity_row := HBoxContainer.new()
	var quantity_label := Label.new()
	quantity_label.text = "Add quantity"
	quantity_row.add_child(quantity_label)
	trade_quantity_input = LineEdit.new()
	trade_quantity_input.text = "1"
	trade_quantity_input.custom_minimum_size = Vector2(65, 0)
	quantity_row.add_child(trade_quantity_input)
	var add_item_button := Button.new()
	add_item_button.text = "Add"
	add_item_button.pressed.connect(_add_trade_offer_item)
	quantity_row.add_child(add_item_button)
	var remove_item_button := Button.new()
	remove_item_button.text = "Remove"
	remove_item_button.pressed.connect(_remove_trade_offer_item)
	quantity_row.add_child(remove_item_button)
	var coins_label := Label.new()
	coins_label.text = "Coins"
	quantity_row.add_child(coins_label)
	trade_coins_input = LineEdit.new()
	trade_coins_input.text = "0"
	trade_coins_input.custom_minimum_size = Vector2(75, 0)
	quantity_row.add_child(trade_coins_input)
	column.add_child(quantity_row)
	trade_offer_items_label = Label.new()
	trade_offer_items_label.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	trade_offer_items_label.custom_minimum_size = Vector2(380, 42)
	column.add_child(trade_offer_items_label)
	var actions := HBoxContainer.new()
	trade_invite_button = Button.new()
	trade_invite_button.text = "Invite"
	trade_invite_button.pressed.connect(func() -> void: _send_trade_action("invite"))
	actions.add_child(trade_invite_button)
	trade_accept_button = Button.new()
	trade_accept_button.text = "Accept"
	trade_accept_button.visible = false
	trade_accept_button.pressed.connect(func() -> void: _send_trade_action("accept"))
	actions.add_child(trade_accept_button)
	trade_update_button = Button.new()
	trade_update_button.text = "Update"
	trade_update_button.visible = false
	trade_update_button.pressed.connect(func() -> void: _send_trade_action("update"))
	actions.add_child(trade_update_button)
	trade_confirm_button = Button.new()
	trade_confirm_button.text = "Confirm"
	trade_confirm_button.visible = false
	trade_confirm_button.pressed.connect(func() -> void: _send_trade_action("confirm"))
	actions.add_child(trade_confirm_button)
	trade_cancel_button = Button.new()
	trade_cancel_button.text = "Cancel"
	trade_cancel_button.visible = false
	trade_cancel_button.pressed.connect(func() -> void: _send_trade_action("cancel"))
	actions.add_child(trade_cancel_button)
	column.add_child(actions)
	trade_status_label = Label.new()
	trade_status_label.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	trade_status_label.custom_minimum_size = Vector2(380, 125)
	column.add_child(trade_status_label)
	var close_button := Button.new()
	close_button.text = "Close"
	close_button.pressed.connect(func() -> void: trade_panel.visible = false)
	column.add_child(close_button)
	auth_panel.get_parent().add_child(trade_panel)
	if not get_viewport().size_changed.is_connected(_layout_trade_panel):
		get_viewport().size_changed.connect(_layout_trade_panel)
	_layout_trade_panel()


func _layout_trade_panel() -> void:
	if trade_panel == null:
		return
	var viewport_width: float = get_viewport().get_visible_rect().size.x
	var panel_width := minf(410.0, maxf(240.0, viewport_width - 24.0))
	trade_panel.offset_left = -panel_width / 2.0
	trade_panel.offset_right = panel_width / 2.0
	trade_panel.custom_minimum_size = Vector2(panel_width, 390)


func _refresh_trade_targets() -> void:
	if trade_target_select == null:
		return
	trade_target_select.clear()
	var ids: Array = remote_player_names.keys()
	ids.sort()
	for character_id in ids:
		trade_target_select.add_item(str(remote_player_names[character_id]))
		trade_target_select.set_item_metadata(trade_target_select.item_count - 1, str(character_id))


func _refresh_trade_items() -> void:
	if trade_item_select == null:
		return
	trade_item_select.clear()
	trade_item_select.add_item("Select item")
	trade_item_select.set_item_metadata(0, "")
	var equipped: Array = current_equipment.values()
	var available: Dictionary = {}
	for item_id in current_inventory.keys():
		if int(current_inventory.get(item_id, 0)) > 0:
			available[str(item_id)] = true
	for item_id in trade_offer_items.keys():
		available[str(item_id)] = true
	var item_ids: Array = available.keys()
	item_ids.sort()
	for item_id in item_ids:
		if equipped.has(item_id):
			continue
		var bag_count := int(current_inventory.get(item_id, 0))
		var offered_count := int(trade_offer_items.get(item_id, 0))
		trade_item_select.add_item(str(item_id).replace("_", " ") + "  (bag " + str(bag_count) + ", offer " + str(offered_count) + ")")
		trade_item_select.set_item_metadata(trade_item_select.item_count - 1, str(item_id))
	_update_trade_offer_items_label()


func _current_own_trade_offer() -> Dictionary:
	var participants: Variant = current_trade.get("participants", [])
	if participants is Array:
		for participant in participants:
			if participant is Dictionary and str(participant.get("characterId", "")) == player_id:
				var offer: Variant = participant.get("offer", {})
				if offer is Dictionary:
					return offer
	return {"items": {}, "coins": 0}


func _add_trade_offer_item() -> void:
	if trade_item_select == null or trade_quantity_input == null:
		return
	var item_id := str(trade_item_select.get_selected_metadata())
	var quantity_text := trade_quantity_input.text.strip_edges()
	if item_id.is_empty():
		trade_status_label.text = "Select an item stack to add."
		return
	if not quantity_text.is_valid_int() or int(quantity_text) < 1:
		trade_status_label.text = "Enter a positive whole-number item quantity."
		return
	var quantity := int(quantity_text)
	if not trade_offer_items.has(item_id) and trade_offer_items.size() >= 20:
		trade_status_label.text = "A trade offer can contain at most 20 item types."
		return
	var current_offer := _current_own_trade_offer()
	var current_items: Dictionary = current_offer.get("items", {})
	var total_available := int(current_inventory.get(item_id, 0)) + int(current_items.get(item_id, 0))
	var new_quantity := int(trade_offer_items.get(item_id, 0)) + quantity
	if new_quantity > total_available:
		trade_status_label.text = "You can only add items currently available after returning your existing offer."
		return
	trade_offer_items[item_id] = new_quantity
	_update_trade_offer_items_label()
	_refresh_trade_items()


func _remove_trade_offer_item() -> void:
	if trade_item_select == null:
		return
	var item_id := str(trade_item_select.get_selected_metadata())
	if item_id.is_empty() or not trade_offer_items.has(item_id):
		trade_status_label.text = "Select an item already in your offer to remove it."
		return
	trade_offer_items.erase(item_id)
	_update_trade_offer_items_label()
	_refresh_trade_items()


func _update_trade_offer_items_label() -> void:
	if trade_offer_items_label == null:
		return
	if trade_offer_items.is_empty():
		trade_offer_items_label.text = "Staged items: none"
		return
	var item_ids: Array = trade_offer_items.keys()
	item_ids.sort()
	var parts: PackedStringArray = []
	for item_id in item_ids:
		parts.append(str(item_id).replace("_", " ") + " ×" + str(trade_offer_items[item_id]))
	trade_offer_items_label.text = "Staged items: " + ", ".join(parts)


func _send_trade_action(action: String) -> void:
	if trade_request == null or trade_pending or player_id.is_empty():
		return
	var url := api_base_url + "/api/v1/account/characters/" + player_id + "/trades"
	var method := HTTPClient.METHOD_POST
	var body := ""
	if action == "invite":
		if trade_target_select == null or trade_target_select.item_count == 0:
			trade_status_label.text = "No nearby player is available to invite."
			return
		var target_id := str(trade_target_select.get_selected_metadata())
		if target_id.is_empty():
			trade_status_label.text = "Select a nearby player first."
			return
		url += "/invite/" + target_id
	elif action == "accept":
		var trade_id := str(current_trade.get("id", ""))
		if trade_id.is_empty():
			return
		url = api_base_url + "/api/v1/account/characters/" + player_id + "/trades/accept/" + trade_id
	elif action == "update" or action == "confirm" or action == "cancel":
		var active_id := str(current_trade.get("id", ""))
		if active_id.is_empty():
			return
		url = api_base_url + "/api/v1/account/characters/" + player_id + "/trades/" + ("offer" if action == "update" else action) + "/" + active_id
		if action == "update":
			method = HTTPClient.METHOD_PUT
	else:
		return

	if action == "invite" or action == "accept" or action == "update":
		var coins_text := trade_coins_input.text.strip_edges()
		if not coins_text.is_valid_int():
			trade_status_label.text = "Enter a whole-number coin offer."
			return
		var coins := int(coins_text)
		if coins < 0:
			trade_status_label.text = "Coin offers cannot be negative."
			return
		var currently_offered := _current_own_trade_offer()
		var available_coins := current_coins + int(currently_offered.get("coins", 0))
		if coins > available_coins:
			trade_status_label.text = "Coin offer exceeds your available balance."
			return
		for item_id in trade_offer_items.keys():
			var available_quantity := int(current_inventory.get(item_id, 0)) + int(currently_offered.get("items", {}).get(item_id, 0))
			if int(trade_offer_items[item_id]) < 1 or int(trade_offer_items[item_id]) > available_quantity:
				trade_status_label.text = "Offer quantity exceeds your available stack for " + str(item_id).replace("_", " ") + "."
				return
		var request_data: Dictionary = {"offer": {"items": trade_offer_items.duplicate(), "coins": coins}}
		if action == "update":
			request_data["revision"] = int(current_trade.get("revision", 0))
		body = JSON.stringify(request_data)
	elif action == "confirm":
		body = JSON.stringify({"revision": int(current_trade.get("revision", 0))})

	trade_action = action
	var error := trade_request.request(url, _api_headers(), method, body)
	if error == OK:
		trade_pending = true
		trade_status_label.text = "Sending trade " + action + "..."


func _on_trade_request_completed(result: int, response_code: int, _headers: PackedStringArray, body: PackedByteArray) -> void:
	trade_pending = false
	if player_id.is_empty():
		return
	if response_code == 401:
		_expire_session()
		return
	var payload: Variant = JSON.parse_string(body.get_string_from_utf8())
	if trade_action == "poll" and response_code == 404:
		current_trade.clear()
		_render_trade_snapshot()
		return
	if result != HTTPRequest.RESULT_SUCCESS or response_code < 200 or response_code >= 300 or not payload is Dictionary:
		var error_code := str(payload.get("error", "trade_unavailable")) if payload is Dictionary else "trade_unavailable"
		if trade_status_label != null:
			trade_status_label.text = "Trade: " + error_code.replace("_", " ")
		return
	current_trade = payload
	_render_trade_snapshot()
	var completed_status := str(current_trade.get("status", ""))
	if trade_action == "invite" or trade_action == "accept" or trade_action == "update" or completed_status == "completed" or completed_status == "cancelled" or completed_status == "expired":
		_refresh_character_after_trade()
	var trade_status := str(current_trade.get("status", ""))
	if trade_button != null and trade_status == "pending":
		trade_button.text = "Trade • 1"
		var own_role := ""
		var partner_name := "nearby player"
		for participant in current_trade.get("participants", []):
			if not participant is Dictionary:
				continue
			if str(participant.get("characterId", "")) == player_id:
				own_role = str(participant.get("role", ""))
			else:
				partner_name = str(participant.get("name", partner_name))
		if own_role == "target":
			status_label.text = "Trade invitation from " + partner_name
		else:
			status_label.text = "Trade invitation sent to " + partner_name
	elif trade_button != null:
		trade_button.text = "Trade"
		if trade_status == "open":
			status_label.text = "Trade open • review both offers and confirm when ready."
		elif trade_status == "completed" or trade_status == "cancelled" or trade_status == "expired":
			status_label.text = "Trade " + trade_status + "."


func _refresh_character_after_trade() -> void:
	if player_id.is_empty() or trade_refresh_pending:
		return
	if trade_refresh_request == null:
		trade_refresh_request = _new_api_request()
		add_child(trade_refresh_request)
		trade_refresh_request.request_completed.connect(_on_trade_character_refreshed)
	var url := api_base_url + "/api/v1/account/characters/" + player_id
	var error := trade_refresh_request.request(url, _api_headers())
	if error == OK:
		trade_refresh_pending = true


func _on_trade_character_refreshed(result: int, response_code: int, _headers: PackedStringArray, body: PackedByteArray) -> void:
	trade_refresh_pending = false
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
	_update_inventory(payload.get("inventory", {}), payload.get("bank", {}), payload.get("coins", current_coins), payload.get("equipment", {}), payload.get("health", current_health))
	_refresh_trade_items()


func _render_trade_snapshot() -> void:
	if trade_status_label == null:
		return
	if current_trade.is_empty():
		trade_status_label.text = "No active trade. Invite a nearby player to begin."
		trade_offer_items.clear()
		trade_editor_trade_id = ""
		trade_editor_revision = -1
		if trade_coins_input != null:
			trade_coins_input.text = "0"
		_update_trade_offer_items_label()
		_refresh_trade_items()
		trade_invite_button.visible = true
		trade_accept_button.visible = false
		trade_update_button.visible = false
		trade_confirm_button.visible = false
		trade_cancel_button.visible = false
		return
	var status := str(current_trade.get("status", "unknown"))
	var revision := int(current_trade.get("revision", 0))
	var trade_id := str(current_trade.get("id", ""))
	var participants: Variant = current_trade.get("participants", [])
	if trade_id != trade_editor_trade_id or revision != trade_editor_revision:
		trade_editor_trade_id = trade_id
		trade_editor_revision = revision
		trade_offer_items.clear()
		if trade_coins_input != null:
			trade_coins_input.text = "0"
		for participant in participants:
			if participant is Dictionary and str(participant.get("characterId", "")) == player_id:
				var own_offer: Variant = participant.get("offer", {})
				if own_offer is Dictionary:
					var staged: Variant = own_offer.get("items", {})
					if staged is Dictionary:
						trade_offer_items = staged.duplicate()
					if trade_coins_input != null:
						trade_coins_input.text = str(int(own_offer.get("coins", 0)))
		_update_trade_offer_items_label()
		_refresh_trade_items()
	var own_role := ""
	var other_name := "another player"
	var own_offer := ""
	var other_offer := ""
	for participant in participants:
		if not participant is Dictionary:
			continue
		var offer: Dictionary = participant.get("offer", {})
		var offer_text := str(offer.get("items", {})) + " • " + str(offer.get("coins", 0)) + " coins"
		if bool(participant.get("confirmed", false)):
			offer_text += " • CONFIRMED"
		if str(participant.get("characterId", "")) == player_id:
			own_role = str(participant.get("role", ""))
			own_offer = offer_text
		else:
			other_name = str(participant.get("name", other_name))
			other_offer = offer_text
	var lines := "Trade with " + other_name + "
Status: " + status.capitalize() + "  •  Revision " + str(revision)
	if not own_offer.is_empty():
		lines += "
Your offer: " + own_offer
	if not other_offer.is_empty():
		lines += "
Their offer: " + other_offer
	trade_status_label.text = lines
	trade_invite_button.visible = false
	trade_accept_button.visible = status == "pending" and own_role == "target"
	trade_update_button.visible = status == "open"
	trade_confirm_button.visible = status == "open"
	trade_cancel_button.visible = status == "open" or status == "pending"


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


func _show_rewards_panel() -> void:
	if rewards_panel != null:
		rewards_panel.visible = not rewards_panel.visible
		if rewards_panel.visible:
			if skills_panel != null:
				skills_panel.visible = false
			if inventory_panel != null:
				inventory_panel.visible = false
			if market_panel != null:
				market_panel.visible = false
			if equipment_panel != null:
				equipment_panel.visible = false
			_request_reward_claims()
		return
	if skills_panel != null:
		skills_panel.visible = false
	if inventory_panel != null:
		inventory_panel.visible = false
	if market_panel != null:
		market_panel.visible = false
	if equipment_panel != null:
		equipment_panel.visible = false
	rewards_panel = PanelContainer.new()
	rewards_panel.position = Vector2(420, 155)
	rewards_panel.custom_minimum_size = Vector2(380, 280)
	var column := VBoxContainer.new()
	rewards_panel.add_child(column)
	var heading := Label.new()
	heading.text = "ENCOUNTER REWARDS"
	heading.add_theme_font_size_override("font_size", 18)
	column.add_child(heading)
	var description := Label.new()
	description.text = "Pending XP and loot are stored until claimed."
	description.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	column.add_child(description)
	var scroll := ScrollContainer.new()
	scroll.custom_minimum_size = Vector2(350, 170)
	scroll.size_flags_vertical = Control.SIZE_EXPAND_FILL
	column.add_child(scroll)
	rewards_claims_column = VBoxContainer.new()
	rewards_claims_column.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	scroll.add_child(rewards_claims_column)
	rewards_status_label = Label.new()
	rewards_status_label.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	rewards_status_label.text = "Loading reward claims..."
	column.add_child(rewards_status_label)
	var refresh_button := Button.new()
	refresh_button.text = "Refresh claims"
	refresh_button.pressed.connect(_request_reward_claims)
	column.add_child(refresh_button)
	auth_panel.get_parent().add_child(rewards_panel)
	_request_reward_claims()


func _ensure_rewards_request() -> void:
	if rewards_request != null:
		return
	rewards_request = _new_api_request()
	add_child(rewards_request)
	rewards_request.request_completed.connect(_on_rewards_request_completed)


func _cancel_rewards_request() -> void:
	if rewards_request == null:
		rewards_pending = false
		return
	if rewards_request.request_completed.is_connected(_on_rewards_request_completed):
		rewards_request.request_completed.disconnect(_on_rewards_request_completed)
	if rewards_request.get_http_client_status() != HTTPClient.STATUS_DISCONNECTED:
		rewards_request.cancel_request()
	rewards_request.queue_free()
	rewards_request = null
	rewards_pending = false


func _request_reward_claims() -> void:
	_ensure_rewards_request()
	if player_id.is_empty() or auth_token.is_empty() or rewards_pending or rewards_request == null:
		return
	rewards_action = "list"
	var url := api_base_url + "/api/v1/account/characters/" + player_id + "/rewards"
	var error := rewards_request.request(url, _api_headers(), HTTPClient.METHOD_GET)
	if error == OK:
		rewards_pending = true
		if rewards_status_label != null:
			rewards_status_label.text = "Loading reward claims..."
	elif rewards_status_label != null:
		rewards_status_label.text = "Could not request reward claims."


func _claim_encounter_reward(enemy_id: String, encounter_id: String) -> void:
	_ensure_rewards_request()
	if player_id.is_empty() or auth_token.is_empty() or rewards_pending or rewards_request == null:
		return
	rewards_action = "claim"
	var url := api_base_url + "/api/v1/account/characters/" + player_id + "/rewards/" + enemy_id + "/" + encounter_id + "/claim"
	var error := rewards_request.request(url, _api_headers(), HTTPClient.METHOD_POST, "")
	if error == OK:
		rewards_pending = true
		rewards_status_label.text = "Claiming encounter reward..."
	else:
		rewards_status_label.text = "Could not submit reward claim."


func _on_rewards_request_completed(result: int, response_code: int, _headers: PackedStringArray, body: PackedByteArray) -> void:
	rewards_pending = false
	if player_id.is_empty():
		return
	if response_code == 401:
		_expire_session()
		return
	var payload: Variant = JSON.parse_string(body.get_string_from_utf8())
	if rewards_action == "list":
		if result != HTTPRequest.RESULT_SUCCESS or response_code != 200 or not payload is Dictionary:
			if rewards_status_label != null:
				rewards_status_label.text = "Could not load reward claims."
			return
		_render_reward_claims(payload.get("claims", []))
		return
	if response_code == 409:
		if rewards_status_label != null:
			var error_code := str(payload.get("error", "reward_claim_failed")) if payload is Dictionary else "reward_claim_failed"
			if error_code == "reward_inventory_full":
				rewards_status_label.text = "Inventory stack full. Free space in that item stack, then retry; the reward remains pending."
			else:
				rewards_status_label.text = "Reward claim conflict: " + error_code
		_request_reward_claims_deferred()
		return
	if result != HTTPRequest.RESULT_SUCCESS or response_code != 200 or not payload is Dictionary:
		if rewards_status_label != null:
			rewards_status_label.text = "Reward claim failed."
		return
	var character: Variant = payload.get("character", {})
	if character is Dictionary:
		_update_inventory(character.get("inventory", {}), character.get("bank", {}),
			character.get("coins", current_coins), character.get("equipment", {}),
			character.get("health", current_health))
	if rewards_status_label != null:
		rewards_status_label.text = "Reward claimed successfully."
	_request_reward_claims_deferred()


func _request_reward_claims_deferred() -> void:
	# Avoid starting a second request from inside HTTPRequest.request_completed.
	call_deferred("_request_reward_claims")


func _render_reward_claims(claims: Variant) -> void:
	if rewards_claims_column == null:
		return
	for child in rewards_claims_column.get_children():
		child.queue_free()
	if not claims is Array or claims.is_empty():
		var empty_label := Label.new()
		empty_label.text = "No encounter rewards are waiting."
		rewards_claims_column.add_child(empty_label)
		if rewards_status_label != null:
			rewards_status_label.text = "Reward list is up to date."
		return
	var pending_count := 0
	for value in claims:
		if not value is Dictionary:
			continue
		var claim: Dictionary = value
		var enemy_id := str(claim.get("enemyId", ""))
		var encounter_id := str(claim.get("encounterId", ""))
		var skill_id := str(claim.get("experienceSkillId", "combat"))
		var xp := int(claim.get("experienceAmount", 0))
		var parent_skill := str(claim.get("parentSkillId", ""))
		var parent_xp := int(claim.get("parentExperienceAmount", 0))
		var loot_items: Array = []
		var raw_loot_items: Variant = claim.get("lootItems", [])
		if raw_loot_items is Array:
			for loot_value in raw_loot_items:
				if loot_value is Dictionary:
					var item_id := str(loot_value.get("itemId", ""))
					var item_quantity := int(loot_value.get("quantity", 0))
					if not item_id.is_empty() and item_quantity > 0:
						loot_items.append({"itemId": item_id, "quantity": item_quantity})
		# Older servers and legacy claim payloads may still expose one item.
		if loot_items.is_empty():
			var legacy_item_id := str(claim.get("lootItemId", ""))
			var legacy_quantity := int(claim.get("lootQuantity", 0))
			if not legacy_item_id.is_empty() and legacy_quantity > 0:
				loot_items.append({"itemId": legacy_item_id, "quantity": legacy_quantity})
		var status := str(claim.get("status", "pending"))
		var summary := Label.new()
		var reward_text := skill_id.capitalize() + " XP " + str(xp)
		if not parent_skill.is_empty() and parent_xp > 0:
			reward_text += " • " + parent_skill.capitalize() + " XP " + str(parent_xp)
		for loot in loot_items:
			reward_text += " • " + str(loot.get("itemId", "")).replace("_", " ").capitalize() + " x" + str(loot.get("quantity", 0))
		summary.text = enemy_id.replace("-", " ").capitalize() + "\n" + reward_text + " • " + status.capitalize()
		summary.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
		rewards_claims_column.add_child(summary)
		if status == "pending":
			pending_count += 1
			var claim_button := Button.new()
			claim_button.text = "Claim reward"
			claim_button.pressed.connect(_claim_encounter_reward.bind(enemy_id, encounter_id))
			rewards_claims_column.add_child(claim_button)
		var separator := HSeparator.new()
		rewards_claims_column.add_child(separator)
	if rewards_status_label != null:
		rewards_status_label.text = str(pending_count) + " pending encounter reward(s)."


func _show_inventory_panel() -> void:
	if rewards_panel != null:
		rewards_panel.visible = false
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
	for skill_id in CORE_SKILL_IDS:
		if not skills.has(skill_id):
			continue
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
	skills_status_label.text = str(CORE_SKILL_IDS.size()) + " core skills • progression updates after server actions"


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
