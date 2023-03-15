#!/bin/bash
set -e

if ! command -v gobadge &>/dev/null; then
    export PATH="$(go env GOPATH)/bin:$PATH"
    if ! command -v gobadge &>/dev/null; then
        go install github.com/AlexBeauchemin/gobadge@latest
    fi
fi

go test -covermode=atomic -coverprofile=coverage-root.out -v -p=1 -count=1 -exec sudo ./...
go tool cover -func=coverage-root.out -o=coverage.out
go tool cover -html coverage-root.out -o=coverage.html
gobadge -filename=coverage.out -green=80 -yellow=50
