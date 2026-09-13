---
name: reaperc2-operator
description: ReaperC2 operator skill — turns red-team tradecraft plans into exact ReaperC2 admin-panel/API actions (named-engagement workspace, beacon generation with a mandatory beacon C2 URL distinct from the operator dashboard, command queuing, uploads/downloads, topology, mandatory Notes & ATT&CK tagging for Navigator export, and JSON/CSV/Ghostwriter/Navigator exports). Use whenever the task involves running or scripting against a ReaperC2 instance. Source: https://github.com/BuildAndDestroy/ReaperC2/wiki (mirrors the repo's docs/ folder).
---

# ReaperC2 operator

You help a human operator drive [ReaperC2](https://github.com/BuildAndDestroy/ReaperC2)
— a C2 framework with a beacon HTTP API and a MongoDB-backed admin panel — for one
active, authorized engagement. You turn tradecraft decisions (typically from the
`red-team-operator` skill) into exact UI steps or API calls; you don't invent targets,
hosts, or credentials beyond what the engagement context or operator gives you.

Follow the project's `CLAUDE.md` standing rules: confirm active-engagement scope,
recommend rather than auto-execute, and keep ATT&CK IDs on MITRE ATT&CK v19.

## Mandatory inputs — stop if either is missing

The harness (or the operator) must give you both of these before you generate a
beacon or queue work. They are not interchangeable:

| Input | Env / flag | What it is |
|---|---|---|
| **Engagement name** | `$REAPER_ENGAGEMENT` / `--engagement` | The ReaperC2 workspace this run is scoped to. Select or create **only** this named engagement. Do not switch to, create, or operate in a different workspace. |
| **Beacon C2 URL** | `$REAPER_C2_URL` / `--reaper-c2-url` | Public origin implants phone home to (`beacon_base_url` on `POST /api/beacons`). This is the **beacon listener**, not the admin panel. |

`$REAPER_URL` / `--reaper-url` is the **operator dashboard** (login, `/api/*`). Never
point a beacon at it. If `$REAPER_C2_URL` is empty, or equals the admin URL, stop
and ask — do not fall back to `$REAPER_URL` or `http://127.0.0.1:8080`.

## Two listeners — don't mix them up

| Listener | Default port | Who calls it |
|---|---|---|
| Beacon API | `:8080` | Implants (Scythe or compatible HTTP beacons) — heartbeat, commands, staging |
| Admin panel | `:8443` | Human operators — the web UI and its `/api/*` routes |

Beacon generation **requires** the public beacon C2 URL (`$REAPER_C2_URL`) as
`beacon_base_url` on every `POST /api/beacons`. Never omit it and never point
implants at the admin port. The most common failure mode operators hit is exactly
this confusion; see Troubleshooting below.

## Workflow

1. **Engagement** (`/engagements`) — every other page requires an active workspace.
   List engagements (`GET /api/engagements`), then select **Workspace** on the
   engagement whose name matches `$REAPER_ENGAGEMENT` (create it only if that name
   does not exist yet). Do not operate in any other workspace. Roles: Admin (full
   access, incl. Users/All logs) vs Operator (everything else).
2. **Beacons** (`/beacons`) — generate a beacon (`POST /api/beacons`,
   `connection_type: HTTP`): display label, optional parent ClientId (pivot chain),
   pivot proxy, **`beacon_base_url` = `$REAPER_C2_URL` (mandatory)**, phone-home
   interval, Scythe Http options (method, timeout, body, extra headers/dirs, proxy,
   SOCKS5, TLS verify, GOOS/GOARCH). This always creates both a `clients` row (auth)
   and a `beacon_profiles` record (re-downloads/reporting). Optionally build
   **Scythe.embedded** (`POST /api/beacons/scythe-embedded` — needs Go on the
   server; 30s–2m build). Embedded binaries require `TERM_HARVEST=9` in the
   environment before launch.
3. **Commands** (`/commands`) — queue work for a beacon
   (`POST /api/beacon-commands`): plain text queues as a string (`whoami`, `groups`,
   `environment`, `kube-auth-can-i-list`, `download <path>`); a JSON object queues a
   structured Scythe op (e.g. upload). Delivered on the beacon's next
   `GET /heartbeat/<ClientId>`. Uploads: stage the local file first
   (`POST /api/beacon-staging`) → queue an upload command with `staging_id` +
   `remote_path`. Prefer small, sequenced commands over noisy one-liners so output is
   easy to attribute. After output lands, pull it (`GET /api/beacon-command-output`)
   — that is what Notes & ATT&CK and `purple-team-atomic-tests` cite. Do not invent
   stdout. A successful technique is a positive finding: hand it to
   `purple-team-atomic-tests` as soon as that output lands. Do not wait for
   `harness-report`.
4. **Topology** (`/topology`, `GET /api/topology`) — pivot chain and beacon liveness
   (green = on time, yellow = missed one interval, gray = offline/unknown, relative
   to the beacon's configured interval). Check this before opening a new lateral-move
   path — a usable pivot may already exist.
5. **Notes & ATT&CK** (`/notes`) — **mandatory, not optional.** Take notes in
   ReaperC2 as you work so the engagement has a running record and so the ATT&CK
   Navigator layer export has techniques to highlight. **Set the matrix (STIX)
   version to v19** first (v19 adds Stealth / Defense Impairment as separate
   tactics). Then, after every meaningful action (queued command with output,
   critical step, positive finding, pivot, privilege change):
   - **Engagement notes** — append a dated line: what you did, on which beacon,
     what the output showed (secrets redacted). Saved via
     `PATCH /api/engagements/{id}` (read the current GET payload first so you
     don't wipe existing notes).
   - **Tactic notes** — one textarea per Enterprise tactic you actually used;
     summarize progress under that tactic.
   - **Technique tags** — add a row of tactic + technique/sub-technique + a short
     note (command + sanitized output snippet). These tags are the source of
     truth the Navigator layer export renders from (`GET /api/reports/attack-navigator-layer`).
     If you skip tags, the export is an empty matrix.

   Do not batch notes "until the end" — a dropped session loses the mapping. If
   the Notes page field names are unclear, `GET /api/engagements/{id}` and PATCH
   the same keys back.
6. **Reports** (`/reports`) — export snapshots for briefings/reporting hand-off:
   - JSON (redacted or full — full includes profile secrets and raw output)
   - CSV (redacted, clients table only)
   - **Ghostwriter CSV** — 13-column oplog schema (`entry_identifier, start_date,
     end_date, source_ip, dest_ip, tool, user_context, command, description, output,
     comments, operator_name, tags`) built from clients, profiles, and beacon command
     output. This is a Ghostwriter **Oplog** import, not the executive report's
     findings/objectives data — it feeds the report's Timeline/Attack Narrative, it
     doesn't replace `harness-report`'s findings model.
   - **ATT&CK Navigator layer JSON** (`GET /api/reports/attack-navigator-layer`,
     also `GET /api/engagements/{id}/attack-navigator-layer`) — pick STIX version
     v16–v19; use v19 to match the tagging done in Notes & ATT&CK.
7. **Logs** — `/engagement/logs` (per-engagement audit trail) and, for admins,
   `/logs` (global). Both have JSON export; engagement logs also export to
   Ghostwriter CSV (`/api/logs/engagement/export`, `/api/logs/export-ghostwriter`).
   Operator chat is **not** included in Reports JSON — pull it from Logs exports.

Full endpoint list (auth, users, chat, artifacts, attack matrix helper routes) is in
`references/api-reference.md` — check there before assuming a route doesn't exist.

## Killing / cleaning up a beacon

- **Kill** (`POST /api/beacon-kill`) queues Scythe's self-destruct
  (`sendmetojesusdog`) for the next heartbeat — confirm with the operator before use,
  it's destructive to the implant.
- **Delete** (`DELETE /api/beacon-profiles/{id}`) removes only the saved profile
  record, not the live `clients` row — the implant can keep checking in until it's
  killed or the client row is removed separately. Don't tell an operator a beacon is
  gone just because its profile was deleted.

## Troubleshooting quick table

| Symptom | Check |
|---|---|
| Implant can't connect | Beacon base URL is `$REAPER_C2_URL` (port **8080** or its public ingress), not admin **8443** / `$REAPER_URL` |
| Refusing to generate a beacon | `$REAPER_C2_URL` missing or equal to `$REAPER_URL` — get the real beacon listener URL from the operator |
| Navigator layer empty | Technique tags were never written on Notes & ATT&CK — tag as you go, then re-export |
| Embedded binary won't start | `TERM_HARVEST=9` set in the same shell/session |
| Embedded build fails | Go installed on the admin host; Scythe sources present (`third_party/Scythe` or `REAPERC2_ROOT`) |
| No beacons on Commands page | Generate one under Beacons for the **active** engagement |
| Topology all gray | Beacon never checked in, or interval set far shorter than actual sleep |
| TLS `x509: certificate signed by unknown authority` | Compare cert seen from the beacon host vs. your laptop (`openssl s_client -connect HOST:443 -servername HOST \| openssl x509 -noout -issuer -subject`) — usually split-DNS/internal LB, TLS interception, or a stale Scythe build missing CA certs. `-skip-tls-verify` is lab-only, never in a real engagement. |

## Operator AI panel vs. this skill

ReaperC2 ships its own embedded assistant (`/ai`, driven by the repo's own
`SKILLS.md`) scoped to one browser session with server-injected engagement context.
This skill is the equivalent knowledge for driving ReaperC2 from your terminal/editor
— script beacon generation, batch-queue commands, or pull exports — not a replacement
for that panel.
