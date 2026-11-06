# Deployment and operations

## Required production settings

- Set `RS_GALAXIES_ENV=production`.
- Set `RS_GALAXIES_DATABASE_URL` to a dedicated PostgreSQL database with a least-privilege application role. The process fails closed if the URL is missing.
- Set `RS_GALAXIES_ADDR` to the intended bind address. Place the service behind a TLS-terminating reverse proxy or load balancer.
- Set `RS_GALAXIES_ALLOWED_ORIGINS` to a comma-separated exact list of browser origins. Leave it empty for same-origin clients; production wildcard CORS is not supported.
- Leave `RS_GALAXIES_ENABLE_REGISTRATION` unset until email verification, password recovery, edge-level abuse controls, account monitoring and account-disable operations are ready.

## Database and backups

- Startup applies versioned migrations under a PostgreSQL advisory transaction lock.
- Back up the database and test restore procedures before upgrades.
- Keep PostgreSQL network access private. Do not expose the database port publicly.
- Monitor connection pool saturation, failed logins, HTTP 5xx responses, session volume, database storage and backup freshness.
- The server stores password hashes and SHA-256 digests of random bearer tokens; do not log request authorization headers or response bodies containing access tokens.

## Security controls and remaining launch gates

Implemented:
- PBKDF2-HMAC-SHA-256 password hashing with random salts and 600,000 iterations.
- 256-bit opaque bearer sessions, 30-day expiry, server-side revocation and a 10-active-session account cap.
- Account ownership predicates in character SQL queries and gameplay transactions.
- Per-process registration/login throttling and exact-origin production CORS.
- Production registration disabled by default; legacy unscoped player endpoints remain disabled.

Still required before public launch:
- Email verification and password recovery.
- Distributed edge rate limiting and abuse monitoring.
- MFA and administrative account disable/recovery workflows.
- Operational alerting, backup restore drills, capacity/load testing and security review.
- Secure client-side token persistence if persistent sign-in is desired. The current Godot client deliberately keeps tokens in memory only.


## Health checks

- Probe `GET /healthz` for process liveness. It deliberately does not depend on PostgreSQL, so a database outage does not trigger a restart loop.
- Probe `GET /readyz` for readiness. When PostgreSQL is configured, it performs a bounded two-second database ping and returns HTTP 503 with a generic response if the dependency is unavailable.
- Route traffic only to ready instances; let the orchestrator use liveness independently.


## Player trading release gate

Player-to-player trade routes are disabled in production unless `RS_GALAXIES_ENABLE_TRADING=true` is explicitly set. Keep the flag unset until PostgreSQL concurrency tests, cancel/confirm and expiry/confirm race tests, escrow refund/restart recovery, real-device client validation and abuse/load review are complete.
