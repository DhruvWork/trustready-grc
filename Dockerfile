# ubuntu:24.04 - pinned to digest for reproducibility (2026-02-05)
ARG BASE_IMAGE=ubuntu:24.04@sha256:c4a8d5503dfb2a3eb8ab5f807da5bc69a85730fb49b5cfca2330194ebcc41c7b
FROM ${BASE_IMAGE}

LABEL org.opencontainers.image.source="https://github.com/DhruvWork/trustready-grc"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.vendor="TrustReady"

RUN useradd -m probo && \
    apt-get update && \
    apt-get upgrade -y && \
    apt-get install -y ca-certificates libcap2-bin && \
    rm -rf /var/lib/apt/lists/*

ARG TARGETPLATFORM=linux/amd64
COPY $TARGETPLATFORM/trustreadyd /usr/local/bin/trustreadyd
COPY $TARGETPLATFORM/trustreadyd-bootstrap /usr/local/bin/trustreadyd-bootstrap
COPY entrypoint.sh /usr/local/bin/entrypoint.sh

RUN chmod +x /usr/local/bin/trustreadyd && \
    chmod +x /usr/local/bin/trustreadyd-bootstrap && \
    chmod +x /usr/local/bin/entrypoint.sh && \
    setcap CAP_NET_BIND_SERVICE=+eip /usr/local/bin/trustreadyd && \
    mkdir -p /etc/trustreadyd && \
    chown probo:probo /etc/trustreadyd

USER probo

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
