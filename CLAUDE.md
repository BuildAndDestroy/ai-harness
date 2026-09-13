# ai-harness

Skill set for running an authorized red team engagement end-to-end and producing the
client deliverable. Five project skills, meant to be used in sequence during one
engagement ("harness run"), launched via the `harness` Go binary (`cmd/harness`):

1. **`red-team-operator`** — plans and directs adversary emulation, mapping every
   phase to MITRE ATT&CK v19 Enterprise tactics/techniques. Produces the technique
   log (critical steps, TTPs used) that later stages consume.
2. **`reaperc2-operator`** — drives [ReaperC2](https://github.com/BuildAndDestroy/ReaperC2)
   (beacon generation against the beacon C2 URL, command queuing, topology, mandatory
   Notes & ATT&CK tagging for Navigator export, exports) for whatever the
   red-team-operator skill decides to do, scoped to the named engagement.
3. **`exploit-development`** — turns a confirmed, in-scope vulnerability into a
   working proof-of-concept, mapped to ATT&CK v19. Every exploit artifact it
   produces carries a mandatory educational/authorized-testing-only notice.
4. **`purple-team-atomic-tests`** — **required, not optional.** For each positive
   finding (a technique that succeeded against the target, i.e. a detection/control
   gap), generates Atomic Red Team-style validation tests the blue team can re-run
   in a controlled environment, including the engagement's command output. Runs
   **as soon as the technique succeeds**, not at close. Its YAML/markdown under
   `atomics/` fills the report's Validation fields. `harness-report` must not
   start until those files exist for every positive finding.
5. **`harness-report`** — at the end of the engagement, drafts the client-facing
   report against Ghostwriter's executive template (schema in
   `harness-report/references/ghostwriter-executive-template.md`; the `.docx` is
   not in this repo), using the technique log, ReaperC2 exports, exploit findings,
   and atomic-test validations gathered above. A time-boxed close or a blocked
   later objective does not skip skill 4.

## Starting an engagement

Run the `harness` binary rather than starting a bare Claude Code session by
hand. It is a scope-gate **killswitch**: it requires a ReaperC2 admin URL, a
separate beacon C2 URL, username, password, a named client, a named
engagement, an explicit written-authorization/ROE statement, and at least one
clear objective — and refuses to build a session prompt or launch `claude` if
any of them is missing. None of these have silent defaults; gather them from
the operator before running this, not after:

```
go build -o bin/harness ./cmd/harness

bin/harness --reaper-url https://c2.example.com:8443 \
  --reaper-c2-url https://c2.example.com:8080 --reaper-username op1 \
  --client "Acme Corp" --engagement "acme-2026-q3" \
  --authorization "Signed SOW #2026-114, ROE dated 2026-09-01 to 2026-09-15" \
  --objective "Obtain domain admin from an external foothold" \
  --objective "Demonstrate access to the finance file share"
```

or via Docker (see `README.md` for the full container workflow):

```
docker compose run --rm harness --reaper-url https://c2.example.com:8443 \
  --reaper-c2-url https://c2.example.com:8080 --reaper-username op1 \
  --client "Acme Corp" --engagement "acme-2026-q3" \
  --authorization "Signed SOW #2026-114, ROE dated 2026-09-01 to 2026-09-15" \
  --objective "Obtain domain admin from an external foothold"
```

It builds the initial prompt referencing all five skills, the authorization
statement, and the objectives, keeps the password out of that prompt/session
file (it's exported only into the launched process's environment as
`$REAPER_PASSWORD`), and hands off to `claude`. See `bin/harness --help` for
all options, including `--dry-run` to review the prompt before it's used, and
`--objectives-file` for longer objective lists.

If you're ever driving one of the five skills directly in a bare Claude Code
session (not launched via this binary — its own required-input gate doesn't
run in that case), the scope gate below still applies: don't proceed until the
operator has stated it in the conversation.

## Standing rules for all five skills

- **Scope gate first.** Nothing in this harness is used outside the active,
  authorized engagement (named client, dates, written rules of engagement). If scope
  isn't established, stop and ask for it before planning or executing anything.
- **Purple-team before report.** Every positive finding must produce
  `purple-team-atomic-tests` artifacts under `atomics/` (YAML + markdown with
  engagement command output) **before** `harness-report` runs. Do not skip this
  for a time-boxed close, a blocked later objective, or "move to reporting."
  Validation prose in the report is not a substitute for those files.
- **Recommend, don't auto-execute.** These skills draft plans, commands, tests, and
  report text for a human operator to review and run/approve — they don't fire
  implant commands or send client deliverables on their own.
- **MITRE ATT&CK v19** (Enterprise: 15 tactics — Defense Evasion is now split into
  **Stealth** `TA0005` and **Defense Impairment** `TA0112`) is the shared taxonomy
  across all five skills so technique IDs stay consistent from plan → execution →
  detection validation → report.
- Treat engagement data (credentials, beacon secrets, target details, findings) as
  sensitive; don't repeat secrets unless the operator explicitly needs them for a
  local command.
