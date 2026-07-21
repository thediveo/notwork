// Copyright 2026 Harald Albrecht.
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

package netkit

import (
	"os"
	"time"

	"github.com/thediveo/spacetest/netns"
	"github.com/vishvananda/netlink"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/success"
)

var _ = Describe("provides transient netkit network interface pairs", Ordered, func() {

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

	It("creates a netkit pair in the same current transient network namespace", func() {
		defer netns.EnterTransient()()

		primary, peer := NewTransient(
			WithRxQueues(4), WithTxQueues(4),
			WithPeerRxQueues(4), WithPeerTxQueues(4))
		Expect(primary).NotTo(BeNil())
		Expect(peer).NotTo(BeNil())
		Expect(primary.Attrs().Name).To(HavePrefix(NetkitPrefix))
		Expect(peer.Attrs().Name).To(HavePrefix(NetkitPrefix))
		Expect(primary.Attrs().Name).NotTo(Equal(peer.Attrs().Name))
		Expect(primary.Attrs().Index).NotTo(BeZero())
		Expect(peer.Attrs().Index).NotTo(BeZero())
		// Check that the network interface pair was in fact created.
		ql := Successful(netlink.LinkByName(primary.Attrs().Name))
		Expect(ql.Attrs().OperState).To(BeEquivalentTo(netlink.OperDown))
		Expect(ql).To(HaveField("IsPrimary()", BeTrue()))
		ql = Successful(netlink.LinkByName(peer.Attrs().Name))
		Expect(ql.Attrs().OperState).To(BeEquivalentTo(netlink.OperDown))
		Expect(ql).To(HaveField("IsPrimary()", BeFalse()))
	})

	It("creates a netkit pair in the two different network namespace", func() {
		nkhostNetnsfd := netns.NewTransient()
		nkpeerNetnsfd := netns.NewTransient()

		primary, peer := NewTransient(InNamespace(nkhostNetnsfd),
			WithPeerNamespace(nkpeerNetnsfd))
		Expect(primary).NotTo(BeNil())
		Expect(peer).NotTo(BeNil())
		Expect(primary.Attrs().Name).To(HavePrefix(NetkitPrefix))
		Expect(peer.Attrs().Name).To(HavePrefix(NetkitPrefix))
		Expect(primary.Attrs().Name).NotTo(Equal(peer.Attrs().Name))
		Expect(primary.Attrs().Index).NotTo(BeZero())
		Expect(peer.Attrs().Index).NotTo(BeZero())
		Expect(netlink.LinkByName(primary.Attrs().Name)).Error().To(HaveOccurred())
		Expect(netlink.LinkByName(peer.Attrs().Name)).Error().To(HaveOccurred())
	})

})
