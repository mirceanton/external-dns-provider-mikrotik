FROM gcr.io/distroless/static-debian12:nonroot
USER 8675:8675
COPY external-dns-provider-mikrotik /
ENTRYPOINT ["/external-dns-provider-mikrotik"]
