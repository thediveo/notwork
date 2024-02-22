/*
Package netns supports running unit tests in separated transient network
namespaces.

This package also helps handling network namespace identiers in form of inode
numbers, as well as getting [netlink.Handle] objects for messing around with
network interfaces, network address configurations, routing, et cetera.

# Usage

The simplest use cased is to just call [EnterTransient] and defer its return
value.

	import "github.com/notwork/netns"

	It("tests something inside a temporary network namespace", func() {
	  defer netns.EnterTransient()() // !!! double ()()
	  // ...
	})

This first locks the calling go routine to its OS-level thread, then creates a
new throw-away network namespace, and finally switches the OS-level thread with
the locked go routine to this new network namespace. Deferring the result of
[EnterTransient] ensures to switch the OS-level thread back to its original
network namespace (usually the host network namespace) and unlocks the thread
from the go routine. If there are no further references alive to the throw-away
network namespace, then the Linux kernel will automatically garbage collect it.

# Advanced

In more complex scenarios, such as testing with multiple throw-away network
namespaces, these can be created without automatically switching into them.
Instead, creating virtual network interfaces can be done by either only
temporarily switching into these network namespaces using [Execute], or by
creating a [NewNetlinkHandle] to carry out RTNETLINK operations on the
handle(s).

The following example uses the first method of switching into the first
throw-away network namespace and then creates a VETH pair of network interfaces.
One end is located in the second network interfaces.

	import (
		"github.com/notwork/netns"
		"github.com/notwork/veth"
	)

	It("tests something inside a temporary network namespace", func() {
		dupondNetns := netns.NewTransient()
		dupontNetns := netns.NewTransient()
		var dupond, dupont netlink.Link
		netns.Execute(dupondNetns, func() {
			dupond, dupont = veth.NewTransient(WithPeerNamespace(dupontNetns))
		})
	})

As for the names of the VETH pair end variables, please refer to [Dupond et
Dupont].

[Dupond et Dupont]: https://en.wikipedia.org/wiki/Thomson_and_Thompson
*/
package netns
