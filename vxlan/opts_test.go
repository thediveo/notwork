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

package vxlan

import (
	"github.com/thediveo/notwork/link"
	"github.com/vishvananda/netlink"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("VXLAN configuration options", func() {

	It("configures VXLAN", func() {
		l := &link.Link{Link: &netlink.Vxlan{}}
		for _, opt := range []Opt{
			InNamespace(42),
			WithID(666),
			WithDestinationPort(12345),
			WithTTL(123),
			WithSourcePorts(66, 6),
		} {
			Expect(opt(l)).To(Succeed())
		}
		Expect(InNamespace(-42)(l)).To(Succeed())
		Expect(l.Link).To(HaveField("Namespace", netlink.NsFd(-42)))
		Expect(l.Link).To(HaveField("VxlanId", 666))
		Expect(l.Link).To(HaveField("Port", 12345))
		Expect(l.Link).To(HaveField("TTL", 123))
		Expect(l.Link).To(HaveField("PortLow", 6))
		Expect(l.Link).To(HaveField("PortHigh", 66))
	})

})
