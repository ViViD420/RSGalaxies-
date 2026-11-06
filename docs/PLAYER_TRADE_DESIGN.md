# Player-to-Player Trade Design

Status: invite/accept/read API slice implemented; escrow offer updates, confirmations, settlement, cancellation and Godot trade UI are not implemented. Do not enable player trading for production until the full lifecycle and required tests are complete.

## Goals

Add a safe, server-authoritative exchange between two characters without exposing inventory or currency mutation to clients. The first implementation should support stackable catalog items and coins, work across two accounts, persist through server restarts, and never rely on a client-provided price or item definition.

## Trade lifecycle

1. **Invite** — an authenticated character invites another visible character in the same zone and within the configured interaction range. Reject self-trades, defeated characters, stale presence, cross-zone targets and duplicate active trades.
2. **Accept** — the target accepts the invitation. The trade becomes open and both participants can stage offers.
3. **Offer update** — each participant may replace their own offer. The server validates item IDs, quantities, stack limits and coin balance. Offered inventory and coins move into durable escrow in the same transaction as the offer revision. The client never writes escrow data directly.
4. **Confirm** — each side confirms the exact current revision. Any offer edit increments the revision and clears both confirmations. The second confirmation attempts settlement.
5. **Settle** — lock the trade and both character rows in deterministic character-ID order. Validate both destination inventories and coin caps before applying either side. Transfer escrow, update both characters, mark the trade completed, and commit once.
6. **Cancel/expire** — either participant may cancel while unsettled. A background or request-triggered expiry path returns escrow to the original owners. Repeated cancel/expire calls must be idempotent.

## Invariants

- Every active trade has exactly two distinct character participants, each bound to an owning account.
- Every escrowed item and coin is backed by an equal debit from its owner's character.
- A participant can only read or mutate a trade in which their character is a participant and their account owns that character.
- A character can have at most one active trade at a time.
- Equipped items are not tradable in the first version; only inventory stacks are eligible.
- Bank contents, skills, XP, health, equipment and account data are never part of a trade payload.
- Item definitions and max stack sizes always come from the server catalog.
- No client-supplied item name, price, stack limit, character owner, account ID or settlement status is trusted.
- Settlement and escrow release are all-or-nothing PostgreSQL transactions.
- A trade is terminal after completion, cancellation or expiry. Terminal trades cannot be reopened or settled again.
- Responses must not expose the other participant's account ID, session information or private account data.

## Persistence shape

Use a trade row containing an opaque ID, both character IDs, both account IDs, status, current revision, expiry, creation/update timestamps, and terminal timestamp. Use a separate participant-offer row keyed by trade ID and character ID with a JSONB item map, coin amount, revision, and confirmation revision. Add foreign keys to characters/accounts, status and nonnegative-value checks, and indexes for participant active-trade lookup and expiry cleanup.

Prefer a normalized escrow ledger table for item and currency movements if it can be introduced without duplicating mutable balances. The authoritative balance remains the character inventory/coins plus the durable open-trade escrow rows. Add constraints to prevent duplicate item rows and negative escrow quantities.

## Transaction and locking rules

- Always lock the trade row before changing an offer or confirmation.
- For two-character settlement/release, lock player rows in ascending character-ID order to prevent opposing trade operations from deadlocking.
- Re-check ownership, status, revision and expiry inside the transaction; never rely on an earlier HTTP read.
- On offer replacement, restore the previous escrow to the owner and debit the new offer within the same transaction.
- Before settlement, simulate both destination inventories and currency balances in memory. If either side would exceed a cap, do not transfer anything and leave the trade open with confirmations reset.
- Keep transactions short; do not make network calls, wait for user input or invoke external services while holding row locks.
- Retry only errors known to be serialization/deadlock failures, and retry the entire transaction with a bounded attempt count.

## API and client contract

Expose account-scoped endpoints for invite, accept, read, update-own-offer, confirm and cancel. Use strict JSON decoding, request body limits, rate limits, stable error codes and generic not-found behavior for trades outside the caller's ownership scope. Every mutating request should include the expected trade revision; stale revisions return a conflict and the latest safe trade snapshot.

The Godot client should display the other character's public name, both staged offers, confirmation state, revision and expiry. Disable controls while a request is pending. Never optimistically display a completed trade before the server confirms settlement. On reconnect, fetch the active trade by character ID and reconcile from the server.

## Required tests before enabling routes

- Cross-account ownership and unauthorized trade lookup/update rejection.
- Self-trade, wrong-zone, out-of-range, stale-presence and defeated-character rejection.
- Offer validation, insufficient inventory/coins and stack-cap checks.
- Offer replacement restores old escrow exactly once and debits new escrow exactly once.
- Any edit invalidates both confirmations.
- Both confirmation orders settle correctly and atomically.
- Concurrent confirm requests settle once; duplicate requests are idempotent.
- Concurrent cancel/confirm and expiry/confirm cannot duplicate or destroy escrow.
- Destination inventory-full or currency-cap failures leave both characters and escrow unchanged.
- Crash/restart recovery preserves open escrow and expiry behavior.
- Race-enabled Go tests and PostgreSQL integration tests pass before production routes are registered.

## Release boundary

Do not enable player-to-player trading until persistence, ownership checks, escrow accounting, transaction locking, concurrency tests and a Godot client path are all complete. The merchant system remains separate and unchanged.
