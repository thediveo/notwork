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
	"time"

	"github.com/onsi/gomega/gleak/goroutine"
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

	When("getting netnsids", func() {

		It("sets it first, when necessary", func() {
			netnsfd := NewTransient()

			// There should not be any nsid for the transient network namespace yet,
			// when seen from our current network namespace.
			Expect(Successful(netlink.GetNetNsIdByFd(netnsfd))).To(Equal(-1))

			nsid := NsID(netnsfd)
			Expect(nsid).NotTo(Equal(-1))
			Expect(NsID(netnsfd)).To(Equal(nsid))
		})

		It("gets a netnsid by path", func() {
			orignetnsfd := Current()
			defer EnterTransient()()

			nsid := NsID(orignetnsfd)
			Expect(nsid).NotTo(Equal(-1))
			Expect(NsID("/proc/1/ns/net")).To(Equal(nsid))
		})

	})

	When("running Ginkgo test leaf nodes", Ordered, func() {

		var gid uint64

		It("gets a first go routine", func() {
			gid = goroutine.Current().ID
			Expect(gid).NotTo(BeZero())
		})

		It("runs this unit test leaf node on a different go routine than the first", func() {
			gid2 := goroutine.Current().ID
			Expect(gid2).NotTo(BeZero())
			Expect(gid2).NotTo(Equal(gid))
		})

	})

})
