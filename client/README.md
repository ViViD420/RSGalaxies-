# RS Galaxies client prototype

This is an early Godot 4 third-person 3D movement and server-connectivity prototype. It is intentionally a small technical test, not the final art direction or game client.

## Run

Open this folder in Godot 4.3 or newer and run the main scene.

- Keyboard: WASD to move, Space to jump.
- Mouse: orbit camera; Escape releases the cursor.
- Touchscreen: on-screen directional pad, USE and jump buttons; drag to orbit the camera.
- Gather: walk within range of the starter tree or stone deposit and press E / USE to collect a resource and gain the matching skill XP.
- Craft: collect two logs, move to the frontier workbench and press E / USE to craft a wooden plank for Crafting XP.
- Bank: move to the starter chest behind the spawn point and press E / USE to deposit one carried starter item, or withdraw one item when your bag has none of those resources.
- Mining and smithing: mine the copper vein behind the bank, smelt two ore into a bar at the frontier forge, then combine three bars and a plank to forge a training sword.
- Merchant: move to the frontier trader stall beside the tree and press E / USE to open the buy/sell panel. Choose a starter item and quantity; the server validates prices, coin balance and stack limits.

On startup the client checks the Go API and presents account login/registration. After authentication it opens an account character picker, lets the player enter an existing character or create another, and creates a starter character if the account has none. It sends the bearer token with account-scoped movement, gathering and crafting requests, and supports character switching, sign-out and re-authentication after session expiry. The token is held in memory only and is not persisted between application launches. The server validates movement against elapsed server time, maximum speed and zone bounds. When running as a Codespaces Web build, the client derives the forwarded API origin automatically. Account login requires PostgreSQL-backed account services: use the provided Codespaces workspace, or start PostgreSQL and set `RS_GALAXIES_DATABASE_URL` before launching the Go server. The in-memory-only development server does not register account routes. In a local native run, start the Go server on port 8080 first; change api_base_url in scripts/main.gd if it uses a different address.

A Web export preset is included in export_presets.cfg. The Codespaces setup installs Godot 4.3 and its export templates so a browser build can be produced from the hosted workspace.

## Current limitations

- Movement is client-predicted with periodic server validation and correction; there is no continuous world simulation or remote-player replication yet.
- The world is a generated test floor with a placeholder capsule character.
- Inventory currently supports starter stackable resources and the first crafted plank item. Authentication is integrated with bearer sessions, but the client does not yet persist tokens securely across launches, support password recovery/MFA, character selection or character deletion.
- There is no equipment, combat, production art or remote-player networking yet.
- This prototype is used to validate camera, movement and touch interaction before committing to a full client architecture.
