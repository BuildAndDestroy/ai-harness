---
name: purple-team-atomic-tests
description: Purple team agent that turns a red team's positive findings (techniques that succeeded against the target — i.e. detection/control gaps) into Atomic Red Team-style atomic tests the blue team can safely re-run in a controlled environment to validate detections. Includes the engagement's actual command and ReaperC2 command output as evidence. Use as soon as a technique succeeds — required before harness-report, not optional, and not only at engagement close. Also use when asked to build/expand an atomic test library.
---

# Purple team — atomic test generation

You turn a red team **positive finding** into one or more **atomic tests**: small,
safe, repeatable procedures that reproduce the same ATT&CK technique in a controlled
environment so the blue team can validate whether their detections actually fire.
This is purple teaming, not more offense — the deliverable is a test the defenders
run against their own instrumented environment, not another production engagement.

Follow the project's `CLAUDE.md` standing rules. In particular: **atomic tests run
against a controlled/lab environment the blue team owns, never back against the
original client production target** without a separate, explicit authorization for
that. Every test needs a cleanup step — leave nothing behind.

## When to run (mandatory)

This skill is **required for every positive finding**, not a close-out extra.

- Run it **as soon as the technique succeeds** — do not wait until objectives are
  done, until LPE is exhausted, or until someone says "write the report."
- Write the files under `atomics/` **before** `harness-report` starts. Validation
  sentences in a report are not this skill's deliverable.
- A time-boxed close, a blocked later objective, or "move to reporting" does
  **not** skip this. If `harness-report` is about to run and `atomics/<id>/`
  is missing for a successful technique, run this skill first.
- Do not mark this skill complete unless those YAML + markdown files exist.

## Input: what makes a finding "positive"

A finding is in scope for this skill when the red team confirmed a technique
**succeeded** and there's something concrete to validate detection against:

- technique ID
- the procedure actually used (commands/API calls/artifacts)
- **the command output from the engagement** — pull it from ReaperC2
  (`GET /api/beacon-command-output`, Commands output history, or the Ghostwriter
  oplog CSV `output` column). Do not invent stdout/stderr.
- the platform
- what was or wasn't observed by defenses at the time (EDR alert, SIEM log, nothing)

If any of that is missing, ask for it rather than inventing detection outcomes or
fabricating command results — the report's Validation section only carries weight if
it's honest about what was tested.

## Output: the atomic test

Use the [Atomic Red Team](https://github.com/redcanaryco/atomic-red-team) YAML
schema — it's the de facto format blue teams already have tooling for (Invoke-
AtomicRedTeam, etc.), and it keys directly on ATT&CK technique/sub-technique IDs, so
it slots straight into the same v19 taxonomy the `red-team-operator` skill uses. Full
schema and a worked example are in `references/atomic-test-format.md`. Every test
must have:

- `attack_technique` — the exact `Txxxx` or `Txxxx.xxx` from the finding (don't
  round up to the parent technique if a sub-technique was actually observed)
- a minimal, platform-appropriate `executor.command` that reproduces the *observed
  behavior*, not a random public PoC — fidelity to what the red team actually did is
  the point
- `executor.cleanup_command` whenever the test creates a file, registry key,
  scheduled task, account, process, or network artifact
- `input_arguments` for anything environment-specific (paths, usernames, hosts) so
  it's portable across the blue team's lab hosts
- `dependencies`/`dependency_executor_name` if the test needs setup (a tool present,
  a file staged) before it can run

The companion `atomics/Txxxx[.xxx]/Txxxx[.xxx].md` **must** include an **Engagement
evidence** section with the exact command(s) that succeeded and the command output
captured during the engagement (secrets redacted). That output is why the finding
is a detection gap — the blue team needs to see what a successful run looked like.

## Tie the test back to the finding

For each finding, also produce the three things the engagement report expects
(these map directly onto the "Observations and Recommendations" section of the
Ghostwriter executive template — see `harness-report`):

1. **Observation** — what happened (one paragraph, non-technical enough for an
   executive summary, technical enough to be unambiguous). Cite the engagement
   command and a sanitized snippet of its output so the finding is grounded in
   what ReaperC2 actually recorded.
2. **Recommendation** — the control or detection change that would have caught or
   stopped it.
3. **Validation** — the atomic test itself (or a short description of it plus where
   the full YAML lives), stated as "re-run this test after implementing the
   recommendation to confirm detection now fires." Point at the Engagement
   evidence output as the expected successful-run signature.

## Where tests live

Default to an `atomics/` directory laid out the same way as the upstream Atomic Red
Team repo: `atomics/T1059.001/T1059.001.md` (human-readable) next to
`atomics/T1059.001/T1059.001.yaml` (machine-readable), one directory per technique.
If this project doesn't have an `atomics/` directory yet, create it there rather
than inventing a different layout. Creating those files is the skill's job, not
an optional extra the operator has to request.

## Safety checklist before handing a test off

- Cleanup command present and actually reverses the test's changes.
- No hardcoded client-identifying hostnames/IPs/credentials from the real engagement
  — parameterize via `input_arguments` with lab-safe defaults.
- `elevation_required` set accurately (don't silently require admin/root without
  saying so).
- Test is scoped to *detection validation*, not to re-demonstrating impact (e.g. a
  ransomware-simulation finding gets a test that stages the technique's precursor
  behavior, not one that actually encrypts anything).
