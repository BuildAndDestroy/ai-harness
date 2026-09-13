---
name: harness-report
description: Reporting agent that drafts the client-facing engagement report at the end of a harness run, following this project's Ghostwriter executive-report template (v3-ghostwriter-executive-document.docx). Use at the end of an engagement AFTER purple-team-atomic-tests has written atomics/ YAML+md for every positive finding. If those files are missing, run that skill first instead of drafting. Also use whenever asked to draft/update the report, fill in findings, or produce a Ghostwriter-compatible JSON/section draft.
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

## Skill-4 gate — stop if atomic tests are missing

`purple-team-atomic-tests` is a required prior step, not optional reporting
color. Before you draft (markdown or JSON):

1. List the engagement's **positive findings** (techniques that succeeded).
2. For each, confirm `atomics/<technique-id>/<technique-id>.yaml` and the
   matching `.md` exist, with an Engagement evidence section citing real
   ReaperC2 command output.
3. If any are missing, **stop**. Run `purple-team-atomic-tests` for those
   findings, then come back. Do not draft Validation by inventing
   `ART-Txxxx-…` names or paraphrasing a test that was never written.

A time-boxed close, a blocked later objective, or an operator "move to
reporting" instruction does **not** waive this. If there were genuinely no
positive findings (no technique succeeded), say so and leave Observations /
Validation empty rather than fabricating tests.

## Gather inputs before drafting

| Report section | Comes from |
|---|---|
| Contacts, Red Team roster, infrastructure, targets | Engagement setup (client POCs, team, domains/servers/cloud IPs used, in-scope targets) — ask the operator if not already established |
| Goals & Objectives, Scope, Whitecards | Engagement planning / ROE — ask if missing, don't invent scope |
| Executive Summary, Methodology, Scenario | Summarized from everything below once it exists — write this section last even though it appears early |
| Attack Narrative (Critical Steps) | `red-team-operator`'s critical-step log for this engagement |
| Observations / Recommendations / Validation | `purple-team-atomic-tests` output per positive finding — Validation text must point at the real `atomics/<id>/<id>.yaml` (and its Engagement evidence command output). Invented `ART-Txxxx-…` labels are not an atomic test |
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
- Default output format: the report as markdown, organized under the template's own
  H1/H2 headings, so the operator can review it before it's transcribed into
  Ghostwriter or the docx. If asked for a Ghostwriter-ready JSON context object
  instead, use the exact variable names from the reference doc as keys.

## When something doesn't fit the template

If the engagement produced content the template has no slot for, say so and propose
where it best fits (usually Observations and Recommendations, or an added Critical
Step) rather than silently dropping it or inventing a new top-level section.
