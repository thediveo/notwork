/*
Package netdevsim helps with creating transient [netdevsim] type virtual network
interfaces for testing purposes. It leverages the [Ginkgo] testing framework and
matching (erm, sic!) [Gomega] matchers.

# Caveat Emptor

From the [kernel perspective], “netdevsim is a test driver which can be used to
exercise driver configuration APIs without requiring capable hardware. [...] We
give no guarantees that netdevsim won’t change in the future in a way which
would break what would normally be considered uAPI.” (uAPI = “userspace API”).

With that made clear, let's trespass and move on...

# Test Long and Prosper

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

For your convenience, you might want to simply call [ensure.Netdevsim] and check
its bool result:
  - true indicates that the netdevsim bus is available (in the worst case by
    loading the required kernel module when necessary),
  - false indicates that either the caller isn't root and thus cannot manage
    netdevsim devices anyway, or the required kernel module could not be loaded.

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

It should now go without saying, but don't run multiple netdevsim-related tests
concurrently.

[netdevsim]: https://docs.kernel.org/process/maintainer-netdev.html#netdevsim
[Ginkgo]: https://github.com/onsi/ginkgo
[Gomega]: https://github.com/onsi/gomega
[DeferCleanup]: https://pkg.go.dev/github.com/onsi/ginkgo/v2#DeferCleanup
[kernel perspective]: https://docs.kernel.org/process/maintainer-netdev.html#netdevsim
*/
package netdevsim
