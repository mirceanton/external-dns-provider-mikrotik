# =================================================================================================
# BUILDER STAGE
# =================================================================================================
FROM golang:1.23-alpine@sha256:44a2d64f00857d544048dd31d8e1fbd885bb90306819f4313d7bc85b87ca04b0 AS builder

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
