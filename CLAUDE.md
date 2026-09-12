# ai-harness

Skill set for running an authorized red team engagement end-to-end and producing the
client deliverable. Five project skills, meant to be used in sequence during one
engagement ("harness run"), launched via the `harness` Go binary (`cmd/harness`):

1. **`red-team-operator`** — plans and directs adversary emulation, mapping every
   phase to MITRE ATT&CK v19 Enterprise tactics/techniques. Produces the technique
   log (critical steps, TTPs used) that later stages consume.
2. **`reaperc2-operator`** — drives [ReaperC2](https://github.com/BuildAndDestroy/ReaperC2)
   (beacon generation, command queuing, topology, Notes & ATT&CK tagging, exports)
   for whatever the red-team-operator skill decides to do.
3. **`exploit-development`** — turns a confirmed, in-scope vulnerability into a
   working proof-of-concept, mapped to ATT&CK v19. Every exploit artifact it
   produces carries a mandatory educational/authorized-testing-only notice.
4. **`purple-team-atomic-tests`** — for each positive finding (a technique that
   succeeded against the target, i.e. a detection/control gap), generates Atomic Red
   Team-style validation tests the blue team can re-run in a controlled environment.
   Its output fills the report's Validation fields.
5. **`harness-report`** — at the end of the engagement, drafts the client-facing
   report against `v3-ghostwriter-executive-document.docx`, using the technique log,
   ReaperC2 exports, exploit findings, and atomic-test validations gathered above.

## Starting an engagement

Run the `harness` binary rather than starting a bare Claude Code session by
hand — it requires a ReaperC2 URL, username, password, and at least one clear
objective, and refuses to proceed without them:

```
go build -o bin/harness ./cmd/harness

bin/harness --reaper-url https://c2.example.com:8443 --reaper-username op1 \
  --client "Acme Corp" --engagement "acme-2026-q3" \
  --objective "Obtain domain admin from an external foothold" \
  --objective "Demonstrate access to the finance file share"
```

or via Docker (see `README.md` for the full container workflow):

```
docker compose run --rm harness --reaper-url https://c2.example.com:8443 \
  --reaper-username op1 --objective "Obtain domain admin from an external foothold"
```

It builds the initial prompt referencing all five skills and the objectives, keeps
the password out of that prompt/session file (it's exported only into the launched
process's environment as `$REAPER_PASSWORD`), and hands off to `claude`. See
`bin/harness --help` for all options, including `--dry-run` to review the prompt
before it's used, and `--objectives-file` for longer objective lists.

## Standing rules for all five skills

- **Scope gate first.** Nothing in this harness is used outside the active,
  authorized engagement (named client, dates, written rules of engagement). If scope
  isn't established, stop and ask for it before planning or executing anything.
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
