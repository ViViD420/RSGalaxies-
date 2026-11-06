# Zone presence and remote character prototype

## Protocol

- Authenticated clients send `POST /api/v1/account/characters/{characterId}/presence/heartbeat` every five seconds.
- The server derives zone and coordinates from the account-owned character row; clients cannot publish arbitrary locations or account identifiers.
- `GET /api/v1/account/characters/{characterId}/presence` returns up to 100 other characters with a live lease in the caller's current zone.
- The lease is 20 seconds. A disconnected client naturally disappears from rosters after expiry; heartbeat writes also prune rows older than five minutes.
- Responses expose only character ID, display name, zone ID, position and last-seen timestamp.

## Persistence and concurrency

- Migration 0008 creates the zone_presence table and indexes by active zone and account.
- Heartbeat locks the owned character row and updates its presence in the same transaction. Repeated writes are coalesced to at most one persisted update per character per second.
- Presence is one row per character. A reconnect refreshes the same row instead of creating duplicate sessions or avatars.
- Roster lookup verifies ownership of the requesting character and never returns account IDs, emails, inventory, skills or session details.

## Client behavior

- The Godot prototype sends a heartbeat and then polls the roster every five seconds.
- Remote characters are rendered as simple non-colliding placeholder avatars with nameplates.
- When a character is no longer in the active roster, its placeholder is removed.
- This is a polling prototype, not realtime replication. Movement is still periodically validated through the separate movement endpoint; interpolation, prediction correction, disconnect recovery and zone handoff remain future work.

## Operational limitations

- Presence is best-effort and lease-based. It is not proof of a live authenticated socket or a guarantee of exact online status.
- The current endpoint is suitable for early multiplayer testing, not high population scale. A realtime transport, bounded zone fan-out, presence sharding and load tests are required before alpha.
