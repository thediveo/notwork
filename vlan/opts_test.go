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

package vlan

import (
	"github.com/vishvananda/netlink"

	"github.com/thediveo/notwork/link"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("VLAN configuration options", func() {

	It("configures VLAN", func() {
		l := &link.Link{Link: &netlink.Vlan{}}
		for _, opt := range []Opt{
			InNamespace(42),
			WithLinkNamespace(42),
			With8021Q(),
			WithReorderingHeaders(),
			WithGVRP(),
			WithMVRP(),
			WithLooseBinding(),
		} {
			Expect(opt(l)).To(Succeed())
		}
		Expect(InNamespace(-42)(l)).To(Succeed())
		Expect(l.Link).To(HaveField("Namespace", netlink.NsFd(-42)))
		Expect(WithLinkNamespace(-666)(l)).To(Succeed())
		Expect(l).To(HaveField("LinkNamespace", netlink.NsFd(-666)))
		Expect(l.Link).To(HaveField("VlanProtocol", netlink.VLAN_PROTOCOL_8021Q))
		Expect(l.Link).To(HaveField("ReorderHdr", neu(true)))
		Expect(l.Link).To(HaveField("Gvrp", neu(true)))
		Expect(l.Link).To(HaveField("Mvrp", neu(true)))
		Expect(l.Link).To(HaveField("LooseBinding", neu(true)))

		for _, opt := range []Opt{
			With8021AD(),
			WithoutReorderingHeaders(),
			WithoutGVRP(),
			WithoutMVRP(),
			WithTightBinding(),
		} {
			Expect(opt(l)).To(Succeed())
		}
		Expect(l.Link).To(HaveField("VlanProtocol", netlink.VLAN_PROTOCOL_8021AD))
		Expect(l.Link).To(HaveField("ReorderHdr", neu(false)))
		Expect(l.Link).To(HaveField("Gvrp", neu(false)))
		Expect(l.Link).To(HaveField("Mvrp", neu(false)))
		Expect(l.Link).To(HaveField("LooseBinding", neu(false)))
	})

})
