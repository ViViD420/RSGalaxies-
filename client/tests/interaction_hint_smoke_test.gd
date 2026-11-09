extends SceneTree

const MAIN_SCRIPT := preload("res://scripts/main.gd")
const PLAYER_SCENE := preload("res://scenes/player.tscn")


func _initialize() -> void:
	call_deferred("_run_interaction_hint_smoke_test")


func _run_interaction_hint_smoke_test() -> void:
	var main_node := Node3D.new()
	main_node.set_script(MAIN_SCRIPT)
	var player := PLAYER_SCENE.instantiate() as CharacterBody3D
	main_node.set("player_node", player)
	main_node.set("player_id", "interaction-hint-test")
	main_node.set("interaction_pending", false)
	main_node.set("interaction_hint_label", Label.new())
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

	player.global_position = Vector3(20.0, 0.0, 0.0)
	main_node.call("_update_interaction_hint")
	if not hint.text.begins_with("Move closer") or not hint.text.contains("Young Frontier Tree"):
		_fail("Out-of-range target did not produce a move-closer prompt: " + hint.text)
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

	print("PASS: contextual interaction hints select the nearest gathering target, show range guidance, expose pending state and hide before character selection.")
	player.queue_free()
	main_node.queue_free()
	await process_frame
	quit(0)


func _fail(message: String) -> void:
	push_error("Interaction hint smoke test failed: " + message)
	call_deferred("_quit_after_cleanup", 1)


func _quit_after_cleanup(exit_code: int) -> void:
	await process_frame
	quit(exit_code)
