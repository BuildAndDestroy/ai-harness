# ai-harness

Claude Code skills and a launcher for running an authorized red team engagement
end-to-end against [ReaperC2](https://github.com/BuildAndDestroy/ReaperC2) and
producing the client deliverable. See `CLAUDE.md` for the full skill roster and
standing rules; this file covers building and running the `harness` launcher.

## What's here

- `.claude/skills/` — five Claude Code skills: `red-team-operator`,
  `reaperc2-operator`, `exploit-development`, `purple-team-atomic-tests`,
  `harness-report`.
- `cmd/harness`, `internal/harness` — the Go launcher that validates engagement
  inputs, builds the initial session prompt, and starts `claude`.
- `v3-ghostwriter-executive-document.docx` — the Ghostwriter report template
  `harness-report` drafts against.
- `Dockerfile`, `docker-compose.yml` — containerized build/run, multi-arch
  (`linux/amd64`, `linux/arm64`).

## Running locally (Go toolchain)

Requires Go 1.26+ and the `claude` CLI on `PATH`.

```
go build -o bin/harness ./cmd/harness

./bin/harness \
  --reaper-url https://c2.example.com:8443 \
  --reaper-username op1 \
  --client "Acme Corp" \
  --engagement "acme-2026-q3" \
  --objective "Obtain domain admin from an external foothold" \
  --objective "Demonstrate access to the finance file share"
```

You'll be prompted for the ReaperC2 password (hidden input) unless you pass
`--reaper-password` or set `REAPER_PASSWORD`. The password is never written to
disk — it lives only in the launcher's process environment and is exported into
`claude`'s environment as `$REAPER_PASSWORD`. The rendered session prompt (client,
engagement, objectives, ReaperC2 URL/username — no secret) is saved to
`sessions/<engagement>.md` for reference.

Use `--dry-run` to see the prompt without launching `claude`, and
`harness --help` for the full flag list (including `--objectives-file` for a
longer objectives list and `--sessions-dir` to change where prompts are saved).

## Running with Docker

```
cp .env.example .env   # fill in REAPER_URL / REAPER_USERNAME / REAPER_PASSWORD / ANTHROPIC_API_KEY
docker compose run --rm harness \
  --client "Acme Corp" --engagement "acme-2026-q3" \
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
docker buildx build --platform linux/amd64,linux/arm64 -t ai-harness:latest --push .
```

(`--push` is required for a genuinely multi-platform result — `docker buildx`
can't `--load` more than one platform into the local daemon at once. Drop
`--platform` to build a single-arch image for your local host with
`make docker-build`.)

## Development

```
go vet ./...
go test ./...
gofmt -l .   # should print nothing
```

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
