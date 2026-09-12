# Atomic Red Team YAML schema (reference)

One file per technique, keyed by technique ID. Top level:

```yaml
attack_technique: T1059.001
display_name: PowerShell
atomic_tests:
  - <atomic test>
  - <atomic test>
```

## Atomic test object

| Field | Required | Notes |
|---|---|---|
| `name` | yes | Short, specific — describes the exact behavior, not just the technique name |
| `auto_generated_guid` | recommended | Stable UUID for tooling that tracks test identity |
| `description` | yes | 1-3 sentences: what it does and why it maps to this finding |
| `supported_platforms` | yes | `windows`, `linux`, `macos` (list) |
| `input_arguments` | if parameterized | map of `name: {description, type, default}` |
| `dependency_executor_name` | if setup needed | e.g. `powershell`, `sh` |
| `dependencies` | if setup needed | list of `{description, prereq_command, get_prereq_command}` |
| `executor` | yes | see below |

## Executor object

```yaml
executor:
  name: sh                 # or powershell, command_prompt, bash, manual
  elevation_required: false
  command: |
    <the actual reproduction — use #{input_argument_name} placeholders>
  cleanup_command: |
    <reverses every change the command made>
```

## Worked example

Finding: red team ran a scheduled cron job for Linux persistence
(`T1053.003 — Scheduled Task/Job: Cron`); host EDR did not alert on the crontab
write.

```yaml
attack_technique: T1053.003
display_name: "Scheduled Task/Job: Cron"
atomic_tests:
  - name: "Persistence via user crontab entry"
    auto_generated_guid: 6e6f2f2a-2f38-4a2c-9b1a-9a2f6c9b5d10
    description: >
      Reproduces the crontab-based persistence observed during the engagement:
      an attacker-controlled command line is added to the current user's crontab
      to re-run on a schedule. Validates whether crontab writes or the resulting
      scheduled execution are alerted on.
    supported_platforms:
      - linux
    input_arguments:
      command_to_run:
        description: Command the "implant" re-executes on schedule (lab-safe stand-in)
        type: string
        default: "/bin/echo atomic-T1053.003-marker >> /tmp/atomic-T1053.003.log"
      schedule:
        description: Cron schedule
        type: string
        default: "* * * * *"
    executor:
      name: sh
      elevation_required: false
      command: |
        (crontab -l 2>/dev/null; echo "#{schedule} #{command_to_run}") | crontab -
      cleanup_command: |
        crontab -l 2>/dev/null | grep -v "atomic-T1053.003" | crontab -
        rm -f /tmp/atomic-T1053.003.log
```

## Naming/location convention

```
atomics/
  T1053.003/
    T1053.003.md     # human-readable: same content, prose form, for the report appendix
    T1053.003.yaml    # machine-readable, loadable by Invoke-AtomicRedTeam-style runners
```

If a finding doesn't map to an existing upstream Atomic Red Team test (many don't —
this is normal for bespoke procedures), write the custom test in this same schema
rather than forcing the finding into an unrelated upstream test just to reuse it.
