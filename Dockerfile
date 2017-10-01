# Build the binary in a separate container
FROM golang:1.9-alpine3.6 AS builder
RUN mkdir -p /go/src/app
WORKDIR /go/src/app
COPY . /go/src/app
RUN go-wrapper install

# Copy provider to final container
FROM alpine:3.6
MAINTAINER Ville Törhönen <ville@torhonen.fi>
ENV TERRAFORM_VERSION 0.10.6
RUN set -x \
    && apk add --no-cache \
        ca-certificates \
    && apk add --no-cache --virtual .build-deps \
        unzip \
        curl \
    && curl -sLo /tmp/terraform.zip https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip \
    && unzip /tmp/terraform.zip -d /usr/local/bin \
    && rm -f /tmp/terraform.zip \
    && apk del .build-deps
COPY --from=builder /go/bin/terraform-provider-upcloud /usr/local/bin/terraform-provider-upcloud
ENTRYPOINT ["/usr/local/bin/terraform"]
