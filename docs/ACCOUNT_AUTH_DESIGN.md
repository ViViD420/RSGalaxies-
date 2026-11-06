# Account authentication and session design

Status: credential primitives implemented; account persistence and HTTP integration are the next milestone. Production player routes remain disabled.

## Security boundaries

- Account identity is separate from character identity. An account may own multiple characters; every character lookup and mutation must be scoped by the authenticated account ID.
- Passwords are never stored or logged. The current credential primitive uses PBKDF2-HMAC-SHA-256 with a per-password 128-bit random salt and 600,000 iterations. Encoded hashes carry an algorithm/version marker and bounded iteration count.
- Session tokens are 256-bit random opaque bearer secrets. Only SHA-256 token digests are stored server-side; the raw token is returned once.
- Session records must have explicit creation, expiry, revocation and last-used timestamps. Logout revokes the server-side record.
- Authentication errors must not reveal whether an account exists. Registration and login request bodies need strict size limits, unknown-field rejection and trailing-JSON rejection.
- Cookie-based browser sessions must use Secure, HttpOnly and SameSite=Lax (or stricter), with CSRF protection on state-changing cookie-authenticated requests. Bearer tokens must not be placed in URLs or logs.
- Rate limiting, email verification, password reset and abuse monitoring are required before public account registration.

## Persistence model

- accounts: stable ID, normalized unique email, password hash, created/updated timestamps, disabled timestamp.
- sessions: token digest primary/unique key, account foreign key, created/expiry/last-used/revoked timestamps, optional user-agent metadata (never a raw credential).
- players: nullable account_id during migration for legacy development rows; indexed account_id. New character creation must require an authenticated account.
- A migration must preserve existing development player data. Production character reads and writes must include account ownership in the SQL predicate, not just check ownership in application memory.

## Enablement gate

Do not expose player APIs in production until account creation/login/logout, session expiry/revocation, account-scoped character CRUD and all gameplay mutation endpoints have integration tests against PostgreSQL. Public player listing and arbitrary player-ID access must never be restored.
