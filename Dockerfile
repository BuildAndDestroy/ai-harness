# syntax=docker/dockerfile:1

## ---- build stage ------------------------------------------------------------
# Builds natively on the platform running `docker buildx build` and cross-compiles
# the Go binary for whatever target platform(s) are requested, so multi-arch
# builds (linux/amd64, linux/arm64) don't need QEMU to run the compiler itself.
FROM --platform=$BUILDPLATFORM golang:1.26-bookworm AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w" -o /out/harness ./cmd/harness

## ---- runtime stage ------------------------------------------------------------
# Node base so the Claude Code CLI (an npm package) can live alongside the
# compiled harness binary — harness execs `claude` to actually run the session.
FROM node:22-bookworm-slim AS runtime

ARG CLAUDE_CODE_VERSION=latest

# apt-get upgrade pulls in Debian security-repo fixes published after this
# base image's layer was snapshotted (e.g. libpcre2-8-0); npm@latest replaces
# the Node image's bundled npm, which carries its own dated transitive deps
# (pacote/tar/sigstore/ip-address/brace-expansion/picomatch) — both matter for
# the Trivy scan in CI, not just functionality.
RUN apt-get update \
    && apt-get upgrade -y \
    && apt-get install -y --no-install-recommends ca-certificates git curl \
    && rm -rf /var/lib/apt/lists/* \
    && npm install -g npm@latest \
    # npm now blocks postinstall scripts by default unless explicitly
    # allow-listed; claude-code's postinstall fetches its native binary, so it
    # must be allowed or the CLI is left non-functional.
    && npm install -g --allow-scripts=@anthropic-ai/claude-code @anthropic-ai/claude-code@${CLAUDE_CODE_VERSION} \
    && npm cache clean --force \
    # claude resolves to a native binary (bin/claude.exe) and needs neither
    # node nor npm at runtime — only npm itself was used to fetch it. Drop
    # npm's own bundled node_modules (its vendored pacote/tar/ip-address/etc.,
    # which trail their own CVEs) now that it's served its purpose.
    && rm -rf /usr/local/lib/node_modules/npm /usr/local/bin/npm /usr/local/bin/npx

RUN useradd --create-home --uid 10001 --shell /usr/sbin/nologin harness

WORKDIR /workspace

COPY --from=builder /out/harness /usr/local/bin/harness
COPY CLAUDE.md ./CLAUDE.md
COPY .claude ./.claude
COPY v3-ghostwriter-executive-document.docx ./v3-ghostwriter-executive-document.docx

RUN mkdir -p /workspace/sessions && chown -R harness:harness /workspace

USER harness

# ANTHROPIC_API_KEY must be provided at runtime for claude to authenticate
# non-interactively; REAPER_URL / REAPER_USERNAME / REAPER_PASSWORD (or the
# matching --reaper-* flags) are required by the harness binary itself.
ENTRYPOINT ["/usr/local/bin/harness"]
CMD ["--help"]
