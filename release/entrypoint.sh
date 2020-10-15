#!/bin/bash

set -eux
vault write -field=token auth/ldap/login/"$VAULT_LOGIN" \
  password=@<(printenv VAULT_LOGIN_PASSWORD | tr -d '\n') > ~/.vault-token

trap "vault token revoke -self" EXIT
"$@"