#!/usr/bin/env bash
# One JSON array line for .github/workflows/acctest.yml matrix "group" values.
#
#   --all              full list from .github/acctest-path-mapping.json
#   git diff ... | $0    PR subset (same file; rules below in the embedded jq)
#
set -euo pipefail
ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
MAP="$ROOT/.github/acctest-path-mapping.json"
JQ_SCRIPT="$ROOT/scripts/acctest-groups-from-diff.jq"
[[ -f "$MAP" ]] || {
  echo "acctest-groups-from-diff: missing $MAP" >&2
  exit 1
}
[[ -f "$JQ_SCRIPT" ]] || {
  echo "acctest-groups-from-diff: missing $JQ_SCRIPT" >&2
  exit 1
}

full_matrix() {
  jq -c '(.groups | keys) + ["unittests"] | unique | sort' "$MAP"
}

[[ "${1:-}" == --all ]] && {
  full_matrix
  exit 0
}

paths=$(jq -cRs 'split("\n") | map(select(length > 0))' <<<"$(cat)")
# -c required: GITHUB_OUTPUT expects one line (echo "groups=$groups").
jq -nc --argjson paths "$paths" --slurpfile m "$MAP" -f "$JQ_SCRIPT"
