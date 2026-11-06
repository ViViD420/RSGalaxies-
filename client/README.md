# RS Galaxies client prototype

This is an early Godot 4 third-person 3D movement and server-connectivity prototype. It is intentionally a small technical test, not the final art direction or game client.

## Run

Open this folder in Godot 4.3 or newer and run the main scene.

- WASD: move
- Space: jump
- Mouse: orbit camera
- Escape: release the mouse cursor
- Click: capture the mouse cursor again

The client checks GET http://127.0.0.1:8080/healthz on startup. Start the Go development server from the server folder first. Change api_base_url in scripts/main.gd if the server is hosted at a different address.

## Current limitations

- The client movement is local-only; the server does not yet own or replicate position.
- The world is a generated test floor with a placeholder capsule character.
- There is no authentication, persistence, inventory, combat, assets or multiplayer networking yet.
- This prototype is used to validate the camera and movement feel before committing to a full client architecture.
