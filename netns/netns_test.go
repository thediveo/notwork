// Copyright 2023 Harald Albrecht.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package netns

import (
	"os"
	"runtime"
	"syscall"
	"time"

	"github.com/onsi/gomega/gleak/goroutine"
	"github.com/thediveo/notwork/dummy"
	"github.com/thediveo/notwork/link"
	"github.com/vishvananda/netlink"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/success"
)

var _ = Describe("transient network namespaces", Ordered, func() {

	BeforeEach(func() {
		if os.Getuid() != 0 {
			Skip("needs root")
		}
		goodfds := Filedescriptors()
		goodgos := Goroutines()
		DeferCleanup(func() {
			Eventually(Goroutines).Within(2 * time.Second).ProbeEvery(250 * time.Millisecond).
				ShouldNot(HaveLeaked(goodgos))
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
		})
	})

	It("returns a fd reference and cleans it up", func() {
		beforefds := []any{}
		for _, fd := range Filedescriptors() {
			beforefds = append(beforefds, fd.FdNo())
		}
		netnsfd := Current()
		afterfds := []any{}
		for _, fd := range Filedescriptors() {
			afterfds = append(afterfds, fd.FdNo())
		}
		Expect(afterfds).To(ConsistOf(append(beforefds, netnsfd)...))
	})

	It("creates, enters, and leaves a transient network namespace", func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		initialNetnsInfo := Successful(os.Stat("/proc/thread-self/ns/net"))

		By("creating and entering a new network namespace")
		f := EnterTransient()
		currentNetnsInfo := Successful(os.Stat("/proc/thread-self/ns/net"))
		Expect(initialNetnsInfo.Sys().(*syscall.Stat_t).Ino).NotTo(
			Equal(currentNetnsInfo.Sys().(*syscall.Stat_t).Ino))

		By("switching back into the original network namespace")
		Expect(f).NotTo(Panic())
		currentNetnsInfo = Successful(os.Stat("/proc/thread-self/ns/net"))
		Expect(initialNetnsInfo.Sys().(*syscall.Stat_t).Ino).To(
			Equal(currentNetnsInfo.Sys().(*syscall.Stat_t).Ino))
	})

	It("creates a transient network namespace without entering it", func() {
		homeIno := CurrentIno()
		Expect(homeIno).NotTo(BeZero())
		Expect(homeIno).To(Equal(Ino("/proc/self/ns/net")))

		netnsfd := NewTransient()
		netnsIno := Ino(netnsfd)
		Expect(netnsIno).NotTo(BeZero())
		Expect(netnsIno).NotTo(Equal(homeIno))
	})

	It("cannot enter an invalid network namespace", func() {
		var msg string
		g := NewGomega(func(message string, callerSkip ...int) {
			msg = message
		})
		execute(g, 0, func() {})
		Expect(msg).To(ContainSubstring("cannot switch into network namespace"))
	})

	It("executes a function in a different network namespace", func() {
		netnsfd := NewTransient()
		netnsIno := Ino(netnsfd)
		var currentnetnsIno uint64
		Execute(netnsfd, func() { currentnetnsIno = Ino("/proc/thread-self/ns/net") })
		Expect(currentnetnsIno).NotTo(BeZero())
		Expect(currentnetnsIno).To(Equal(netnsIno))
	})

	It("returns a netlink handle for a network namespace fd reference", func() {
		netnsfd := NewTransient()
		var dmy netlink.Link
		Execute(netnsfd, func() {
			dmy = dummy.NewTransient()
		})
		h := NewNetlinkHandle(netnsfd)
		defer h.Close()
		Expect(Successful(h.LinkByName(dmy.Attrs().Name)).Attrs().Name).
			To(Equal(dmy.Attrs().Name))
	})

	It("creates a dummy network interface in a different network namespace, obeying LinkAttrs.Namespace", func() {
		defer EnterTransient()()
		othernetnsfd := NewTransient()
		dmytempl := &netlink.Dummy{
			LinkAttrs: netlink.LinkAttrs{
				Namespace: netlink.NsFd(othernetnsfd),
			},
		}
		dmy := link.NewTransient(dmytempl, "dmy-")
		Expect(netlink.LinkByName(dmy.Attrs().Name)).Error().To(HaveOccurred())
		h := NewNetlinkHandle(othernetnsfd)
		defer h.Close()
		Expect(Successful(h.LinkByName(dmy.Attrs().Name)).Attrs().Name).
			To(Equal(dmy.Attrs().Name))
	})

	It("cannot create a MACVLAN when the parent/master isn't in the current network namespace", func() {
		// We need to create three separate new network namespaces in order to
		// exactly know their configuration: only a lo(nely) lo at the
		// beginning, with index 1. Otherwise, we could accidentally match an
		// existing parent network interface eligible as MACVLAN parent...
		parentnetnsfd := NewTransient()
		macvlannetnsfd := NewTransient()
		var dmy netlink.Link
		Execute(parentnetnsfd, func() {
			dmy = link.NewTransient(&netlink.Dummy{}, "dmy-")
		})

		defer EnterTransient()()
		// the dummy is in parentnetnsfd, probably with index 2 or so; here in
		// this new network namespace, we only have lo with index 1. Nothing
		// else, no dummy with index 1.
		Expect(netlink.LinkAdd(&netlink.Macvlan{
			LinkAttrs: netlink.LinkAttrs{
				ParentIndex: dmy.Attrs().Index,
				Namespace:   netlink.NsFd(macvlannetnsfd),
			},
		})).NotTo(Succeed())
	})

	It("creates a MACVLAN with a parent/master in a different network namespace", func() {
		macvlannetnsfd := NewTransient()
		defer EnterTransient()()
		dmy := link.NewTransient(&netlink.Dummy{}, "dmy-")
		Expect(dmy.Attrs().Index).NotTo(BeZero())
		mcvlan := link.NewTransient(&netlink.Macvlan{
			LinkAttrs: netlink.LinkAttrs{
				ParentIndex: dmy.Attrs().Index,
				Namespace:   netlink.NsFd(macvlannetnsfd),
			},
		}, "mc-")
		Expect(netlink.LinkByName(mcvlan.Attrs().Name)).Error().To(HaveOccurred())
		var l netlink.Link
		var err error
		Execute(macvlannetnsfd, func() {
			l, err = netlink.LinkByName(mcvlan.Attrs().Name)
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(l.Attrs().Name).To(Equal(mcvlan.Attrs().Name))
	})

	It("creates a VETH pair in two other network namespaces", func() {
		netnsA := NewTransient()
		netnsB := NewTransient()
		defer EnterTransient()()
		vethA := link.NewTransient(&netlink.Veth{
			LinkAttrs: netlink.LinkAttrs{
				Namespace: netlink.NsFd(netnsA),
			},
			PeerNamespace: netlink.NsFd(netnsB),
		}, "veth-")
		var err error
		Execute(netnsA, func() { _, err = netlink.LinkByName(vethA.Attrs().Name) })
		Expect(err).NotTo(HaveOccurred())
		Execute(netnsB, func() { _, err = netlink.LinkByName(vethA.(*netlink.Veth).PeerName) })
		Expect(err).NotTo(HaveOccurred())
	})

	When("running Ginkgo test leaf nodes", Ordered, func() {

		var gid uint64

		It("gets a first go routine", func() {
			gid = goroutine.Current().ID
			Expect(gid).NotTo(BeZero())
		})

		It("runs this unit test on a different go routine", func() {
			gid2 := goroutine.Current().ID
			Expect(gid2).NotTo(BeZero())
			Expect(gid2).NotTo(Equal(gid))
		})

	})

})
