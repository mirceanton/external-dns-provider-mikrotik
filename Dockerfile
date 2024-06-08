# =================================================================================================
# BUILDER STAGE
# =================================================================================================
FROM golang:1.22-alpine@sha256:9bdd5692d39acc3f8d0ea6f81327f87ac6b473dd29a2b6006df362bff48dd1f8 as builder

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
