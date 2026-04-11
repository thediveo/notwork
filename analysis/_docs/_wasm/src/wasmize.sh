#!/bin/bash
set -e

for dir in */; do
  echo "${dir}main.go ⇒ ${dir%/}.wasm.gz"
  GOOS=js GOARCH=wasm go build -o "../${dir%/}.wasm" "${dir}main.go"
  (cd ".." && gzip -9 -f "${dir%/}.wasm")
done
