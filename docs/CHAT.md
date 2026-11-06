# Local Zone Chat

## Current behavior

- Chat messages are visible to authenticated characters in the same world zone.
- Sending and reading require account-owned character routes; the server derives the zone and sender identity from the character record rather than trusting client-supplied values.
- Messages are trimmed, control characters are removed, and the content is limited to 240 Unicode characters.
- Each character has a two-second send cooldown enforced transactionally in PostgreSQL.
- Reads use an ascending message-ID cursor and return at most 50 messages per request.
- The Godot client polls every two seconds, appends only messages newer than its last cursor, and holds the access token in memory.

## API

- GET /api/v1/account/characters/{characterId}/chat/local?after={messageId}&limit=50
- POST /api/v1/account/characters/{characterId}/chat/local with JSON {"message":"Hello, zone!"}

## Persistence and safety

- Migration 0007 creates the chat_messages table and zone/cursor and sender/cooldown indexes.
- A send transaction locks the owned character row, checks its most recent message timestamp, and inserts the message before commit.
- The zone is taken from the locked character state. Clients cannot post into another zone or impersonate another character.
- Message history is currently retained without an automated expiry policy. Add retention, moderation/reporting, block lists and operational abuse controls before a public launch.

## Limitations

This is polling-based local chat, not a realtime socket service. Direct messages, friends, groups, guild channels, moderation tooling, profanity policy and global chat are not implemented.
