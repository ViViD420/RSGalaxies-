# RS Galaxies HTTP API

The world service listens on port 8080 by default. Responses are JSON unless the status is 204.

## Health

- `GET /healthz` — process liveness; does not depend on PostgreSQL.
- `GET /readyz` — dependency readiness; performs a two-second bounded PostgreSQL ping when configured and returns HTTP 503 without leaking connection details if unavailable.

## Public catalogs

- `GET /api/v1/world/catalog` — consolidated skills, zones, gathering nodes, items/prices, enemies, recipes, stations and starting coin balance.
- `GET /api/v1/skills`
- `GET /api/v1/world/zones`
- `GET /api/v1/world/nodes`
- `GET /api/v1/items`
- `GET /api/v1/market/items` — item definitions and merchant prices.
- `GET /api/v1/combat/enemies` — enemy spawn definitions.
- `GET /api/v1/crafting/recipes`
- `GET /api/v1/crafting/stations`

## Authentication

Account routes are registered when PostgreSQL-backed account services are configured.

### Register

`POST /api/v1/auth/register`

```json
{"email":"player@example.com","password":"a-long-unique-passphrase"}
```

Returns the account and a one-time opaque `accessToken`. The token must be kept secret by the client. Passwords must be 12–1024 bytes. Registration is available in development; production registration is disabled unless `RS_GALAXIES_ENABLE_REGISTRATION=true`.

### Login

`POST /api/v1/auth/login`

Uses the same request body as registration. Unknown email and incorrect password intentionally return the same 401 response.

### Authenticated account routes

Send `Authorization: Bearer <accessToken>`:

- `GET /api/v1/auth/me`
- `POST /api/v1/auth/logout` — revokes the current session; returns 204.
- `GET /api/v1/account/characters`
- `POST /api/v1/account/characters` with `{"name":"Wayfarer"}`
- `GET /api/v1/account/characters/{characterId}`
- `GET /api/v1/account/characters/{characterId}/skills`
- `POST /api/v1/account/characters/{characterId}/skills/{skillId}/experience` with `{"amount":25}`
- `POST /api/v1/account/characters/{characterId}/move` with `{"x":1.0,"z":2.0}`
- `POST /api/v1/account/characters/{characterId}/gather/{nodeId}`
- `POST /api/v1/account/characters/{characterId}/craft/{recipeId}`
- `POST /api/v1/account/characters/{characterId}/bank/deposit` with `{"itemId":"softwood_log","quantity":10}`
- `POST /api/v1/account/characters/{characterId}/bank/withdraw` with `{"itemId":"softwood_log","quantity":10}`
- `POST /api/v1/account/characters/{characterId}/market/buy/{itemId}` with `{"quantity":1}`
- `POST /api/v1/account/characters/{characterId}/market/sell/{itemId}` with `{"quantity":1}`
- `POST /api/v1/account/characters/{characterId}/equipment/equip/{itemId}`
- `POST /api/v1/account/characters/{characterId}/equipment/unequip/{slot}`
- `POST /api/v1/account/characters/{characterId}/combat/attack/{enemyId}`
- `POST /api/v1/account/characters/{characterId}/rest` — recover at the safe spawn, or respawn there after defeat.
- `GET /api/v1/account/characters/{characterId}/chat/local?after={messageId}&limit=50`
- `POST /api/v1/account/characters/{characterId}/chat/local` with `{"message":"Hello, zone!"}`.
- `POST /api/v1/account/characters/{characterId}/inventory/use/{itemId}` — consume an owned item effect such as a healing potion.
- `POST /api/v1/account/characters/{characterId}/presence/heartbeat` — refresh this owned character's 20-second zone-presence lease; returns 204.
- `GET /api/v1/account/characters/{characterId}/presence` — list up to 100 other active characters in the same zone, exposing only character ID, display name, zone and position.
- `GET /api/v1/account/characters/{characterId}/world/enemies` — list authoritative enemy instance positions, health, behavior and telegraph/respawn timestamps for the character's current zone. Requires account ownership.
- `POST /api/v1/account/characters/{characterId}/trades/invite/{targetCharacterId}` with `{"offer":{"items":{"softwood_log":2},"coins":10}}` — create a nearby-player invitation and escrow the initiator's offer.
- `POST /api/v1/account/characters/{characterId}/trades/accept/{tradeId}` with `{"offer":{"items":{"rough_stone":1},"coins":5}}` — accept and escrow the target's offer.
- `GET /api/v1/account/characters/{characterId}/trades/active` — recover the character's current trade after reconnect.
- `GET /api/v1/account/trades/{tradeId}` — read a trade only when the authenticated account owns one of its participants.
- `PUT /api/v1/account/characters/{characterId}/trades/offer/{tradeId}` with `{"revision":2,"offer":{"items":{},"coins":0}}` — replace the caller's escrowed offer; invalidates both confirmations.
- `POST /api/v1/account/characters/{characterId}/trades/confirm/{tradeId}` with `{"revision":3}` — confirm the exact offer revision; the second confirmation settles atomically.
- `POST /api/v1/account/characters/{characterId}/trades/cancel/{tradeId}` — cancel and return both participants' escrow.

Trade invitations require a fresh target presence lease and both characters to be active in the same zone within six world units. Each trade expires after five minutes; escrow is returned atomically on cancellation or expiry. Offers are limited to 20 distinct catalog items, and the server reserves inventory-stack and currency capacity so escrow can always be refunded.

A new local-chat reader starts with `after=0` and receives the latest 50 zone messages in chronological order; later requests fetch newer IDs from the returned cursor. A PostgreSQL-backed world worker advances enemy AI every 250 ms. It selects fresh character presence, locks the character row before the enemy spawn row, and commits enemy movement/behavior and any completed telegraph damage plus Defense XP atomically. Multiple server instances serialize on the same rows. Every character operation is scoped by the authenticated account in its SQL query or transaction. A character belonging to another account is reported as not found.

New characters start with 100 coins; starter merchant prices are part of the public item catalog. Purchases and sales atomically update the owned character's coin balance and inventory. Sessions expire after 30 days. The server stores only a SHA-256 digest of the random 256-bit token. Each account is limited to 10 active sessions; creating another prunes the oldest. Logout revokes the digest server-side.

## Production configuration

- `RS_GALAXIES_ENV=production`
- `RS_GALAXIES_DATABASE_URL` is mandatory; startup fails closed without it.
- `RS_GALAXIES_ALLOWED_ORIGINS` is a comma-separated exact-origin allowlist for browser clients. Leave empty if the client is same-origin.
- Keep `RS_GALAXIES_ENABLE_REGISTRATION` unset until email verification, edge-level rate limiting, monitoring and account recovery operations are ready.
- Legacy unscoped `/api/v1/players` routes remain disabled in production.

The application adds a small per-process IP/path throttle for registration and login. It is not a substitute for distributed rate limiting at the trusted edge. Account recovery, email verification, MFA and administrative account-disable tooling are not implemented yet.
