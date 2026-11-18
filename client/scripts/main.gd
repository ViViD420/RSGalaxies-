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
var current_max_health := 100
var current_action := 100
var current_max_action := 100
var current_mind := 100
var current_max_mind := 100
var current_constitution := 10
var current_stamina := 10
var current_willpower := 10
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
var recipe_catalog: Array = []
var crafting_button: Button
var crafting_panel: PanelContainer
var crafting_station_filter: OptionButton
var crafting_recipes_column: VBoxContainer
var crafting_status_label: Label
var crafting_selected_station := "all"
var stats_button: Button
var stats_panel: PanelContainer
var stats_summary_label: Label
var hud_health_bar: ProgressBar
var hud_action_bar: ProgressBar
var hud_mind_bar: ProgressBar
var guide_button: Button
var guide_panel: PanelContainer
var guide_filter_select: OptionButton
var guide_entries_column: VBoxContainer
var guide_selected_filter := "all"
var quick_heal_button: Button
var settings_button: Button
var settings_panel: PanelContainer
var settings_status_label: Label
var tracked_waypoint_name := ""
var tracked_waypoint_position := Vector3.ZERO
var tracked_waypoint_active := false
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
		at.y = _terrain_height_at(at.x, at.z)
		match str(node.get("resourceId", "")):
			"softwood_log":
				_add_tree(at)
			"rough_stone":
				_add_stone_deposit(at)
			"copper_ore":
				_add_copper_vein(at)
			"wild_herb":
				_add_herb_patch(at)
			"frontier_grain":
				_add_crop_plot(at)
			"salvaged_scrap":
				_add_scrap_pile(at)
			"iron_ore":
				_add_iron_vein(at)
			"titanium_ore":
				_add_titanium_outcrop(at)
			"aether_crystal":
				_add_aether_crystal_node(at)
			"cobalt_ore":
				_add_cobalt_vein(at)
			"durasteel_ore":
				_add_durasteel_fissure(at)
			"luminite_gem":
				_add_luminite_geode(at)
			"mythril_ore":
				_add_mythril_vein(at)
			"orichalcum_ore":
				_add_orichalcum_vein(at)
			"obsidian_shard":
				_add_obsidian_outcrop(at)
			"void_quartz":
				_add_void_quartz_monolith(at)
			"neutronium_ore":
				_add_neutronium_vein(at)
			"starsteel_ore":
				_add_starsteel_meteorite(at)
			"electrum_ore":
				_add_electrum_vein(at)
			"solarite_crystal":
				_add_solarite_spire(at)
			"chronite_ore":
				_add_chronite_anomaly(at)
			"abyssal_ore":
				_add_abyssal_chasm(at)
			"nova_prism":
				_add_nova_prism_matrix(at)
			"ironwood_log":
				_add_ironwood_tree(at)
			"petrified_starwood":
				_add_starwood_tree(at)
			"glowcap_fungus":
				_add_glowcap_cluster(at)
			"sunblossom_petal":
				_add_sunblossom_patch(at)
			"tuber_root":
				_add_tuber_bed(at)
			"starberry_cluster":
				_add_starberry_arbor(at)
			"circuit_board":
				_add_circuit_relay(at)
			"quantum_capacitor":
				_add_orbital_pod_wreckage(at)
			"aether_bough_timber":
				_add_elderwood_tree(at)
			"voidbark_heartwood":
				_add_voidbark_tree(at)
			"cryo_lotus_petal":
				_add_frostlotus_bloom(at)
			"astral_mycelium":
				_add_astralspore_colony(at)
			"auric_grain_sheaf":
				_add_goldengrain_field(at)
			"celestial_nectar_fruit":
				_add_celestialfruit_orchard(at)
			"titan_actuator_coil":
				_add_servoarray_chassis(at)
			"singularity_core_cell":
				_add_singularity_wreckage(at)
			"runic_banite_ore":
				_add_runic_banite_monolith(at)
			"elder_wroshyr_timber":
				_add_elder_wroshyr_tree(at)
			"sovereign_torstol_leaf":
				_add_torstol_bloom_patch(at)
			"spirit_muja_fruit":
				_add_spirit_muja_arbor(at)
			"hypergate_relic_core":
				_add_rakatan_hypercore_relic(at)
			"orikalkum_beskar_ore":
				_add_orikalkum_beskar_vein(at)
			"felucian_magic_log":
				_add_felucian_magic_copse(at)
			"snapdragon_kolto_frond":
				_add_snapdragon_kolto_glade(at)
			"cursed_blight_plum":
				_add_cursed_blight_trellis(at)
			"gree_aether_matrix":
				_add_gree_aether_conduit(at)
			"primordial_animica_ore":
				_add_primordial_animica_monolith(at)
			"elder_bloodwood_timber":
				_add_bloodwood_brylark_tree(at)
			"fellstalk_ryll_blossom":
				_add_fellstalk_ryll_sanctuary(at)
			"golden_dragon_cactus":
				_add_golden_dragon_conservatory(at)
			"starforge_omnium_core":
				_add_starforge_reliquary(at)
			"promethium_stygium_ore":
				_add_promethium_stygium_rift(at)
			"umbaran_crystal_timber":
				_add_umbaran_crystal_grove(at)
			"arbuck_zevrona_herb":
				_add_arbuck_zevrona_thicket(at)
			"obsidian_lava_lychee":
				_add_mustafar_lavafruit_conservatory(at)
			"massassi_chronote_relic":
				_add_massassi_chronote_obelisk(at)
			"primal_kyber_ore":
				_add_primal_kyber_monolith(at)
			"solak_anima_heartwood":
				_add_solak_anima_bough(at)
			"bloodweed_vong_spore":
				_add_bloodweed_vong_hive(at)
			"celestial_mortis_melon":
				_add_golden_ambrosia_sanctuary(at)
			"infinity_gate_matrix":
				_add_infinity_holocron_vault(at)
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


func _add_crop_plot(at: Vector3) -> void:
	var plot := Node3D.new()
	plot.position = at
	var soil_material := StandardMaterial3D.new()
	soil_material.albedo_color = Color(0.26, 0.18, 0.11)
	var stalk_material := StandardMaterial3D.new()
	stalk_material.albedo_color = Color(0.76, 0.64, 0.22)
	var bed := MeshInstance3D.new()
	var bed_mesh := BoxMesh.new()
	bed_mesh.size = Vector3(1.8, 0.14, 1.8)
	bed.mesh = bed_mesh
	bed.position.y = 0.07
	bed.material_override = soil_material
	plot.add_child(bed)
	for row in [-0.5, 0.0, 0.5]:
		for col in [-0.5, 0.0, 0.5]:
			var stalk := MeshInstance3D.new()
			var stalk_mesh := CylinderMesh.new()
			stalk_mesh.top_radius = 0.04
			stalk_mesh.bottom_radius = 0.06
			stalk_mesh.height = 0.68
			stalk.mesh = stalk_mesh
			stalk.position = Vector3(row, 0.42, col)
			stalk.material_override = stalk_material
			plot.add_child(stalk)
	var label := Label3D.new()
	label.text = "Frontier Crop Plot"
	label.position.y = 1.05
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	plot.add_child(label)
	add_child(plot)


func _add_scrap_pile(at: Vector3) -> void:
	var pile := Node3D.new()
	pile.position = at
	var hull_material := StandardMaterial3D.new()
	hull_material.albedo_color = Color(0.38, 0.41, 0.45)
	var conduit_material := StandardMaterial3D.new()
	conduit_material.albedo_color = Color(0.24, 0.56, 0.68)
	for index in range(3):
		var plate := MeshInstance3D.new()
		var plate_mesh := BoxMesh.new()
		plate_mesh.size = Vector3(0.85, 0.18, 0.62)
		plate.mesh = plate_mesh
		plate.position = Vector3(float(index - 1) * 0.28, 0.18 + float(index) * 0.12, float(index % 2) * 0.18 - 0.09)
		plate.rotation = Vector3(0.18 * float(index), 0.65 * float(index), -0.15)
		plate.material_override = hull_material
		pile.add_child(plate)
	var core := MeshInstance3D.new()
	var core_mesh := CylinderMesh.new()
	core_mesh.top_radius = 0.16
	core_mesh.bottom_radius = 0.16
	core_mesh.height = 0.52
	core.mesh = core_mesh
	core.position = Vector3(0.0, 0.42, 0.08)
	core.rotation.z = 0.45
	core.material_override = conduit_material
	pile.add_child(core)
	var label := Label3D.new()
	label.text = "Debris Scrap Pile"
	label.position.y = 1.05
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	pile.add_child(label)
	add_child(pile)


func _add_iron_vein(at: Vector3) -> void:
	var deposit := Node3D.new()
	deposit.position = at
	var slate_material := StandardMaterial3D.new()
	slate_material.albedo_color = Color(0.24, 0.25, 0.28)
	var iron_material := StandardMaterial3D.new()
	iron_material.albedo_color = Color(0.46, 0.30, 0.26)
	for index in range(4):
		var boulder := MeshInstance3D.new()
		var mesh := SphereMesh.new()
		mesh.radius = 0.50 + float(index % 2) * 0.14
		mesh.height = mesh.radius * 1.65
		boulder.mesh = mesh
		boulder.position = Vector3(float(index % 2) * 0.58 - 0.28, 0.36, float(index / 2) * 0.52 - 0.26)
		boulder.material_override = slate_material
		deposit.add_child(boulder)
	var seam := MeshInstance3D.new()
	var seam_mesh := BoxMesh.new()
	seam_mesh.size = Vector3(0.65, 0.42, 0.38)
	seam.mesh = seam_mesh
	seam.position = Vector3(0.0, 0.58, -0.15)
	seam.rotation = Vector3(0.2, 0.4, -0.15)
	seam.material_override = iron_material
	deposit.add_child(seam)
	var label := Label3D.new()
	label.text = "Deep Iron Vein"
	label.position.y = 1.25
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	deposit.add_child(label)
	add_child(deposit)


func _add_titanium_outcrop(at: Vector3) -> void:
	var outcrop := Node3D.new()
	outcrop.position = at
	var base_material := StandardMaterial3D.new()
	base_material.albedo_color = Color(0.32, 0.35, 0.40)
	var titanium_material := StandardMaterial3D.new()
	titanium_material.albedo_color = Color(0.74, 0.80, 0.88)
	titanium_material.roughness = 0.35
	for index in range(3):
		var spire := MeshInstance3D.new()
		var spire_mesh := CylinderMesh.new()
		spire_mesh.top_radius = 0.14
		spire_mesh.bottom_radius = 0.42
		spire_mesh.height = 1.25 + float(index) * 0.22
		spire.mesh = spire_mesh
		spire.position = Vector3(float(index - 1) * 0.42, spire_mesh.height * 0.48, float(index % 2) * 0.22 - 0.11)
		spire.rotation.z = float(index - 1) * 0.16
		spire.material_override = titanium_material if index == 1 else base_material
		outcrop.add_child(spire)
	var label := Label3D.new()
	label.text = "Titanium Outcrop"
	label.position.y = 1.75
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	outcrop.add_child(label)
	add_child(outcrop)


func _add_aether_crystal_node(at: Vector3) -> void:
	var cluster := Node3D.new()
	cluster.position = at
	var crystal_material := StandardMaterial3D.new()
	crystal_material.albedo_color = Color(0.28, 0.78, 0.92)
	crystal_material.emission_enabled = true
	crystal_material.emission = Color(0.14, 0.48, 0.64)
	crystal_material.roughness = 0.2
	for index in range(4):
		var shard := MeshInstance3D.new()
		var shard_mesh := CylinderMesh.new()
		shard_mesh.top_radius = 0.04
		shard_mesh.bottom_radius = 0.24
		shard_mesh.height = 1.15 + float(index % 2) * 0.35
		shard.mesh = shard_mesh
		var angle := float(index) * TAU / 4.0
		shard.position = Vector3(cos(angle) * 0.24, shard_mesh.height * 0.46, sin(angle) * 0.24)
		shard.rotation = Vector3(sin(angle) * 0.22, angle, cos(angle) * 0.22)
		shard.material_override = crystal_material
		cluster.add_child(shard)
	var label := Label3D.new()
	label.text = "Aether Crystal Formation"
	label.position.y = 1.75
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	cluster.add_child(label)
	add_child(cluster)


func _add_cobalt_vein(at: Vector3) -> void:
	var ridge := Node3D.new()
	ridge.position = at
	var rock_mat := StandardMaterial3D.new()
	rock_mat.albedo_color = Color(0.22, 0.26, 0.34)
	var cobalt_mat := StandardMaterial3D.new()
	cobalt_mat.albedo_color = Color(0.20, 0.46, 0.86)
	for index in range(3):
		var slab := MeshInstance3D.new()
		var slab_mesh := BoxMesh.new()
		slab_mesh.size = Vector3(0.75, 0.55 + float(index) * 0.18, 0.52)
		slab.mesh = slab_mesh
		slab.position = Vector3(float(index - 1) * 0.38, slab_mesh.size.y * 0.5, float(index % 2) * 0.16)
		slab.rotation.y = float(index) * 0.35
		slab.material_override = cobalt_mat if index == 1 else rock_mat
		ridge.add_child(slab)
	var label := Label3D.new()
	label.text = "Cobalt Ridge Seam"
	label.position.y = 1.35
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	ridge.add_child(label)
	add_child(ridge)


func _add_durasteel_fissure(at: Vector3) -> void:
	var fissure := Node3D.new()
	fissure.position = at
	var dark_mat := StandardMaterial3D.new()
	dark_mat.albedo_color = Color(0.18, 0.20, 0.22)
	var steel_mat := StandardMaterial3D.new()
	steel_mat.albedo_color = Color(0.54, 0.58, 0.62)
	steel_mat.roughness = 0.4
	for index in range(4):
		var pillar := MeshInstance3D.new()
		var pillar_mesh := CylinderMesh.new()
		pillar_mesh.top_radius = 0.22
		pillar_mesh.bottom_radius = 0.36
		pillar_mesh.height = 1.1 + float(index % 2) * 0.35
		pillar.mesh = pillar_mesh
		pillar.position = Vector3(float(index % 2) * 0.52 - 0.26, pillar_mesh.height * 0.48, float(index / 2) * 0.52 - 0.26)
		pillar.material_override = steel_mat if index % 2 == 0 else dark_mat
		fissure.add_child(pillar)
	var label := Label3D.new()
	label.text = "Durasteel Fissure"
	label.position.y = 1.75
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	fissure.add_child(label)
	add_child(fissure)


func _add_luminite_geode(at: Vector3) -> void:
	var geode := Node3D.new()
	geode.position = at
	var gem_mat := StandardMaterial3D.new()
	gem_mat.albedo_color = Color(0.92, 0.76, 0.26)
	gem_mat.emission_enabled = true
	gem_mat.emission = Color(0.55, 0.42, 0.12)
	for index in range(4):
		var prism := MeshInstance3D.new()
		var prism_mesh := CylinderMesh.new()
		prism_mesh.top_radius = 0.05
		prism_mesh.bottom_radius = 0.26
		prism_mesh.height = 1.25
		prism.mesh = prism_mesh
		var angle := float(index) * TAU / 4.0
		prism.position = Vector3(cos(angle) * 0.26, 0.58, sin(angle) * 0.26)
		prism.rotation = Vector3(sin(angle) * 0.25, angle, cos(angle) * 0.25)
		prism.material_override = gem_mat
		geode.add_child(prism)
	var label := Label3D.new()
	label.text = "Luminite Geode Cluster"
	label.position.y = 1.75
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	geode.add_child(label)
	add_child(geode)


func _add_mythril_vein(at: Vector3) -> void:
	var vein := Node3D.new()
	vein.position = at
	var mythril_mat := StandardMaterial3D.new()
	mythril_mat.albedo_color = Color(0.68, 0.88, 0.96)
	mythril_mat.emission_enabled = true
	mythril_mat.emission = Color(0.22, 0.42, 0.54)
	mythril_mat.roughness = 0.25
	for index in range(3):
		var spire := MeshInstance3D.new()
		var spire_mesh := CylinderMesh.new()
		spire_mesh.top_radius = 0.10
		spire_mesh.bottom_radius = 0.38
		spire_mesh.height = 1.45 + float(index) * 0.2
		spire.mesh = spire_mesh
		spire.position = Vector3(float(index - 1) * 0.40, spire_mesh.height * 0.48, 0.0)
		spire.rotation.z = float(index - 1) * 0.18
		spire.material_override = mythril_mat
		vein.add_child(spire)
	var label := Label3D.new()
	label.text = "Mythril Star-Vein"
	label.position.y = 1.95
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	vein.add_child(label)
	add_child(vein)


func _add_orichalcum_vein(at: Vector3) -> void:
	var node := Node3D.new()
	node.position = at
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(0.88, 0.44, 0.18)
	mat.roughness = 0.35
	for index in range(3):
		var col := MeshInstance3D.new()
		var mesh := CylinderMesh.new()
		mesh.top_radius = 0.14
		mesh.bottom_radius = 0.36
		mesh.height = 1.3 + float(index) * 0.2
		col.mesh = mesh
		col.position = Vector3(float(index - 1) * 0.38, mesh.height * 0.48, 0.0)
		col.material_override = mat
		node.add_child(col)
	var label := Label3D.new()
	label.text = "Orichalcum Vent Seam"
	label.position.y = 1.85
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	node.add_child(label)
	add_child(node)


func _add_obsidian_outcrop(at: Vector3) -> void:
	var node := Node3D.new()
	node.position = at
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(0.12, 0.10, 0.16)
	mat.roughness = 0.15
	for index in range(4):
		var shard := MeshInstance3D.new()
		var mesh := BoxMesh.new()
		mesh.size = Vector3(0.38, 1.25 + float(index % 2) * 0.35, 0.28)
		shard.mesh = mesh
		shard.position = Vector3(float(index - 1.5) * 0.32, mesh.size.y * 0.48, float(index % 2) * 0.18)
		shard.rotation = Vector3(0.15, float(index) * 0.4, -0.18)
		shard.material_override = mat
		node.add_child(shard)
	var label := Label3D.new()
	label.text = "Volcanic Obsidian Outcrop"
	label.position.y = 1.85
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	node.add_child(label)
	add_child(node)


func _add_void_quartz_monolith(at: Vector3) -> void:
	var node := Node3D.new()
	node.position = at
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(0.58, 0.24, 0.86)
	mat.emission_enabled = true
	mat.emission = Color(0.32, 0.10, 0.52)
	var obelisk := MeshInstance3D.new()
	var mesh := BoxMesh.new()
	mesh.size = Vector3(0.65, 1.85, 0.65)
	obelisk.mesh = mesh
	obelisk.position.y = 0.92
	obelisk.rotation.y = 0.4
	obelisk.material_override = mat
	node.add_child(obelisk)
	var label := Label3D.new()
	label.text = "Void Quartz Monolith"
	label.position.y = 2.15
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	node.add_child(label)
	add_child(node)


func _add_neutronium_vein(at: Vector3) -> void:
	var node := Node3D.new()
	node.position = at
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(0.28, 0.34, 0.42)
	mat.roughness = 0.3
	for index in range(3):
		var mass := MeshInstance3D.new()
		var mesh := SphereMesh.new()
		mesh.radius = 0.56
		mesh.height = 1.1
		mass.mesh = mesh
		mass.position = Vector3(float(index - 1) * 0.42, 0.52, float(index % 2) * 0.2)
		mass.material_override = mat
		node.add_child(mass)
	var label := Label3D.new()
	label.text = "Neutronium Crater Vein"
	label.position.y = 1.55
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	node.add_child(label)
	add_child(node)


func _add_starsteel_meteorite(at: Vector3) -> void:
	var node := Node3D.new()
	node.position = at
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(0.84, 0.88, 0.96)
	mat.emission_enabled = true
	mat.emission = Color(0.35, 0.42, 0.65)
	mat.roughness = 0.2
	var core := MeshInstance3D.new()
	var mesh := SphereMesh.new()
	mesh.radius = 0.78
	mesh.height = 1.35
	core.mesh = mesh
	core.position.y = 0.68
	core.material_override = mat
	node.add_child(core)
	var label := Label3D.new()
	label.text = "Fallen Star-Metal Meteorite"
	label.position.y = 1.85
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	node.add_child(label)
	add_child(node)


func _add_electrum_vein(at: Vector3) -> void:
	var node := Node3D.new()
	node.position = at
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(0.94, 0.82, 0.32)
	mat.roughness = 0.25
	for index in range(3):
		var bar := MeshInstance3D.new()
		var mesh := CylinderMesh.new()
		mesh.top_radius = 0.12
		mesh.bottom_radius = 0.34
		mesh.height = 1.35 + float(index) * 0.18
		bar.mesh = mesh
		bar.position = Vector3(float(index - 1) * 0.36, mesh.height * 0.48, 0.0)
		bar.material_override = mat
		node.add_child(bar)
	var label := Label3D.new()
	label.text = "Electrum Crest Vein"
	label.position.y = 1.85
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	node.add_child(label)
	add_child(node)


func _add_solarite_spire(at: Vector3) -> void:
	var node := Node3D.new()
	node.position = at
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(0.98, 0.56, 0.14)
	mat.emission_enabled = true
	mat.emission = Color(0.62, 0.28, 0.06)
	var spire := MeshInstance3D.new()
	var mesh := CylinderMesh.new()
	mesh.top_radius = 0.06
	mesh.bottom_radius = 0.48
	mesh.height = 2.0
	spire.mesh = mesh
	spire.position.y = 1.0
	spire.material_override = mat
	node.add_child(spire)
	var label := Label3D.new()
	label.text = "Solarite Sun-Spire"
	label.position.y = 2.25
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	node.add_child(label)
	add_child(node)


func _add_chronite_anomaly(at: Vector3) -> void:
	var node := Node3D.new()
	node.position = at
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(0.32, 0.88, 0.78)
	mat.emission_enabled = true
	mat.emission = Color(0.14, 0.48, 0.42)
	for index in range(3):
		var ring := MeshInstance3D.new()
		var mesh := BoxMesh.new()
		mesh.size = Vector3(0.62, 0.62, 0.62)
		ring.mesh = mesh
		ring.position = Vector3(0.0, 0.55 + float(index) * 0.45, 0.0)
		ring.rotation = Vector3(0.4 * float(index), 0.6 * float(index), 0.3)
		ring.material_override = mat
		node.add_child(ring)
	var label := Label3D.new()
	label.text = "Chronite Rift Anomaly"
	label.position.y = 2.15
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	node.add_child(label)
	add_child(node)


func _add_abyssal_chasm(at: Vector3) -> void:
	var node := Node3D.new()
	node.position = at
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(0.14, 0.16, 0.24)
	mat.roughness = 0.2
	for index in range(4):
		var monolith := MeshInstance3D.new()
		var mesh := BoxMesh.new()
		mesh.size = Vector3(0.45, 1.45 + float(index % 2) * 0.35, 0.45)
		monolith.mesh = mesh
		monolith.position = Vector3(float(index - 1.5) * 0.36, mesh.size.y * 0.48, 0.0)
		monolith.material_override = mat
		node.add_child(monolith)
	var label := Label3D.new()
	label.text = "Abyssal Adamant Chasm"
	label.position.y = 2.05
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	node.add_child(label)
	add_child(node)


func _add_nova_prism_matrix(at: Vector3) -> void:
	var node := Node3D.new()
	node.position = at
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(0.96, 0.42, 0.78)
	mat.emission_enabled = true
	mat.emission = Color(0.58, 0.18, 0.44)
	var crystal := MeshInstance3D.new()
	var mesh := SphereMesh.new()
	mesh.radius = 0.68
	mesh.height = 1.65
	crystal.mesh = mesh
	crystal.position.y = 0.85
	crystal.material_override = mat
	node.add_child(crystal)
	var label := Label3D.new()
	label.text = "Nova Core Matrix"
	label.position.y = 2.05
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	node.add_child(label)
	add_child(node)


func _add_ironwood_tree(at: Vector3) -> void:
	var node := Node3D.new()
	node.position = at
	var trunk_mat := StandardMaterial3D.new()
	trunk_mat.albedo_color = Color(0.32, 0.22, 0.16)
	var trunk := MeshInstance3D.new()
	var trunk_mesh := CylinderMesh.new()
	trunk_mesh.top_radius = 0.28
	trunk_mesh.bottom_radius = 0.42
	trunk_mesh.height = 2.6
	trunk.mesh = trunk_mesh
	trunk.position.y = 1.3
	trunk.material_override = trunk_mat
	node.add_child(trunk)
	var canopy_mat := StandardMaterial3D.new()
	canopy_mat.albedo_color = Color(0.18, 0.42, 0.28)
	var canopy := MeshInstance3D.new()
	var canopy_mesh := SphereMesh.new()
	canopy_mesh.radius = 1.4
	canopy_mesh.height = 2.6
	canopy.mesh = canopy_mesh
	canopy.position.y = 3.0
	canopy.material_override = canopy_mat
	node.add_child(canopy)
	var label := Label3D.new()
	label.text = "Ironwood Timber Stand"
	label.position.y = 4.5
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	node.add_child(label)
	add_child(node)


func _add_starwood_tree(at: Vector3) -> void:
	var node := Node3D.new()
	node.position = at
	var trunk_mat := StandardMaterial3D.new()
	trunk_mat.albedo_color = Color(0.46, 0.44, 0.54)
	var trunk := MeshInstance3D.new()
	var trunk_mesh := CylinderMesh.new()
	trunk_mesh.top_radius = 0.30
	trunk_mesh.bottom_radius = 0.46
	trunk_mesh.height = 2.8
	trunk.mesh = trunk_mesh
	trunk.position.y = 1.4
	trunk.material_override = trunk_mat
	node.add_child(trunk)
	var canopy_mat := StandardMaterial3D.new()
	canopy_mat.albedo_color = Color(0.38, 0.62, 0.86)
	canopy_mat.emission_enabled = true
	canopy_mat.emission = Color(0.14, 0.28, 0.46)
	var canopy := MeshInstance3D.new()
	var canopy_mesh := SphereMesh.new()
	canopy_mesh.radius = 1.45
	canopy_mesh.height = 2.7
	canopy.mesh = canopy_mesh
	canopy.position.y = 3.2
	canopy.material_override = canopy_mat
	node.add_child(canopy)
	var label := Label3D.new()
	label.text = "Petrified Starwood Grove"
	label.position.y = 4.7
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	node.add_child(label)
	add_child(node)


func _add_glowcap_cluster(at: Vector3) -> void:
	var node := Node3D.new()
	node.position = at
	var cap_mat := StandardMaterial3D.new()
	cap_mat.albedo_color = Color(0.24, 0.86, 0.76)
	cap_mat.emission_enabled = true
	cap_mat.emission = Color(0.10, 0.45, 0.38)
	var cap := MeshInstance3D.new()
	var cap_mesh := SphereMesh.new()
	cap_mesh.radius = 0.48
	cap_mesh.height = 0.55
	cap.mesh = cap_mesh
	cap.position.y = 0.52
	cap.material_override = cap_mat
	node.add_child(cap)
	var label := Label3D.new()
	label.text = "Bioluminescent Glowcap Cluster"
	label.position.y = 1.25
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	node.add_child(label)
	add_child(node)


func _add_sunblossom_patch(at: Vector3) -> void:
	var node := Node3D.new()
	node.position = at
	var bloom_mat := StandardMaterial3D.new()
	bloom_mat.albedo_color = Color(0.96, 0.72, 0.18)
	bloom_mat.emission_enabled = true
	bloom_mat.emission = Color(0.45, 0.28, 0.04)
	var bloom := MeshInstance3D.new()
	var bloom_mesh := SphereMesh.new()
	bloom_mesh.radius = 0.46
	bloom_mesh.height = 0.60
	bloom.mesh = bloom_mesh
	bloom.position.y = 0.50
	bloom.material_override = bloom_mat
	node.add_child(bloom)
	var label := Label3D.new()
	label.text = "Solaris Sunblossom Patch"
	label.position.y = 1.25
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	node.add_child(label)
	add_child(node)


func _add_tuber_bed(at: Vector3) -> void:
	var node := Node3D.new()
	node.position = at
	var soil_mat := StandardMaterial3D.new()
	soil_mat.albedo_color = Color(0.34, 0.24, 0.14)
	var bed := MeshInstance3D.new()
	var bed_mesh := BoxMesh.new()
	bed_mesh.size = Vector3(1.5, 0.24, 1.5)
	bed.mesh = bed_mesh
	bed.position.y = 0.12
	bed.material_override = soil_mat
	node.add_child(bed)
	var label := Label3D.new()
	label.text = "Irrigated Heart-Tuber Bed"
	label.position.y = 1.15
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	node.add_child(label)
	add_child(node)


func _add_starberry_arbor(at: Vector3) -> void:
	var node := Node3D.new()
	node.position = at
	var vine_mat := StandardMaterial3D.new()
	vine_mat.albedo_color = Color(0.68, 0.24, 0.56)
	var arbor := MeshInstance3D.new()
	var arbor_mesh := BoxMesh.new()
	arbor_mesh.size = Vector3(1.4, 1.1, 0.9)
	arbor.mesh = arbor_mesh
	arbor.position.y = 0.55
	arbor.material_override = vine_mat
	node.add_child(arbor)
	var label := Label3D.new()
	label.text = "Hydroponic Starberry Arbor"
	label.position.y = 1.45
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	node.add_child(label)
	add_child(node)


func _add_circuit_relay(at: Vector3) -> void:
	var node := Node3D.new()
	node.position = at
	var relay_mat := StandardMaterial3D.new()
	relay_mat.albedo_color = Color(0.28, 0.52, 0.58)
	var pylon := MeshInstance3D.new()
	var pylon_mesh := BoxMesh.new()
	pylon_mesh.size = Vector3(0.75, 1.35, 0.75)
	pylon.mesh = pylon_mesh
	pylon.position.y = 0.68
	pylon.material_override = relay_mat
	node.add_child(pylon)
	var label := Label3D.new()
	label.text = "Abandoned Relay Array"
	label.position.y = 1.65
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	node.add_child(label)
	add_child(node)


func _add_orbital_pod_wreckage(at: Vector3) -> void:
	var node := Node3D.new()
	node.position = at
	var hull_mat := StandardMaterial3D.new()
	hull_mat.albedo_color = Color(0.52, 0.56, 0.66)
	var pod := MeshInstance3D.new()
	var pod_mesh := SphereMesh.new()
	pod_mesh.radius = 0.78
	pod_mesh.height = 1.30
	pod.mesh = pod_mesh
	pod.position.y = 0.60
	pod.material_override = hull_mat
	node.add_child(pod)
	var label := Label3D.new()
	label.text = "Crashed Orbital Pod Wreckage"
	label.position.y = 1.65
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	node.add_child(label)
	add_child(node)


func _add_elderwood_tree(at: Vector3) -> void:
	var node := Node3D.new()
	node.position = at
	var trunk_mat := StandardMaterial3D.new()
	trunk_mat.albedo_color = Color(0.62, 0.66, 0.74)
	var trunk := MeshInstance3D.new()
	var trunk_mesh := CylinderMesh.new()
	trunk_mesh.top_radius = 0.32
	trunk_mesh.bottom_radius = 0.48
	trunk_mesh.height = 3.0
	trunk.mesh = trunk_mesh
	trunk.position.y = 1.5
	trunk.material_override = trunk_mat
	node.add_child(trunk)
	var canopy_mat := StandardMaterial3D.new()
	canopy_mat.albedo_color = Color(0.32, 0.82, 0.76)
	canopy_mat.emission_enabled = true
	canopy_mat.emission = Color(0.10, 0.34, 0.32)
	var canopy := MeshInstance3D.new()
	var canopy_mesh := SphereMesh.new()
	canopy_mesh.radius = 1.55
	canopy_mesh.height = 2.8
	canopy.mesh = canopy_mesh
	canopy.position.y = 3.4
	canopy.material_override = canopy_mat
	node.add_child(canopy)
	var label := Label3D.new()
	label.text = "Ancient Aether-Bough Tree"
	label.position.y = 5.0
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	node.add_child(label)
	add_child(node)


func _add_voidbark_tree(at: Vector3) -> void:
	var node := Node3D.new()
	node.position = at
	var trunk_mat := StandardMaterial3D.new()
	trunk_mat.albedo_color = Color(0.16, 0.14, 0.24)
	var trunk := MeshInstance3D.new()
	var trunk_mesh := CylinderMesh.new()
	trunk_mesh.top_radius = 0.36
	trunk_mesh.bottom_radius = 0.54
	trunk_mesh.height = 3.2
	trunk.mesh = trunk_mesh
	trunk.position.y = 1.6
	trunk.material_override = trunk_mat
	node.add_child(trunk)
	var canopy_mat := StandardMaterial3D.new()
	canopy_mat.albedo_color = Color(0.58, 0.26, 0.82)
	canopy_mat.emission_enabled = true
	canopy_mat.emission = Color(0.28, 0.10, 0.44)
	var canopy := MeshInstance3D.new()
	var canopy_mesh := SphereMesh.new()
	canopy_mesh.radius = 1.65
	canopy_mesh.height = 2.9
	canopy.mesh = canopy_mesh
	canopy.position.y = 3.6
	canopy.material_override = canopy_mat
	node.add_child(canopy)
	var label := Label3D.new()
	label.text = "Abyssal Voidbark Great-Tree"
	label.position.y = 5.2
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	node.add_child(label)
	add_child(node)


func _add_frostlotus_bloom(at: Vector3) -> void:
	var node := Node3D.new()
	node.position = at
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(0.68, 0.92, 0.98)
	mat.emission_enabled = true
	mat.emission = Color(0.22, 0.52, 0.68)
	var bloom := MeshInstance3D.new()
	var mesh := SphereMesh.new()
	mesh.radius = 0.52
	mesh.height = 0.64
	bloom.mesh = mesh
	bloom.position.y = 0.52
	bloom.material_override = mat
	node.add_child(bloom)
	var label := Label3D.new()
	label.text = "Cryo-Lotus Ridge Bloom"
	label.position.y = 1.35
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	node.add_child(label)
	add_child(node)


func _add_astralspore_colony(at: Vector3) -> void:
	var node := Node3D.new()
	node.position = at
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(0.84, 0.46, 0.96)
	mat.emission_enabled = true
	mat.emission = Color(0.42, 0.18, 0.56)
	var colony := MeshInstance3D.new()
	var mesh := SphereMesh.new()
	mesh.radius = 0.56
	mesh.height = 0.72
	colony.mesh = mesh
	colony.position.y = 0.56
	colony.material_override = mat
	node.add_child(colony)
	var label := Label3D.new()
	label.text = "Astral Spore Colony"
	label.position.y = 1.40
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	node.add_child(label)
	add_child(node)


func _add_goldengrain_field(at: Vector3) -> void:
	var node := Node3D.new()
	node.position = at
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(0.94, 0.78, 0.24)
	var field := MeshInstance3D.new()
	var mesh := BoxMesh.new()
	mesh.size = Vector3(1.65, 0.65, 1.65)
	field.mesh = mesh
	field.position.y = 0.33
	field.material_override = mat
	node.add_child(field)
	var label := Label3D.new()
	label.text = "Terraced Auric Grain Field"
	label.position.y = 1.35
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	node.add_child(label)
	add_child(node)


func _add_celestialfruit_orchard(at: Vector3) -> void:
	var node := Node3D.new()
	node.position = at
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(0.96, 0.54, 0.32)
	mat.emission_enabled = true
	mat.emission = Color(0.42, 0.20, 0.10)
	var arbor := MeshInstance3D.new()
	var mesh := BoxMesh.new()
	mesh.size = Vector3(1.5, 1.25, 1.0)
	arbor.mesh = mesh
	arbor.position.y = 0.63
	arbor.material_override = mat
	node.add_child(arbor)
	var label := Label3D.new()
	label.text = "Celestial Nectar Orchard"
	label.position.y = 1.60
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	node.add_child(label)
	add_child(node)


func _add_servoarray_chassis(at: Vector3) -> void:
	var node := Node3D.new()
	node.position = at
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(0.46, 0.52, 0.58)
	var chassis := MeshInstance3D.new()
	var mesh := BoxMesh.new()
	mesh.size = Vector3(1.1, 1.2, 0.9)
	chassis.mesh = mesh
	chassis.position.y = 0.60
	chassis.material_override = mat
	node.add_child(chassis)
	var label := Label3D.new()
	label.text = "Derelict Mech Chassis"
	label.position.y = 1.65
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	node.add_child(label)
	add_child(node)


func _add_singularity_wreckage(at: Vector3) -> void:
	var node := Node3D.new()
	node.position = at
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(0.36, 0.68, 0.96)
	mat.emission_enabled = true
	mat.emission = Color(0.16, 0.36, 0.62)
	var core := MeshInstance3D.new()
	var mesh := SphereMesh.new()
	mesh.radius = 0.82
	mesh.height = 1.45
	core.mesh = mesh
	core.position.y = 0.72
	core.material_override = mat
	node.add_child(core)
	var label := Label3D.new()
	label.text = "Shattered Warp Core Wreckage"
	label.position.y = 1.85
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	node.add_child(label)
	add_child(node)


func _add_runic_banite_monolith(at: Vector3) -> void:
	var node := Node3D.new()
	node.position = at
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(0.18, 0.78, 0.86)
	mat.emission_enabled = true
	mat.emission = Color(0.08, 0.42, 0.52)
	mat.roughness = 0.22
	var spire := MeshInstance3D.new()
	var mesh := CylinderMesh.new()
	mesh.top_radius = 0.24
	mesh.bottom_radius = 0.72
	mesh.height = 2.25
	spire.mesh = mesh
	spire.position.y = 1.12
	spire.material_override = mat
	node.add_child(spire)
	var label := Label3D.new()
	label.text = "Runic Banite Monolith"
	label.position.y = 2.55
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	node.add_child(label)
	add_child(node)


func _add_elder_wroshyr_tree(at: Vector3) -> void:
	var node := Node3D.new()
	node.position = at
	var trunk_mat := StandardMaterial3D.new()
	trunk_mat.albedo_color = Color(0.36, 0.22, 0.12)
	var canopy_mat := StandardMaterial3D.new()
	canopy_mat.albedo_color = Color(0.16, 0.68, 0.44)
	canopy_mat.emission_enabled = true
	canopy_mat.emission = Color(0.06, 0.28, 0.16)
	var trunk := MeshInstance3D.new()
	var trunk_mesh := CylinderMesh.new()
	trunk_mesh.top_radius = 0.42
	trunk_mesh.bottom_radius = 0.68
	trunk_mesh.height = 3.1
	trunk.mesh = trunk_mesh
	trunk.position.y = 1.55
	trunk.material_override = trunk_mat
	node.add_child(trunk)
	var canopy := MeshInstance3D.new()
	var canopy_mesh := SphereMesh.new()
	canopy_mesh.radius = 1.95
	canopy_mesh.height = 2.8
	canopy.mesh = canopy_mesh
	canopy.position.y = 3.65
	canopy.material_override = canopy_mat
	node.add_child(canopy)
	var label := Label3D.new()
	label.text = "Elder Wroshyr-Heart Great-Tree"
	label.position.y = 5.35
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	node.add_child(label)
	add_child(node)


func _add_torstol_bloom_patch(at: Vector3) -> void:
	var node := Node3D.new()
	node.position = at
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(0.24, 0.92, 0.56)
	mat.emission_enabled = true
	mat.emission = Color(0.10, 0.48, 0.24)
	var bloom := MeshInstance3D.new()
	var mesh := SphereMesh.new()
	mesh.radius = 0.72
	mesh.height = 1.15
	bloom.mesh = mesh
	bloom.position.y = 0.58
	bloom.material_override = mat
	node.add_child(bloom)
	var label := Label3D.new()
	label.text = "Torstol-Bacta Sovereign Patch"
	label.position.y = 1.55
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	node.add_child(label)
	add_child(node)


func _add_spirit_muja_arbor(at: Vector3) -> void:
	var node := Node3D.new()
	node.position = at
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(0.98, 0.78, 0.24)
	mat.emission_enabled = true
	mat.emission = Color(0.46, 0.34, 0.08)
	var arbor := MeshInstance3D.new()
	var mesh := BoxMesh.new()
	mesh.size = Vector3(1.65, 1.45, 1.15)
	arbor.mesh = mesh
	arbor.position.y = 0.72
	arbor.material_override = mat
	node.add_child(arbor)
	var label := Label3D.new()
	label.text = "Spirit-Muja Sacred Arbor"
	label.position.y = 1.85
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	node.add_child(label)
	add_child(node)


func _add_rakatan_hypercore_relic(at: Vector3) -> void:
	var node := Node3D.new()
	node.position = at
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(0.92, 0.34, 0.78)
	mat.emission_enabled = true
	mat.emission = Color(0.48, 0.14, 0.40)
	var relic := MeshInstance3D.new()
	var mesh := CylinderMesh.new()
	mesh.top_radius = 0.55
	mesh.bottom_radius = 0.75
	mesh.height = 1.85
	relic.mesh = mesh
	relic.position.y = 0.92
	relic.material_override = mat
	node.add_child(relic)
	var label := Label3D.new()
	label.text = "Ancient Hyper-Gate Relic"
	label.position.y = 2.25
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	node.add_child(label)
	add_child(node)


func _add_orikalkum_beskar_vein(at: Vector3) -> void:
	var node := Node3D.new()
	node.position = at
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(0.84, 0.24, 0.18)
	mat.emission_enabled = true
	mat.emission = Color(0.42, 0.08, 0.06)
	var vein := MeshInstance3D.new()
	var mesh := CylinderMesh.new()
	mesh.top_radius = 0.28
	mesh.bottom_radius = 0.65
	mesh.height = 1.95
	vein.mesh = mesh
	vein.position.y = 0.98
	vein.material_override = mat
	node.add_child(vein)
	var label := Label3D.new()
	label.text = "Mandalorian Orikalkum Vein"
	label.position.y = 2.35
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	node.add_child(label)
	add_child(node)


func _add_felucian_magic_copse(at: Vector3) -> void:
	var node := Node3D.new()
	node.position = at
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(0.28, 0.88, 0.82)
	mat.emission_enabled = true
	mat.emission = Color(0.10, 0.42, 0.40)
	var cap := MeshInstance3D.new()
	var mesh := SphereMesh.new()
	mesh.radius = 1.45
	mesh.height = 2.15
	cap.mesh = mesh
	cap.position.y = 1.45
	cap.material_override = mat
	node.add_child(cap)
	var label := Label3D.new()
	label.text = "Felucian Magic-Spore Copse"
	label.position.y = 2.95
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	node.add_child(label)
	add_child(node)


func _add_snapdragon_kolto_glade(at: Vector3) -> void:
	var node := Node3D.new()
	node.position = at
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(0.92, 0.28, 0.48)
	mat.emission_enabled = true
	mat.emission = Color(0.45, 0.10, 0.22)
	var frond := MeshInstance3D.new()
	var mesh := SphereMesh.new()
	mesh.radius = 0.76
	mesh.height = 1.25
	frond.mesh = mesh
	frond.position.y = 0.62
	frond.material_override = mat
	node.add_child(frond)
	var label := Label3D.new()
	label.text = "Snapdragon Kolto-Fern Glade"
	label.position.y = 1.65
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	node.add_child(label)
	add_child(node)


func _add_cursed_blight_trellis(at: Vector3) -> void:
	var node := Node3D.new()
	node.position = at
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(0.56, 0.22, 0.78)
	mat.emission_enabled = true
	mat.emission = Color(0.24, 0.08, 0.36)
	var trellis := MeshInstance3D.new()
	var mesh := BoxMesh.new()
	mesh.size = Vector3(1.65, 1.50, 1.15)
	trellis.mesh = mesh
	trellis.position.y = 0.75
	trellis.material_override = mat
	node.add_child(trellis)
	var label := Label3D.new()
	label.text = "Cursed Blight-Plum Trellis"
	label.position.y = 1.90
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	node.add_child(label)
	add_child(node)


func _add_gree_aether_conduit(at: Vector3) -> void:
	var node := Node3D.new()
	node.position = at
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(0.24, 0.76, 0.96)
	mat.emission_enabled = true
	mat.emission = Color(0.10, 0.38, 0.56)
	var pylon := MeshInstance3D.new()
	var mesh := BoxMesh.new()
	mesh.size = Vector3(0.95, 2.05, 0.95)
	pylon.mesh = mesh
	pylon.position.y = 1.02
	pylon.material_override = mat
	node.add_child(pylon)
	var label := Label3D.new()
	label.text = "Ancient Gree Aether-Conduit"
	label.position.y = 2.45
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	node.add_child(label)
	add_child(node)


func _add_primordial_animica_monolith(at: Vector3) -> void:
	var node := Node3D.new()
	node.position = at
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(0.96, 0.78, 0.22)
	mat.emission_enabled = true
	mat.emission = Color(0.52, 0.38, 0.08)
	var monolith := MeshInstance3D.new()
	var mesh := CylinderMesh.new()
	mesh.top_radius = 0.18
	mesh.bottom_radius = 0.72
	mesh.height = 2.35
	monolith.mesh = mesh
	monolith.position.y = 1.18
	monolith.material_override = mat
	node.add_child(monolith)
	var label := Label3D.new()
	label.text = "Primordial Animica-Phrik Monolith"
	label.position.y = 2.75
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	node.add_child(label)
	add_child(node)


func _add_bloodwood_brylark_tree(at: Vector3) -> void:
	var node := Node3D.new()
	node.position = at
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(0.76, 0.14, 0.20)
	mat.emission_enabled = true
	mat.emission = Color(0.36, 0.04, 0.08)
	var tree := MeshInstance3D.new()
	var mesh := SphereMesh.new()
	mesh.radius = 1.85
	mesh.height = 2.75
	tree.mesh = mesh
	tree.position.y = 2.15
	tree.material_override = mat
	node.add_child(tree)
	var label := Label3D.new()
	label.text = "Wilderness Bloodwood-Brylark Patriarch"
	label.position.y = 3.95
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	node.add_child(label)
	add_child(node)


func _add_fellstalk_ryll_sanctuary(at: Vector3) -> void:
	var node := Node3D.new()
	node.position = at
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(0.68, 0.28, 0.94)
	mat.emission_enabled = true
	mat.emission = Color(0.34, 0.12, 0.52)
	var bloom := MeshInstance3D.new()
	var mesh := SphereMesh.new()
	mesh.radius = 0.84
	mesh.height = 1.35
	bloom.mesh = mesh
	bloom.position.y = 0.68
	bloom.material_override = mat
	node.add_child(bloom)
	var label := Label3D.new()
	label.text = "Nightsister Fellstalk-Ryll Sanctuary"
	label.position.y = 1.75
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	node.add_child(label)
	add_child(node)


func _add_golden_dragon_conservatory(at: Vector3) -> void:
	var node := Node3D.new()
	node.position = at
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(0.98, 0.62, 0.18)
	mat.emission_enabled = true
	mat.emission = Color(0.52, 0.28, 0.06)
	var cactus := MeshInstance3D.new()
	var mesh := CylinderMesh.new()
	mesh.top_radius = 0.42
	mesh.bottom_radius = 0.56
	mesh.height = 1.80
	cactus.mesh = mesh
	cactus.position.y = 0.90
	cactus.material_override = mat
	node.add_child(cactus)
	var label := Label3D.new()
	label.text = "Sovereign Dragon-Cactus Conservatory"
	label.position.y = 2.20
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	node.add_child(label)
	add_child(node)


func _add_starforge_reliquary(at: Vector3) -> void:
	var node := Node3D.new()
	node.position = at
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(0.96, 0.88, 0.36)
	mat.emission_enabled = true
	mat.emission = Color(0.58, 0.46, 0.14)
	var core := MeshInstance3D.new()
	var mesh := SphereMesh.new()
	mesh.radius = 0.92
	mesh.height = 1.85
	core.mesh = mesh
	core.position.y = 1.05
	core.material_override = mat
	node.add_child(core)
	var label := Label3D.new()
	label.text = "Celestial Star-Forge Singularity Reliquary"
	label.position.y = 2.45
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	node.add_child(label)
	add_child(node)


func _add_promethium_stygium_rift(at: Vector3) -> void:
	var node := Node3D.new()
	node.position = at
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(0.68, 0.14, 0.28)
	mat.emission_enabled = true
	mat.emission = Color(0.32, 0.04, 0.14)
	var rock := MeshInstance3D.new()
	var mesh := CylinderMesh.new()
	mesh.top_radius = 0.28
	mesh.bottom_radius = 0.88
	mesh.height = 1.9
	rock.mesh = mesh
	rock.position.y = 0.95
	rock.material_override = mat
	node.add_child(rock)
	var label := Label3D.new()
	label.text = "Promethium-Stygium Prism Rift"
	label.position.y = 2.25
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	node.add_child(label)
	add_child(node)


func _add_umbaran_crystal_grove(at: Vector3) -> void:
	var node := Node3D.new()
	node.position = at
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(0.24, 0.68, 0.88)
	mat.emission_enabled = true
	mat.emission = Color(0.08, 0.28, 0.42)
	var trunk := MeshInstance3D.new()
	var mesh := CylinderMesh.new()
	mesh.top_radius = 0.34
	mesh.bottom_radius = 0.56
	mesh.height = 2.25
	trunk.mesh = mesh
	trunk.position.y = 1.12
	trunk.material_override = mat
	node.add_child(trunk)
	var label := Label3D.new()
	label.text = "Umbaran Crystal-Pine Grove"
	label.position.y = 2.55
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	node.add_child(label)
	add_child(node)


func _add_arbuck_zevrona_thicket(at: Vector3) -> void:
	var node := Node3D.new()
	node.position = at
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(0.34, 0.86, 0.54)
	mat.emission_enabled = true
	mat.emission = Color(0.10, 0.36, 0.22)
	var bloom := MeshInstance3D.new()
	var mesh := SphereMesh.new()
	mesh.radius = 0.72
	mesh.height = 1.15
	bloom.mesh = mesh
	bloom.position.y = 0.62
	bloom.material_override = mat
	node.add_child(bloom)
	var label := Label3D.new()
	label.text = "Arbuck-Zevrona Spore Thicket"
	label.position.y = 1.65
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	node.add_child(label)
	add_child(node)


func _add_mustafar_lavafruit_conservatory(at: Vector3) -> void:
	var node := Node3D.new()
	node.position = at
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(0.92, 0.38, 0.14)
	mat.emission_enabled = true
	mat.emission = Color(0.44, 0.14, 0.04)
	var cactus := MeshInstance3D.new()
	var mesh := CylinderMesh.new()
	mesh.top_radius = 0.42
	mesh.bottom_radius = 0.52
	mesh.height = 1.65
	cactus.mesh = mesh
	cactus.position.y = 0.82
	cactus.material_override = mat
	node.add_child(cactus)
	var label := Label3D.new()
	label.text = "Mustafar Obsidian Lavafruit Conservatory"
	label.position.y = 2.05
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	node.add_child(label)
	add_child(node)


func _add_massassi_chronote_obelisk(at: Vector3) -> void:
	var node := Node3D.new()
	node.position = at
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(0.58, 0.26, 0.84)
	mat.emission_enabled = true
	mat.emission = Color(0.24, 0.08, 0.42)
	var obelisk := MeshInstance3D.new()
	var mesh := BoxMesh.new()
	mesh.size = Vector3(0.95, 1.95, 0.95)
	obelisk.mesh = mesh
	obelisk.position.y = 0.98
	obelisk.material_override = mat
	node.add_child(obelisk)
	var label := Label3D.new()
	label.text = "Massassi Chronote Obelisk Vault"
	label.position.y = 2.35
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	node.add_child(label)
	add_child(node)


func _add_primal_kyber_monolith(at: Vector3) -> void:
	var node := Node3D.new()
	node.position = at
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(0.94, 0.22, 0.36)
	mat.emission_enabled = true
	mat.emission = Color(0.52, 0.08, 0.18)
	var crystal := MeshInstance3D.new()
	var mesh := CylinderMesh.new()
	mesh.top_radius = 0.22
	mesh.bottom_radius = 0.96
	mesh.height = 2.35
	crystal.mesh = mesh
	crystal.position.y = 1.18
	crystal.material_override = mat
	node.add_child(crystal)
	var label := Label3D.new()
	label.text = "Primal Kyber-Novite Monolith"
	label.position.y = 2.68
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	node.add_child(label)
	add_child(node)


func _add_solak_anima_bough(at: Vector3) -> void:
	var node := Node3D.new()
	node.position = at
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(0.28, 0.88, 0.72)
	mat.emission_enabled = true
	mat.emission = Color(0.10, 0.42, 0.32)
	var tree := MeshInstance3D.new()
	var mesh := CylinderMesh.new()
	mesh.top_radius = 0.46
	mesh.bottom_radius = 0.72
	mesh.height = 2.65
	tree.mesh = mesh
	tree.position.y = 1.32
	tree.material_override = mat
	node.add_child(tree)
	var label := Label3D.new()
	label.text = "Solak Anima-Wroshyr Worldtree Bough"
	label.position.y = 2.95
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	node.add_child(label)
	add_child(node)


func _add_bloodweed_vong_hive(at: Vector3) -> void:
	var node := Node3D.new()
	node.position = at
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(0.84, 0.14, 0.34)
	mat.emission_enabled = true
	mat.emission = Color(0.42, 0.04, 0.16)
	var hive := MeshInstance3D.new()
	var mesh := SphereMesh.new()
	mesh.radius = 0.82
	mesh.height = 1.35
	hive.mesh = mesh
	hive.position.y = 0.72
	hive.material_override = mat
	node.add_child(hive)
	var label := Label3D.new()
	label.text = "Wilderness Bloodweed-Vong Spore Hive"
	label.position.y = 1.85
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	node.add_child(label)
	add_child(node)


func _add_golden_ambrosia_sanctuary(at: Vector3) -> void:
	var node := Node3D.new()
	node.position = at
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(0.96, 0.82, 0.24)
	mat.emission_enabled = true
	mat.emission = Color(0.48, 0.36, 0.08)
	var dome := MeshInstance3D.new()
	var mesh := SphereMesh.new()
	mesh.radius = 0.88
	mesh.height = 1.55
	dome.mesh = mesh
	dome.position.y = 0.82
	dome.material_override = mat
	node.add_child(dome)
	var label := Label3D.new()
	label.text = "Celestial Mortis-Ambrosia Melon Sanctuary"
	label.position.y = 2.05
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	node.add_child(label)
	add_child(node)


func _add_infinity_holocron_vault(at: Vector3) -> void:
	var node := Node3D.new()
	node.position = at
	var mat := StandardMaterial3D.new()
	mat.albedo_color = Color(0.42, 0.86, 0.98)
	mat.emission_enabled = true
	mat.emission = Color(0.18, 0.44, 0.62)
	var core := MeshInstance3D.new()
	var mesh := SphereMesh.new()
	mesh.radius = 0.98
	mesh.height = 1.96
	core.mesh = mesh
	core.position.y = 1.12
	core.material_override = mat
	node.add_child(core)
	var label := Label3D.new()
	label.text = "Kwa Infinity-Gate Holocron Vault"
	label.position.y = 2.55
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED
	label.font_size = 22
	node.add_child(label)
	add_child(node)


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
		at.y = _terrain_height_at(at.x, at.z)
		_add_enemy_visual(enemy_id, at, str(enemy.get("name", "Frontier Creature")), int(enemy.get("maxHealth", 10)))

func _add_enemy_visual(enemy_id: String, at: Vector3, enemy_name: String, max_health: int) -> void:
	var creature := Node3D.new()
	creature.name = "Enemy_" + enemy_id
	creature.position = at
	var is_boss := enemy_id == "stoneback-colossus" or enemy_id == "krayt-jad-sovereign" or enemy_id == "tzkal-mythosaur-emperor" or enemy_id == "zamorak-zillo-apocalypse"
	var hide_material := StandardMaterial3D.new()
	if enemy_id == "zamorak-zillo-apocalypse":
		hide_material.albedo_color = Color(0.72, 0.08, 0.24)
		hide_material.emission_enabled = true
		hide_material.emission = Color(0.48, 0.04, 0.18)
	elif enemy_id == "tzkal-mythosaur-emperor":
		hide_material.albedo_color = Color(0.88, 0.64, 0.16)
		hide_material.emission_enabled = true
		hide_material.emission = Color(0.46, 0.24, 0.04)
	elif enemy_id == "krayt-jad-sovereign":
		hide_material.albedo_color = Color(0.82, 0.22, 0.14)
		hide_material.emission_enabled = true
		hide_material.emission = Color(0.42, 0.10, 0.04)
	elif is_boss:
		hide_material.albedo_color = Color(0.27, 0.29, 0.30)
	elif enemy_id.begins_with("canyon-stalker"):
		hide_material.albedo_color = Color(0.58, 0.42, 0.22)
	elif enemy_id.begins_with("mire-drake"):
		hide_material.albedo_color = Color(0.18, 0.46, 0.32)
	elif enemy_id.begins_with("ash-hound"):
		hide_material.albedo_color = Color(0.42, 0.18, 0.16)
	elif enemy_id.begins_with("frost-prowler"):
		hide_material.albedo_color = Color(0.68, 0.82, 0.92)
	elif enemy_id.begins_with("ruin-sentinel"):
		hide_material.albedo_color = Color(0.44, 0.48, 0.54)
	elif enemy_id.begins_with("aether-wraith"):
		hide_material.albedo_color = Color(0.42, 0.26, 0.68)
	elif enemy_id.begins_with("feral-drone"):
		hide_material.albedo_color = Color(0.62, 0.54, 0.26)
	elif enemy_id.begins_with("thornback-behemoth"):
		hide_material.albedo_color = Color(0.34, 0.42, 0.22)
	elif enemy_id.begins_with("obsidian-wyrm"):
		hide_material.albedo_color = Color(0.18, 0.14, 0.22)
	elif enemy_id.begins_with("void-reaver"):
		hide_material.albedo_color = Color(0.54, 0.18, 0.48)
	elif enemy_id.begins_with("star-golem"):
		hide_material.albedo_color = Color(0.72, 0.80, 0.94)
	elif enemy_id.begins_with("dune-scorpion"):
		hide_material.albedo_color = Color(0.76, 0.58, 0.28)
	elif enemy_id.begins_with("crystalline-mantis"):
		hide_material.albedo_color = Color(0.28, 0.84, 0.62)
	elif enemy_id.begins_with("magma-crusher"):
		hide_material.albedo_color = Color(0.72, 0.24, 0.12)
	elif enemy_id.begins_with("rift-specter"):
		hide_material.albedo_color = Color(0.36, 0.68, 0.92)
	elif enemy_id.begins_with("apex-devourer"):
		hide_material.albedo_color = Color(0.58, 0.12, 0.24)
	elif enemy_id.begins_with("ironbark-treant"):
		hide_material.albedo_color = Color(0.28, 0.52, 0.30)
	elif enemy_id.begins_with("cryo-stalker"):
		hide_material.albedo_color = Color(0.56, 0.88, 0.98)
	elif enemy_id.begins_with("solar-drake"):
		hide_material.albedo_color = Color(0.94, 0.58, 0.16)
	elif enemy_id.begins_with("chrono-stalker"):
		hide_material.albedo_color = Color(0.62, 0.44, 0.92)
	elif enemy_id.begins_with("abyssal-goliath"):
		hide_material.albedo_color = Color(0.18, 0.16, 0.28)
	elif enemy_id.begins_with("dune-krayt-drake"):
		hide_material.albedo_color = Color(0.84, 0.68, 0.34)
	elif enemy_id.begins_with("dathomir-hexweaver"):
		hide_material.albedo_color = Color(0.58, 0.18, 0.68)
	elif enemy_id.begins_with("acklay-reaper"):
		hide_material.albedo_color = Color(0.22, 0.72, 0.48)
	elif enemy_id.begins_with("gorax-behemoth"):
		hide_material.albedo_color = Color(0.46, 0.34, 0.24)
	elif enemy_id.begins_with("terentatek-mauler"):
		hide_material.albedo_color = Color(0.68, 0.16, 0.22)
	elif enemy_id.begins_with("felucian-bull-rancor"):
		hide_material.albedo_color = Color(0.24, 0.68, 0.58)
	elif enemy_id.begins_with("mandalore-basilisk"):
		hide_material.albedo_color = Color(0.62, 0.66, 0.74)
	elif enemy_id.begins_with("mythosaur-reaper"):
		hide_material.albedo_color = Color(0.82, 0.24, 0.36)
	elif enemy_id.begins_with("yuuzhan-vonduun-stalker"):
		hide_material.albedo_color = Color(0.28, 0.78, 0.42)
	elif enemy_id.begins_with("yavin-massassi-warlord"):
		hide_material.albedo_color = Color(0.76, 0.22, 0.18)
	elif enemy_id.begins_with("hk-nex-praetorian"):
		hide_material.albedo_color = Color(0.48, 0.24, 0.78)
	elif enemy_id.begins_with("zillo-vorago-behemoth"):
		hide_material.albedo_color = Color(0.34, 0.42, 0.54)
	else:
		hide_material.albedo_color = Color(0.48, 0.30, 0.18) if max_health < 20 else Color(0.32, 0.25, 0.18)
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

	var bars_row := HBoxContainer.new()
	bars_row.add_theme_constant_override("separation", 6)
	column.add_child(bars_row)
	hud_health_bar = ProgressBar.new()
	hud_health_bar.max_value = 100
	hud_health_bar.value = 100
	hud_health_bar.show_percentage = false
	hud_health_bar.custom_minimum_size = Vector2(90, 12)
	hud_health_bar.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	bars_row.add_child(hud_health_bar)
	hud_action_bar = ProgressBar.new()
	hud_action_bar.max_value = 100
	hud_action_bar.value = 100
	hud_action_bar.show_percentage = false
	hud_action_bar.custom_minimum_size = Vector2(90, 12)
	hud_action_bar.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	bars_row.add_child(hud_action_bar)
	hud_mind_bar = ProgressBar.new()
	hud_mind_bar.max_value = 100
	hud_mind_bar.value = 100
	hud_mind_bar.show_percentage = false
	hud_mind_bar.custom_minimum_size = Vector2(90, 12)
	hud_mind_bar.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	bars_row.add_child(hud_mind_bar)

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
	stats_button = _create_action_button(action_toolbar, "Vitals", _show_stats_panel)
	stats_button.visible = false
	bag_button = _create_action_button(action_toolbar, "Bag", _show_inventory_panel)
	bag_button.visible = false
	crafting_button = _create_action_button(action_toolbar, "Craft", func() -> void: _show_crafting_panel())
	crafting_button.visible = false
	if stats_button != null:
		stats_button.visible = false
	if guide_button != null:
		guide_button.visible = false
	if quick_heal_button != null:
		quick_heal_button.visible = false
	if settings_button != null:
		settings_button.visible = false
	if settings_panel != null:
		settings_panel.visible = false
	if stats_panel != null:
		stats_panel.visible = false
	if guide_panel != null:
		guide_panel.visible = false
	guide_button = _create_action_button(action_toolbar, "Atlas", _show_guide_panel)
	guide_button.visible = false
	quick_heal_button = _create_action_button(action_toolbar, "Med/Stim", _use_best_consumable)
	quick_heal_button.visible = false
	rest_button = _create_action_button(action_toolbar, "Rest", _rest_character)
	rest_button.visible = false
	equipment_button = _create_action_button(action_toolbar, "Gear", _show_equipment_panel)
	equipment_button.visible = false
	settings_button = _create_action_button(action_toolbar, "Options", _show_settings_panel)
	settings_button.visible = false
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
		var recipes: Variant = payload.get("recipes", [])
		if recipes is Array:
			recipe_catalog = recipes
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


func _select_forge_recipe() -> String:
	if int(current_inventory.get("primal_kyber_bar", 0)) >= 2 and int(current_inventory.get("zillo_leviathan_sovereign_horn", 0)) >= 1 and int(current_inventory.get("zillo_vorago_tectonic_plate", 0)) >= 1:
		return "zillo-tectonic-bulwark-cuirass"
	if int(current_inventory.get("zamorak_nihilus_chaos_core", 0)) >= 1 and int(current_inventory.get("zillo_leviathan_sovereign_horn", 0)) >= 1 and int(current_inventory.get("primal_kyber_bar", 0)) >= 1:
		return "nihilus-chaos-sovereign-crown"
	if int(current_inventory.get("primal_kyber_bar", 0)) >= 2 and int(current_inventory.get("zamorak_nihilus_chaos_core", 0)) >= 1 and int(current_inventory.get("infinity_gate_matrix", 0)) >= 1:
		return "ek-zekkil-kyber-greatsaber"
	if int(current_inventory.get("primal_kyber_bar", 0)) >= 1 and int(current_inventory.get("zillo_vorago_tectonic_plate", 0)) >= 2 and int(current_inventory.get("solak_anima_heartwood", 0)) >= 1:
		return "solak-zillo-sovereign-halberd"
	if int(current_inventory.get("primal_kyber_bar", 0)) >= 2 and int(current_inventory.get("zillo_vorago_tectonic_plate", 0)) >= 2:
		return "primal-beskar-juggernaut-plate"
	if int(current_inventory.get("primal_kyber_bar", 0)) >= 2 and int(current_inventory.get("zillo_vorago_tectonic_plate", 0)) >= 1 and int(current_inventory.get("solak_anima_heartwood", 0)) >= 1:
		return "primal-zillo-seismic-maul"
	if int(current_inventory.get("primal_kyber_ore", 0)) >= 2 and int(current_inventory.get("promethium_stygium_bar", 0)) >= 1:
		return "primal-kyber-bar"
	if int(current_inventory.get("promethium_stygium_bar", 0)) >= 2 and int(current_inventory.get("massassi_chronote_relic", 0)) >= 1 and int(current_inventory.get("massassi_kalgerion_horn", 0)) >= 1:
		return "stygium-shadow-lightsaber"
	if int(current_inventory.get("promethium_stygium_bar", 0)) >= 1 and int(current_inventory.get("vonduun_amphistaff_chitin", 0)) >= 2 and int(current_inventory.get("umbaran_crystal_timber", 0)) >= 1:
		return "amphistaff-noxious-glaive"
	if int(current_inventory.get("promethium_stygium_bar", 0)) >= 2 and int(current_inventory.get("massassi_kalgerion_horn", 0)) >= 2:
		return "vestments-massassi-warplate"
	if int(current_inventory.get("promethium_stygium_bar", 0)) >= 1 and int(current_inventory.get("vonduun_amphistaff_chitin", 0)) >= 2:
		return "cryptbloom-vonduun-greathelm"
	if int(current_inventory.get("promethium_stygium_bar", 0)) >= 2 and int(current_inventory.get("massassi_kalgerion_horn", 0)) >= 1 and int(current_inventory.get("umbaran_crystal_timber", 0)) >= 1:
		return "kalgerion-massassi-warmaul"
	if int(current_inventory.get("promethium_stygium_ore", 0)) >= 2 and int(current_inventory.get("primordial_animica_bar", 0)) >= 1:
		return "promethium-stygium-bar"
	if int(current_inventory.get("primordial_animica_bar", 0)) >= 2 and int(current_inventory.get("mythosaur_leviathan_crest", 0)) >= 1 and int(current_inventory.get("basilisk_beskar_core", 0)) >= 1:
		return "justiciar-mandalore-hauberk"
	if int(current_inventory.get("tzkal_infernal_heart", 0)) >= 1 and int(current_inventory.get("mythosaur_leviathan_crest", 0)) >= 1 and int(current_inventory.get("primordial_animica_bar", 0)) >= 1:
		return "zuk-infernal-beskar-crown"
	if int(current_inventory.get("primordial_animica_bar", 0)) >= 2 and int(current_inventory.get("tzkal_infernal_heart", 0)) >= 1 and int(current_inventory.get("starforge_omnium_core", 0)) >= 1:
		return "zaros-starforge-godsword"
	if int(current_inventory.get("primordial_animica_bar", 0)) >= 1 and int(current_inventory.get("mythosaur_sanguine_horn", 0)) >= 2 and int(current_inventory.get("elder_bloodwood_timber", 0)) >= 1:
		return "vitur-mythosaur-scythe"
	if int(current_inventory.get("primordial_animica_bar", 0)) >= 2 and int(current_inventory.get("basilisk_beskar_core", 0)) >= 2:
		return "masterwork-animica-warplate"
	if int(current_inventory.get("primordial_animica_bar", 0)) >= 2 and int(current_inventory.get("mythosaur_sanguine_horn", 0)) >= 1 and int(current_inventory.get("elder_bloodwood_timber", 0)) >= 1:
		return "elder-maul-of-vitur"
	if int(current_inventory.get("primordial_animica_ore", 0)) >= 2 and int(current_inventory.get("orikalkum_beskar_bar", 0)) >= 1:
		return "primordial-animica-bar"
	if int(current_inventory.get("orikalkum_beskar_bar", 0)) >= 2 and int(current_inventory.get("gree_aether_matrix", 0)) >= 1 and int(current_inventory.get("terentatek_blood_spine", 0)) >= 1:
		return "korasi-darksaber-edge"
	if int(current_inventory.get("orikalkum_beskar_bar", 0)) >= 1 and int(current_inventory.get("terentatek_blood_spine", 0)) >= 2 and int(current_inventory.get("felucian_magic_log", 0)) >= 1:
		return "guthan-terentatek-warspear"
	if int(current_inventory.get("orikalkum_beskar_bar", 0)) >= 2 and int(current_inventory.get("terentatek_blood_spine", 0)) >= 2:
		return "malevolent-beskar-cuirass"
	if int(current_inventory.get("orikalkum_beskar_bar", 0)) >= 1 and int(current_inventory.get("felucian_rancor_hide", 0)) >= 2:
		return "barrows-cortosis-greathelm"
	if int(current_inventory.get("orikalkum_beskar_bar", 0)) >= 2 and int(current_inventory.get("terentatek_blood_spine", 0)) >= 1 and int(current_inventory.get("felucian_magic_log", 0)) >= 1:
		return "dharok-rancor-greataxe"
	if int(current_inventory.get("orikalkum_beskar_ore", 0)) >= 2 and int(current_inventory.get("runic_banite_ore", 0)) >= 1:
		return "orikalkum-beskar-bar"
	if int(current_inventory.get("ancient_krayt_dragon_scale", 0)) >= 2 and int(current_inventory.get("ichor_spirit_sigil", 0)) >= 1 and int(current_inventory.get("runic_banite_bar", 0)) >= 1:
		return "elysian-spirit-cuirass"
	if int(current_inventory.get("sovereign_jad_fire_core", 0)) >= 1 and int(current_inventory.get("ancient_krayt_dragon_scale", 0)) >= 1 and int(current_inventory.get("runic_banite_bar", 0)) >= 1:
		return "infernal-jad-greathelm"
	if int(current_inventory.get("runic_banite_bar", 0)) >= 2 and int(current_inventory.get("krayt_wyrm_pearl", 0)) >= 1 and int(current_inventory.get("hypergate_relic_core", 0)) >= 1:
		return "krayt-fang-whip-blade"
	if int(current_inventory.get("runic_banite_bar", 0)) >= 1 and int(current_inventory.get("noxious_reaper_limb", 0)) >= 2 and int(current_inventory.get("elder_wroshyr_timber", 0)) >= 1:
		return "noxious-acklay-halberd"
	if int(current_inventory.get("runic_banite_bar", 0)) >= 2 and int(current_inventory.get("bandosian_gorax_tasset", 0)) >= 2:
		return "torva-gorax-warplate"
	if int(current_inventory.get("runic_banite_bar", 0)) >= 2 and int(current_inventory.get("bandosian_gorax_tasset", 0)) >= 1 and int(current_inventory.get("elder_wroshyr_timber", 0)) >= 1:
		return "gorax-titan-maul"
	if int(current_inventory.get("runic_banite_ore", 0)) >= 2 and int(current_inventory.get("nova_prism", 0)) >= 1:
		return "runic-banite-bar"
	if int(current_inventory.get("chronite_bar", 0)) >= 2 and int(current_inventory.get("ancient_heartwood_sap", 0)) >= 1 and int(current_inventory.get("voidbark_heartwood", 0)) >= 1:
		return "sovereign-greatsword"
	if int(current_inventory.get("abyssal_bar", 0)) >= 1 and int(current_inventory.get("abyssal_behemoth_plate", 0)) >= 2 and int(current_inventory.get("chronite_bar", 0)) >= 1:
		return "dreadnought-aegis-helm"
	if int(current_inventory.get("starsteel_bar", 0)) >= 2 and int(current_inventory.get("solaris_fire_gland", 0)) >= 1 and int(current_inventory.get("aether_bough_timber", 0)) >= 1:
		return "solaris-halberd"
	if int(current_inventory.get("neutronium_bar", 0)) >= 2 and int(current_inventory.get("titan_actuator_coil", 0)) >= 1 and int(current_inventory.get("glacial_core_shard", 0)) >= 1:
		return "juggernaut-war-plate"
	if int(current_inventory.get("neutronium_bar", 0)) >= 2 and int(current_inventory.get("titan_actuator_coil", 0)) >= 1 and int(current_inventory.get("aether_bough_timber", 0)) >= 1:
		return "overlord-siege-maul"
	if int(current_inventory.get("iron_bar", 0)) >= 2 and int(current_inventory.get("boar_tusk", 0)) >= 2 and int(current_inventory.get("ironwood_log", 0)) >= 1:
		return "colossus-breaker-greataxe"
	if int(current_inventory.get("iron_bar", 0)) >= 2 and int(current_inventory.get("ironwood_log", 0)) >= 1:
		return "vanguard-cleaver"
	if int(current_inventory.get("titanium_bar", 0)) >= 2 and int(current_inventory.get("sentinel_core", 0)) >= 1 and int(current_inventory.get("salvaged_power_cell", 0)) >= 1:
		return "titanium-carbine"
	if int(current_inventory.get("stalker_chitin", 0)) >= 3 and int(current_inventory.get("iron_bar", 0)) >= 2:
		return "chitin-cuirass"
	if int(current_inventory.get("iron_bar", 0)) >= 3 and int(current_inventory.get("hound_fang", 0)) >= 2 and int(current_inventory.get("wooden_plank", 0)) >= 1:
		return "iron-greatsword"
	if int(current_inventory.get("titanium_ore", 0)) >= 2 and int(current_inventory.get("iron_bar", 0)) >= 1:
		return "titanium-bar"
	if int(current_inventory.get("iron_ore", 0)) >= 2:
		return "iron-bar"
	if int(current_inventory.get("copper_bar", 0)) >= 2 and int(current_inventory.get("boar_tusk", 0)) >= 2 and int(current_inventory.get("wooden_plank", 0)) >= 1:
		return "boar-tusk-spear"
	if int(current_inventory.get("salvaged_power_cell", 0)) >= 1 and int(current_inventory.get("copper_bar", 0)) >= 2 and int(current_inventory.get("wooden_plank", 0)) >= 1:
		return "frontier-rifle"
	if int(current_inventory.get("copper_bar", 0)) >= 2 and int(current_inventory.get("rat_tooth", 0)) >= 2 and int(current_inventory.get("wooden_plank", 0)) >= 1:
		return "rat-tooth-blade"
	if int(current_inventory.get("copper_bar", 0)) >= 3 and int(current_inventory.get("wooden_plank", 0)) >= 1:
		return "copper-sword"
	return "copper-bar"


func _select_workbench_recipe() -> String:
	if int(current_inventory.get("solak_anima_heartwood", 0)) >= 2 and int(current_inventory.get("infinity_gate_matrix", 0)) >= 1 and int(current_inventory.get("zamorak_nihilus_chaos_core", 0)) >= 1:
		return "zaryte-solak-singularity-rifle"
	if int(current_inventory.get("zamorak_nihilus_chaos_core", 0)) >= 1 and int(current_inventory.get("infinity_gate_matrix", 0)) >= 1 and int(current_inventory.get("solak_anima_heartwood", 0)) >= 2:
		return "omnipower-chaos-holostaff"
	if int(current_inventory.get("primal_kyber_bar", 0)) >= 1 and int(current_inventory.get("zillo_leviathan_sovereign_horn", 0)) >= 1 and int(current_inventory.get("infinity_gate_matrix", 0)) >= 1:
		return "vorago-seismic-kyber-talons"
	if int(current_inventory.get("solak_anima_heartwood", 0)) >= 2 and int(current_inventory.get("infinity_gate_matrix", 0)) >= 1 and int(current_inventory.get("zillo_vorago_tectonic_plate", 0)) >= 1:
		return "bolg-infinity-bowcaster"
	if int(current_inventory.get("infinity_gate_matrix", 0)) >= 1 and int(current_inventory.get("hk_nex_assassin_core", 0)) >= 1 and int(current_inventory.get("primal_kyber_bar", 0)) >= 1:
		return "praetorian-infinity-blasters"
	if int(current_inventory.get("bloodweed_vong_spore", 0)) >= 2 and int(current_inventory.get("celestial_mortis_melon", 0)) >= 2 and int(current_inventory.get("zamorak_nihilus_chaos_core", 0)) >= 1:
		return "supreme-elder-mortis-ambrosia"
	if int(current_inventory.get("infinity_gate_matrix", 0)) >= 1 and int(current_inventory.get("bloodweed_vong_spore", 0)) >= 2 and int(current_inventory.get("bacta_gel", 0)) >= 1:
		return "infinity-kyber-med-core"
	if int(current_inventory.get("umbaran_crystal_timber", 0)) >= 2 and int(current_inventory.get("massassi_chronote_relic", 0)) >= 1 and int(current_inventory.get("promethium_stygium_bar", 0)) >= 1:
		return "decimation-umbaran-sniper"
	if int(current_inventory.get("umbaran_crystal_timber", 0)) >= 2 and int(current_inventory.get("massassi_chronote_relic", 0)) >= 1 and int(current_inventory.get("arbuck_zevrona_herb", 0)) >= 1:
		return "virtus-chronote-scepter"
	if int(current_inventory.get("vonduun_amphistaff_chitin", 0)) >= 2 and int(current_inventory.get("promethium_stygium_bar", 0)) >= 1 and int(current_inventory.get("massassi_chronote_relic", 0)) >= 1:
		return "cinderbane-teras-gauntlets"
	if int(current_inventory.get("umbaran_crystal_timber", 0)) >= 2 and int(current_inventory.get("massassi_chronote_relic", 0)) >= 1 and int(current_inventory.get("vonduun_amphistaff_chitin", 0)) >= 1:
		return "pernix-umbaran-bowcaster"
	if int(current_inventory.get("massassi_chronote_relic", 0)) >= 2 and int(current_inventory.get("hk_nex_assassin_core", 0)) >= 1 and int(current_inventory.get("promethium_stygium_bar", 0)) >= 1:
		return "shadow-hk-dual-disruptors"
	if int(current_inventory.get("obsidian_lava_lychee", 0)) >= 2 and int(current_inventory.get("arbuck_zevrona_herb", 0)) >= 1 and int(current_inventory.get("bacta_gel", 0)) >= 1:
		return "vulcan-lava-lychee-nectar"
	if int(current_inventory.get("arbuck_zevrona_herb", 0)) >= 2 and int(current_inventory.get("bacta_gel", 0)) >= 1 and int(current_inventory.get("massassi_chronote_relic", 0)) >= 1:
		return "arbuck-stygium-trauma-stim"
	if int(current_inventory.get("elder_bloodwood_timber", 0)) >= 2 and int(current_inventory.get("starforge_omnium_core", 0)) >= 1 and int(current_inventory.get("tzkal_infernal_heart", 0)) >= 1:
		return "bofa-bloodwood-railgun"
	if int(current_inventory.get("tzkal_infernal_heart", 0)) >= 1 and int(current_inventory.get("starforge_omnium_core", 0)) >= 1 and int(current_inventory.get("elder_bloodwood_timber", 0)) >= 2:
		return "fractured-armadyl-holostaff"
	if int(current_inventory.get("primordial_animica_bar", 0)) >= 1 and int(current_inventory.get("mythosaur_leviathan_crest", 0)) >= 1 and int(current_inventory.get("starforge_omnium_core", 0)) >= 1:
		return "telos-anima-talons"
	if int(current_inventory.get("elder_bloodwood_timber", 0)) >= 2 and int(current_inventory.get("starforge_omnium_core", 0)) >= 1 and int(current_inventory.get("mythosaur_sanguine_horn", 0)) >= 1:
		return "eldritch-starforge-bowcaster"
	if int(current_inventory.get("starforge_omnium_core", 0)) >= 1 and int(current_inventory.get("basilisk_beskar_core", 0)) >= 1 and int(current_inventory.get("primordial_animica_bar", 0)) >= 1:
		return "leng-starforge-blasters"
	if int(current_inventory.get("fellstalk_ryll_blossom", 0)) >= 2 and int(current_inventory.get("golden_dragon_cactus", 0)) >= 2 and int(current_inventory.get("tzkal_infernal_heart", 0)) >= 1:
		return "elder-overload-dragon-salve"
	if int(current_inventory.get("starforge_omnium_core", 0)) >= 1 and int(current_inventory.get("fellstalk_ryll_blossom", 0)) >= 2 and int(current_inventory.get("bacta_gel", 0)) >= 1:
		return "starforge-nanite-rejuvenator"
	if int(current_inventory.get("felucian_magic_log", 0)) >= 2 and int(current_inventory.get("gree_aether_matrix", 0)) >= 1 and int(current_inventory.get("orikalkum_beskar_bar", 0)) >= 1:
		return "crystal-seren-longrifle"
	if int(current_inventory.get("felucian_magic_log", 0)) >= 2 and int(current_inventory.get("gree_aether_matrix", 0)) >= 1 and int(current_inventory.get("snapdragon_kolto_frond", 0)) >= 1:
		return "ahrim-holocron-staff"
	if int(current_inventory.get("felucian_rancor_hide", 0)) >= 2 and int(current_inventory.get("orikalkum_beskar_bar", 0)) >= 1 and int(current_inventory.get("gree_aether_matrix", 0)) >= 1:
		return "goliath-shock-cestus"
	if int(current_inventory.get("felucian_magic_log", 0)) >= 2 and int(current_inventory.get("gree_aether_matrix", 0)) >= 1 and int(current_inventory.get("felucian_rancor_hide", 0)) >= 1:
		return "karil-quarrel-repeater"
	if int(current_inventory.get("gree_aether_matrix", 0)) >= 2 and int(current_inventory.get("orikalkum_beskar_bar", 0)) >= 1:
		return "ascension-bryar-carbine"
	if int(current_inventory.get("cursed_blight_plum", 0)) >= 2 and int(current_inventory.get("snapdragon_kolto_frond", 0)) >= 1 and int(current_inventory.get("bacta_gel", 0)) >= 1:
		return "guthix-blight-plum-restoration"
	if int(current_inventory.get("snapdragon_kolto_frond", 0)) >= 2 and int(current_inventory.get("bacta_gel", 0)) >= 1 and int(current_inventory.get("gree_aether_matrix", 0)) >= 1:
		return "kolto-barrows-trauma-injector"
	if int(current_inventory.get("elder_wroshyr_timber", 0)) >= 2 and int(current_inventory.get("sovereign_jad_fire_core", 0)) >= 1 and int(current_inventory.get("hypergate_relic_core", 0)) >= 1:
		return "twisted-wroshyr-longbow"
	if int(current_inventory.get("ichor_spirit_sigil", 0)) >= 1 and int(current_inventory.get("sovereign_jad_fire_core", 0)) >= 1 and int(current_inventory.get("elder_wroshyr_timber", 0)) >= 2:
		return "tumeken-ichor-scepter"
	if int(current_inventory.get("runic_banite_bar", 0)) >= 1 and int(current_inventory.get("ancient_krayt_dragon_scale", 0)) >= 2 and int(current_inventory.get("hypergate_relic_core", 0)) >= 1:
		return "beskar-dragon-claws"
	if int(current_inventory.get("elder_wroshyr_timber", 0)) >= 2 and int(current_inventory.get("hypergate_relic_core", 0)) >= 1 and int(current_inventory.get("noxious_reaper_limb", 0)) >= 1:
		return "zaryte-bowcaster-cannon"
	if int(current_inventory.get("hypergate_relic_core", 0)) >= 1 and int(current_inventory.get("krayt_wyrm_pearl", 0)) >= 1 and int(current_inventory.get("runic_banite_bar", 0)) >= 1:
		return "westar-armadyl-blaster"
	if int(current_inventory.get("sovereign_torstol_leaf", 0)) >= 2 and int(current_inventory.get("spirit_muja_fruit", 0)) >= 1 and int(current_inventory.get("sovereign_jad_fire_core", 0)) >= 1:
		return "overload-bacta-flask"
	if int(current_inventory.get("spirit_muja_fruit", 0)) >= 2 and int(current_inventory.get("sovereign_torstol_leaf", 0)) >= 1 and int(current_inventory.get("bacta_gel", 0)) >= 1:
		return "saradomin-muja-banquet"
	if int(current_inventory.get("bacta_gel", 0)) >= 2 and int(current_inventory.get("hypergate_relic_core", 0)) >= 1 and int(current_inventory.get("sovereign_torstol_leaf", 0)) >= 1:
		return "rakatan-stim-infuser"
	if int(current_inventory.get("voidbark_heartwood", 0)) >= 2 and int(current_inventory.get("singularity_core_cell", 0)) >= 1 and int(current_inventory.get("chronite_bar", 0)) >= 1:
		return "elderwood-longbow"
	if int(current_inventory.get("aether_bough_timber", 0)) >= 1 and int(current_inventory.get("astral_mycelium", 0)) >= 2 and int(current_inventory.get("solaris_fire_gland", 0)) >= 1:
		return "sunfire-focus-scepter"
	if int(current_inventory.get("chronite_bar", 0)) >= 1 and int(current_inventory.get("temporal_phase_tendril", 0)) >= 2 and int(current_inventory.get("singularity_core_cell", 0)) >= 1:
		return "chronite-kinetic-gauntlets"
	if int(current_inventory.get("aether_bough_timber", 0)) >= 2 and int(current_inventory.get("singularity_core_cell", 0)) >= 1 and int(current_inventory.get("neutronium_bar", 0)) >= 1:
		return "abyssal-artillery-cannon"
	if int(current_inventory.get("electrum_bar", 0)) >= 2 and int(current_inventory.get("glacial_core_shard", 0)) >= 1 and int(current_inventory.get("cryo_lotus_petal", 0)) >= 1:
		return "cryo-pulse-blaster"
	if int(current_inventory.get("celestial_nectar_fruit", 0)) >= 2 and int(current_inventory.get("astral_mycelium", 0)) >= 1 and int(current_inventory.get("ancient_heartwood_sap", 0)) >= 1:
		return "astral-ambrosia-elixir"
	if int(current_inventory.get("auric_grain_sheaf", 0)) >= 2 and int(current_inventory.get("cryo_lotus_petal", 0)) >= 1 and int(current_inventory.get("bacta_gel", 0)) >= 1:
		return "auric-harvest-feast"
	if int(current_inventory.get("ironwood_log", 0)) >= 2 and int(current_inventory.get("circuit_board", 0)) >= 1 and int(current_inventory.get("copper_bar", 0)) >= 1:
		return "skirmisher-repeater-bow"
	if int(current_inventory.get("ironwood_log", 0)) >= 2 and int(current_inventory.get("aether_crystal", 0)) >= 1:
		return "crystal-channeled-rod"
	if int(current_inventory.get("tuber_root", 0)) >= 2 and int(current_inventory.get("glowcap_fungus", 0)) >= 1:
		return "hearty-tuber-stew"
	if int(current_inventory.get("bacta_gel", 0)) >= 2 and int(current_inventory.get("stalker_venom_sac", 0)) >= 1 and int(current_inventory.get("aether_crystal", 0)) >= 1:
		return "advanced-medpack"
	if int(current_inventory.get("bacta_gel", 0)) >= 1 and int(current_inventory.get("frontier_grain", 0)) >= 2 and int(current_inventory.get("aether_crystal", 0)) >= 1:
		return "vitality-stim-pack"
	if int(current_inventory.get("bacta_gel", 0)) >= 1 and int(current_inventory.get("hound_fang", 0)) >= 1 and int(current_inventory.get("wild_herb", 0)) >= 2:
		return "adrenal-action-booster"
	if int(current_inventory.get("aether_crystal", 0)) >= 1 and int(current_inventory.get("herbal_bundle", 0)) >= 1 and int(current_inventory.get("wild_herb", 0)) >= 1:
		return "neural-mind-tonic"
	if int(current_inventory.get("bacta_gel", 0)) >= 1 and int(current_inventory.get("herbal_bundle", 0)) >= 1:
		return "field-medpack"
	if int(current_inventory.get("wild_herb", 0)) >= 2 and int(current_inventory.get("rough_stone", 0)) >= 1:
		return "bacta-gel"
	if int(current_inventory.get("salvaged_power_cell", 0)) >= 1 and int(current_inventory.get("wooden_plank", 0)) >= 2 and int(current_inventory.get("wild_herb", 0)) >= 1:
		return "resonance-focus"
	if int(current_inventory.get("salvaged_scrap", 0)) >= 2 and int(current_inventory.get("copper_bar", 0)) >= 1 and int(current_inventory.get("wooden_plank", 0)) >= 1:
		return "frontier-pistol"
	if int(current_inventory.get("salvaged_scrap", 0)) >= 2 and int(current_inventory.get("copper_bar", 0)) >= 1:
		return "salvaged-power-cell"
	if int(current_inventory.get("frontier_grain", 0)) >= 2 and int(current_inventory.get("wild_herb", 0)) >= 1:
		return "frontier-ration"
	if int(current_inventory.get("herbal_bundle", 0)) >= 1 and int(current_inventory.get("wild_herb", 0)) >= 1:
		return "herbal-tea"
	if int(current_inventory.get("wild_herb", 0)) >= 3 and int(current_inventory.get("wooden_plank", 0)) >= 1:
		return "herbal-bundle"
	return "basic-plank"


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
		var recipe_id := _select_forge_recipe()
		interaction_type = "craft"
		status_label.text = "Using frontier forge (" + recipe_id.replace("-", " ") + ")..."
		error = gather_request.request(
			api_base_url + "/api/v1/account/characters/" + player_id + "/craft/" + recipe_id,
			_api_headers(),
			HTTPClient.METHOD_POST,
			""
		)
	elif nearest_station_id == "frontier-workbench" and nearest_station_distance <= 3.0 and nearest_station_distance <= nearest_node_distance and nearest_station_distance <= enemy_distance:
		var recipe_id := _select_workbench_recipe()
		interaction_type = "craft"
		status_label.text = "Using frontier workbench (" + recipe_id.replace("-", " ") + ")..."
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
	else:
		_report_interaction_request_error(error)


func _report_interaction_request_error(request_error: int) -> void:
	if status_label != null:
		status_label.text = "Interaction: request could not start (" + error_string(request_error) + ")"


func _on_gather_completed(result: int, response_code: int, _headers: PackedStringArray, body: PackedByteArray) -> void:
	interaction_pending = false
	if result != HTTPRequest.RESULT_SUCCESS:
		status_label.text = "Interaction: server connection lost"
		return
	if response_code == 401:
		_expire_session()
		return
	var response_json := JSON.new()
	var parse_error := response_json.parse(body.get_string_from_utf8())
	if parse_error != OK or not response_json.data is Dictionary:
		status_label.text = "Interaction: invalid server response"
		return
	var payload: Dictionary = response_json.data
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
		if interaction_type == "craft":
			if crafting_status_label != null:
				crafting_status_label.text = message
		if interaction_type == "combat":
			if error_code == "combat_cooldown":
				status_label.text = "Combat: wait for your next attack"
			elif error_code == "enemy_respawning":
				status_label.text = "Combat: this creature is recovering"
			elif error_code == "player_defeated":
				status_label.text = "Defeated • use Rest to respawn"
		return
	var player_data: Dictionary = payload.get("character", payload.get("player", {}))
	_sync_vital_attributes_from_character(player_data)
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
			if player_node != null and is_instance_valid(player_node):
				player_node.call("play_defeat_animation")
			combat_message = "You were defeated • use Rest to respawn"
			if movement_timer != null:
				movement_timer.stop()
		status_label.text = "Combat: " + combat_message
		_refresh_skills_panel()
		return
	if interaction_type == "use_item":
		if player_node != null and is_instance_valid(player_node):
			player_node.call("play_use_item_animation")
		var item_id := str(payload.get("itemId", "minor_healing_potion"))
		_update_inventory(player_data.get("inventory", {}), player_data.get("bank", {}), player_data.get("coins", current_coins), player_data.get("equipment", {}), payload.get("health", player_data.get("health", current_health)))
		status_label.text = "Used " + item_id.replace("_", " ") + " • +" + str(payload.get("healed", 0)) + " health"
		return
	if interaction_type == "rest":
		if player_node != null and is_instance_valid(player_node):
			player_node.call("play_respawn_animation")
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
		if player_node != null and is_instance_valid(player_node):
			player_node.call("play_craft_animation")
		var crafting: Dictionary = payload.get("crafting", {})
		var output_name := "crafted item"
		var outputs: Array = crafting.get("outputs", [])
		if not outputs.is_empty():
			var output_id := str(outputs[0].get("itemId", ""))
			for item in market_catalog:
				if item is Dictionary and str(item.get("id", "")) == output_id:
					output_name = str(item.get("name", output_id))
		var craft_msg := "Crafted " + output_name + " • +" + str(crafting.get("experience", 0)) + " " + str(crafting.get("skillId", "crafting")) + " XP"
		status_label.text = craft_msg
		if crafting_status_label != null:
			crafting_status_label.text = craft_msg
	else:
		if player_node != null and is_instance_valid(player_node):
			player_node.call("play_gather_animation")
		var gathering: Dictionary = payload.get("gathering", {})
		var item_name := str(gathering.get("itemId", "resource")).replace("_", " ")
		for item in market_catalog:
			if item is Dictionary and str(item.get("id", "")) == str(gathering.get("itemId", "")):
				item_name = str(item.get("name", item_name))
		status_label.text = "Gathered " + item_name + " • +" + str(gathering.get("experience", 0)) + " " + str(gathering.get("skillId", "")) + " XP"
	_update_inventory(player_data.get("inventory", {}), player_data.get("bank", {}), player_data.get("coins", current_coins), player_data.get("equipment", {}), player_data.get("health", current_health))
	if crafting_panel != null and crafting_panel.visible:
		_refresh_crafting_recipe_list()
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


func _sync_vital_attributes_from_character(character_data: Variant) -> void:
	if not character_data is Dictionary:
		return
	var data: Dictionary = character_data
	if data.has("action"):
		current_action = int(data.get("action", current_action))
	if data.has("mind"):
		current_mind = int(data.get("mind", current_mind))
	if data.has("constitution"):
		current_constitution = max(10, int(data.get("constitution", 10)))
	if data.has("stamina"):
		current_stamina = max(10, int(data.get("stamina", 10)))
	if data.has("willpower"):
		current_willpower = max(10, int(data.get("willpower", 10)))
	var buff_h := int(data.get("maxHealthBuff", 0))
	var buff_a := int(data.get("maxActionBuff", 0))
	var buff_m := int(data.get("maxMindBuff", 0))
	current_max_health = game_max_health + buff_h + max(0, current_constitution - 10) * 2
	current_max_action = 100 + buff_a + max(0, current_stamina - 10) * 2
	current_max_mind = 100 + buff_m + max(0, current_willpower - 10) * 2


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
	if hud_health_bar != null:
		hud_health_bar.max_value = max(1, current_max_health)
		hud_health_bar.value = clamp(current_health, 0, current_max_health)
	if hud_action_bar != null:
		hud_action_bar.max_value = max(1, current_max_action)
		hud_action_bar.value = clamp(current_action, 0, current_max_action)
	if hud_mind_bar != null:
		hud_mind_bar.max_value = max(1, current_max_mind)
		hud_mind_bar.value = clamp(current_mind, 0, current_max_mind)
	if stats_panel != null and stats_panel.visible:
		_refresh_stats_panel()
	var hp_regen := 1 + int(current_constitution / 5)
	inventory_label.text = "HP %d/%d (+%d/5s) • ACT %d/%d • MND %d/%d • CON %d STA %d WIL %d • Coins %d • ATK %d DEF %d" % [
		current_health, current_max_health, hp_regen,
		current_action, current_max_action,
		current_mind, current_max_mind,
		current_constitution, current_stamina, current_willpower,
		current_coins, current_attack_bonus, current_defense_bonus
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
	if event is InputEventKey and event.pressed and not event.echo:
		if event.keycode == KEY_ESCAPE:
			Input.mouse_mode = Input.MOUSE_MODE_VISIBLE
			return
		if player_id.is_empty():
			return
		var focus_owner := get_viewport().gui_get_focus_owner() if get_viewport() != null else null
		if focus_owner is LineEdit:
			return
		match event.physical_keycode:
			KEY_I:
				_show_inventory_panel()
			KEY_K:
				_toggle_skills_panel()
			KEY_C:
				_show_crafting_panel()
			KEY_V:
				_show_stats_panel()
			KEY_G:
				_show_guide_panel()
			KEY_H:
				_use_best_consumable()
			KEY_O:
				_show_settings_panel()
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
	_sync_vital_attributes_from_character(character_data)
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
	if crafting_panel != null:
		crafting_panel.visible = false
	logout_button.visible = true
	crafting_button.visible = true
	stats_button.visible = true
	guide_button.visible = true
	quick_heal_button.visible = true
	settings_button.visible = true
	if settings_panel != null:
		settings_panel.visible = false
	if stats_panel != null:
		stats_panel.visible = false
	if guide_panel != null:
		guide_panel.visible = false
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
	if crafting_panel != null:
		crafting_panel.visible = false
	crafting_button.visible = false
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
	if crafting_panel != null:
		crafting_panel.visible = false
	if crafting_button != null:
		crafting_button.visible = false
	if stats_button != null:
		stats_button.visible = false
	if guide_button != null:
		guide_button.visible = false
	if quick_heal_button != null:
		quick_heal_button.visible = false
	if settings_button != null:
		settings_button.visible = false
	if settings_panel != null:
		settings_panel.visible = false
	if stats_panel != null:
		stats_panel.visible = false
	if guide_panel != null:
		guide_panel.visible = false
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
	if crafting_panel != null:
		crafting_panel.visible = false
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
	if crafting_panel != null:
		crafting_panel.visible = false
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
			if crafting_panel != null:
				crafting_panel.visible = false
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
	if crafting_panel != null:
		crafting_panel.visible = false
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


func _use_best_consumable() -> void:
	if player_id.is_empty() or auth_token.is_empty() or interaction_pending or gather_request == null:
		return
	var priority_items := [
		"phoenix_resuscitation_pack",
		"nanite_repair_injector",
		"surgical_trauma_kit",
		"advanced_medpack",
		"field_medpack",
		"minor_healing_potion",
		"frontier_ration",
		"herbal_tea",
		"celestial_harmony_draught",
		"overdrive_omni_stim",
		"titan_constitution_elixir",
		"vitality_stim_pack",
		"adrenal_action_booster",
		"neural_mind_tonic"
	]
	var chosen := ""
	for item_id in priority_items:
		if int(current_inventory.get(item_id, 0)) > 0:
			chosen = item_id
			break
	if chosen.is_empty():
		status_label.text = "Med/Stim: no medical packs or enhancers in bag"
		return
	interaction_type = "use_item"
	var error := gather_request.request(
		api_base_url + "/api/v1/account/characters/" + player_id + "/items/use/" + chosen,
		_api_headers(),
		HTTPClient.METHOD_POST,
		""
	)
	if error == OK:
		interaction_pending = true
		status_label.text = "Using " + chosen.replace("_", " ") + "..."
	else:
		_report_interaction_request_error(error)


func _show_settings_panel() -> void:
	if skills_panel != null:
		skills_panel.visible = false
	if inventory_panel != null:
		inventory_panel.visible = false
	if market_panel != null:
		market_panel.visible = false
	if equipment_panel != null:
		equipment_panel.visible = false
	if rewards_panel != null:
		rewards_panel.visible = false
	if crafting_panel != null:
		crafting_panel.visible = false
	if stats_panel != null:
		stats_panel.visible = false
	if guide_panel != null:
		guide_panel.visible = false
	if settings_panel != null:
		settings_panel.visible = not settings_panel.visible
		return
	settings_panel = PanelContainer.new()
	settings_panel.position = Vector2(390, 125)
	settings_panel.custom_minimum_size = Vector2(400, 320)
	var column := VBoxContainer.new()
	settings_panel.add_child(column)
	var heading := Label.new()
	heading.text = "OPTIONS, CAMERA & CONTROLS"
	heading.add_theme_font_size_override("font_size", 18)
	column.add_child(heading)
	var sens_row := HBoxContainer.new()
	column.add_child(sens_row)
	var sens_down := Button.new()
	sens_down.text = "Look Sens -"
	sens_down.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	sens_down.pressed.connect(func() -> void:
		if player_node != null and is_instance_valid(player_node):
			var cur := float(player_node.get("mouse_sensitivity"))
			player_node.set("mouse_sensitivity", maxf(0.0008, cur - 0.0005))
			settings_status_label.text = "Look sensitivity: %.4f" % float(player_node.get("mouse_sensitivity"))
	)
	sens_row.add_child(sens_down)
	var sens_up := Button.new()
	sens_up.text = "Look Sens +"
	sens_up.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	sens_up.pressed.connect(func() -> void:
		if player_node != null and is_instance_valid(player_node):
			var cur := float(player_node.get("mouse_sensitivity"))
			player_node.set("mouse_sensitivity", minf(0.0080, cur + 0.0005))
			settings_status_label.text = "Look sensitivity: %.4f" % float(player_node.get("mouse_sensitivity"))
	)
	sens_row.add_child(sens_up)
	var zoom_row := HBoxContainer.new()
	column.add_child(zoom_row)
	var first_person_btn := Button.new()
	first_person_btn.text = "First-Person View"
	first_person_btn.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	first_person_btn.pressed.connect(func() -> void:
		if player_node != null and is_instance_valid(player_node):
			player_node.call("_adjust_camera_zoom", -20.0)
			settings_status_label.text = "Camera: First-Person View"
	)
	zoom_row.add_child(first_person_btn)
	var third_person_btn := Button.new()
	third_person_btn.text = "Third-Person View"
	third_person_btn.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	third_person_btn.pressed.connect(func() -> void:
		if player_node != null and is_instance_valid(player_node):
			player_node.call("_adjust_camera_zoom", 5.0)
			player_node.call("reset_camera_behind_character")
			settings_status_label.text = "Camera: Reset Third-Person View"
	)
	zoom_row.add_child(third_person_btn)
	var hotkeys_label := Label.new()
	hotkeys_label.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	hotkeys_label.text = (
		"Controls & Hotkeys:\n"
		+ "  • WASD / Touch Pad: Move  •  Space / JUMP: Jump\n"
		+ "  • E / USE: Gather, Craft, Attack, Bank, Trader\n"
		+ "  • R: Toggle Autorun  •  CapsLock: Walk/Run Pace\n"
		+ "  • Home: Reset Camera Behind Character\n"
		+ "  • I: Bag/Bank  •  K: Skills  •  C: Crafting\n"
		+ "  • V: Vitals/Stats  •  G: World Atlas  •  H: Quick Med/Stim"
	)
	column.add_child(hotkeys_label)
	settings_status_label = Label.new()
	settings_status_label.text = "Adjust camera, view mode, or review keybindings."
	column.add_child(settings_status_label)
	var close_button := Button.new()
	close_button.text = "Close"
	close_button.pressed.connect(func() -> void: settings_panel.visible = false)
	column.add_child(close_button)
	if auth_panel != null and auth_panel.get_parent() != null:
		auth_panel.get_parent().add_child(settings_panel)
	else:
		add_child(settings_panel)


func _show_guide_panel() -> void:
	if skills_panel != null:
		skills_panel.visible = false
	if inventory_panel != null:
		inventory_panel.visible = false
	if market_panel != null:
		market_panel.visible = false
	if equipment_panel != null:
		equipment_panel.visible = false
	if rewards_panel != null:
		rewards_panel.visible = false
	if crafting_panel != null:
		crafting_panel.visible = false
	if stats_panel != null:
		stats_panel.visible = false
	if guide_panel != null:
		guide_panel.visible = not guide_panel.visible
		if guide_panel.visible:
			_refresh_guide_panel()
		return
	guide_panel = PanelContainer.new()
	guide_panel.position = Vector2(390, 120)
	guide_panel.custom_minimum_size = Vector2(420, 350)
	var column := VBoxContainer.new()
	guide_panel.add_child(column)
	var heading := Label.new()
	heading.text = "FRONTIER WORLD ATLAS & DIRECTORY"
	heading.add_theme_font_size_override("font_size", 18)
	column.add_child(heading)
	guide_filter_select = OptionButton.new()
	guide_filter_select.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	guide_filter_select.add_item("All World Locations")
	guide_filter_select.set_item_metadata(0, "all")
	guide_filter_select.add_item("Gathering & Mining Nodes")
	guide_filter_select.set_item_metadata(1, "nodes")
	guide_filter_select.add_item("Crafting Stations")
	guide_filter_select.set_item_metadata(2, "stations")
	guide_filter_select.add_item("Monsters & Enemies")
	guide_filter_select.set_item_metadata(3, "enemies")
	guide_filter_select.item_selected.connect(func(idx: int) -> void:
		guide_selected_filter = str(guide_filter_select.get_item_metadata(idx))
		_refresh_guide_panel()
	)
	column.add_child(guide_filter_select)
	var scroll := ScrollContainer.new()
	scroll.custom_minimum_size = Vector2(390, 230)
	scroll.size_flags_vertical = Control.SIZE_EXPAND_FILL
	column.add_child(scroll)
	guide_entries_column = VBoxContainer.new()
	guide_entries_column.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	scroll.add_child(guide_entries_column)
	var close_button := Button.new()
	close_button.text = "Close"
	close_button.pressed.connect(func() -> void: guide_panel.visible = false)
	column.add_child(close_button)
	if auth_panel != null and auth_panel.get_parent() != null:
		auth_panel.get_parent().add_child(guide_panel)
	else:
		add_child(guide_panel)
	_refresh_guide_panel()


func _refresh_guide_panel() -> void:
	if guide_entries_column == null:
		return
	for child in guide_entries_column.get_children():
		child.queue_free()
	var player_pos := player_node.global_position if player_node != null and is_instance_valid(player_node) else Vector3.ZERO
	if guide_selected_filter == "all" or guide_selected_filter == "stations":
		for station in station_catalog:
			if station is Dictionary:
				var pos: Dictionary = station.get("position", {})
				var sx := float(pos.get("x", 0.0))
				var sz := float(pos.get("z", 0.0))
				var dist := Vector2(player_pos.x - sx, player_pos.z - sz).length()
				var lbl := Label.new()
				lbl.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
				lbl.text = "[Station] %s • (%.0f, %.0f) • %.1fm away" % [str(station.get("name", "Station")), sx, sz, dist]
				guide_entries_column.add_child(lbl)
	if guide_selected_filter == "all" or guide_selected_filter == "nodes":
		for node in gathering_catalog:
			if node is Dictionary:
				var pos: Dictionary = node.get("position", {})
				var nx := float(pos.get("x", 0.0))
				var nz := float(pos.get("z", 0.0))
				var dist := Vector2(player_pos.x - nx, player_pos.z - nz).length()
				var lbl := Label.new()
				lbl.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
				lbl.text = "[%s Lv %d] %s • (%.0f, %.0f) • %.1fm" % [str(node.get("skillId", "gather")).capitalize(), int(node.get("requiredLevel", 1)), str(node.get("name", "Node")), nx, nz, dist]
				guide_entries_column.add_child(lbl)
	if guide_selected_filter == "all" or guide_selected_filter == "enemies":
		for enemy in enemy_catalog:
			if enemy is Dictionary:
				var pos: Dictionary = enemy.get("position", {})
				var ex := float(pos.get("x", 0.0))
				var ez := float(pos.get("z", 0.0))
				var dist := Vector2(player_pos.x - ex, player_pos.z - ez).length()
				var lbl := Label.new()
				lbl.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
				lbl.text = "[Enemy HP %d • ATK %d] %s • (%.0f, %.0f) • %.1fm" % [int(enemy.get("maxHealth", 10)), int(enemy.get("attack", 1)), str(enemy.get("name", "Creature")), ex, ez, dist]
				guide_entries_column.add_child(lbl)


func _show_stats_panel() -> void:
	if skills_panel != null:
		skills_panel.visible = false
	if inventory_panel != null:
		inventory_panel.visible = false
	if market_panel != null:
		market_panel.visible = false
	if equipment_panel != null:
		equipment_panel.visible = false
	if rewards_panel != null:
		rewards_panel.visible = false
	if crafting_panel != null:
		crafting_panel.visible = false
	if stats_panel != null:
		stats_panel.visible = not stats_panel.visible
		if stats_panel.visible:
			_refresh_stats_panel()
		return
	stats_panel = PanelContainer.new()
	stats_panel.position = Vector2(400, 130)
	stats_panel.custom_minimum_size = Vector2(390, 310)
	var column := VBoxContainer.new()
	stats_panel.add_child(column)
	var heading := Label.new()
	heading.text = "CHARACTER VITALS & ATTRIBUTES"
	heading.add_theme_font_size_override("font_size", 18)
	column.add_child(heading)
	stats_summary_label = Label.new()
	stats_summary_label.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	column.add_child(stats_summary_label)
	var pace_button := Button.new()
	pace_button.text = "Toggle Walk / Run Pace (CapsLock)"
	pace_button.pressed.connect(func() -> void:
		if player_node != null and is_instance_valid(player_node):
			player_node.call("toggle_walk_mode")
			_refresh_stats_panel()
	)
	column.add_child(pace_button)
	var autorun_button := Button.new()
	autorun_button.text = "Toggle Autorun (R)"
	autorun_button.pressed.connect(func() -> void:
		if player_node != null and is_instance_valid(player_node):
			player_node.call("toggle_autorun")
			_refresh_stats_panel()
	)
	column.add_child(autorun_button)
	var close_button := Button.new()
	close_button.text = "Close"
	close_button.pressed.connect(func() -> void: stats_panel.visible = false)
	column.add_child(close_button)
	if auth_panel != null and auth_panel.get_parent() != null:
		auth_panel.get_parent().add_child(stats_panel)
	else:
		add_child(stats_panel)
	_refresh_stats_panel()


func _refresh_stats_panel() -> void:
	if stats_summary_label == null:
		return
	var hp_regen := 1 + int(current_constitution / 5)
	var act_regen := 2 + int(current_stamina / 5)
	var mnd_regen := 2 + int(current_willpower / 5)
	var pace_text := "RUN (100%)"
	if player_node != null and is_instance_valid(player_node) and bool(player_node.call("is_walk_mode")):
		pace_text = "WALK (55%)"
	stats_summary_label.text = (
		"Primary HAM Pools:\n"
		+ "  • Health: %d / %d  (+%d HP / 5s)\n" % [current_health, current_max_health, hp_regen]
		+ "  • Action: %d / %d  (+%d ACT / 5s)\n" % [current_action, current_max_action, act_regen]
		+ "  • Mind: %d / %d  (+%d MND / 5s)\n\n" % [current_mind, current_max_mind, mnd_regen]
		+ "Secondary Attributes:\n"
		+ "  • Constitution (Health Regen & Cap): %d\n" % current_constitution
		+ "  • Stamina (Action Regen & Cap): %d\n" % current_stamina
		+ "  • Willpower (Mind Regen & Cap): %d\n\n" % current_willpower
		+ "Combat & Movement:\n"
		+ "  • Attack Bonus: +%d  •  Defense Bonus: +%d\n" % [current_attack_bonus, current_defense_bonus]
		+ "  • Locomotion Pace: %s" % pace_text
	)


func _show_crafting_panel(station_id: String = "") -> void:
	if stats_panel != null:
		stats_panel.visible = false
	if skills_panel != null:
		skills_panel.visible = false
	if inventory_panel != null:
		inventory_panel.visible = false
	if market_panel != null:
		market_panel.visible = false
	if equipment_panel != null:
		equipment_panel.visible = false
	if rewards_panel != null:
		rewards_panel.visible = false
	if not station_id.is_empty():
		crafting_selected_station = station_id
	if crafting_panel != null:
		if station_id.is_empty():
			crafting_panel.visible = not crafting_panel.visible
		else:
			crafting_panel.visible = true
		if crafting_panel.visible:
			_sync_crafting_station_filter()
			_refresh_crafting_recipe_list()
		return
	crafting_panel = PanelContainer.new()
	crafting_panel.position = Vector2(400, 130)
	crafting_panel.custom_minimum_size = Vector2(410, 340)
	var column := VBoxContainer.new()
	crafting_panel.add_child(column)
	var heading := Label.new()
	heading.text = "CRAFTING SCHEMATICS"
	heading.add_theme_font_size_override("font_size", 18)
	column.add_child(heading)
	crafting_station_filter = OptionButton.new()
	crafting_station_filter.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	crafting_station_filter.add_item("All Stations")
	crafting_station_filter.set_item_metadata(0, "all")
	crafting_station_filter.add_item("Frontier Workbench")
	crafting_station_filter.set_item_metadata(1, "frontier-workbench")
	crafting_station_filter.add_item("Frontier Forge")
	crafting_station_filter.set_item_metadata(2, "frontier-forge")
	crafting_station_filter.item_selected.connect(func(idx: int) -> void:
		crafting_selected_station = str(crafting_station_filter.get_item_metadata(idx))
		_refresh_crafting_recipe_list()
	)
	column.add_child(crafting_station_filter)
	var scroll := ScrollContainer.new()
	scroll.custom_minimum_size = Vector2(380, 210)
	scroll.size_flags_vertical = Control.SIZE_EXPAND_FILL
	column.add_child(scroll)
	crafting_recipes_column = VBoxContainer.new()
	crafting_recipes_column.size_flags_horizontal = Control.SIZE_EXPAND_FILL
	scroll.add_child(crafting_recipes_column)
	crafting_status_label = Label.new()
	crafting_status_label.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
	crafting_status_label.text = "Move within 3m of the required station to craft."
	column.add_child(crafting_status_label)
	var close_button := Button.new()
	close_button.text = "Close"
	close_button.pressed.connect(func() -> void: crafting_panel.visible = false)
	column.add_child(close_button)
	if auth_panel != null and auth_panel.get_parent() != null:
		auth_panel.get_parent().add_child(crafting_panel)
	else:
		add_child(crafting_panel)
	_sync_crafting_station_filter()
	_refresh_crafting_recipe_list()


func _sync_crafting_station_filter() -> void:
	if crafting_station_filter == null:
		return
	for idx in range(crafting_station_filter.item_count):
		if str(crafting_station_filter.get_item_metadata(idx)) == crafting_selected_station:
			crafting_station_filter.select(idx)
			return


func _refresh_crafting_recipe_list() -> void:
	if crafting_recipes_column == null:
		return
	for child in crafting_recipes_column.get_children():
		child.queue_free()
	var matched := 0
	for value in recipe_catalog:
		if not value is Dictionary:
			continue
		var recipe: Dictionary = value
		var station_id := str(recipe.get("stationId", ""))
		if crafting_selected_station != "all" and station_id != crafting_selected_station:
			continue
		matched += 1
		var recipe_id := str(recipe.get("id", ""))
		var recipe_name := str(recipe.get("name", recipe_id))
		var skill_id := str(recipe.get("skillId", "crafting")).capitalize()
		var req_level := int(recipe.get("requiredLevel", 1))
		var xp := int(recipe.get("experience", 0))
		var ing_parts: Array[String] = []
		var has_all_ingredients := true
		var ingredients: Variant = recipe.get("ingredients", [])
		if ingredients is Array:
			for ing in ingredients:
				if ing is Dictionary:
					var item_id := str(ing.get("itemId", ""))
					var need := int(ing.get("quantity", 1))
					var have := int(current_inventory.get(item_id, 0))
					if have < need:
						has_all_ingredients = false
					ing_parts.append(item_id.replace("_", " ") + " " + str(have) + "/" + str(need))
		var row_label := Label.new()
		row_label.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
		row_label.text = recipe_name + " (" + skill_id + " Lv " + str(req_level) + " • +" + str(xp) + " XP)\nRequires: " + ", ".join(ing_parts)
		crafting_recipes_column.add_child(row_label)
		var craft_btn := Button.new()
		craft_btn.text = "Craft " + recipe_name if has_all_ingredients else "Missing ingredients (" + recipe_name + ")"
		craft_btn.disabled = not has_all_ingredients
		craft_btn.pressed.connect(_submit_craft_recipe.bind(recipe_id))
		crafting_recipes_column.add_child(craft_btn)
		var sep := HSeparator.new()
		crafting_recipes_column.add_child(sep)
	if matched == 0:
		var empty_lbl := Label.new()
		empty_lbl.text = "No crafting recipes available for this station."
		crafting_recipes_column.add_child(empty_lbl)


func _submit_craft_recipe(recipe_id: String) -> void:
	if player_id.is_empty() or auth_token.is_empty() or interaction_pending or gather_request == null:
		return
	interaction_type = "craft"
	var error := gather_request.request(
		api_base_url + "/api/v1/account/characters/" + player_id + "/craft/" + recipe_id,
		_api_headers(),
		HTTPClient.METHOD_POST,
		""
	)
	if error == OK:
		interaction_pending = true
		status_label.text = "Crafting " + recipe_id.replace("-", " ") + "..."
		if crafting_status_label != null:
			crafting_status_label.text = "Crafting " + recipe_id.replace("-", " ") + "..."
	else:
		_report_interaction_request_error(error)


func _show_inventory_panel() -> void:
	if crafting_panel != null:
		crafting_panel.visible = false
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
		var weapon_id := str(current_equipment.get("main_hand", ""))
		var weapon := Node3D.new()
		weapon.position = Vector3(0.48, 0.95, -0.05)
		weapon.rotation_degrees.z = -22.0
		var primary := MeshInstance3D.new()
		var primary_mesh := BoxMesh.new()
		if weapon_id == "boar_tusk_spear":
			primary_mesh.size = Vector3(0.08, 1.35, 0.08)
			primary.position.y = 0.55
		elif weapon_id == "frontier_pistol":
			primary_mesh.size = Vector3(0.10, 0.22, 0.36)
			primary.position = Vector3(0.0, 0.10, -0.14)
		elif weapon_id == "frontier_rifle":
			primary_mesh.size = Vector3(0.10, 0.20, 0.82)
			primary.position = Vector3(0.0, 0.12, -0.32)
		elif weapon_id == "resonance_focus":
			primary_mesh.size = Vector3(0.09, 1.25, 0.09)
			primary.position.y = 0.50
			var focus_orb := MeshInstance3D.new()
			var orb_mesh := SphereMesh.new()
			orb_mesh.radius = 0.14
			orb_mesh.height = 0.28
			focus_orb.mesh = orb_mesh
			focus_orb.position.y = 1.18
			var orb_mat := StandardMaterial3D.new()
			orb_mat.albedo_color = Color(0.28, 0.68, 0.88)
			focus_orb.material_override = orb_mat
			weapon.add_child(focus_orb)
		else:
			primary_mesh.size = Vector3(0.10, 0.85, 0.07)
			primary.position.y = 0.42
		primary.mesh = primary_mesh
		primary.material_override = metal
		weapon.add_child(primary)
		var grip := MeshInstance3D.new()
		var grip_mesh := BoxMesh.new()
		grip_mesh.size = Vector3(0.12, 0.24, 0.12)
		grip.mesh = grip_mesh
		grip.position.y = -0.12
		grip.material_override = leather
		weapon.add_child(grip)
		equipment_visual_root.add_child(weapon)
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
	if crafting_panel != null:
		crafting_panel.visible = false
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
		var unlock_summary := _describe_skill_progression_unlocks(str(skill_id), level)
		if not unlock_summary.is_empty():
			var unlock_label := Label.new()
			unlock_label.text = unlock_summary
			unlock_label.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
			unlock_label.add_theme_font_size_override("font_size", 11)
			skill_column.add_child(unlock_label)
		skills_list.add_child(skill_column)
	skills_status_label.text = str(CORE_SKILL_IDS.size()) + " core skills • progression updates after server actions"


func _describe_skill_progression_unlocks(skill_id: String, current_level: int) -> String:
	var recipe_station_by_output: Dictionary = {}
	for recipe in recipe_catalog:
		if recipe is Dictionary:
			var station_label := "Forge" if str(recipe.get("stationId", "")) == "frontier-forge" else "Workbench"
			var req_sk := str(recipe.get("skillId", "")).capitalize()
			var req_lv := int(recipe.get("requiredLevel", 1))
			var outputs: Variant = recipe.get("outputs", [])
			if outputs is Array:
				for out in outputs:
					if out is Dictionary:
						recipe_station_by_output[str(out.get("itemId", ""))] = "%s (%s Lv %d)" % [station_label, req_sk, req_lv]
	var unlocked: Array[String] = []
	var next_unlock := ""
	var next_level := 999
	for node in gathering_catalog:
		if node is Dictionary and str(node.get("skillId", "")) == skill_id:
			var req := int(node.get("requiredLevel", 1))
			var entry := "Lv %d Harvest: %s" % [req, str(node.get("name", node.get("id", "")))]
			if req <= current_level:
				unlocked.append(entry)
			elif req < next_level:
				next_level = req
				next_unlock = entry
	for recipe in recipe_catalog:
		if recipe is Dictionary and str(recipe.get("skillId", "")) == skill_id:
			var req := int(recipe.get("requiredLevel", 1))
			var st := "Forge" if str(recipe.get("stationId", "")) == "frontier-forge" else "Workbench"
			var entry := "Lv %d %s: %s" % [req, st, str(recipe.get("name", recipe.get("id", "")))]
			if req <= current_level:
				unlocked.append(entry)
			elif req < next_level:
				next_level = req
				next_unlock = entry
	for item in market_catalog:
		if item is Dictionary and str(item.get("requiredSkillId", "")) == skill_id and str(item.get("equipmentSlot", "")) != "":
			var req := int(item.get("requiredSkillLevel", 1))
			var item_id := str(item.get("id", ""))
			var source := str(recipe_station_by_output.get(item_id, "Monster Drop / Merchant"))
			var entry := "Lv %d Equip: %s [%s]" % [req, str(item.get("name", item_id)), source]
			if req <= current_level:
				unlocked.append(entry)
			elif req < next_level:
				next_level = req
				next_unlock = entry
	var lines: Array[String] = []
	if unlocked.size() > 0:
		lines.append("Unlocked -> " + unlocked[unlocked.size() - 1])
	if not next_unlock.is_empty():
		lines.append("Next -> " + next_unlock)
	return "\n".join(lines)


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
