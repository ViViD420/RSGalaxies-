extends SceneTree

const MAIN_SCRIPT := preload("res://scripts/main.gd")
# Use the real player scene (controller script + nodes) so main.gd is exercised
# against the same player contract it relies on in production.
const PLAYER_SCENE := preload("res://scenes/player.tscn")
var test_main_node: Node3D


func _initialize() -> void:
	call_deferred("_run_interaction_hint_smoke_test")


func _run_interaction_hint_smoke_test() -> void:
	var main_node := Node3D.new()
	test_main_node = main_node
	main_node.set_script(MAIN_SCRIPT)
	var hint_label := Label.new()
	main_node.add_child(hint_label)
	var status_test_label := Label.new()
	main_node.add_child(status_test_label)
	main_node.set("status_label", status_test_label)
	var player := PLAYER_SCENE.instantiate() as CharacterBody3D
	root.add_child(player)
	main_node.set("player_node", player)
	main_node.set("player_id", "interaction-hint-test")
	main_node.set("interaction_pending", false)
	main_node.set("interaction_hint_label", hint_label)
	main_node.set("station_catalog", [
		{"id": "frontier-workbench", "name": "Frontier Workbench", "position": {"x": 0.0, "z": 4.0}}
	])
	main_node.set("enemy_catalog", [])
	main_node.set("enemy_targets", {})
	main_node.set("enemy_available", {})
	main_node.set("gathering_catalog", [
		{"id": "tree-001", "name": "Young Frontier Tree", "position": {"x": 2.0, "z": 0.0}}
	])

	main_node.call("_update_interaction_hint")
	var hint := main_node.get("interaction_hint_label") as Label
	if not hint.visible or not hint.text.contains("Gather Young Frontier Tree"):
		_fail("Nearby gathering target did not produce a usable gather prompt: " + hint.text)
		return

	player.position = Vector3(20.0, 0.0, 0.0)
	main_node.call("_update_interaction_hint")
	if not hint.text.begins_with("Move closer"):
		_fail("Out-of-range nearest interaction did not produce a move-closer prompt: " + hint.text)
		return

	main_node.set("interaction_pending", true)
	main_node.call("_update_interaction_hint")
	if hint.text != "Working on your interaction...":
		_fail("Pending interaction did not communicate its busy state.")
		return

	main_node.set("player_id", "")
	main_node.call("_update_interaction_hint")
	if hint.visible:
		_fail("Interaction hint remained visible without an authenticated character.")
		return

	main_node.set("interaction_pending", true)
	main_node.call("_on_gather_completed", HTTPRequest.RESULT_SUCCESS, 200, PackedStringArray(), PackedByteArray("not json".to_utf8_buffer()))
	if status_test_label.text != "Interaction: invalid server response":
		_fail("Malformed successful interaction response was not surfaced clearly: " + status_test_label.text)
		return
	main_node.call("_report_interaction_request_error", ERR_BUSY)
	if not status_test_label.text.begins_with("Interaction: request could not start"):
		_fail("Interaction request startup failure was not surfaced: " + status_test_label.text)
		return

	# Validate workbench and forge recipe priority selection for farming/scavenging/weapon loops.
	main_node.set("current_inventory", {"frontier_grain": 2, "wild_herb": 1})
	if str(main_node.call("_select_workbench_recipe")) != "frontier-ration":
		_fail("Workbench recipe selector did not prioritize frontier-ration when grain and herb are present.")
		return
	main_node.set("current_inventory", {"salvaged_scrap": 2, "copper_bar": 1})
	if str(main_node.call("_select_workbench_recipe")) != "salvaged-power-cell":
		_fail("Workbench recipe selector did not prioritize salvaged-power-cell.")
		return
	main_node.set("current_inventory", {"salvaged_power_cell": 1, "copper_bar": 2, "wooden_plank": 1})
	if str(main_node.call("_select_forge_recipe")) != "frontier-rifle":
		_fail("Forge recipe selector did not prioritize frontier-rifle when power cell, copper and plank are present.")
		return

	# Validate Crafting Schematics panel builds, filters by station, and cleans up.
	main_node.set("recipe_catalog", [
		{"id": "basic-plank", "name": "Basic Wooden Plank", "skillId": "crafting", "requiredLevel": 1, "stationId": "frontier-workbench", "ingredients": [{"itemId": "softwood_log", "quantity": 2}], "outputs": [{"itemId": "wooden_plank", "quantity": 1}], "experience": 30},
		{"id": "copper-bar", "name": "Smelt Copper Bar", "skillId": "smithing", "requiredLevel": 1, "stationId": "frontier-forge", "ingredients": [{"itemId": "copper_ore", "quantity": 2}], "outputs": [{"itemId": "copper_bar", "quantity": 1}], "experience": 35}
	])
	main_node.call("_show_crafting_panel", "frontier-forge")
	var craft_panel := main_node.get("crafting_panel") as PanelContainer
	var craft_column := main_node.get("crafting_recipes_column") as VBoxContainer
	if craft_panel == null or not craft_panel.visible or craft_column == null or craft_column.get_child_count() == 0:
		_fail("Crafting Schematics panel did not build and populate for frontier-forge.")
		return

	# Validate new medpack/buff/alloy recipe selectors and HAM/Secondary attribute sync.
	main_node.set("current_inventory", {"bacta_gel": 1, "herbal_bundle": 1})
	if str(main_node.call("_select_workbench_recipe")) != "field-medpack":
		_fail("Workbench recipe selector did not prioritize field-medpack.")
		return
	main_node.set("current_inventory", {"bacta_gel": 1, "frontier_grain": 2, "aether_crystal": 1})
	if str(main_node.call("_select_workbench_recipe")) != "vitality-stim-pack":
		_fail("Workbench recipe selector did not prioritize vitality-stim-pack.")
		return
	main_node.set("current_inventory", {"iron_ore": 2})
	if str(main_node.call("_select_forge_recipe")) != "iron-bar":
		_fail("Forge recipe selector did not prioritize iron-bar.")
		return
	main_node.call("_sync_vital_attributes_from_character", {"action": 95, "mind": 90, "constitution": 15, "stamina": 12, "willpower": 14, "maxHealthBuff": 35})
	if int(main_node.get("current_max_health")) != 145 or int(main_node.get("current_constitution")) != 15:
		_fail("HAM/Secondary attribute sync did not compute expected max health and constitution.")
		return

	# Validate Character Vitals & Attributes panel builds and renders HAM + Secondary attributes.
	main_node.call("_show_stats_panel")
	var vitals_panel := main_node.get("stats_panel") as PanelContainer
	var vitals_label := main_node.get("stats_summary_label") as Label
	if vitals_panel == null or not vitals_panel.visible or vitals_label == null or not vitals_label.text.contains("Constitution"):
		_fail("Character Vitals & Attributes panel did not build or render Constitution summary.")
		return

	# Validate Frontier World Atlas & Directory panel builds and populates station/node entries.
	main_node.call("_show_guide_panel")
	var atlas_panel := main_node.get("guide_panel") as PanelContainer
	var atlas_column := main_node.get("guide_entries_column") as VBoxContainer
	if atlas_panel == null or not atlas_panel.visible or atlas_column == null or atlas_column.get_child_count() == 0:
		_fail("Frontier World Atlas & Directory panel did not build or populate entries.")
		return

	# Validate Options, Camera & Controls panel builds cleanly.
	main_node.call("_show_settings_panel")
	var opt_panel := main_node.get("settings_panel") as PanelContainer
	if opt_panel == null or not opt_panel.visible:
		_fail("Options, Camera & Controls panel did not build or become visible.")
		return

	print("PASS: contextual interaction hints select the nearest gathering target, show range guidance, expose pending state, hide before character selection, select station recipes, build Crafting Schematics, Character Vitals, World Atlas and Options panels, sync HAM/Secondary attributes and report malformed/startup errors.")
	player.queue_free()
	main_node.free()
	await process_frame
	quit(0)


func _fail(message: String) -> void:
	push_error("Interaction hint smoke test failed: " + message)
	for child in root.get_children():
		child.queue_free()
	if test_main_node != null and is_instance_valid(test_main_node):
		test_main_node.free()
	call_deferred("_quit_after_cleanup", 1)


func _quit_after_cleanup(exit_code: int) -> void:
	await process_frame
	quit(exit_code)
