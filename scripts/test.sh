#!/usr/bin/env bash
set -euo pipefail

go clean -testcache
go test ./... -race -coverprofile=coverage.out