#!/usr/bin/env bash
set -euo pipefail

COVER_FILE="${1:-coverage.out}"
MIN_COVERAGE="${2:-${COVERAGE_MIN:-85}}"

if [[ ! -f "${COVER_FILE}" ]]; then
  echo "coverage file not found: ${COVER_FILE}" >&2
  exit 1
fi

TOTAL=$(go tool cover -func="${COVER_FILE}" | awk '/^total:/ {gsub("%","",$3); print $3}')
if [[ -z "${TOTAL}" ]]; then
  echo "failed to parse total coverage from ${COVER_FILE}" >&2
  exit 1
fi

if awk -v total="${TOTAL}" -v min="${MIN_COVERAGE}" 'BEGIN { exit !(total+0 >= min+0) }'; then
  echo "coverage gate passed: ${TOTAL}% >= ${MIN_COVERAGE}%"
  exit 0
fi

echo "coverage gate failed: ${TOTAL}% < ${MIN_COVERAGE}%" >&2
exit 1
