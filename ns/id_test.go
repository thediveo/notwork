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

package ns

import (
	"os"
	"time"

	"github.com/thediveo/spacetest/netns"
	"github.com/vishvananda/netlink"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/success"
)

var _ = Describe("transient network namespaces", Ordered, func() {

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

	It("sets it first, when necessary", func() {
		netnsfd := netns.NewTransient()

		// There should not be any nsid for the transient network namespace yet,
		// when seen from our current network namespace.
		Expect(Successful(netlink.GetNetNsIdByFd(netnsfd))).To(Equal(-1))

		nsid := ID(netnsfd)
		Expect(nsid).NotTo(Equal(-1))
		Expect(ID(netnsfd)).To(Equal(nsid))
	})

	It("gets a netnsid by path", func() {
		orignetnsfd := netns.Current()
		defer netns.EnterTransient()()

		nsid := ID(orignetnsfd)
		Expect(nsid).NotTo(Equal(-1))
		Expect(ID("/proc/1/ns/net")).To(Equal(nsid))
	})

})
