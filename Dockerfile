FROM alpine:3.20.3@sha256:beefdbd8a1da6d2915566fde36db9db0b524eb737fc57cd1367effd16dc0d06d
USER 8675:8675
COPY external-dns-provider-mikrotik /
ENTRYPOINT ["/external-dns-provider-mikrotik"]