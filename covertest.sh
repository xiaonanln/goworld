#!/usr/bin/env bash

set -e
echo "" > coverage.txt

PKG=$(go list ./... | grep -v /vendor/)

for d in $PKG; do
	go test -race -coverprofile=profile.out -covermode=atomic $d
	if [ -f profile.out ]; then
		cat profile.out >> coverage.txt
		rm profile.out
	fi
done
