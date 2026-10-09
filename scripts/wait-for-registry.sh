#!/usr/bin/env bash
# Poll a Terraform-protocol registry until it lists a provider version.
#
# Usage:
#   wait-for-registry.sh <name> <versions-url> <version> [timeout-seconds] [interval-seconds]
#
# <version> may include a leading v (v0.3.1). HTTP 404 means the provider is not
# listed yet and fails immediately. OpenTofu requires a one-time browser
# submission; see docs/guides/releasing.md.

set -euo pipefail

NAME="${1:?name}"
VERSIONS_URL="${2:?versions url}"
VERSION="${3:?version}"
TIMEOUT_SECONDS="${4:-1200}"
INTERVAL_SECONDS="${5:-30}"

VERSION="${VERSION#v}"
TMP="$(mktemp)"
trap 'rm -f "$TMP"' EXIT

deadline=$((SECONDS + TIMEOUT_SECONDS))

fail_not_listed() {
  echo "${NAME} Registry does not list vetal-ca-org/voipms (HTTP 404)." >&2
  if [[ "${NAME}" == "OpenTofu" ]]; then
    cat >&2 <<'EOF'

OpenTofu listings are not created from CI. Submit once in a browser
(the GitHub issue form UI only; gh/API issues are closed unprocessed):

  1. Provider: https://github.com/opentofu/registry/issues/new?template=provider.yml
     Provider Repository: vetal-ca-org/terraform-provider-voipms

  2. After that merge, signing key:
     https://github.com/opentofu/registry/issues/new?template=provider_key.yml
     Provider Namespace: vetal-ca-org
     Public org membership on vetal-ca-org must be visible.
     Provider GPG Key:  gpg --armor --export F3ADF9C3A8C694B3

Then re-run the failed Release jobs.
EOF
  fi
  exit 1
}

has_version() {
  jq -e --arg v "$VERSION" '.versions[]? | select(.version == $v)' "$TMP" >/dev/null
}

while true; do
  http_code="$(curl -sS -L --retry 3 --retry-delay 2 -o "$TMP" -w "%{http_code}" "$VERSIONS_URL" || echo "000")"
  case "$http_code" in
    404)
      fail_not_listed
      ;;
    200)
      if has_version; then
        echo "${NAME} Registry lists ${VERSION}."
        exit 0
      fi
      echo "${NAME} Registry is listed; waiting for version ${VERSION}."
      ;;
    *)
      echo "HTTP ${http_code} from ${VERSIONS_URL}; retrying."
      ;;
  esac
  remaining=$((deadline - SECONDS))
  if (( remaining <= 0 )); then
    echo "Timed out after ${TIMEOUT_SECONDS}s waiting for ${NAME} Registry to ingest ${VERSION}." >&2
    exit 1
  fi
  if (( INTERVAL_SECONDS < remaining )); then
    sleep "$INTERVAL_SECONDS"
  else
    sleep "$remaining"
  fi
done
