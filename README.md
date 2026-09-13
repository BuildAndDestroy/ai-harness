# ai-harness

Claude Code skills and a launcher for running an authorized red team engagement
end-to-end against [ReaperC2](https://github.com/BuildAndDestroy/ReaperC2) and
producing the client deliverable. See `CLAUDE.md` for the full skill roster and
standing rules; this file covers building and running the `harness` launcher.

## What's here

- `.claude/skills/` — five Claude Code skills, used in sequence:
  `red-team-operator`, `reaperc2-operator`, `exploit-development`,
  `purple-team-atomic-tests`, `harness-report`. `purple-team-atomic-tests` is
  required for every positive finding (real files under `atomics/`) and must
  finish before `harness-report` drafts; a time-boxed close does not skip it.
- `cmd/harness`, `internal/harness` — the Go launcher that validates engagement
  inputs, builds the initial session prompt, and starts `claude`.
- `v3-ghostwriter-executive-document.docx` — the Ghostwriter report template
  `harness-report` drafts against.
- `Dockerfile`, `docker-compose.yml` — containerized build/run, multi-arch
  (`linux/amd64`, `linux/arm64`).

## Engagement intake template

Gather this from the operator/client before running a build — nothing here has
a silent default, and the launcher refuses to start without the required
fields. Fill in a copy of this and hand it back:

```
Client:                 <named client, or "Internal Lab" if self-authorized>
Engagement name:        <short slug, e.g. acme-2026-q3>
Authorization / ROE:    <SOW #, ROE dates, or an explicit self-authorization
                         statement that you own/control every in-scope system>
Objective(s):           <one or more concrete goals>

ReaperC2 admin dashboard URL:   <e.g. https://127.0.0.1:8443/login>
ReaperC2 username:              <e.g. aiuser>
ReaperC2 password:              <supply via --reaper-password prompt or
                                 REAPER_PASSWORD env var — do not paste it into
                                 chat, a session file, or a committed file>
ReaperC2 beacon C2 FQDN/URL:    <e.g. https://metrics.example.com — must differ
                                 from the admin dashboard URL above>

Target(s) in scope:
  - Target webapp URL:          <e.g. http://10.0.20.75:30280/>
  - Target app credentials:     <e.g. admin / password, if authorized to use>
  - (any other in-scope hosts/creds)
```

Maps to the flags below as: Client → `--client`, Engagement name →
`--engagement`, Authorization/ROE → `--authorization`, ReaperC2 admin dashboard
→ `--reaper-url`, ReaperC2 beacon C2 FQDN → `--reaper-c2-url`, ReaperC2
username → `--reaper-username`, ReaperC2 password → `--reaper-password` /
`REAPER_PASSWORD` (never a bare CLI arg in a shared shell). There's no
dedicated flag for target webapp URL/credentials — fold them into
`--objective` text (e.g. `--objective "Obtain password or flag.txt from
http://10.0.20.75:30280/, credentials admin:password"`) so the session has
that context; the `red-team-operator` skill treats it as part of the
authorized scope, not a separate secret channel.

## Running locally (Go toolchain)

Requires Go 1.26+ and the `claude` CLI on `PATH`.

```
go build -o bin/harness ./cmd/harness

./bin/harness \
  --reaper-url https://c2.example.com:8443 \
  --reaper-c2-url https://c2.example.com:8080 \
  --reaper-username op1 \
  --client "Acme Corp" \
  --engagement "acme-2026-q3" \
  --authorization "Signed SOW #2026-114, ROE dated 2026-09-01 to 2026-09-15" \
  --objective "Obtain domain admin from an external foothold" \
  --objective "Demonstrate access to the finance file share"
```

`--reaper-url` is the operator dashboard; `--reaper-c2-url` is the beacon
listener implants phone home to — they must differ. `--engagement` scopes the
run to that ReaperC2 workspace and is required.

`--client`, `--authorization`, and `--objective` are a scope-gate killswitch:
the binary refuses to build a session prompt or launch `claude` unless a named
client, a written authorization/ROE statement (a SOW reference, or an explicit
self-authorization statement for a lab you own), and at least one objective are
all given explicitly — none of them default to a placeholder. Gather them
before running this, not after.

You'll be prompted for the ReaperC2 password (hidden input) unless you pass
`--reaper-password` or set `REAPER_PASSWORD`. The password is never written to
disk — it lives only in the launcher's process environment and is exported into
`claude`'s environment as `$REAPER_PASSWORD`. The rendered session prompt (client,
engagement, authorization, objectives, ReaperC2 admin/C2 URLs and username — no
secret) is saved to `sessions/<engagement>.md` for reference.

Use `--dry-run` to see the prompt without launching `claude`, and
`harness --help` for the full flag list (including `--objectives-file` for a
longer objectives list and `--sessions-dir` to change where prompts are saved).

## Running with Docker

```
cp .env.example .env   # fill in REAPER_URL / REAPER_C2_URL / REAPER_USERNAME / REAPER_PASSWORD / REAPER_ENGAGEMENT / ANTHROPIC_API_KEY
docker compose run --rm harness \
  --client "Acme Corp" --engagement "acme-2026-q3" \
  --authorization "Signed SOW #2026-114, ROE dated 2026-09-01 to 2026-09-15" \
  --objective "Obtain domain admin from an external foothold"
```

The image bundles the compiled `harness` binary and the Claude Code CLI
(`@anthropic-ai/claude-code`, installed via npm). `ANTHROPIC_API_KEY` is required
for `claude` to authenticate inside the container — there's no browser available
for an interactive OAuth login there. `./sessions` is bind-mounted into the
container so session prompt files persist on the host.

To build the image directly instead of via compose:

```
docker build -t ai-harness:local .
```

## Multi-arch image build

The `Dockerfile` cross-compiles the Go binary natively from the build host
(`--platform=$BUILDPLATFORM` on the builder stage) and pulls the matching
Node.js base image per target platform, so a single `buildx` invocation produces
both architectures without emulating the Go compiler under QEMU:

```
docker buildx build --platform linux/amd64,linux/arm64 -t ai-harness:latest .
```

This verifies the Go binary and image build for both platforms. Nothing is
pushed — CI does the same (`push: false`). `docker buildx` can't `--load` more
than one platform into the local daemon at once; drop `--platform` to build a
single-arch image for your local host with `make docker-build`.

## Development

```
go vet ./...
go test ./...
gofmt -l .   # should print nothing
```

### Reproducing a Trivy CI failure locally

`.github/workflows/trivy.yml` runs with `format: sarif`, which — unlike Trivy's
default table output — writes results only to the SARIF file (visible in the repo's
Security tab), not to the job log. To see the actual vulnerability table when the
job fails:

```
docker build -t ai-harness:scan .

docker run --rm \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v trivy-cache:/root/.cache/ \
  ghcr.io/aquasecurity/trivy:0.36.0 image \
  --severity CRITICAL,HIGH --ignore-unfixed --format table \
  ai-harness:scan
```

(Match the version to whatever `trivy.yml`'s `aquasecurity/trivy-action` tag pins,
so local results match CI.)

`make build`, `make test`, `make vet`, `make fmt-check`, `make docker-build`, and
`make docker-buildx` wrap the above.

## Security notes

- The ReaperC2 password is only ever held in process memory / environment
  variables — the launcher never writes it to a file, and the skills are
  instructed to never print, log, or echo it back.
- Session prompt files under `sessions/` (git-ignored) do contain the client
  name, engagement name, and objectives — treat that directory as engagement-
  confidential.
- `docker-compose.yml` reads secrets from a git-ignored `.env` file; don't commit
  one.
