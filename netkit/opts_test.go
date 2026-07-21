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
	"github.com/vishvananda/netlink"

	"github.com/thediveo/notwork/link"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("netkit configuration options", func() {

	It("configures netkit", func() {
		l := &link.Link{Link: &Netkit{
			Netkit: netlink.Netkit{},
			Peer:   netlink.LinkAttrs{},
		}}
		for _, opt := range []Opt{
			InNamespace(42),
			WithPeerNamespace(666),
			WithIPMode(),
			WithPolicy(netlink.NETKIT_POLICY_BLACKHOLE),
			WithScrub(netlink.NETKIT_SCRUB_DEFAULT),
			WithHeadroom(42),
			WithTailroom(666),
			WithRxQueues(2),
			WithTxQueues(3),
			WithPeerRxQueues(4),
			WithPeerTxQueues(5),
		} {
			Expect(opt(l)).To(Succeed())
		}
		Expect(l.Link).To(HaveField("Attrs().Namespace", netlink.NsFd(42)))
		Expect(l.Link).To(HaveField("Peer.Namespace", netlink.NsFd(666)))
		Expect(l.Link).To(HaveField("Mode", netlink.NETKIT_MODE_L3))
		Expect(l.Link).To(HaveField("Policy", netlink.NETKIT_POLICY_BLACKHOLE))
		Expect(l.Link).To(HaveField("Scrub", netlink.NETKIT_SCRUB_DEFAULT))
		Expect(l.Link).To(HaveField("DesiredHeadroom", uint16(42)))
		Expect(l.Link).To(HaveField("DesiredTailroom", uint16(666)))
		Expect(l.Link).To(HaveField("NumRxQueues", 2))
		Expect(l.Link).To(HaveField("NumTxQueues", 3))
		Expect(l.Link).To(HaveField("Peer.NumRxQueues", 4))
		Expect(l.Link).To(HaveField("Peer.NumTxQueues", 5))

		for _, opt := range []Opt{
			WithL2Mode(),
			WithPeerPolicy(netlink.NETKIT_POLICY_BLACKHOLE),
			WithPeerScrub(netlink.NETKIT_SCRUB_DEFAULT),
		} {
			Expect(opt(l)).To(Succeed())
		}
		Expect(l.Link).To(HaveField("Mode", netlink.NETKIT_MODE_L2))
		Expect(l.Link).To(HaveField("PeerPolicy", netlink.NETKIT_POLICY_BLACKHOLE))
		Expect(l.Link).To(HaveField("PeerScrub", netlink.NETKIT_SCRUB_DEFAULT))
	})

})
