#!/bin/sh

sudo apt-get update
sudo apt-get install -y apache2-utils podman

cat > config.yml <<EOF
version: 0.1
delete:
  enabled: true
http:
  addr: 0.0.0.0:80
proxy:
  remoteurl: https://registry-1.docker.io
  ttl: 168h
storage:
  filesystem: {}
EOF

mkdir -p auth
htpasswd -Bbn token "${token}" > auth/htpasswd

podman run -d -p 80:80 \
  -v "$(pwd)/config.yml:/etc/distribution/config.yml" \
  -v "$(pwd)/auth:/auth" \
  -e REGISTRY_AUTH=htpasswd \
  -e "REGISTRY_AUTH_HTPASSWD_REALM=Registry Realm" \
  -e REGISTRY_AUTH_HTPASSWD_PATH=/auth/htpasswd \
  --name registry \
  registry:3
