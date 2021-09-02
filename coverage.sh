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

# delete $OUT file, ignoring if it exists
rm -f "$OUT"

# run tests with coverprofile
function tests() {
  go test -v -race -coverprofile=$OUT
}

case "$COVER" in
  func)
    tests
    go tool cover -func=$OUT
    ;;
  html)
    tests
    go tool cover -html=$OUT
    ;;
  *)
    usage
    ;;
esac
