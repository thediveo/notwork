/*
Package notwork helps unit tests to create transient virtual network interfaces.
It leverages both the [Ginkgo] testing framework and matchting (erm, sic!)
[Gomega] matchers.

# Usage

To create a transient [MACVLAN] network interface with a dummy-type parent
network interface for the duration of a test (node):

	import (
	    "github.com/thediveo/notwork/dummy"
	    "github.com/thediveo/notwork/macvlan"

	    . "github.com/onsi/ginkgo/v2"
	    . "github.com/onsi/gomega"
	)

	var _ = Describe("some testing", func() {

	    It("creates a transient MACVLAN with a dummy parent", func() {
	        mcvlan := macvlan.NewTransient(dummy.NewTransient())
	    })

	})

The MACVLAN and dummy network interfaces will automatically removed at the end
of the test (node) they are created in, regardless of success or failure.

If creating the virtual network interfaces fails, the test will immediately
fail.

# Disclaimer

This module suffers from overzealous sub-packaging.

[Ginkgo]: https://github.com/onsi/ginkgo
[Gomega]: https://github.com/onsi/gomega
[MACVLAN]: https://developers.redhat.com/blog/2018/10/22/introduction-to-linux-interfaces-for-virtual-networking#macvlan
*/
package notwork
