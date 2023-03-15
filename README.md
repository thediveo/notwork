# `notwork`

[![PkgGoDev](https://pkg.go.dev/badge/github.com/thediveo/notwork)](https://pkg.go.dev/github.com/thediveo/notwork)
[![GitHub](https://img.shields.io/github/license/thediveo/notwork)](https://img.shields.io/github/license/thediveo/notwork)
![build and test](https://github.com/thediveo/notwork/workflows/build%20and%20test/badge.svg?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/thediveo/whalewatcher)](https://goreportcard.com/report/github.com/thediveo/notwork)
![Coverage](https://img.shields.io/badge/Coverage-94.6%25-brightgreen)

A tiny package to help with creating transient Linux virtual network elements
for testing purposes. It leverages both the
[Ginkgo](https://github.com/onsi/ginkgo) testing framework and matching (erm,
sic!) [Gomega](https://github.com/onsi/gomega) matchers.

## Make Targets

- `make`: lists all targets.
- `make coverage`: runs all tests with coverage and then **updates the coverage
  badge in `README.md`**.
- `make pkgsite`: installs [`x/pkgsite`](https://golang.org/x/pkgsite/cmd/pkgsite), as
  well as the [`browser-sync`](https://www.npmjs.com/package/browser-sync) and
  [`nodemon`](https://www.npmjs.com/package/nodemon) npm packages first, if not
  already done so. Then runs the `pkgsite` and hot reloads it whenever the
  documentation changes.
- `make report`: installs
  [`@gojp/goreportcard`](https://github.com/gojp/goreportcard) if not yet done
  so and then runs it on the code base.
- `make test`: runs **all** tests (as root), always.
- `make vuln`: installs
  [`x/vuln/cmd/govulncheck`](https://golang.org/x/vuln/cmd/govulncheck) and then
  runs it.

## Copyright and License

Copyright 2023 Harald Albrecht, licensed under the Apache License, Version 2.0.
