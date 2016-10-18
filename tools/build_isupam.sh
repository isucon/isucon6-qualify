#!/bin/sh
OS_TARGETS="linux darwin freebsd windows"

for os in $OS_TARGETS; do
  if [ "$os" == "windows" ]; then
    GOOS=$os go build -ldflags "-s -w" -o bin/isupam_$os.exe ./cmd/isupam
  else
    GOOS=$os go build -ldflags "-s -w" -o bin/isupam_$os ./cmd/isupam
  fi
done
