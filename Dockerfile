# =================================================================================================
# BUILDER STAGE
# =================================================================================================
FROM golang:1.22-alpine@sha256:562abef817ebfee7ce2c56b4b503c080ccfe8c4497549370c0d2e0c929849fe4 AS builder

ARG VERSION=dev
ARG REVISION=dev

WORKDIR /build
COPY . .

RUN go build -ldflags "-s -w -X main.Version=${VERSION} -X main.Gitsha=${REVISION}" -o webhook


# =================================================================================================
# PRODUCTION STAGE
# =================================================================================================
FROM scratch
USER 8675:8675
COPY --from=builder --chmod=555 /build/webhook /external-dns-mikrotik-webhook
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/external-dns-mikrotik-webhook"]
