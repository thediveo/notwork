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

package macvlan

import (
	"os"
	"time"

	"github.com/thediveo/notwork/dummy"
	"github.com/thediveo/notwork/netdevsim"
	"github.com/thediveo/notwork/netdevsim/ensure"
	"github.com/thediveo/notwork/netns"
	"github.com/vishvananda/netlink"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/success"
)

var _ = Describe("provides transient MACVLAN network interfaces", Ordered, func() {

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

	It("creates a MACVLAN with a dummy parent using legacy API", func() {
		defer netns.EnterTransient()()
		_ = CreateTransient(dummy.NewTransientUp())
	})

	It("creates a MACVLAN with a dummy parent and a configuration option", func() {
		defer netns.EnterTransient()()
		_ = NewTransient(dummy.NewTransientUp(), WithMode(netlink.MACVLAN_MODE_BRIDGE))
	})

	It("finds a hardware NIC in up state", func() {
		defer netns.EnterTransient()()
		if !ensure.Netdevsim() {
			Skip("cannot use netdevsim as suitable fake HW device")
		}
		// In order to be able to run this test inside a devcontainer, we need
		// some netdev that classifies as "HW" device, even if virtual. Virtual
		// netdevs such as "VETH" do not classify, unfortunately.
		_, fakehwndev := netdevsim.NewTransient()
		Expect(netlink.LinkSetUp(fakehwndev[0])).To(Succeed())
		parent := LocateHWParent()
		Expect(parent).NotTo(BeNil())
	})

	It("creates a MACVLAN with its parent in a different network namespace", func() {
		dmyNetnsfd := netns.NewTransient()
		dmy := dummy.NewTransient(dummy.InNamespace(dmyNetnsfd))

		destNetnsfd := netns.NewTransient()
		mcvlan := NewTransient(dmy,
			InNamespace(destNetnsfd),
			WithLinkNamespace(dmyNetnsfd))
		Expect(mcvlan.Attrs().Index).NotTo(BeZero())

		destnlh := netns.NewNetlinkHandle(destNetnsfd)
		Expect(Successful(destnlh.LinkByName(mcvlan.Attrs().Name))).To(
			HaveField("Attrs().Index", mcvlan.Attrs().Index))
	})

	DescribeTable("comparing links by OperState",
		func(lops1, lops2 int, expected int) {
			Expect(compareLinksByOperState(
				&netlink.Device{LinkAttrs: netlink.LinkAttrs{OperState: netlink.LinkOperState(lops1)}},
				&netlink.Device{LinkAttrs: netlink.LinkAttrs{OperState: netlink.LinkOperState(lops2)}},
			)).To(Equal(expected))
		},
		Entry("same state", netlink.OperUp, netlink.OperUp, 0),
		Entry("a before b", netlink.OperUp, netlink.OperUnknown, -1),
		Entry("a after b", netlink.OperUnknown, netlink.OperUp, 1),
	)

})
