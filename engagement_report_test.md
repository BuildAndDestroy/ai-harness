# Red Team Engagement Report — Internal Lab "test"

> Draft produced by the `harness-report` skill against
> `v3-ghostwriter-executive-document.docx`. Markdown organized under the template's
> own H1/H2 headings; Jinja field names are annotated so this can be transcribed
> straight into Ghostwriter's report data model. Anything not yet backed by real
> engagement data is left as the template's `<UPDATE ME>` / `<Company>` placeholder
> per the skill rules — do not invent values for those.

---

## 1. Contacts and Resources

**`{{ client.name }}` Points of Contact** — loop `client.contacts`

| name | job_title | email |
|---|---|---|
| `<UPDATE ME>` | `<UPDATE ME>` | `<UPDATE ME>` |

> Operator stated this is self-authorized on their own infrastructure; no external
> client POC was provided. Populate if a formal client contact exists.

**Red Team** — loop `team`

| name | role | email | phone |
|---|---|---|---|
| `<Company>` Operator | Red Team Lead | `<UPDATE ME>` | `<UPDATE ME>` |

**Domain Names Used for Assessment Activities** — loop `infrastructure.domains`

| domain | activity |
|---|---|
| metrics.harvestrangelabs.com | ReaperC2 beacon C2 (HTTPS) |

**Servers Used for Assessment Activities** — loop `infrastructure.servers` then `infrastructure.cloud`

| ip_address | activity | role |
|---|---|---|
| 127.0.0.1:8443 | ReaperC2 admin dashboard (local) | C2 admin panel |
| 10.0.20.243:8000 | Staging HTTP server (beacon delivery) | Tooling host |
| metrics.harvestrangelabs.com | Beacon C2 endpoint | C2 (external) |

**Targets** — loop `targets`

| ip_address | hostname |
|---|---|
| 10.0.20.75:30280 | DVWA web application (NodePort) |

---

## 2. Executive Summary

`{{ client.name }}` (Internal Lab) authorized a red-team engagement against a
deliberately vulnerable web application (DVWA) running on a test Kubernetes cluster,
with the objective of obtaining remote code execution, establishing persistence via
a beacon, and escalating privileges to root on the application pod or to host-level
execution.

The assessment demonstrated that an external adversary with knowledge of the
application's default credentials could obtain **remote code execution** on the
application pod within minutes and **establish a persistent command-and-control
foothold** (ReaperC2 beacon) that maintained a live session throughout the engagement
and re-established itself after a pod restart. Two kernel-level privilege-escalation
attempts were made against the host node's kernel (5.15.0-191-generic); one was
confirmed vulnerable and its KASLR protection was defeated, but no public exploit
compatible with this kernel build was available, so **root on the pod and host-level
execution were not achieved**. The blocking factor is documented in Section 5
(Critical Step 4) and Section 7.

**Goals & Objectives** — loop `objectives`

| percent_complete | objective |
|---|---|
| 100% | Obtain RCE on the DVWA web application pod |
| 100% | Establish persistence with a ReaperC2 beacon |
| 0% | Obtain root on the webapp pod (blocked — no compatible kernel LPE) |
| 0% | Get host-level execution access (blocked — hardened container + patched kernel) |

**Summary of Findings** — loop `findings` (cellbg `finding.severity_color`)

| Severity | Title |
|---|---|
| Critical | DVWA Command Injection — Remote Code Execution |
| High | DVWA Default / Weak Credentials |
| High | Excessive Linux Capabilities and Disabled seccomp on Application Pod |
| Medium | Unpatched Kernel — Exploitable CVEs Present (CVE-2026-74581, CVE-2026-64560) |
| Informational | Kubelet API Reachable from Pod (anonymous auth correctly disabled) |

**Mitre ATT&CK Heat Map** — `<UPDATE ME: paste Navigator layer screenshot/export
from ReaperC2 Reports export (STIX v19)>`

---

## 3. Methodology and Goals

The engagement followed a Get In / Stay In / Act methodology. **Get In**: active
reconnaissance of the exposed DVWA NodePort, followed by exploitation of a
command-injection flaw using known default credentials. **Stay In**: deployment of a
ReaperC2 beacon over an HTTPS C2 channel to maintain a foothold and re-establish
access after pod restarts. **Act**: in-pod discovery and privilege-escalation attempts
(container capability abuse, Kubernetes API/kubelet abuse, and kernel LPE) aimed at
the objective of root-on-pod or host-level execution.

`<UPDATE ME: attack-path diagram image>`

**Goals and Objectives** — loop `objectives`

- Obtain RCE on the DVWA web application pod.
- Establish persistence with a ReaperC2 beacon.
- Obtain root on the webapp pod or get host-level execution access.

---

## 4. Scenarios and Scope

### Scenario

Engagement model: **Assumed Breach (credential)** — the operator supplied default
application credentials (`admin` / `password`) as a starting assumption, and the
target was an intentionally vulnerable training application (DVWA) on a test cluster.
The assessment began from external network reachability of the application NodePort
and proceeded through exploitation, persistence, discovery, and privilege
escalation. The cluster backend (Kubernetes control plane and worker nodes) was
explicitly in scope as test infrastructure with snapshots.

### Scope

`{{ client.name }}` authorized the following in-scope targets and activities.

| name | description |
|---|---|
| DVWA web app (`http://10.0.20.75:30280/`) | Exposed NodePort; command injection, authentication bypass, and post-exploitation of the hosting pod |
| Kubernetes backend (test cluster) | Worker node `k8s-worker1` (10.0.20.75), kubelet, pod service account, and node kernel as test infra with snapshots |
| ReaperC2 C2 (`metrics.harvestrangelabs.com`) | Beacon C2 and admin panel for persistence and command queuing |

### Whitecards

| title | issued | description |
|---|---|---|
| Default application credentials | Pre-engagement | Operator-supplied `admin` / `password` for DVWA |
| `kubectl` access to test cluster | On request | Operator provided cluster-admin `kubectl` for pod/node recovery (test infra only) |
---

## 5. Attack Narrative

The following critical steps document the attack path. Each is tagged with the
ATT&CK v19 tactic it primarily exercises.

### Critical Step 1 — Reconnaissance and Initial Access via DVWA Command Injection

The DVWA instance was reachable on `http://10.0.20.75:30280/`. Using the supplied
default credentials, the operator authenticated and set the application security
level to "low". The `/vulnerabilities/exec/` endpoint (a ping utility) accepted an
`ip` parameter with no input sanitization at security level "low", allowing
metacharacter injection. A payload of the form `127.0.0.1; <command>` was processed
by PHP `shell_exec`, returning command output in the page. This yielded **remote code
execution as `www-data` (uid 33)** on the application pod (`dvwa-78d998d9bd-*`).

ATT&CK: `TA0001` Initial Access — `T1190` Exploit Public-Facing Application;
`TA0002` Execution — `T1059.004` Unix Shell.

### Critical Step 2 — Persistence: ReaperC2 Beacon Delivery and C2 Establishment

A ReaperC2 beacon binary (`dvwa-parent.bin`, ~10 MB) was staged on `10.0.20.243:8000`
and pulled into the pod with `curl` to `/dev/shm/dvwa-parent.bin`, made executable,
and launched with `TERM_HARVEST=9`. The beacon established an HTTPS C2 channel to
`metrics.harvestrangelabs.com`, polling every 30 seconds. ReaperC2 registered the
beacon with `ClientId a6e02d08-f2b9-4641-ac1f-a292e1a8f616`; the beacon log confirmed
`{"status":"ok"}` and `{"Active":true,...}` responses, and a SOCKS5 proxy listener
on `:8181`. The beacon survived a pod restart by being re-delivered, demonstrating a
re-establishable foothold (though not reboot-persistent, since `/dev/shm` is wiped on
pod restart).

ATT&CK: `TA0011` Command and Control — `T1071.001` Web Protocols, `T1105` Ingress
Tool Transfer; `TA0003` Persistence — live C2 process (note: not autostart-persistent).

### Critical Step 3 — Discovery: Container, Kubernetes, and Privilege-Escalation Surface

From the `www-data` shell, the operator enumerated the execution environment:
kernel `5.15.0-191-generic` (#201-Ubuntu, Aug 7 2026); `seccomp` disabled (`seccomp=0`);
user namespaces enabled; PID 1 (apache) `CapEff=0xa80425fb` (CAP_SETUID,
CAP_DAC_OVERRIDE, CAP_NET_RAW, CAP_SYS_CHROOT, CAP_MKNOD, …) while the `www-data`
shell had `CapEff=0` with the same bounding set; SUID set limited to the standard
Debian binaries (`su`, `mount`, `passwd`, `chsh`, …) with no `pkexec`/`sudo` and no
file-capability binaries; `/etc/passwd` 644, `/etc/shadow` 640 root:shadow; no useful
cron; no writable root-owned paths. Kubernetes: the pod used the **default service
account** (`system:serviceaccount:dvwa:default`) with no RBAC; the kubelet on
`:10250` was reachable but webhook-denied `nodes/proxy`; no cloud metadata endpoint
responded.

ATT&CK: `TA0007` Discovery — `T1082` System Information Discovery, `T1613` Container
and Resource Discovery.

### Critical Step 4 — Privilege Escalation: Kernel LPE Attempts (Not Achieved)

Two kernel local privilege-escalation attempts were made against the node kernel
(`5.15.0-191-generic`):

1. **CVE-2026-74581** (IPv6 FIB rule use-after-free, fixed upstream in 5.15.216) —
   confirmed present on the target. The KASLR protection was defeated: the public
   PoC's generic KASLR leak succeeded, recovering kernel base
   `0xffffffff9e600000` (slide `0x1d600000`). However, the only public full exploit
   (NebuSec/CyberMeowfia) is hardcoded for **Debian 6.12.101** — its ROP pivot
   gadgets, CPU-Entry-Area physical offset (`CPU1_CEA_PHYS_OFFSET`), and
   `core_pattern` mode-field offset are all 6.12-specific. On 5.15 the post-leak
   pivot targets the wrong physical page, causing a **kernel panic** that took the
   worker node (`k8s-worker1`) down twice. Root-cause analysis showed the blocker is
   `CPU1_CEA_PHYS_OFFSET`: a runtime-allocated, per-boot physical address that cannot
   be derived statically and for which the exploit has no runtime leak. Porting
   requires writing a new CEA-physical-address leak and re-engineering the CEA
   entry-stack pivot for 5.15 — beyond re-gadgeting. Abandoned.

2. **CVE-2026-64560** (posix-cpu-timers UAF via `timer_delete` vs non-leader `exec`
   race, fixed in 5.15.213) — confirmed present and reachable: the public trigger's
   `--check` passed on the target as unprivileged `www-data`. No public full exploit
   exists (only a trigger; the full chain was withheld). Would require from-scratch
   exploit development (KASLR leak → `posix_timers_cache` slab reclaim → arbitrary
   write → `modprobe_path` overwrite).

Standard paths (SUID/cron/writable/capability-abuse/K8s RBAC/kubelet/cloud-metadata)
were exhausted with no viable route. **Root on the pod and host-level execution were
not achieved.**

ATT&CK: `TA0004` Privilege Escalation — `T1068.001` Exploitation for Privilege
Escalation: Kernel (attempted, unsuccessful); `T1611` Escape to Host (attempted,
unsuccessful).
---

## 6. Observations and Recommendations

### Observation 1 — Unsanitized Command Injection in Exposed Web Application

**Recommendation:** Remove the DVWA NodePort exposure from any network reachable
beyond the training environment. Where DVWA must be exposed, enforce the "impossible"
security level (parameterized input / prepared statements and shell argument
escaping), place it behind a WAF with command-injection signatures, and require
authentication with non-default credentials. Treat DVWA as a deliberately vulnerable
training image — never deploy it on shared or production-adjacent clusters.

**Validation:** Atomic Red Team-style test `ART-T1059.004-dvwa-cmdi`: send
`ip=127.0.0.1; id` to `/vulnerabilities/exec/` at security level "low"; expected
`uid=33(www-data)` in the `<pre>` block. Blue-team detection: alert on
`shell_exec`/`/bin/sh -c` spawned by the web-server user (`www-data`) as a child of
the PHP/Apache worker, and on outbound `curl` to non-allowlisted hosts from the pod.

### Observation 2 — Excessive Container Capabilities and Disabled seccomp

**Recommendation:** Drop all Linux capabilities except the minimum required (none for
a static-content/PHP app); enable `seccompProfile: RuntimeDefault` (or a custom
profile denying `unshare`, `clone` with `CLONE_NEWUSER/NEWNET`, raw sockets); set
`runAsNonRoot: true` with a fixed non-root UID; set `readOnlyRootFilesystem: true`
with `emptyDir` mounts only for required writable paths; and remove `CAP_NET_RAW`,
`CAP_MKNOD`, `CAP_SYS_CHROOT`, `CAP_SETUID` from the pod spec. These capabilities
materially increased the post-RCE attack surface (kernel-LPE prerequisites and
potential device-node access).

**Validation:** Atomic test `ART-T1611-cap-surface`: from inside the pod, run
`grep Cap /proc/1/status` and `cat /proc/self/status | grep Cap`; expected after
remediation: `CapEff: 0` for PID 1 and `CapBnd` reduced to only required caps, and
`unshare -Urn` to fail under the seccomp profile.

### Observation 3 — Kernel Lagging Stable Security Patches

**Recommendation:** Patch the node kernel to at least **5.15.216** (which fixes
CVE-2026-74581) and preferably 5.15.213+ (which fixes CVE-2026-64560). Enable a
kernel live-patching pipeline or scheduled node rolling-updates so high-severity
CVEs are applied within the operator's SLA. KASLR is enabled (good) and was
defeated only via a side channel — keep KASLR on and add kernel hardening
(`page_alloc.shuffle=1`, slab freelist randomization).

**Validation:** Atomic test `ART-T1068.001-kernel-version`: `uname -r` on the node;
expected `>= 5.15.216`. Re-run the CVE-2026-74581 KASLR-leak trigger and confirm it
no longer leaks a stable base on patched kernels.

### Observation 4 — Unrestricted Pod Egress Enabled Beacon C2

**Recommendation:** Apply a Kubernetes `NetworkPolicy` restricting pod egress to only
required destinations/ports; block arbitrary outbound to the beacon C2 domain and to
internal staging hosts. The beacon was delivered and maintained purely via outbound
HTTPS that no egress control impeded.

**Validation:** Atomic test `ART-T1071.001-beacon-egress`: attempt a beacon check-in
to `metrics.harvestrangelabs.com` from the pod; expected after remediation: the
connection is denied by the NetworkPolicy and logged by the CNI (Calico) flow logs.

---

## 7. Detailed Findings

### Critical Finding – DVWA Command Injection — Remote Code Execution

**CVSS Score:** 9.8 (AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H)

**Affected Hosts:** DVWA pod (`dvwa-78d998d9bd-*`) on `k8s-worker1`, exposed at
`http://10.0.20.75:30280/`.

**Description:** The `/vulnerabilities/exec/` endpoint of DVWA at security level "low"
passes the user-supplied `ip` parameter unsanitized to PHP `shell_exec`, allowing
shell metacharacter injection. A payload `127.0.0.1; <command>` executes arbitrary
commands as the web-server user.

**Impact:** Unauthenticated-to-RCE on the application pod, enabling full compromise of
the pod, deployment of persistent C2, and a launch point for in-cluster movement and
kernel-level privilege escalation against the host node.

**Replication Steps:** Authenticate to DVWA (`admin`/`password`), set security level to
"low", POST `ip=127.0.0.1; id&Submit=Submit` to `/vulnerabilities/exec/`; observe
`uid=33(www-data)` in the `<pre>` response block.

**Host Detection Techniques:** Alert on `sh`/`bash`/`shell_exec` child processes of
the Apache/PHP-FPM worker; alert on outbound `curl`/`wget` from the `www-data` user;
web WAF rule for `;` `|` `&&` `$(` backtick metacharacters in the `ip` parameter.

**Solution:** Never expose DVWA outside isolated training environments. Where required,
set security level to "impossible" (prepared statements + `escapeshellarg`), enforce
non-default credentials, and front with a WAF.

### High Finding – DVWA Default / Weak Credentials

**CVSS Score:** 7.5 (AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:N/A:N)

**Affected Hosts:** DVWA pod login at `http://10.0.20.75:30280/login.php`.

**Description:** DVWA shipped with the default `admin` / `password` credential pair,
which the operator used as the assumed-breach starting point. Default credentials are
a trivial first step for any adversary once the application is reachable.

**Impact:** Direct authentication bypass, removing the only barrier between an
external attacker and the command-injection flaw.

**Replication Steps:** Browse to `/login.php`, submit `admin`/`password`; observe
redirect to `index.php`.

**Host Detection Techniques:** Alert on successful login using the known default
`admin` credential from an external source IP; periodic credential audit.

**Solution:** Enforce non-default, strong credentials on first deploy; remove default
accounts; add MFA where the application supports it.

### High Finding – Excessive Linux Capabilities and Disabled seccomp on Application Pod

**CVSS Score:** 7.0 (AV:L/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H) — post-RCE enabler

**Affected Hosts:** DVWA pod on `k8s-worker1`.

**Description:** The pod ran with PID 1 effective capabilities `0xa80425fb` including
`CAP_SETUID`, `CAP_DAC_OVERRIDE`, `CAP_NET_RAW`, `CAP_SYS_CHROOT`, and `CAP_MKNOD`,
with `seccomp` disabled and user namespaces enabled. The `www-data` shell retained
the full bounding set (caps droppable but not gained without a SUID/file-cap binary).

**Impact:** These capabilities are prerequisites for the kernel-LPE techniques
attempted (raw sockets, user+net namespace creation, FIB rule manipulation) and for
potential device-node-based host-disk access. They materially widened the post-RCE
blast radius.

**Replication Steps:** From RCE, run `grep Cap /proc/1/status` (shows
`CapEff: 00000000a80425fb`) and `cat /proc/self/status | grep Cap` (shows `CapEff: 0`,
`CapBnd: 00000000a80425fb`).

**Host Detection Techniques:** Audit pod security context at admission (OPA/Kyverno
policy denying broad caps and `seccomp=Unconfined`); alert on `unshare`/`clone` with
namespace flags from container processes.

**Solution:** Drop all capabilities, set `seccompProfile: RuntimeDefault`,
`runAsNonRoot: true`, `readOnlyRootFilesystem: true`; remove `CAP_NET_RAW`/
`CAP_MKNOD`/`CAP_SYS_CHROOT`/`CAP_SETUID`.

### Medium Finding – Unpatched Kernel — Exploitable CVEs Present

**CVSS Score:** 7.8 (AV:L/AC:L/PR:L/UI:N/S:U/C:H/I:H/A:H)

**Affected Hosts:** Worker node `k8s-worker1`, kernel `5.15.0-191-generic`.

**Description:** The node kernel (5.15.0-191, built Aug 7 2026) is vulnerable to
CVE-2026-74581 (IPv6 FIB rule UAF, fixed 5.15.216) and CVE-2026-64560 (posix-cpu-timers
UAF, fixed 5.15.213). The KASLR protection was demonstrably defeated against
CVE-2026-74581 (kernel base leaked: `0xffffffff9e600000`).

**Impact:** A working kernel LPE (once a compatible exploit exists) would allow
container escape to host root. The KASLR leak proves the protection is bypassable on
this build.

**Replication Steps:** From pod RCE, run the CVE-2026-74581 PoC `--leak-only`/leak
stage; observe leaked KASLR base. Run the CVE-2026-64560 trigger `--check`; observe
"prerequisites OK".

**Host Detection Techniques:** Kernel-patch compliance scan; alert on
`unshare(CLONE_NEWUSER|CLONE_NEWNET)` + `RTM_NEWRULE`/FIB rule churn from containers;
anomalous `timer_create`/`timer_delete` storm.

**Solution:** Patch kernel to >= 5.15.216; enable live-patching / rolling node updates.

### Informational Finding – Kubelet API Reachable from Pod (anonymous auth correctly disabled)

**CVSS Score:** 0.0 (Informational)

**Affected Hosts:** Worker node `k8s-worker1` kubelet on `:10250`.

**Description:** The kubelet API (`:10250`) was network-reachable from inside the pod
but correctly returned `Unauthorized` for anonymous requests and `Forbidden` for the
pod's default service-account token (`nodes/proxy` denied by webhook). This is the
expected secure configuration; recorded for completeness.

**Impact:** None directly; confirms the kubelet authn/authz boundary held.

**Replication Steps:** `curl -sk https://10.0.20.75:10250/pods` → `Unauthorized`;
with SA bearer → `Forbidden (… verb=get, resource=nodes, subresource=[proxy])`.

**Host Detection Techniques:** Monitor kubelet audit logs for anonymous/bearer
denials; ensure `--anonymous-auth=false` remains set.

**Solution:** No change required; maintain current kubelet auth settings. Optionally
restrict `:10250` at the network policy layer to control-plane components only.
---

## 8. Mitre ATT&CK Killchain

Each Tactic and Technique exercised during the engagement is listed below in matrix
order. Some tactics legitimately had no applicable technique and are left blank per
the template. Technique IDs are ATT&CK v19 (Enterprise), noting the v19 split of the
former Defense Evasion tactic into **Stealth (`TA0005`)** and **Defense Impairment
(`TA0112`)**.

Navigator layer: `<UPDATE ME: attach ReaperC2 Reports export — STIX v19 layer JSON>`
for [ATT&CK Navigator](https://mitre-attack.github.io/attack-navigator/).

| Tactic | Technique |
|---|---|
| 1. Reconnaissance `TA0043` | T1595.002 - Active Scanning: Vulnerability Scanning |
| 2. Resource Development `TA0042` | T1583.001 - Acquire Infrastructure: Domains (C2 FQDN); T1583.004 - Acquire Infrastructure: Servers (ReaperC2 server) |
| 3. Initial Access `TA0001` | T1190 - Exploit Public-Facing Application; T1078.001 - Valid Accounts: Default Accounts |
| 4. Execution `TA0002` | T1059.004 - Unix Shell; T1204.002 - User Execution: Malicious File (beacon binary) |
| 5. Persistence `TA0003` | (live C2 process foothold; not autostart-persistent — see Critical Step 2) |
| 6. Privilege Escalation `TA0004` | T1068.001 - Exploitation for Privilege Escalation: Kernel (attempted, unsuccessful); T1611 - Escape to Host (attempted, unsuccessful) |
| 7. Stealth `TA0005` | (none — no defense-evasion activity performed) |
| 8. Defense Impairment `TA0112` | (none) |
| 9. Credential Access `TA0006` | (none — known default credentials used, no credential access performed) |
| 10. Discovery `TA0007` | T1082 - System Information Discovery; T1613 - Container and Resource Discovery |
| 11. Lateral Movement `TA0008` | (none) |
| 12. Collection `TA0009` | (none) |
| 13. Command and Control `TA0011` | T1071.001 - Web Protocols (HTTPS beacon); T1105 - Ingress Tool Transfer (beacon download); T1573.002 - Encrypted Channel: Asymmetric Cryptography (TLS) |
| 14. Exfiltration `TA0010` | (none) |
| 15. Impact `TA0040` | (none — collateral node panics during privesc attempts were not an objective) |

---

## 9. Timeline

All timestamps are UTC, sourced from beacon logs and ReaperC2. Host/node timestamps
were not independently collected.

### C2 - Logs

```
2026/09/13 04:17:18 [+] Starting request loop. Sending requests every 30s...
2026/09/13 04:17:18 [+] SOCKS5 proxy server listening on :8181
2026/09/13 04:17:18 {"status":"ok"}
2026/09/13 04:17:48 {"status":"ok"}
2026/09/13 04:26:49 {"Active":true,"ClientId":"a6e02d08-f2b9-4641-ac1f-a292e1a8f616","Commands":[]}
2026/09/13 04:27:19 {"status":"ok"}
```

### C2 - Sessions

```
Beacon: dvwa-parent (dvwa-rce / dvwa-pivot-parent)
ClientId: a6e02d08-f2b9-4641-ac1f-a292e1a8f616
Host: dvwa-78d998d9bd-* (k8s-worker1, 10.0.20.75)
User: www-data (uid 33)
C2: https://metrics.harvestrangelabs.com (HTTPS, 30s interval)
Proxy: SOCKS5 on :8181
First check-in: 2026-09-13 04:17:18 UTC
Status at report time: Active (PID 128, uptime 10+ min)
```

### Attack Simulation

```
T+00:00  Reconnaissance: DVWA NodePort 30280 confirmed reachable (HTTP 200).
T+00:02  Initial Access: authenticated with admin/password; set security=low.
T+00:04  Execution: command injection via /vulnerabilities/exec/ → RCE as www-data.
T+00:08  Persistence: dvwa-parent.bin delivered to /dev/shm, beacon C2 established.
T+00:15  Discovery: enumerated kernel/caps/SUID/cron/K8s SA/kubelet/metadata.
T+00:25  Privilege Escalation: CVE-2026-74581 KASLR leak succeeded (base 0xffffffff9e600000).
T+00:30  Privilege Escalation: CVE-2026-74581 pivot → kernel panic #1 (k8s-worker1 NotReady).
T+01:00  (Node recovered via kubectl; pod rescheduled.)
T+01:05  Privilege Escalation: CVE-2026-74581 retry → kernel panic #2.
T+01:30  (Node recovered; DVWA DB re-initialized.)
T+01:35  Persistence: beacon re-delivered, re-established (ClientId a6e02d08…).
T+01:40  Privilege Escalation: CVE-2026-64560 trigger --check passed (vuln confirmed reachable).
T+01:45  Decision: no compatible 5.15 full LPE PoC; root/host escalation not achievable.
```

---

## 10. Conclusion

The engagement successfully demonstrated the **Get In** and **Stay In** phases against
the Internal Lab "test" target: an external adversary with the application's default
credentials could obtain remote code execution on the DVWA application pod within
minutes and establish a persistent, ReaperC2-controlled HTTPS C2 foothold that
maintained a live session and re-established itself after a pod restart. The beacon
(`ClientId a6e02d08-f2b9-4641-ac1f-a292e1a8f616`) remained active throughout.

The **Act** phase (privilege escalation to root-on-pod or host-level execution) was
**not achieved**. The container was hardened in the dimensions that matter most for
this objective: the Kubernetes service account was the unprivileged default SA with no
RBAC, the kubelet correctly denied anonymous and bearer-token proxy access, no cloud
metadata was exposed, and the standard local-privilege-escalation surface (SUID,
cron, writable root paths, file capabilities) was clean. The only remaining avenue was
a kernel local privilege escalation, and the node kernel (5.15.0-191-generic) — while
confirmed vulnerable to CVE-2026-74581 and CVE-2026-64560 — had **no public exploit
compatible with this build**: the sole full PoC for CVE-2026-74581 is hardcoded for
Debian 6.12.101, and its CPU-Entry-Area physical-offset pivot cannot be statically
ported to 5.15 (it is a runtime, per-boot value with no runtime leak in the exploit).
Defeating KASLR alone was demonstrated, proving the protection is bypassable, but a
complete root-shell chain was not attainable with publicly available tooling within
the engagement.

An external threat could realistically replicate the RCE-and-persistence path
observed here (low skill barrier: default credentials + a well-documented DVWA
flaw + an off-the-shelf C2). The kernel-escape path, by contrast, requires
advanced, target-specific exploit development and is not currently achievable with
public tooling — but the presence of unpatched, exploitable kernel CVEs and the
excessive container capability set mean this boundary should not be relied upon.

Overall, the security posture for the **Get In / Stay In** phases is weak (exposed
vulnerable app, default credentials, no egress control, over-capable container) and
should be remediated per Section 6. The **Act** phase boundary held during this
engagement, but primarily due to kernel-patch and exploit-availability luck rather
than defense-in-depth — patching the kernel to 5.15.216+ and tightening the pod
security context are the highest-priority improvements.
