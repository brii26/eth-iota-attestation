#!/bin/sh
set -eu
umask 077

DEPLOYMENT_FILE=/deployment/ethereum-private.json
export DEPLOYMENT_FILE

write_key() {
  destination="$1"
  if [ -s "$destination" ]; then
    return
  fi
  temporary="$(mktemp "$(dirname "$destination")/private.XXXXXX")"
  {
    printf '0x'
    od -An -N32 -tx1 /dev/urandom | tr -d ' \n'
    printf '\n'
  } > "$temporary"
  mv "$temporary" "$destination"
}

case "${1:-}" in
  deploy)
    write_key /admin-key/private.key
    write_key /oracle-key/private.key
    ETHEREUM_ADMIN_PRIVATE_KEY="$(cat /admin-key/private.key)"
    ETHEREUM_ORACLE_PRIVATE_KEY="$(cat /oracle-key/private.key)"
    export ETHEREUM_ADMIN_PRIVATE_KEY ETHEREUM_ORACLE_PRIVATE_KEY
    if [ -s "$DEPLOYMENT_FILE" ]; then
      verify-private
    else
      export FUND_DEV_ACCOUNTS=1
      deploy-private
    fi
    chmod 0644 "$DEPLOYMENT_FILE"
    ;;
  oracle)
    ETHEREUM_ORACLE_PRIVATE_KEY="$(cat /oracle-key/private.key)"
    ETHEREUM_PRIVATE_CONTRACT_ADDRESS="$(sed -n 's/^[[:space:]]*"contractAddress": *"\(0x[0-9a-fA-F]*\)"[, ]*$/\1/p' "$DEPLOYMENT_FILE")"
    if [ "${#ETHEREUM_PRIVATE_CONTRACT_ADDRESS}" -ne 42 ]; then
      printf '%s\n' 'Invalid private deployment address' >&2
      exit 1
    fi
    export ETHEREUM_ORACLE_PRIVATE_KEY ETHEREUM_PRIVATE_CONTRACT_ADDRESS
    exec attestation-oracle
    ;;
  *)
    printf '%s\n' 'Choose deploy or oracle' >&2
    exit 2
    ;;
esac
