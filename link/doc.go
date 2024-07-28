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
the “current network namespace” (but see below). This current network namespace
is the one the OS-level thread executing the calling go routine is currently
attached to.

Please note that [github.com/thediveo/notwork/netns.EnterTransient]
automatically locks an OS-level thread to its go routine in order to avoid nasty
surprises. In a similar way, [github.com/thediveo/notwork/netns.Execute] also
temporarily locks the executing OS-level thread to its calling go routine for
the duration of the function to be executed in a different network namespace.

Next, please note that [netlink.Attrs.Namespace] can be set to a [netlink.NsFd]
wrapping the file descriptor (number) of a (“destination”) network namespace
different from the current network namespace when calling [NewTransient]. In
this case, the newly created virtual network interface will end up in the
referenced network namespace instead of the current network namespace. While
this looks like it was created in the “destination” network namespace, the Linux
kernel actually creates the new network interface in the current network
namespace (but see below) and then moves it to its final resting place.

However, when creating virtual network interfaces related to existing network
interfaces – such as in case of MACVLAN –, then the current network namespace
normally defines where this “related” network interface is. For instance,
[netlink.LinkAttrs.ParentIndex] references the parent network interface of a
MACVLAN in the current namespace. Yet, the MACVLAN network interface will be
created in a network namespace different from the current network namespace, if
[netlink.LinkAttrs.Namespace] has been set.

In case of VETH pairs of network interfaces a third network namespace comes into
play, courtesy of [netlink.Veth.PeerNamespace]. However, the current network
namespace doesn't play any role here anymore, as long as both
[netlink.Attrs.Namespace] and [netlink.Veth.PeerNamespace] are set. If either or
both of these network namespace fields are unset (nil), then the current network
namespace applies.

# WithLinkNamespace

Wrapping a [netlink.Link] using [Wrap] associates that link information with yet
another network namespace, the so-called “link” network namespace in Linux
kernel parlance. While this is an ambigious choice of word, we keep with it (see
[rtnl_newlink_create]) to hopefully not cause even more confusing by inventing
our own terminology.

In short, passing such a wrapped link to [NewTransient] or using
[FromLinkNamespace] instructs it to create the new network interface not from
the perspective of the current network namespace but from the “link” network
namespace specified when wrappikng the original [netlink.Link] information.

# (RT)NETLINK Background

The short story: It's messy.

The long story: The RTNETLINK ABI/API allows to specify the (“destination”)
network namespace where a new virtual network interface is to be created, using
either an IFLA_NET_NS_PID, IFLA_NET_NS_FD, or IFLA_TARGET_NETNSID attribute. To
be more precise, the network interface is first created in the network namespace
a RTNETLINK socket connects to, and only automatically moved to the
“destination” network namespace afterwards.

However, some virtual network interfaces, such as MACVLANs, need a reference to
their parent/master network interface when created. This “link” reference
(kernel terminology can be ambiguous) is taken in the RTNETLINK socket's network
namespace, not the “destination” network namespace.

Now, for reasons of symmetry and in order to basically use any arbitrary
RTNETLINK socket that happens to lie around to create new virtual network
interface in any other arbitrary network namespace, such “link” references
should be namespaced independently of the “destination” and “socket” network
namespaces.

Technically, namespacing parent/master references is done using
IFLA_LINK_NETNSID attributes (that are not to be confused with
IFLA_TARGET_NETNSID attributes).

# vishvananda/netlink

Unfortunately, the vishvananda [netlink] package doesn't support
IFLA_LINK_NETNSID. We thus emulate the intended behavior by switching first into
a “link” network namespace, and then create the virtual network interface there
so that parent/master references are correctly interpreted. As usual, the kernel
then moves the newly created network interface to its “destination” network
namespace.

In order to keep the existing netlink.Link-based API this package thus
optionally wraps them into [Link] objects, where these wrapper objects carry the
information about the wanted “link” network namespace.

We don't envision users of the notwork module directly working with [Link]
objects though. Instead, we assume that module users prefer to work with the
WithLinkNamespace options provided by the individual network interface creation
functions.

[Ginkgo]: https://github.com/onsi/ginkgo
[Gomega]: https://github.com/onsi/gomega
[DeferCleanup]: https://pkg.go.dev/github.com/onsi/ginkgo/v2#DeferCleanup
[rtnl_newlink_create]: https://elixir.bootlin.com/linux/v6.10/source/net/core/rtnetlink.c#L3456
*/
package link
