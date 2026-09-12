# ReaperC2 API reference

Grouped by area. Admin-panel routes require a signed-in session (cookie from
`/login`, plus `/login/mfa` if TOTP is enabled) unless noted. Beacon routes
authenticate implants via `X-Client-Id` / `X-API-Secret` headers, not operator
sessions.

## Auth & account

| Method | Path | Purpose |
|---|---|---|
| GET/POST | `/login` | Operator sign-in |
| GET/POST | `/login/mfa` | TOTP step (if enabled) |
| POST | `/logout` | End session |
| GET/POST | `/api/account` | Current operator's account info |
| POST | `/api/account/password` | Change password |
| POST | `/api/account/totp/begin` | Start TOTP enrollment |
| POST | `/api/account/totp/verify` | Confirm TOTP enrollment |
| POST | `/api/account/totp/disable` | Disable TOTP |
| POST | `/api/account/totp/cancel` | Cancel in-progress enrollment |

## Users (admin only)

| Method | Path | Purpose |
|---|---|---|
| GET/POST | `/api/users` | List / create operator accounts |
| GET/PATCH/DELETE | `/api/users/{username}` | Manage one account |
| POST | `/api/users/{username}/password` | Reset a user's password |

## Engagements

| Method | Path | Purpose |
|---|---|---|
| GET/POST | `/api/engagements` | List / create engagements |
| GET | `/api/engagements/active` | Currently active workspace for this session |
| GET/PATCH | `/api/engagements/{id}` | Read / update one engagement (engagement notes, tactic notes, technique tags, status, haul, room, dates+name+operators if admin). GET first, then PATCH the same keys — this is how the AI operator records Notes & ATT&CK as work happens. |
| GET | `/api/engagements/{id}/attack-navigator-layer` | Navigator layer JSON scoped to this engagement |

## Beacons & profiles

| Method | Path | Purpose |
|---|---|---|
| POST | `/api/beacons` | Generate a beacon (`connection_type: HTTP` + form fields). **`beacon_base_url` is mandatory** and must be the beacon listener (`$REAPER_C2_URL`), never the admin panel URL. Creates `clients` row + `beacon_profiles` record |
| POST | `/api/beacons/scythe-embedded` | Build & download a Scythe.embedded binary from a profile's saved Http options |
| GET | `/api/beacon-profiles` | List saved profiles |
| DELETE | `/api/beacon-profiles/{id}` | Delete a profile record (does **not** remove the live client) |
| POST | `/api/beacon-kill` | Queue self-destruct (`sendmetojesusdog`) for next heartbeat |
| GET | `/api/beacon-presence` | Liveness data (backs Topology coloring) |
| GET | `/heartbeat/<ClientId>` | **Beacon-side**: implant check-in; returns pending `Commands` |
| POST | `/receive/<ClientId>` | **Beacon-side**: implant posts command output |

## Commands & artifacts

| Method | Path | Purpose |
|---|---|---|
| GET/POST | `/api/beacon-commands` | List pending queue / queue a new command (string or `command_obj` JSON) |
| GET | `/api/beacon-command-output` | Stored output history for a beacon — this is the engagement command output `purple-team-atomic-tests` must cite |
| POST | `/api/beacon-staging` | Stage a local file for upload → returns `staging_id` |
| GET | `/api/beacon-artifacts` | List file artifacts (staged uploads + beacon downloads) |
| GET | `/api/beacon-artifacts/{id}` | Metadata for one artifact |
| GET | `/api/beacon-artifacts/{id}/file` | Download artifact bytes (GridFS-backed) |

## Topology, ATT&CK matrix helpers, chat

| Method | Path | Purpose |
|---|---|---|
| GET | `/api/topology` | Pivot graph + liveness for the active engagement |
| GET | `/api/attack/matrix-tactics` | Tactic list for the selected matrix (STIX) version — use this to confirm what a given ReaperC2 instance calls v19's Stealth/Defense Impairment split |
| GET | `/api/attack/matrix-techniques` | Technique/sub-technique list for a tactic + matrix version |
| GET/POST | `/api/chat/messages` | Engagement chat (not included in Reports JSON — use Logs exports) |
| GET/POST | `/api/ai/chat`, `/api/ai/status` | ReaperC2's own embedded operator-AI panel (separate from this skill) |

## Reports & exports

| Method | Path | Purpose |
|---|---|---|
| GET | `/api/reports/export` | JSON (redacted/full) or CSV (redacted, clients only) — query params `format`, `redact` |
| GET | `/api/reports/export-ghostwriter` | Ghostwriter oplog CSV (13-column schema, see SKILL.md) |
| GET | `/api/reports/attack-navigator-layer` | Navigator layer JSON — query param `version` (STIX v16–v19) |

## Logs

| Method | Path | Purpose |
|---|---|---|
| GET | `/api/logs` | Global audit log (admin) |
| GET | `/api/logs/export` | Global audit log export |
| GET | `/api/logs/engagement` | Per-engagement audit log |
| GET | `/api/logs/engagement/export` | Per-engagement export |
| GET | `/api/logs/export-ghostwriter` | Engagement audit trail as Ghostwriter oplog CSV |

## Ghostwriter CSV column schema (oplog import)

```
entry_identifier, start_date, end_date, source_ip, dest_ip, tool, user_context,
command, description, output, comments, operator_name, tags
```

`source_ip`/`dest_ip` use the literal string `ReaperC2` for the C2 server side of an
entry. This is Ghostwriter's **Oplog** shape — separate from the executive report's
findings/objectives/scope data model that `harness-report` fills in.
