# Account authentication and session design

Status: implemented foundation. PostgreSQL accounts/sessions, bearer authentication, account-owned character CRUD and account-scoped gameplay transactions are implemented. Production registration remains disabled by default; production legacy player-ID routes remain disabled.

## Security boundaries

- Account identity is separate from character identity. An account may own multiple characters; every character lookup and mutation must be scoped by the authenticated account ID.
- Passwords are never stored or logged. The current credential primitive uses PBKDF2-HMAC-SHA-256 with a per-password 128-bit random salt and 600,000 iterations. Encoded hashes carry an algorithm/version marker and bounded iteration count.
- Session tokens are 256-bit random opaque bearer secrets. Only SHA-256 token digests are stored server-side; the raw token is returned once.
- Session records have creation, fixed 30-day expiry, revocation and last-used timestamps. Logout revokes the server-side record. The repository caps each account at 10 active sessions, prunes the oldest active session when the cap is exceeded, and removes expired or long-revoked sessions when a new session is issued.
- Authentication errors must not reveal whether an account exists; login performs a dummy PBKDF2 verification for unknown/disabled accounts to reduce response-time enumeration. Registration and login request bodies need strict size limits, unknown-field rejection and trailing-JSON rejection.
- Cookie-based browser sessions must use Secure, HttpOnly and SameSite=Lax (or stricter), with CSRF protection on state-changing cookie-authenticated requests. Bearer tokens must not be placed in URLs or logs.
- Rate limiting, email verification, password reset and abuse monitoring are required before public account registration.

## Persistence model

- accounts: stable ID, normalized unique email, password hash, created/updated timestamps, disabled timestamp.
- sessions: token digest primary/unique key, account foreign key, created/expiry/last-used/revoked timestamps, optional user-agent metadata (never a raw credential).
- players: nullable account_id during migration for legacy development rows; indexed account_id. New character creation must require an authenticated account.
- Migration 0002 adds accounts, sessions and nullable player ownership without deleting legacy development characters. Production character reads and writes include account ownership in SQL predicates.
- Account-scoped operations now include XP awards, movement, gathering and crafting, each using player row locks and transactions; gathering also locks shared node state.

## HTTP behavior

- `POST /api/v1/auth/register` and `POST /api/v1/auth/login` accept JSON `{email,password}` and return an opaque bearer token once. Tokens expire after 30 days; logout revokes the current token.
- `GET /api/v1/auth/me` validates the bearer session.
- `GET/POST /api/v1/account/characters` and `GET /api/v1/account/characters/{id}` are account-scoped. Gameplay endpoints are nested under `/api/v1/account/characters/{id}` for skills, XP, movement, gathering and crafting.
- All authenticated account routes require `Authorization: Bearer <token>`. Player ownership is part of each SQL query/transaction.
- Registration is available in development; production registration is disabled unless `RS_GALAXIES_ENABLE_REGISTRATION=true`. Keep it disabled until email verification and edge-level abuse controls are configured.
- The API applies a small per-process IP/path auth throttle (5 registration attempts and 10 login attempts per minute). This is only a local safety net; public deployments still need a trusted edge rate limiter and monitoring.
- Development CORS is permissive. Production CORS is opt-in per exact origin via comma-separated `RS_GALAXIES_ALLOWED_ORIGINS`; do not use `*` for production.
- Account recovery, email verification, MFA, administrative account disable tooling and distributed rate limiting are not implemented.

## Enablement gate

The account-owned routes can be used in production with PostgreSQL. Keep the legacy public player-ID routes disabled permanently. Before enabling public registration, add email verification, account recovery, edge-level rate limiting, monitoring and an operational account-disable/recovery process. Never expose a global player list or unscoped player-ID mutation endpoint.
