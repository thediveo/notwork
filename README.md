# `notwork`

[![PkgGoDev](https://pkg.go.dev/badge/github.com/thediveo/notwork)](https://pkg.go.dev/github.com/thediveo/notwork)
[![GitHub](https://img.shields.io/github/license/thediveo/notwork)](https://img.shields.io/github/license/thediveo/notwork)
![build and test](https://github.com/thediveo/notwork/workflows/build%20and%20test/badge.svg?branch=master)
[![goroutines](https://img.shields.io/badge/go%20routines-not%20leaking-success)](https://pkg.go.dev/github.com/onsi/gomega/gleak)
[![file descriptors](https://img.shields.io/badge/file%20descriptors-not%20leaking-success)](https://pkg.go.dev/github.com/thediveo/fdooze)
[![Go Report Card](https://goreportcard.com/badge/github.com/thediveo/whalewatcher)](https://goreportcard.com/report/github.com/thediveo/notwork)
![Coverage](https://img.shields.io/badge/Coverage-93.8%25-brightgreen)

A tiny package to help with creating transient Linux virtual network elements
for testing purposes, without having to deal with the tedious details of proper
and robust cleanup.

`notwork` leverages the
[vishvananda/netlink](https://github.com/vishvananda/netlink) module for
RTNETLINK communication, as well as the [Ginkgo](https://github.com/onsi/ginkgo)
testing framework with [Gomega](https://github.com/onsi/gomega) matchers.

## Usage Example

Usually, you don't want to trash around in the host's network namespace.
Insteaed, let's trash around in a transient "throw-away" network namespace, just
created for testing purposes and removed at the end of the current test.

In it, we create a transient MACVLAN network interface with a dummy-type parent
network interface, also only for the duration of the current test (node):

```go
import (
    "github.com/thediveo/notwork/dummy"
    "github.com/thediveo/notwork/macvlan"

    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
)

var _ = Describe("some testing", func() {

    It("creates a transient MACVLAN with a dummy parent in a throw-away network namespace", func() {
        defer netns.EnterTransient()() // !!! double ()()
        // current temporary network namespace will be remove automatically
        // at the end of this test.

        mcvlan := macvlan.NewTransient(dummy.NewTransient())
        // ...virtual network interface will be automatically removed
        // at the end of this test.
    })

})
```

## VETH Pair Ends in Different Network Namespaces

With the previous example under our black notwork belts, let's create a VETH
pair of network interfaces that connect _two transient_ network namespaces.

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
