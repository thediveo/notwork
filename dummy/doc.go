/*
Package dummy helps with creating transient [dummy] type virtual network
interfaces for testing purposes. It leverages the [Ginkgo] testing framework and
matching (erm, sic!) [Gomega] matchers.

The "dummy" network interfaces created by this package are transient because
they automatically get removed at the end of the a test (spec, block/group,
suite, et cetera) using Ginkgo's [DeferCleanup].

[dummy]: https://tldp.org/LDP/nag/node72.html#SECTION007770000
[Ginkgo]: https://github.com/onsi/ginkgo
[Gomega]: https://github.com/onsi/gomega
[DeferCleanup]: https://pkg.go.dev/github.com/onsi/ginkgo/v2#DeferCleanup
*/
package dummy
