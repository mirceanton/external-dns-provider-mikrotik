# =================================================================================================
# BUILDER STAGE
# =================================================================================================
FROM golang:1.22-alpine@sha256:ace6cc3fe58d0c7b12303c57afe6d6724851152df55e08057b43990b927ad5e8 as builder

ARG PKG=github.com/mirceanton/external-dns-provider-mikrotik
ARG VERSION=dev
ARG REVISION=dev

WORKDIR /build
COPY . .

RUN go build -ldflags "-s -w -X main.Version=${VERSION} -X main.Gitsha=${REVISION}" ./cmd/webhook


# =================================================================================================
# PRODUCTION STAGE
# =================================================================================================
FROM scratch
USER 8675:8675
COPY --from=builder --chmod=555 /build/webhook /external-dns-mikrotik-webhook
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/external-dns-mikrotik-webhook"]
