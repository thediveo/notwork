/*
Package netdevsim helps with creating transient [netdevsim] type virtual network
interfaces for testing purposes. It leverages the [Ginkgo] testing framework and
matching (erm, sic!) [Gomega] matchers.

Technically, a netdevsim in the strict sense is a device on the “netdevsim” bus.
A netdevsim bus device then has a number of ports, where these ports corresponds
with netdevs/network interfaces provided by the netdevsim device. For
simplicity, we'll call these network interface also “netdevsim” network
interfaces, like their bus devices.

The "netdevsim" network interfaces created by this package are transient because
they automatically get removed at the end of the a test (spec, block/group,
suite, et cetera) using Ginkgo's [DeferCleanup].

Since Linux kernel 6.9+ two “port” network interfaces of netdevsims can be
linked together, similar to “veth” pairs.

# Caveats

On at least some Linux distributions, you might need to explicitly modprobe the
“netdevsim” module. For what it's worth, the Linux kernel self-tests also
modprobe the netdevsim module.

# Background Information

The lifecycle management of netdevsim network interfaces can't be done through
the usual NETLINK link API, as with other virtual network interfaces, such as
“veth”. Instead, the following sysfs-located pseudo files must be used:

  - /sys/bus/netdevsim/new_device
  - /sys/bus/netdevsim/del_device

Unfortunately, when creating a new netdevsim instance via the pseudo file, we
don't get the information about the names (or indices) of the newly created
network interfaces. This package thus hides the complexity in mapping netdevsim
instances to their particular network names and indices.

To make things worse, the only currently known way to map netdevsim ports to
their network interfaces is via /sys/bus/netdevsim/devices/net/, with the
associated problems around sysfs and non-adaptive network namespacing.

This package works around the situation by creating netdevsims with multiple
ports piecemeal-wise, picking up the newly created ports piece by piece.

[netdevsim]: https://docs.kernel.org/process/maintainer-netdev.html#netdevsim
[Ginkgo]: https://github.com/onsi/ginkgo [Gomega]:
https://github.com/onsi/gomega [DeferCleanup]:
https://pkg.go.dev/github.com/onsi/ginkgo/v2#DeferCleanup
*/
package netdevsim
