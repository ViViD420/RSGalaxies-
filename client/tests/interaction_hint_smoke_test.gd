extends SceneTree

const MAIN_SCRIPT := preload("res://scripts/main.gd")
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
	var player := CharacterBody3D.new()
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

	print("PASS: contextual interaction hints select the nearest gathering target, show range guidance, expose pending state, hide before character selection and report malformed/startup errors.")
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
