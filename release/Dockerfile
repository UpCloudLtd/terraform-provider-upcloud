ARG GO_VERSION=1.17
FROM golang:$GO_VERSION
SHELL ["/bin/bash", "-c"]
RUN apt-get update && apt-get install -y unzip

ARG GORELEASER_URL=https://github.com/goreleaser/goreleaser/releases/download/v1.8.3/goreleaser_Linux_x86_64.tar.gz
ARG GORELEASER_SHA256=304fa012709d12800528b124c9dbeabdcf8918f5e77b3877916e705798ed7962
WORKDIR /go/goreleaser
RUN set -x && \
    GORELEASER=$(basename $GORELEASER_URL) && \
    curl -L $GORELEASER_URL > ./$GORELEASER && \
    sha256sum -c <(echo "$GORELEASER_SHA256 $GORELEASER") && \
    tar xfzv $GORELEASER && \
    mv goreleaser /usr/local/bin

ARG VAULT_URL=https://releases.hashicorp.com/vault/1.5.4/vault_1.5.4_linux_amd64.zip
ARG VAULT_SHA256=50156e687b25b253a63c83b649184c79a1311f862c36f4ba16fd020ece4ed3b3
ARG VAULT_GPG_FINGERPRINT=C874011F0AB405110D02105534365D9472D7468F
ARG VAULT_SUMFILE_URL=https://releases.hashicorp.com/vault/1.5.4/vault_1.5.4_SHA256SUMS
ARG VAULT_SUMFILE_SIG_URL=https://releases.hashicorp.com/vault/1.5.4/vault_1.5.4_SHA256SUMS.sig
COPY hashicorp.asc /usr/share/keyrings/
WORKDIR /go/vault
RUN set -x && \
    VAULT=$(basename $VAULT_URL) && \
    VAULT_SUMFILE=$(basename $VAULT_SUMFILE_URL) && \
    VAULT_SUMFILE_SIG=$(basename $VAULT_SUMFILE_SIG_URL) && \
    curl -L $VAULT_URL > $VAULT && \
    curl -L $VAULT_SUMFILE_URL > $VAULT_SUMFILE && \
    curl -L $VAULT_SUMFILE_SIG_URL > $VAULT_SUMFILE_SIG && \
    rm -rf ~/.gnupg && \
    gpg --import /usr/share/keyrings/hashicorp.asc && \
    gpg --list-keys && \
    gpg --check-signatures $VAULT_GPG_FINGERPRINT && \
    gpg --verify $VAULT_SUMFILE_SIG  $VAULT_SUMFILE && \
    sha256sum -c <(grep $VAULT_SHA256 $VAULT_SUMFILE) && \
    unzip $VAULT && \
    mv vault /usr/local/bin

ADD entrypoint.sh /
ENV VAULT_ADDR="" VAULT_LOGIN="" VAULT_LOGIN_PASSWORD="" VAULT_SIGNER_PATH=""
ENV GITHUB_TOKEN="" GITHUB_OWNER=""

WORKDIR /go/src
ENTRYPOINT ["/entrypoint.sh"]
