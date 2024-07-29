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
	"time"

	"github.com/thediveo/notwork/netns"
	"github.com/vishvananda/netlink"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/success"
)

var _ = Describe("provides transient VETH network interface pairs", Ordered, func() {

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

	It("creates a VETH pair in the same current transient network namespace", func() {
		defer netns.EnterTransient()()

		dupond, dupont := NewTransient()
		Expect(dupond).NotTo(BeNil())
		Expect(dupont).NotTo(BeNil())
		Expect(dupond.Attrs().Name).To(HavePrefix(VethPrefix))
		Expect(dupont.Attrs().Name).To(HavePrefix(VethPrefix))
		Expect(dupond.Attrs().Name).NotTo(Equal(dupont.Attrs().Name))
		Expect(dupond.Attrs().Index).NotTo(BeZero())
		Expect(dupont.Attrs().Index).NotTo(BeZero())
		// Check that the network interface pair was in fact created.
		ql := Successful(netlink.LinkByName(dupond.Attrs().Name))
		Expect(ql.Attrs().OperState).NotTo(Equal(netlink.OperDown))
		ql = Successful(netlink.LinkByName(dupont.Attrs().Name))
		Expect(ql.Attrs().OperState).NotTo(Equal(netlink.OperDown))
	})

	It("creates a VETH pair with the first end in a different network namespace, but with the peer in the current(!) network namespace", func() {
		netnsfd := netns.NewTransient()

		dupond, dupont := NewTransient(InNamespace(netnsfd))
		Expect(dupond).NotTo(BeNil())
		Expect(dupont).NotTo(BeNil())
		Expect(dupond.Attrs().Name).To(HavePrefix(VethPrefix))
		Expect(dupont.Attrs().Name).To(HavePrefix(VethPrefix))
		Expect(dupond.Attrs().Name).NotTo(Equal(dupont.Attrs().Name))
		Expect(dupond.Attrs().Index).NotTo(BeZero())
		Expect(dupont.Attrs().Index).NotTo(BeZero())
		Expect(netlink.LinkByName(dupond.Attrs().Name)).Error().To(HaveOccurred())
		Expect(netlink.LinkByName(dupont.Attrs().Name)).Error().NotTo(HaveOccurred())
	})

	It("creates a VETH pair in the same other network namespace", func() {
		netnsfd := netns.NewTransient()

		dupond, dupont := NewTransient(
			InNamespace(netnsfd), WithPeerNamespace(netnsfd))
		Expect(dupond).NotTo(BeNil())
		Expect(dupont).NotTo(BeNil())
		Expect(dupond.Attrs().Name).To(HavePrefix(VethPrefix))
		Expect(dupont.Attrs().Name).To(HavePrefix(VethPrefix))
		Expect(dupond.Attrs().Name).NotTo(Equal(dupont.Attrs().Name))
		Expect(dupond.Attrs().Index).NotTo(BeZero())
		Expect(dupont.Attrs().Index).NotTo(BeZero())
		Expect(netlink.LinkByName(dupond.Attrs().Name)).Error().To(HaveOccurred())
		Expect(netlink.LinkByName(dupont.Attrs().Name)).Error().To(HaveOccurred())
	})

	It("creates a VETH pair in the two different network namespace", func() {
		dupondNetnsfd := netns.NewTransient()
		dupontNetnsfd := netns.NewTransient()

		dupond, dupont := NewTransient(
			InNamespace(dupondNetnsfd),
			WithPeerNamespace(dupontNetnsfd))
		Expect(dupond).NotTo(BeNil())
		Expect(dupont).NotTo(BeNil())
		Expect(dupond.Attrs().Name).To(HavePrefix(VethPrefix))
		Expect(dupont.Attrs().Name).To(HavePrefix(VethPrefix))
		Expect(dupond.Attrs().Name).NotTo(Equal(dupont.Attrs().Name))
		Expect(dupond.Attrs().Index).NotTo(BeZero())
		Expect(dupont.Attrs().Index).NotTo(BeZero())
		Expect(netlink.LinkByName(dupond.Attrs().Name)).Error().To(HaveOccurred())
		Expect(netlink.LinkByName(dupont.Attrs().Name)).Error().To(HaveOccurred())
	})

})
