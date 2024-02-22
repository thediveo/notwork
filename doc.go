/*
Package notwork helps unit tests to create transient virtual network interfaces.
It leverages both the [Ginkgo] testing framework and matchting (erm, sic!)
[Gomega] matchers.

# Usage Example

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

If creating any of the virtual network interfaces fails, the test will
immediately fail.

# Using Throw-Away Network Namespaces

Even better, don't trash around the host network namespace, but instead use a
throw-away network namespace that is separate from the host network namespace.

	import (
		"github.com/thediveo/notwork/dummy"
		"github.com/thediveo/notwork/macvlan"
		"github.com/thediveo/notwork/netns"

		. "github.com/onsi/ginkgo/v2"
		. "github.com/onsi/gomega"
	)

	var _ = Describe("some isolated testing", func() {

		It("creates a transient MACVLAN with a dummy parent inside a throw-away netns", func() {
			defer netns.EnterTransient()()
			mcvlan := macvlan.NewTransient(dummy.NewTransient())
		})

	})

Please pay attention to the double “()()” when deferring
[github.com/thediveo/notwork/netns.EnterTransient].

# VETH Pair Ends in Different Network Namespaces

With the previous examples under our black notwork belts, let's create a VETH
pair of network interfaces that connect two transient network namespaces.

	import (
		"github.com/notwork/netns"
		"github.com/notwork/veth"
	)

	It("connects two temporary network namespaces", func() {
		dupondNetns := netns.NewTransient()
		dupontNetns := netns.NewTransient()
		var dupond, dupont netlink.Link
		netns.Execute(dupondNetns, func() {
			dupond, dupont = veth.NewTransient(WithPeerNamespace(dupontNetns))
		})
	})

As for the names of the VETH pair end variables, please refer to [Dupond et
Dupont].

# Known Limitations

This module suffers from overzealous sub-packaging.

[Ginkgo]: https://github.com/onsi/ginkgo
[Gomega]: https://github.com/onsi/gomega
[MACVLAN]: https://developers.redhat.com/blog/2018/10/22/introduction-to-linux-interfaces-for-virtual-networking#macvlan
[Dupond et Dupont]: https://en.wikipedia.org/wiki/Thomson_and_Thompson
*/
package notwork
