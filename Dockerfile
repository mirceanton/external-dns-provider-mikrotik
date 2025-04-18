FROM alpine:3.21.3
USER 8675:8675
COPY external-dns-provider-mikrotik /
ENTRYPOINT ["/external-dns-provider-mikrotik"]