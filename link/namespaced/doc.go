/*
Package namespaced supports namespacing netlink.Link link objects for network
interface creation to a so-called “link” network namespace (to use Linux kernel
terminolgy) to reference parent/master links in other network namespaces
correctly.

# Background

It's messy.

The RTNETLINK ABI/API allows to specify the (“destination”) network namespace
where a new virtual network interface is to be created, using either an
IFLA_NET_NS_PID, IFLA_NET_NS_FD, or IFLA_TARGET_NETNSID attribute. To be more
precise, the network interface is first created in the network namespace a
RTNETLINK socket connects to, and only automatically moved to the “destination”
network namespace afterwards.

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

Unfortunately, the [netlink] package doesn't support IFLA_LINK_NETNSID. We thus
emulate the intended behavior by switching first into a “link” network
namespace, and then create the virtual network interface there so that
parent/master references are correctly interpreted. As usual, the kernel then
moves the newly created network interface to its “destination” network
namespace.

In order to keep the existing netlink.Link-based API this package thus
optionally wraps them into [Link] objects, where these wrapper objects carry the
information about the wanted “link” network namespace.

We don't envision users of the notwork module directly working with [Link]
objects though. Instead, we assume that module users prefer to work with the
WithLinkNamespace options.
*/
package namespaced
