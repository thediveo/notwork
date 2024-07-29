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

package netns

import (
	"os"
	"time"

	"github.com/vishvananda/netlink"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/success"
)

const (
	dummyNifName = "dmy-random-name"
)

var _ = Describe("netlink handles", func() {

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

	It("fails with invalid network namespace fd", func() {
		var r any
		func() {
			defer func() { r = recover() }()
			g := NewGomega(func(message string, callerSkip ...int) {
				panic(message)
			})
			_ = newNetlinkHandle(g, 0)
		}()
		Expect(r).To(ContainSubstring("failed to set into network namespace 0 while creating netlink socket: invalid argument"))
	})

	It("correctly connects to a transient network namespace", func() {
		netnsfd := NewTransient()
		dmy := &netlink.Dummy{
			LinkAttrs: netlink.LinkAttrs{
				Name:      dummyNifName, // avoid circular import of link package
				Namespace: netlink.NsFd(netnsfd),
			},
		}
		Expect(netlink.LinkAdd(dmy)).To(Succeed())
		Expect(dmy.Index).To(BeZero(), "someone finally fixed netlink.LinkAdd")

		netnsh := NewNetlinkHandle(netnsfd)
		Expect(Successful(netnsh.LinkList())).
			To(ConsistOf(
				HaveField("LinkAttrs.Name", "lo"),
				HaveField("LinkAttrs", And(
					HaveField("Name", dummyNifName),
					HaveField("Index", Not(BeZero()))))),
				"missing dummy link, it's somewhere else?!")

		hostnetnsh := NewNetlinkHandle(Current())
		Expect(Successful(hostnetnsh.LinkList())).
			NotTo(ContainElement(HaveField("LinkAttrs", And(
				HaveField("Name", dummyNifName),
				HaveField("Index", Not(BeZero()))))),
				"host netns contains dummy link when it shoudln't")
	})

})
