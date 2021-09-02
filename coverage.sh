#!/usr/bin/env bash

set -eu

function usage() {
  echo "Usage: [OUT=filename] $0 [html|func]"
  exit 1
}

# OUT={filename}
: "${OUT:=coverage.out}"

# COVER default to func
COVER="${1:-func}"

# delete out file, ignoring if it exists
rm -f "$OUT"

# run tests with coverprofile
go test -v -race -coverprofile=$OUT

case "$COVER" in
  func)
    go tool cover -func=$OUT
    ;;
  html)
    go tool cover -html=$OUT
    ;;
  *)
    usage
    ;;
esac
