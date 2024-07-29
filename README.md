# `notwork`

[![PkgGoDev](https://pkg.go.dev/badge/github.com/thediveo/notwork)](https://pkg.go.dev/github.com/thediveo/notwork)
[![GitHub](https://img.shields.io/github/license/thediveo/notwork)](https://img.shields.io/github/license/thediveo/notwork)
![build and test](https://github.com/thediveo/notwork/workflows/build%20and%20test/badge.svg?branch=master)
[![goroutines](https://img.shields.io/badge/go%20routines-not%20leaking-success)](https://pkg.go.dev/github.com/onsi/gomega/gleak)
[![file descriptors](https://img.shields.io/badge/file%20descriptors-not%20leaking-success)](https://pkg.go.dev/github.com/thediveo/fdooze)
[![Go Report Card](https://goreportcard.com/badge/github.com/thediveo/whalewatcher)](https://goreportcard.com/report/github.com/thediveo/notwork)
![Coverage](https://img.shields.io/badge/Coverage-85.7%25-brightgreen)

A tiny package to help with creating transient Linux virtual network elements
for testing purposes. It leverages both the
[Ginkgo](https://github.com/onsi/ginkgo) testing framework and matching (erm,
sic!) [Gomega](https://github.com/onsi/gomega) matchers.

## Usage Example

To create a transient MACVLAN network interface with a dummy-type parent network interface for the duration of a test (node):

```go
import (
    "github.com/thediveo/notwork/dummy"
    "github.com/thediveo/notwork/macvlan"

    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
)

var _ = Describe("some testing", func() {

    It("creates a transient MACVLAN with a dummy parent", func() {
        mcvlan := macvlan.NewTransient(dummy.NewTransient())
        // ...virtual network interface will be automatically removed
        // at the end of this test.
    })

})
```

## Using Throw-Away Network Namespaces

Even better, don't trash around the host network namespace, but instead use a
throw-away network namespace that is separate from the host network namespace.

```go
import (
    "github.com/thediveo/notwork/dummy"
    "github.com/thediveo/notwork/macvlan"
    "github.com/thediveo/notwork/netns"

    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
)

var _ = Describe("some isolated testing", func() {

    It("creates a transient MACVLAN with a dummy parent inside a throw-away netns", func() {
        defer netns.EnterTransient()() // !!! double ()()
        // we're not in a new transient network namespace and there's just
        // a lonely lo at this time.
        mcvlan := macvlan.NewTransient(dummy.NewTransient())
    })

})
```

## VETH Pair Ends in Different Network Namespaces

With the previous examples under our black notwork belts, let's create a VETH
pair of network interfaces that connect two transient network namespaces.

```go
import (
  "github.com/notwork/netns"
  "github.com/notwork/veth"
)

var _ = Describe("some isolated testing", func() {

	It("connects two temporary network namespaces", func() {
		dupondNetns := netns.NewTransient() // create, but don't enter
		dupontNetns := netns.NewTransient() // create, but don't enter
		dupond, dupont := veth.NewTransient(InNamespace(dupondNetns), WithPeerNamespace(dupontNetns))
	})

})
```

As for the names of the VETH pair end variables, please refer to [Dupond et
Dupont](https://en.wikipedia.org/wiki/Thomson_and_Thompson).


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
