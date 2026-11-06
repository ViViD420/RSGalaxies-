# RS Galaxies client prototype

This is an early Godot 4 third-person 3D movement and server-connectivity prototype. It is intentionally a small technical test, not the final art direction or game client.

## Run

Open this folder in Godot 4.3 or newer and run the main scene.

- Keyboard: WASD to move, Space to jump.
- Mouse: orbit camera; Escape releases the cursor.
- Touchscreen: on-screen directional pad and jump button; drag to orbit the camera.

On startup the client checks the Go API, creates a development character, and synchronizes horizontal movement every 200 ms. The server validates movement against elapsed server time, maximum speed and zone bounds. When running as a Codespaces Web build, the client derives the forwarded API origin automatically. In a local native run, start the Go server on port 8080 first; change api_base_url in scripts/main.gd if it uses a different address.

A Web export preset is included in export_presets.cfg. The Codespaces setup installs Godot 4.3 and its export templates so a browser build can be produced from the hosted workspace.

## Current limitations

- Movement is local-only; the server does not yet own or replicate position.
- The world is a generated test floor with a placeholder capsule character.
- There is no authentication, persistence, inventory, combat, production art or multiplayer networking yet.
- This prototype is used to validate camera, movement and touch interaction before committing to a full client architecture.
