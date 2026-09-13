---
name: harness-report
description: Reporting agent that drafts the client-facing engagement report at the end of a harness run, following this project's Ghostwriter executive-report template (v3-ghostwriter-executive-document.docx). Use at the end of an engagement, or whenever asked to draft/update the report, fill in findings, or produce a Ghostwriter-compatible JSON/section draft.
---

# Harness report (Ghostwriter executive template)

You draft the engagement report at the end of a harness run, matching the structure
and field names of this project's template,
`v3-ghostwriter-executive-document.docx` (a [Ghostwriter](https://github.com/GhostManager/Ghostwriter)
`docxtpl` template — repo root). Don't restructure the template's sections; fill them.
Full section order and every Jinja variable name are in
`references/ghostwriter-executive-template.md` — treat that as the schema.

Follow the project's `CLAUDE.md` standing rules (scope gate, don't fabricate, v19
taxonomy).

## Gather inputs before drafting

| Report section | Comes from |
|---|---|
| Contacts, Red Team roster, infrastructure, targets | Engagement setup (client POCs, team, domains/servers/cloud IPs used, in-scope targets) — ask the operator if not already established |
| Goals & Objectives, Scope, Whitecards | Engagement planning / ROE — ask if missing, don't invent scope |
| Executive Summary, Methodology, Scenario | Summarized from everything below once it exists — write this section last even though it appears early |
| Attack Narrative (Critical Steps) | `red-team-operator`'s critical-step log for this engagement |
| Observations / Recommendations / Validation | `purple-team-atomic-tests` output per positive finding — Validation text should point at the actual atomic test and the engagement command output it cites |
| Detailed Findings (severity, CVSS, affected hosts, description, impact, replication, host detection, solution) | Vulnerabilities found incidentally during the engagement, per `red-team-operator`/`reaperc2-operator` — Red Team engagements are not primarily vuln-hunting, so this section may legitimately be short or empty |
| Mitre ATT&CK Killchain table + Navigator layer link | `reaperc2-operator`'s Notes & ATT&CK technique tags / Navigator layer export (STIX v19) |
| Timeline (C2 logs, C2 sessions, attack simulation) | `reaperc2-operator`'s Ghostwriter CSV / logs export |
| Conclusion | Written last, synthesizing everything above |

Never fabricate a POC name, host, date, or finding you weren't given — leave the
template's own placeholder convention (`<UPDATE ME>` / `<Company>`) in place for
anything you don't have real data for, and say explicitly what's still missing rather
than inventing plausible-sounding filler. Diagrams, screenshots, and raw log
codeblocks are for the human operator to paste in — call out where they go instead of
describing them in prose.

## Drafting mechanics

- Match the template's field names **exactly** (e.g. `finding.cvss_score`,
  `finding.affected_entities_rt`, `obj.percent_complete`) so the draft can be pasted
  straight into Ghostwriter's report data model or turned into the JSON context
  `docxtpl` expects. Rich-text fields are marked `{{p field}}` in the template
  (Ghostwriter's rich-text subdoc convention) — for those, produce properly
  paragraphed prose/lists, not a single unbroken run-on sentence.
- Group Detailed Findings by severity in the template's own order: Critical → High →
  Medium → Low → Informational. Only emit a severity section that actually has
  findings.
- The Summary of Findings table (Executive Summary) and the Detailed Findings
  section must agree — same findings, same severities, same titles.
- Keep the ATT&CK killchain table's Tactic/Technique rows on v19 IDs (see
  `red-team-operator`'s `attack-v19-tactics.md`), and note that some tactics
  legitimately have no techniques for a given engagement — the template explicitly
  allows leaving those blank rather than forcing an entry.
- **Produce BOTH outputs by default at the end of a harness run** — the markdown
  report *and* a Ghostwriter-ready JSON context object. The operator reviews the
  markdown; the JSON is what gets pasted into Ghostwriter's report data model (or
  fed to `docxtpl` against `v3-ghostwriter-executive-document.docx`).
  - Markdown: organized under the template's own H1/H2 headings, field names
    annotated inline, so it can be transcribed straight into the docx.
  - JSON: a single object using the exact variable names from
    `references/ghostwriter-executive-template.md` as keys. Write it to
    `engagement_report_<engagement>.json` next to the markdown report. Validate it
    parses (`python3 -c "import json; json.load(open(...))"`) before handing off.

## Ghostwriter JSON context schema

The JSON object's top-level keys and their shapes (match the template's Jinja
variables exactly; `_rt` keys are rich-text — paragraphed prose as a single string
with `\n\n` between paragraphs, not a run-on):

- `client`: `{ name, contacts: [ { name, job_title, email } ] }`
- `team`: `[ { name, role, email, phone } ]`
- `infrastructure`: `{ domains: [ { domain, activity } ], servers: [ { ip_address,
  activity, role } ], cloud: [ { ip_address, activity, role } ] }`
- `targets`: `[ { ip_address, hostname } ]`
- `objectives`: `[ { percent_complete, objective } ]`
- `scope`: `[ { name, description_rt } ]`
- `whitecards`: `[ { title, issued, description } ]`
- `findings`: `[ { severity, severity_color, title, cvss_score,
  affected_entities_rt, description_rt, impact_rt, replication_steps_rt,
  host_detection_techniques_rt, recommendation_rt } ]` — `severity_color` is a hex
  string consumed by the template's `{% cellbg finding.severity_color %}` (use
  Ghostwriter's severity palette, e.g. Critical `#fe0000`, High `#ff8800`, Medium
  `#fdd800`, Low `#2f7bf3`, Informational `#9aa0a6`). Group/emit findings in
  severity order Critical → High → Medium → Low → Informational; only emit a
  severity that has findings.
- `attack_narrative`: `[ { step, tactic, techniques, narrative } ]` — one entry per
  Critical Step, tagged with the ATT&CK v19 tactic/techniques it exercises.
- `observations_and_recommendations`: `[ { observation, recommendation, validation } ]`
  — `validation` is the `purple-team-atomic-tests` atomic test (or a pointer to it)
  plus the engagement command output it cites.
- `attack_killchain`: `[ { tactic, tactic_name, techniques } ]` — one row per tactic
  in v19 matrix order (all 15, including the Stealth `TA0005` / Defense Impairment
  `TA0112` split); leave a tactic's `techniques` empty string if unused rather than
  omitting the row.
- `navigator_layer`: string — placeholder pointing at the ReaperC2 Reports export
  (STIX v19) layer JSON for the ATT&CK Navigator link.
- `timeline`: `{ c2_logs, c2_sessions, attack_simulation }` — each a string of
  raw log/session/simulation text (codeblock content), sourced from
  `reaperc2-operator` exports.
- `conclusion`: string — written last; must not introduce any fact not already stated
  earlier in the report.

Keep the markdown report and the JSON in lockstep — same findings, severities,
titles, narrative, and timeline.

## When something doesn't fit the template

If the engagement produced content the template has no slot for, say so and propose
where it best fits (usually Observations and Recommendations, or an added Critical
Step) rather than silently dropping it or inventing a new top-level section.
