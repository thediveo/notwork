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

package netlink

import (
	"os"
	"time"

	"github.com/thediveo/notwork/link"
	"github.com/thediveo/notwork/netns"
	"github.com/vishvananda/netlink"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/success"
)

var _ = Describe("netlink network namespace handling", func() {

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

	Context("vishvananda/netlink", func() {

		It("uses a new (RT)NETLINK socket for each package method call and closes upon returning from the call", func() {
			netnsfd := netns.NewTransient()

			Expect(len(Successful(netlink.LinkList()))).
				To(BeNumerically(">", 1), "your 'host' needs more than just 'lo'")

			netns.Execute(netnsfd, func() {
				Expect(Successful(netlink.LinkList())).
					To(HaveLen(1), "did reuse a socket when it shoudln't")
			})

			Expect(len(Successful(netlink.LinkList()))).
				To(BeNumerically(">", 1), "did reuse a socket when it shouldn't")
		})

		It("fails to fetch the index of a new link created in a different network namespace", func() {
			netnsfd := netns.NewTransient()

			l := &netlink.Dummy{
				LinkAttrs: netlink.LinkAttrs{
					Name:      link.RandomNifname("dmy-"),
					Namespace: netlink.NsFd(netnsfd),
				},
			}
			Expect(netlink.LinkAdd(l)).To(Succeed())
			Expect(l.Attrs().Index).To(BeZero(), "someone finally fixed netlink.LinkAdd!")

			Expect(Successful(
				netns.NewNetlinkHandle(netnsfd).LinkByName(l.Name))).To(
				HaveField("Attrs()", And(
					HaveField("Name", l.Attrs().Name),
					HaveField("Index", Not(BeZero())))))
		})

	})

})
