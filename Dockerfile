# =================================================================================================
# BUILDER STAGE
# =================================================================================================
FROM golang:1.22-alpine@sha256:1a478681b671001b7f029f94b5016aed984a23ad99c707f6a0ab6563860ae2f3 AS builder

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
