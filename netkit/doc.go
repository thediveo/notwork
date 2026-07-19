/*
Package netkit helps with creating transient netkit network interfaces. It
leverages the [Ginkgo] testing framework and matching (erm, sic!) [Gomega]
matchers.

These netkit network interfaces are transient because they automatically get
removed at the end of the a test (spec, block/group, suite, et cetera) using
Ginkgo's [DeferCleanup].

# netkit Information

Unfortunately, netkit information is rather sparse. A good starting point might
be:

  - [Introduction to Linux Netkit interfaces – with a grain of eBPF]
    (Jean-Tiare Le Bigot's blog post 1/2).
  - [Creating a Linux “Yogurt-phone” – with netkit and a grain of eBPF]
    (Jean-Tiare Le Bigot's blog post 2/2).
  - Daniel Borkmann's [BPF Programmable Netdevice] talk slide deck.

Beyond these the [kernel drivers/net/netkit.c] source will be with you!

[Ginkgo]: https://github.com/onsi/ginkgo
[Gomega]: https://github.com/onsi/gomega
[DeferCleanup]: https://pkg.go.dev/github.com/onsi/ginkgo/v2#DeferCleanup
[Introduction to Linux Netkit interfaces – with a grain of eBPF]: https://blog.yadutaf.fr/2025/07/01/introduction-to-linux-netkit-interfaces-with-a-grain-of-ebpf/
[Creating a Linux “Yogurt-phone” – with netkit and a grain of eBPF]: https://blog.yadutaf.fr/2025/09/16/creating-a-yogurt-phone-with-netkit-ebpf/
[BPF Programmable Netdevice]: https://lpc.events/event/17/contributions/1581/attachments/1292/2602/lpc_netkit_devs.pdf
[kernel drivers/net/netkit.c]: https://elixir.bootlin.com/linux/v7.1.4/source/drivers/net/netkit.c
*/
package netkit
