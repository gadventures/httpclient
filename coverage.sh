#/bin/sh
rm coverage.out
go test -coverprofile=coverage.out -race -v
if [ -z "$1" ]; then
  go tool cover -func=coverage.out
else
  go tool cover -html=coverage.out
fi
