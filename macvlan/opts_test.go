// Copyright 2024 Harald Albrecht.
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
	"github.com/thediveo/notwork/link"
	"github.com/vishvananda/netlink"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MACVLAN configuration options", func() {

	It("configures a different network namespace", func() {
		l := &link.Link{Link: &netlink.Macvlan{}}
		Expect(InNamespace(-42)(l)).To(Succeed())
		Expect(l.Link).To(HaveField("Namespace", netlink.NsFd(-42)))
	})

	It("configures a different link network namespace", func() {
		l := &link.Link{Link: &netlink.Macvlan{}}
		Expect(WithLinkNamespace(-42)(l)).To(Succeed())
		Expect(l.LinkNamespace).To(Equal(netlink.NsFd(-42)))
	})

	It("configures the mode", func() {
		l := &link.Link{Link: &netlink.Macvlan{}}
		Expect(WithMode(netlink.MACVLAN_MODE_VEPA)(l)).To(Succeed())
		Expect(l.Link).To(HaveField("Mode", netlink.MACVLAN_MODE_VEPA))
	})

})
