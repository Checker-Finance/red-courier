#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "Usage: $0 <config.yaml>"
  exit 1
fi

go run ./cmd/courier validate --config "$1"