# MITRE ATT&CK v19 (Enterprise) reference

Released 2026-04-28. Latest agile update as of this writing: v19.2 (2026-08). Scale:
15 tactics, 222 techniques, 475 sub-techniques, ~178 groups, ~949 software entries,
~59 campaigns (counts drift slightly release to release — treat as approximate).

## The 15 Enterprise tactics, in matrix (kill-chain) order

| # | Tactic | ID |
|---|---|---|
| 1 | Reconnaissance | TA0043 |
| 2 | Resource Development | TA0042 |
| 3 | Initial Access | TA0001 |
| 4 | Execution | TA0002 |
| 5 | Persistence | TA0003 |
| 6 | Privilege Escalation | TA0004 |
| 7 | **Stealth** | TA0005 |
| 8 | **Defense Impairment** | TA0112 |
| 9 | Credential Access | TA0006 |
| 10 | Discovery | TA0007 |
| 11 | Lateral Movement | TA0008 |
| 12 | Collection | TA0009 |
| 13 | Command and Control | TA0011 |
| 14 | Exfiltration | TA0010 |
| 15 | Impact | TA0040 |

## What changed in v19 (the part worth remembering)

- **Defense Evasion is retired as a tactic.** It's split into:
  - **Stealth (`TA0005`, inherits the old ID)** — techniques where the adversary
    blends into legitimate behavior: obfuscated files/info, execution guardrails,
    process injection, indicator removal, masquerading.
  - **Defense Impairment (`TA0112`, new)** — techniques where the adversary actively
    disables or degrades a control: impair defenses (killing EDR/AV), disabling
    logging, modifying firewall rules, disabling cloud logging.
  - When tagging old reports/techniques written under pre-v19 "Defense Evasion",
    re-classify into whichever of the two actually matches the behavior — don't just
    relabel everything as Stealth.
- ICS ATT&CK gained sub-techniques for the first time.
- Mobile ATT&CK gained the beginnings of Detection Strategies.
- New AI-enabled and social-engineering technique coverage was added (notably
  including a campaign entry for AI-orchestrated operations and a software entry for
  malware that queries an LLM at runtime) — worth checking if the engagement's threat
  model involves AI-assisted tradecraft.

## Sub-techniques and IDs

Most techniques have sub-techniques (`Txxxx.xxx`). Always prefer the most specific
sub-technique that matches observed/planned behavior over the parent technique alone
— it's what the report's "Detailed Findings" and ATT&CK killchain table expect (see
the `harness-report` skill), and what ReaperC2's Notes & ATT&CK technique tags and
Navigator layer export key on.

## Looking up exact IDs

Don't guess technique/sub-technique IDs from memory beyond common ones — verify
against:
- [attack.mitre.org](https://attack.mitre.org) (technique pages, matrix browser)
- [ATT&CK Navigator](https://mitre-attack.github.io/attack-navigator/) — also where
  Navigator layer JSON (from ReaperC2's Reports/Notes & ATT&CK export, STIX version
  set to v19) gets visualized
- The active ReaperC2 engagement's **Notes & ATT&CK** page, which drives its own
  tactic/technique picker off a selectable matrix (STIX) version — set it to v19 to
  match this reference

## Navigator layer JSON shape (for hand-off to reporting/exports)

A layer is JSON with a `techniques` array of `{ techniqueID, tactic, score, color,
comment, enabled }` objects plus `versions.attack` (matrix version, e.g. `"19"`).
ReaperC2 builds this from tagged techniques (green `#74c476` highlight, per-technique
comment) — that's the same JSON that becomes the report's "Mitre ATT&CK Killchain"
section link/attachment.
