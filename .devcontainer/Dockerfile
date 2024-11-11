## ================================================================================================
# Utility versions
## ================================================================================================
ARG TASKFILE_VERSION=v3.38.0@sha256:308c4f5be86bffae3f956cbd7225c4fec69b0e7a89012120b818a10df45d7c59
ARG GOLINT_VERSION=v1.61.0-alpine@sha256:61e2d68adc792393fcb600340fe5c28059638d813869d5b4c9502392a2fb4c96
ARG GORELEASER_VERSION=v2.4.4@sha256:1f5ae36e41ede8b994a93c66e5c1fb1b1111de2ce88f25b6378e341771ed963e
ARG K9S_VERSION=v0.32.6@sha256:2d2410af2f8166dd2b67a63d2f172404a49d9f1aec9fc3719c85ccedc8bc660d

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
FROM mcr.microsoft.com/devcontainers/go:1.23-bookworm@sha256:2e00578e7c526e76ebd51597260928d8542d27c17ffe6390d79200638c51a3e6 AS workspace

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
