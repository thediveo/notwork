/*
Package link helps with creating transient virtual network interfaces of various
types for testing purposes. This package leverages the [Ginkgo] testing
framework and matching (erm, sic!) [Gomega] matchers to assert that creation of
new virtual network interfaces succeeds as expected.

The network interfaces created by this package are transient because they
automatically get removed at the end of the a test – a spec, block/group, suite,
et cetera – using Ginkgo's [DeferCleanup].

# Network Namespace Roulette

Please see [github.com/thediveo/notwork/netns] for details on how to create and
work with multiple network namespaces different from the initial/host network
namespace.

First, [NewTransient] for creating new virtual network interfaces always acts in
the “current network namespace” (but see below). This current network
namespace is the one the OS-level thread executing the calling go routine is
currently attached to.

Please note that [github.com/thediveo/notwork/netns.EnterTransient]
automatically locks an OS-level thread to its go routine in order to avoid nasty
surprises. In a similar way, [github.com/thediveo/notwork/netns.Execute] also
temporarily locks the executing OS-level thread to its calling go routine for
the duration of the function to be executed in a different network namespace.

Next, please note that [netlink.Attrs.Namespace] can be set to a [netlink.NsFd]
wrapping the file descriptor (number) of a network namespace different from the
current network namespace when calling [NewTransient]. In this case, the newly
created virtual network interface will be created in the referenced network
namespace instead of the current network namespace.

However, when creating virtual network interfaces related to existing network
interfaces – such as in case of MACVLAN –, then the current network namespace
defines where this “related” network interface is. For instance,
[netlink.LinkAttrs.ParentIndex] references the parent network interface of a
MACVLAN in the current namespace. Yet, the MACVLAN network interface will be
created in a network namespace different from the current network namespace,
if [netlink.LinkAttrs.Namespace] has been set.

In case of VETH pairs of network interfaces a third network namespace comes
into play, courtesy of [netlink.Veth.PeerNamespace]. However, the current
network namespace doesn't play any role here anymore, as long as both
[netlink.Attrs.Namespace] and [netlink.Veth.PeerNamespace] are set. If either or
both of these network namespace fields are unset (nil), then the current network
namespace applies.

[Ginkgo]: https://github.com/onsi/ginkgo
[Gomega]: https://github.com/onsi/gomega
[DeferCleanup]: https://pkg.go.dev/github.com/onsi/ginkgo/v2#DeferCleanup
*/
package link
