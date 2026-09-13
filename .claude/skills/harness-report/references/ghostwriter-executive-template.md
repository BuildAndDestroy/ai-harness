# Ghostwriter executive template — structure & field reference

Schema for Ghostwriter's executive `docxtpl` report. The `.docx` is not in this
repo; operators keep it in Ghostwriter (or a local copy). `{{ var }}` is plain
text, `{{p var }}` is a rich-text subdoc (paragraphed prose/lists, not a single
line), `{%tr for x in y %} ... {%tr endfor %}` repeats a table row,
`{% cellbg finding.severity_color %}` colors a cell by severity. Section order
below is the order in the document — keep it.

## 1. Contacts and Resources (Heading 1)

- `{{ client.name }} Points of Contact` table — loop `client.contacts`: `name`,
  `job_title`, `email`
- `Red Team` table — loop `team`: `name`, `role`, `email`, `phone`
- `Domain Names Used for Assessment Activities` table — loop
  `infrastructure.domains`: `domain`, `activity`
- `Servers Used for Assessment Activities` table — loop `infrastructure.servers`
  (`ip_address`, `activity`, `role`) then loop `infrastructure.cloud` (same fields)
- `Targets` table — loop `targets`: `ip_address`, `hostname`

## 2. Executive Summary (Heading 1)

- Narrative paragraph referencing `{{ client.name }}` and the engagement.
- `Goals & Objectives` table — loop `objectives`: `percent_complete`, `objective`
- Prose intro to observations, then a plain bullet summary (written prose, not a
  templated loop in this template)
- `Summary of Findings` table — loop `findings`: cell background
  `{% cellbg finding.severity_color %}` + `finding.severity`, and `finding.title`
- `Mitre ATT&CK Heat Map` — image (Navigator layer screenshot/export), pasted by the
  operator, not generated as text

## 3. Methodology and Goals (Heading 1)

- Standard methodology paragraph (Get In / Stay In / Act phases; OSINT → enumeration
  → exploitation → attack). Customize the environment name, not the phase structure.
- Placeholder for an attack-path diagram image.
- `Goals and Objectives` list — loop `objectives`: `objective`

## 4. Scenarios and Scope (Heading 1)

- **Scenario** (Heading 2) — prose: engagement model (Full Engagement / Assumed
  Breach / Custom Breach), starting assumptions, phase-by-phase high-level summary.
- **Scope** (Heading 2) — intro referencing `{{ client.name }}`, then a table looping
  `scope`: `name`, `{{p description_rt }}`
- **Whitecards** (Heading 2) — intro prose, then a table looping `whitecards`:
  `title`, `issued`, `description`

## 5. Attack Narrative (Heading 1)

- Intro prose distinguishing Critical Steps from Observations.
- One Heading 2 subsection per critical step (`Critical Step 1`, `Critical Step 2`,
  …) — free text plus a "glyph"/icon placeholder. Not templated/looped in this
  template; each step is authored directly. Tag the ATT&CK tactic used, per
  `red-team-operator`'s log.

## 6. Observations and Recommendations (Heading 1)

- Intro prose.
- Per observation (not looped — repeat the block per finding): `Observation N`, then
  nested `Recommendation`, then nested `Validation`. This is exactly where
  `purple-team-atomic-tests` output lands: Validation = the atomic test (or a
  pointer to it).

## 7. Detailed Findings (Heading 1)

- Intro prose.
- Four (really five) near-identical looped blocks, one per severity, each filtering
  `findings` by `finding.severity == 'Critical' | 'High' | 'Medium' | 'Low' |
  'Informational'`. Per finding:
  - Heading: `<Severity> Finding – {{ finding.title }}`
  - `CVSS Score: {{ finding.cvss_score }}`
  - `Affected Hosts: {{p finding.affected_entities_rt }}`
  - `Description: {{p finding.description_rt }}`
  - `Impact: {{p finding.impact_rt }}`
  - `Replication Steps: {{p finding.replication_steps_rt }}`
  - `Host Detection Techniques: {{p finding.host_detection_techniques_rt }}`
  - `Solution: {{p finding.recommendation_rt }}`

  Only render a severity's block if at least one finding has that severity.

## 8. Mitre ATT&CK Killchain (Heading 1)

- Intro prose ("each Tactic and Technique outlined below... some Tactics may not
  apply").
- Link to [ATT&CK Navigator](https://mitre-attack.github.io/attack-navigator/) +
  placeholder for the Navigator JSON layer link/file (from `reaperc2-operator`'s
  Reports export, STIX version v19).
- `Tactic` / `Technique` table, one row group per tactic in matrix order (see
  `red-team-operator`'s `attack-v19-tactics.md` for the current 15, including the
  v19 Stealth/Defense Impairment split), technique rows formatted
  `Txxxx[.xxx] - Name[: Sub-technique name]`. Leave a tactic's technique cell blank
  if unused rather than omitting the tactic row.

## 9. Timeline (Heading 1)

- Intro noting all timestamps are UTC, sourced from C2 logs, host logs, Ghostwriter,
  Vectr.
- **C2 - Logs** (Heading 2) — codeblock of raw logs (paste from `reaperc2-operator`
  exports)
- **C2 - Sessions** (Heading 2) — codeblock of session data
- **Attack Simulation** (Heading 2) — timestamped simulation output, if not already
  covered by C2 logs

## 10. Conclusion (Heading 1)

- Synthesis prose: what was found, whether an external threat could replicate the
  path, skill level required, tooling used (publicly available vs. privately built),
  and overall assessment of security posture improvement. Write this last, after
  every other section is populated — it should not introduce any fact not already
  stated earlier in the report.

## Placeholder conventions to preserve when data is missing

- `<Company>` — the red team's own org name (not the client) — leave as-is unless
  told otherwise.
- `<UPDATE ME ...>` — literal instruction text in the template for the human
  operator (diagrams, screenshots, narrative specifics). Keep these visible in the
  draft rather than deleting them, so the operator knows what's still open.
