/*
Package link helps with creating transient type virtual network interfaces of
various types for testing purposes. It leverages the [Ginkgo] testing framework
and matching (erm, sic!) [Gomega] matchers.

The network interfaces created by this package are transient because they
automatically get removed at the end of the a test (spec, block/group, suite, et
cetera) using Ginkgo's [DeferCleanup].

[Ginkgo]: https://github.com/onsi/ginkgo
[Gomega]: https://github.com/onsi/gomega
[DeferCleanup]: https://pkg.go.dev/github.com/onsi/ginkgo/v2#DeferCleanup
*/
package link
