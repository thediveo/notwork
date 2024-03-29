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
	"github.com/vishvananda/netlink"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
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
			Eventually(Filedescriptors).Within(2 * time.Second).ProbeEvery(250 * time.Millisecond).
				ShouldNot(HaveLeakedFds(goodfds))
		})
	})

	It("creates a MACVLAN with a dummy parent", func() {
		_ = CreateTransient(dummy.NewTransientUp())
	})

	It("finds a hardware NIC in up state", func() {
		parent := LocateHWParent()
		Expect(parent).NotTo(BeNil())
	})

	When("using options", func() {

		It("configures a different netns", func() {
			l := &netlink.Macvlan{}
			Expect(InNamespace(-42)(l)).To(Succeed())
			Expect(l.Namespace).To(Equal(netlink.NsFd(-42)))
		})

		It("configures the mode", func() {
			l := &netlink.Macvlan{}
			Expect(WithMode(netlink.MACVLAN_MODE_VEPA)(l)).To(Succeed())
			Expect(l.Mode).To(Equal(netlink.MACVLAN_MODE_VEPA))
		})

	})

})
