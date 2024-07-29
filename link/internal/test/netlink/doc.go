/*
Package netlink contains general netlink behavioral tests (assertions) and
documents gotchas in the github.com/vishvananda/netlink module.

# Netlink package (RT)NETLINK Handles and the Package Handle

The [netlink package] API uses a so-called “package handle” for its
package-level (RT)NETLINK functions: this is a package-wide instance of a zero
value [netlink.Handle]. The [netlink.Handle] type is described as “an handle for
the netlink requests on a specific network namespace”.

Whenever calling package methods that use this “package handle”, a temporary
(RT)NETLINK socket will be set up just for the duration of the method execution
and upon returning to the caller torn down (closed) again. Luckily for us, this
behavior ensures correct operation when calling netlink package methods from a
(hopefully OS-level thread-locked) Go routine currently attached to some other
network namespace, or when switching in and out of different network namespaces.
It is slightly inefficient but eventually doesn't matter in unit tests.

In contrast, non-zero value [netlink.Handle] objects obtained via
[netlink.NewHandleAt] create (RT)NETLINK sockets on demand and then keep these
sockets until the handle gets explicitly closed.

# Netlink Package Handle Creation

For instance, to create a netlink.Handle connected to another network namespace,
in our fail-safe form:

	import (
	    "github.com/thediveo/notwork/netns"
	)

	netnsfd := netns.NewTransient() // ...no explicit close necessary
	h := netns.NewNetlinkHandle(netnsfd) // ...no explicit close necessary

# Creating (“Adding”) a New Link

When creating a new network interface (“link”) using [netlink.LinkAdd] the
interface index will be automatically updated upon successful creation, by
calling [netlink.Handle.LinkByName]. Now, [netlink.LinkByName] and
[netlink.Handle.LinkByName] work on the RTNETLINK connection and interface name,
but have no way to pass the netlink.LinkAttrs.Namespace field that has been used
when creating the link.

The notwork module thus needs to work around this “quirk”.

[netlink package]: https://github.com/vishvananda/netlink
*/
package netlink
