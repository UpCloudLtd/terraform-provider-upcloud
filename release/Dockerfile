ARG GO_VERSION=1.14
FROM golang:$GO_VERSION
SHELL ["/bin/bash", "-c"]
RUN apt-get update && apt-get install unzip

ARG GORELEASER_URL=https://github.com/goreleaser/goreleaser/releases/download/v0.145.0/goreleaser_Linux_x86_64.tar.gz
ARG GORELEASER_SHA256=2ea97ea7d5c1210470d23944188987d85c51768b3d7d3cf7bccb76393b57c44a
WORKDIR /go/goreleaser
RUN set -x && \
    GORELEASER=$(basename $GORELEASER_URL) && \
    curl -L $GORELEASER_URL > ./$GORELEASER && \
    sha256sum -c <(echo "$GORELEASER_SHA256 $GORELEASER") && \
    tar xfzv $GORELEASER && \
    mv goreleaser /usr/local/bin

ARG VAULT_URL=https://releases.hashicorp.com/vault/1.5.4/vault_1.5.4_linux_amd64.zip
ARG VAULT_SHA256=50156e687b25b253a63c83b649184c79a1311f862c36f4ba16fd020ece4ed3b3
ARG VAULT_GPG_KEY=51852D87348FFC4C
ARG VAULT_GPG_FINGERPRINT=91A6E7F85D05C65630BEF18951852D87348FFC4C
ARG VAULT_SUMFILE_URL=https://releases.hashicorp.com/vault/1.5.4/vault_1.5.4_SHA256SUMS
ARG VAULT_SUMFILE_SIG_URL=https://releases.hashicorp.com/vault/1.5.4/vault_1.5.4_SHA256SUMS.sig
WORKDIR /go/vault
RUN set -x && \
    VAULT=$(basename $VAULT_URL) && \
    VAULT_SUMFILE=$(basename $VAULT_SUMFILE_URL) && \
    VAULT_SUMFILE_SIG=$(basename $VAULT_SUMFILE_SIG_URL) && \
    curl -L $VAULT_URL > $VAULT && \
    curl -L $VAULT_SUMFILE_URL > $VAULT_SUMFILE && \
    curl -L $VAULT_SUMFILE_SIG_URL > $VAULT_SUMFILE_SIG && \
    rm -rf ~/.gnupg && \
    gpg --receive-keys $VAULT_GPG_KEY && \
    gpg --check-signatures $VAULT_GPG_FINGERPRINT && \
    gpg --verify $VAULT_SUMFILE_SIG && \
    sha256sum -c <(grep $VAULT_SHA256 $VAULT_SUMFILE) && \
    unzip $VAULT && \
    mv vault /usr/local/bin

ADD entrypoint.sh /
ENV VAULT_ADDR="" VAULT_LOGIN="" VAULT_LOGIN_PASSWORD="" VAULT_SIGNER_PATH=""
ENV GITHUB_TOKEN="" GITHUB_OWNER=""

WORKDIR /go/src
ENTRYPOINT ["/entrypoint.sh"]