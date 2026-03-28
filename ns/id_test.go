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
	"math/rand"
	"os"
	"runtime"
	"syscall"
	"time"

	"github.com/thediveo/caps"
	"github.com/thediveo/spacetest/netns"
	"github.com/thediveo/testily/concur"
	"github.com/vishvananda/netlink"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/success"
)

var _ = Describe("transient network namespaces", Ordered, func() {

	BeforeEach(func() {
		goodfds := Filedescriptors()
		goodgos := Goroutines()
		DeferCleanup(func() {
			Eventually(Goroutines).Within(2 * time.Second).ProbeEvery(250 * time.Millisecond).
				ShouldNot(HaveLeaked(goodgos))
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
		})
	})

	When("root", Ordered, func() {

		BeforeAll(func() {
			if os.Getuid() != 0 {
				Skip("needs root")
			}
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

		Context("kernel", func() {

			It("throws a dedicated errno for duplicate nsids", func() {
				defer netns.EnterTransient()()
				netnsfd := netns.Current()
				netnsid := int(rand.Int31())
				Expect(Successful(netlink.GetNetNsIdByFd(netnsfd))).To(Equal(-1))
				Expect(netlink.SetNetNsIdByFd(netnsfd, netnsid)).To(Succeed())
				Expect(netlink.SetNetNsIdByFd(netnsfd, netnsid)).To(MatchError(syscall.EEXIST))
			})

			It("throws another errno when lacking capability", func() {
				defer netns.EnterTransient()()
				netnsfd := netns.Current()
				Expect(Successful(netlink.GetNetNsIdByFd(netnsfd))).To(Equal(-1))

				errch := concur.PassWhenGone(func() error {
					runtime.LockOSThread()

					// drop all effective capabilities on this throw-away thread.
					craps := Successful(caps.OfThisTask())
					craps.Effective.Clear()
					Expect(caps.SetForThisTask(craps)).To(Succeed())

					return InterceptGomegaFailure(func() { _ = ID(netnsfd) })
				})
				// !!! since InterceptGomegaFailures modifies the Default
				// gomega, we must not use the Default gomega here.
				NewGomega(Fail).Eventually(errch).WithTimeout(2 * time.Second).Should(
					Receive(MatchError(ContainSubstring("cannot assign any new nsid"))))
			})

		})

	})

	When("not root", Ordered, func() {

		BeforeAll(func() {
			if os.Getuid() == 0 {
				Skip("no root")
			}
		})

		Context("kernel", func() {

			It("rejects setting nsids", func() {
				Expect(netlink.SetNetNsIdByFd(netns.Current(), 12345 /* random number */)).Error().To(
					MatchError(syscall.EPERM))
			})

		})

	})

})
