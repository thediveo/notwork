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

package link

import (
	"os"
	"runtime"
	"time"

	"github.com/thediveo/notwork/netns"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/thediveo/success"
)

var _ = Describe("creates transient network interfaces", func() {

	const dummyPrefix = "dmy-"

	BeforeEach(func() {
		if os.Getuid() != 0 {
			Skip("needs root")
		}
	})

	When("creating random network interface names", func() {

		It("creates a random name with a prefix", func() {
			const prefix = "prefix-"
			nifname := base62Nifname(prefix)
			Expect(nifname).To(HaveLen(maxNifnameLen))
			Expect(nifname).To(HavePrefix(prefix))
		})

		It("respects length restrictions", func() {
			oldfail := fail
			var msg string
			fail = func(message string, callerSkip ...int) {
				msg = message
				panic("canary")
			}
			Expect(func() {
				_ = base62Nifname("a-very-long-prefix-that-breaks-the-box")
			}).To(PanicWith("canary"))
			fail = oldfail
			Expect(msg).To(HavePrefix("cannot create random network interface name"))
		})

	})

	Context("creating transient network interfaces and registering them for destruction", Ordered, func() {

		var dl netlink.Link

		It("creates a transient network interface with a random suffix ", func() {
			templ := &netlink.Dummy{}
			dl = NewTransient(templ, dummyPrefix)
			Expect(dl.Attrs().Name).NotTo(BeEmpty())
			Expect(dl.Attrs().Name).NotTo(Equal(templ.Name), "missing random suffix")

			// Check that the network interface was in fact created.
			ql := Successful(netlink.LinkByName(dl.Attrs().Name))
			Expect(ql.Attrs().Index).To(Equal(dl.Attrs().Index))
		})

		It("has no transient network interface anymore", func() {
			// The network interfaces created in the above step should have been
			// deleted by now.
			Expect(netlink.LinkByName(dl.Attrs().Name)).Error().To(
				MatchError("Link not found"))
		})

		It("properly creates a transient network interface in a different network namespace", func() {
			netnsfd := netns.NewTransient()
			templ := &netlink.Dummy{
				LinkAttrs: netlink.LinkAttrs{
					Namespace: netlink.NsFd(netnsfd),
				},
			}
			var dl netlink.Link
			DeferCleanup(func() {
				var err error
				netns.Execute(netnsfd, func() {
					_, err = netlink.LinkByName(dl.Attrs().Name)
				})
				Expect(err).To(HaveOccurred(), "network interface wasn't removed")
			})
			dl = NewTransient(templ, dummyPrefix)

			Expect(netlink.LinkByName(dl.Attrs().Name)).Error().To(HaveOccurred())
			var netnsdl netlink.Link
			var err error
			netns.Execute(netnsfd, func() {
				netnsdl, err = netlink.LinkByName(dl.Attrs().Name)
			})
			Expect(err).NotTo(HaveOccurred(), "network interface went missing")
			Expect(netnsdl.Attrs().Name).To(Equal(dl.Attrs().Name))
		})

		It("rejects invalid network namespace references", func() {
			templ := &netlink.Dummy{
				LinkAttrs: netlink.LinkAttrs{
					Namespace: "42",
				},
			}
			oldfail := fail
			var msg string
			fail = func(message string, callerSkip ...int) {
				msg = message
				panic("canary")
			}
			Expect(func() {
				_ = NewTransient(templ, dummyPrefix)
			}).To(PanicWith("canary"))
			fail = oldfail
			Expect(msg).To(Equal("link.Attrs().Namespace reference must be nil or a netlink.NsFd"))
		})

	})

	It("creates two independent transient network interfaces", func() {
		templ := &netlink.Dummy{}
		dl1 := NewTransient(templ, dummyPrefix)
		dl2 := NewTransient(templ, dummyPrefix)
		Expect(dl1.Attrs().Name).NotTo(Equal(dl2.Attrs().Name))
	})

	It("fails the spec on failure", func() {
		oldfail := fail
		var msg string
		fail = func(message string, callerSkip ...int) {
			msg = message
			panic("canary")
		}
		Expect(func() {
			_ = NewTransient(&netlink.Macvlan{ /* no parent */ }, "ohno-")
		}).To(PanicWith("canary"))
		fail = oldfail
		Expect(msg).To(MatchRegexp(`cannot create a transient network interface .*, reason: invalid argument`))
	})

	It("removes a transient network interface in a different network namespace", func() {
		if os.Geteuid() != 0 {
			Skip("needs root")
		}

		By("creating a new network namespace")
		runtime.LockOSThread()
		netnsfd := netns.Current()
		DeferCleanup(func() {
			if err := unix.Setns(netnsfd, 0); err != nil {
				panic(err)
			}
			runtime.UnlockOSThread()
		})
		Expect(unix.Unshare(unix.CLONE_NEWNET)).To(Succeed())

		By("creating a transient network interface")
		_ = NewTransient(&netlink.Dummy{}, dummyPrefix)
	})

	Context("ensuring that network interfaces are operationally up", func() {

		It("doesn't accept multiple optional durations", func() {
			var r any
			func() {
				defer func() {
					r = recover()
				}()
				EnsureUp(nil, time.Millisecond, time.Millisecond)
			}()
			Expect(r).To(ContainSubstring("single optional maximum wait duration"))
		})

		It("times out", func() {
			// work around circular import
			dmy := NewTransient(&netlink.Dummy{}, "tst-")
			var msg string
			g := NewGomega(func(message string, callerSkip ...int) {
				msg = message
			})
			ensureUp(g, dmy, 100*time.Millisecond)
			Expect(msg).To(ContainSubstring("Timed out after 0."))

			ensureUp(g, dmy)
			Expect(msg).To(ContainSubstring("Timed out after 2."))
		})

		It("waits for operationally up", func() {
			// work around circular import
			dmy := NewTransient(&netlink.Dummy{}, "tst-")
			mcvlan := NewTransient(&netlink.Macvlan{
				LinkAttrs: netlink.LinkAttrs{
					ParentIndex: dmy.Attrs().Index,
				},
				Mode: netlink.MACVLAN_MODE_BRIDGE,
			}, "tst-")
			var msg string
			g := NewGomega(func(message string, callerSkip ...int) {
				msg = message
			})
			_ = netlink.LinkSetUp(mcvlan)
			ensureUp(g, mcvlan)
			Expect(msg).To(BeEmpty())

		})

	})

})
