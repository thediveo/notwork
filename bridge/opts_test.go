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

package bridge

import (
	"time"

	"github.com/vishvananda/netlink"

	"github.com/thediveo/notwork/internal/auxvalues"
	"github.com/thediveo/notwork/internal/neu"
	"github.com/thediveo/notwork/link"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("bridge configuration options", func() {

	It("configures bridge", func() {
		l := &link.Link{Link: &netlink.Bridge{}}
		for _, opt := range []Opt{
			InNamespace(42),
			WithLinkNamespace(666),
			WithVLANFiltering(),
			WithMulticastSnooping(),
			WithAgeing(3*time.Second + 1*time.Millisecond),
			WithHello(5*time.Second + 2*time.Millisecond),
		} {
			Expect(opt(l)).To(Succeed())
		}
		Expect(l.Link).To(HaveField("Namespace", netlink.NsFd(42)))
		Expect(l).To(HaveField("LinkNamespace", netlink.NsFd(666)))
		Expect(l.Link).To(HaveField("VlanFiltering", neu.Value(true)))
		Expect(l.Link).To(HaveField("MulticastSnooping", neu.Value(true)))
		Expect(l.Link).To(HaveField("AgeingTime", neu.Value(uint32(3*auxvalues.CLKTCK()))))
		Expect(l.Link).To(HaveField("HelloTime", neu.Value(uint32(5*auxvalues.CLKTCK()))))

		Expect(WithoutVLANFiltering()(l)).To(Succeed())
		Expect(l.Link).To(HaveField("VlanFiltering", neu.Value(false)))
		Expect(WithoutMulticastSnooping()(l)).To(Succeed())
		Expect(l.Link).To(HaveField("MulticastSnooping", neu.Value(false)))

		Expect(WithAgeing(-1 * time.Second)(l)).NotTo(Succeed())
		Expect(WithHello(-1 * time.Second)(l)).NotTo(Succeed())
	})

})
