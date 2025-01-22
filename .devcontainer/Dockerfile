## ================================================================================================
# Utility versions
## ================================================================================================
ARG TASKFILE_VERSION=v3.38.0@sha256:308c4f5be86bffae3f956cbd7225c4fec69b0e7a89012120b818a10df45d7c59
ARG GOLINT_VERSION=v1.63.4-alpine@sha256:e1768ec667a10f7712d38abc484d18660d8c32caf3a252b8531ea77ff09de575
ARG GORELEASER_VERSION=v2.6.1@sha256:8577ee8e351783b8c89ce26c973c3dbd4f854121e0c5456893ea4c9a3a6d76ce
ARG K9S_VERSION=v0.32.7@sha256:a967991e635dc8327c25c9d20ea71089a9cad66793013c5675c81a1f6c79292b

## ================================================================================================
# "Build" stage for utilities with docker images already present
## ================================================================================================
FROM ghcr.io/mirceanton/taskfile:${TASKFILE_VERSION} AS taskfile
FROM golangci/golangci-lint:${GOLINT_VERSION} AS golangci-lint
FROM goreleaser/goreleaser:${GORELEASER_VERSION} AS goreleaser
FROM docker.io/derailed/k9s:${K9S_VERSION} AS k9s

## ================================================================================================
## Main image
## ================================================================================================
FROM mcr.microsoft.com/devcontainers/go:1.23-bookworm@sha256:a417a341a2a8648db7bf8527b86364848362a2c8dc150993c8a4cc2b53b4ec47 AS workspace

# Install additional tools
COPY --from=taskfile /task /usr/local/bin/task
COPY --from=k9s /bin/k9s /usr/local/bin/k9s
COPY --from=golangci-lint /usr/bin/golangci-lint /usr/local/bin/golangci-lint
COPY --from=goreleaser /usr/bin/goreleaser /usr/local/bin/goreleaser

# Generate bash completions
RUN golangci-lint completion bash | sudo tee /etc/bash_completion.d/golangci-lint.bash > /dev/null
RUN goreleaser completion bash | sudo tee /etc/bash_completion.d/goreleaser.bash > /dev/null

USER vscode
ENTRYPOINT [ "/bin/bash", "-l", "-c" ]
