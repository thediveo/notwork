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

package veth

import (
	"os"

	"github.com/thediveo/notwork/netns"
	"github.com/vishvananda/netlink"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/thediveo/success"
)

var _ = Describe("provides transient VETH network interface pairs", Ordered, func() {

	BeforeEach(func() {
		if os.Getuid() != 0 {
			Skip("needs root")
		}
	})

	It("creates a VETH pair in the same network namespace", func() {
		dupond, dupont := NewTransient()
		Expect(dupond).NotTo(BeNil())
		Expect(dupont).NotTo(BeNil())
		Expect(dupond.Attrs().Name).To(HavePrefix(VethPrefix))
		Expect(dupont.Attrs().Name).To(HavePrefix(VethPrefix))
		Expect(dupond.Attrs().Name).NotTo(Equal(dupont.Attrs().Name))
		// Check that the network interface pair was in fact created.
		ql := Successful(netlink.LinkByIndex(dupond.Attrs().Index))
		Expect(ql.Attrs().OperState).NotTo(Equal(netlink.OperDown))
		ql = Successful(netlink.LinkByIndex(dupont.Attrs().Index))
		Expect(ql.Attrs().OperState).NotTo(Equal(netlink.OperDown))
	})

	It("creates a VETH pair in the different network namespaces", func() {
		dupondNetns := netns.NewTransient()
		dupontNetns := netns.NewTransient()
		var dupond, dupont netlink.Link
		netns.Execute(dupondNetns, func() {
			dupond, dupont = NewTransient(WithPeerNamespace(dupontNetns))
		})
		Expect(dupond).NotTo(BeNil())
		Expect(dupont).NotTo(BeNil())
		Expect(dupond.Attrs().Name).To(HavePrefix(VethPrefix))
		Expect(dupont.Attrs().Name).To(HavePrefix(VethPrefix))
		Expect(dupond.Attrs().Name).NotTo(Equal(dupont.Attrs().Name))
		Expect(netlink.LinkByName(dupond.Attrs().Name)).Error().To(HaveOccurred())
		Expect(netlink.LinkByName(dupont.Attrs().Name)).Error().To(HaveOccurred())
	})

})
