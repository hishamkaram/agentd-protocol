#!/usr/bin/env bash
# Shared Go coverage gate for the AgentD workspace. Fail-CLOSED: any inability
# to determine a numeric total coverage (missing profile, malformed profile,
# empty `go tool cover` output) exits non-zero rather than silently passing.
#
# Usage: check-go-coverage.sh [profile] [minimum]
#   profile  coverage profile path (default: coverage.out)
#   minimum  required percent      (default: 85)
#
# Env overrides (generic name first, legacy cloud name as fallback):
#   AGENTD_COVERAGE_MIN / AGENTD_CLOUD_COVERAGE_MIN      -> minimum
#   AGENTD_COVERAGE_EXCLUDE / AGENTD_CLOUD_COVERAGE_EXCLUDE -> line-exclude regex
#     (default excludes sqlc-generated cloud code; harmless in repos lacking it)
set -euo pipefail

profile="${1:-coverage.out}"
minimum="${AGENTD_COVERAGE_MIN:-${AGENTD_CLOUD_COVERAGE_MIN:-${2:-85}}}"
exclude_regex="${AGENTD_COVERAGE_EXCLUDE:-${AGENTD_CLOUD_COVERAGE_EXCLUDE:-/internal/store/postgres/db/}}"

if [[ ! -f "$profile" ]]; then
  echo "coverage profile not found: $profile" >&2
  exit 2
fi

filtered_profile="$(mktemp)"
trap 'rm -f "$filtered_profile"' EXIT

awk -v exclude_regex="$exclude_regex" '
  NR == 1 { print; next }
  $0 !~ exclude_regex { print }
' "$profile" > "$filtered_profile"

total="$(
  GOWORK="${GOWORK:-off}" go tool cover -func="$filtered_profile" |
    awk '/^total:/ {gsub("%", "", $3); print $3}'
)"

if [[ -z "$total" ]]; then
  echo "could not determine total coverage from $profile" >&2
  exit 2
fi

awk -v total="$total" -v minimum="$minimum" 'BEGIN {
  if ((total + 0) < (minimum + 0)) {
    printf("coverage %.1f%% is below required %.1f%%\n", total, minimum) > "/dev/stderr"
    exit 1
  }
  printf("coverage %.1f%% >= required %.1f%%\n", total, minimum)
}'
