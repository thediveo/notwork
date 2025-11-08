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

package netdevsim

import (
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/mdlayher/devlink"
	"github.com/thediveo/notwork/netdevsim/ensure"
	"github.com/thediveo/notwork/netns"
	"github.com/vishvananda/netlink"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/success"
)

var _ = Describe("netdevsim network interfaces", Ordered, func() {

	DescribeTable("finds the correct lowest netdevsim ID",
		func(sysfsroot string, expected int) {
			Expect(int(lowestUnusedID(sysfsroot))).To(Equal(expected))
		},
		Entry("none", "./_test/none", 0),
		Entry("0", "./_test/zero", 1),
		Entry("1", "./_test/one", 0),
		Entry("0-1-3", "./_test/zero-one-three", 2),
	)

	Context("messing around", func() {

		BeforeAll(func() {
			if !ensure.Netdevsim() {
				Skip("needs kernel module netdevsim")
			}
		})

		BeforeEach(func() {
			goodfds := Filedescriptors()
			goodgos := Goroutines()
			DeferCleanup(func() {
				Eventually(Goroutines).Within(2 * time.Second).ProbeEvery(100 * time.Millisecond).
					ShouldNot(HaveLeaked(goodgos))
				Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
			})
		})

		It("has netdevsim loaded", func() {
			Expect(HasNetdevsim()).To(BeTrue())
		})

		Context("listing port nifnames", func() {

			It("returns an empty list for a non-existing netdevsim device", func() {
				defer netns.EnterTransient()()
				cl := Successful(devlink.New())
				defer func() { _ = cl.Close() }()
				Expect(portNifnames(cl, 666)).To(BeEmpty())
			})

			It("returns a list of network interface names for the ports of a netdevsim device", func() {
				defer netns.EnterTransient()()

				id := lowestUnusedID("/")
				Expect(os.WriteFile(netdevsimRoot+"/new_device",
					fmt.Appendf(nil, "%d 2 1", id), 0)).To(Succeed())
				defer func() {
					Expect(os.WriteFile(netdevsimRoot+"/del_device",
						fmt.Appendf(nil, "%d", id), 0)).To(Succeed())
				}()

				cl := Successful(devlink.New())
				defer func() { _ = cl.Close() }()
				Expect(portNifnames(cl, id)).To(Equal([]string{"eth0", "eth1"}))
			})

		})

		Context("creating netdevsims", func() {

			It("creates a one-port netdevsim", func() {
				defer netns.EnterTransient()()

				_, portnifs := NewTransient(
					WithPorts(1),
					WithRxTxQueueCountEach(1))
				Expect(portnifs).To(HaveLen(1))
				Expect(portnifs[0]).To(And(
					HaveField("Attrs().Name", HavePrefix(NetdevsimPrefix)),
					HaveField("Type()", "device")))
				Expect(portnifs[0].Attrs().Name).To(HavePrefix(NetdevsimPrefix))
				Expect(Successful(net.Interfaces())).To(
					ContainElement(HaveField("Name", portnifs[0].Attrs().Name)))
			})

			It("creates a netdevsim with VFs", func() {
				defer netns.EnterTransient()()

				_, portnifs := NewTransient(
					WithPorts(1),
					WithRxTxQueueCountEach(1),
					WithMaxVFs(4))
				pf := Successful(netlink.LinkByName(portnifs[0].Attrs().Name))
				Expect(pf.Attrs().Vfs).To(HaveLen(4))
			})

			It("creates a multi-port netdevsim", func() {
				defer netns.EnterTransient()()

				_, portnifs := NewTransient(WithPorts(3), WithRxTxQueueCountEach(1))
				Expect(portnifs).To(HaveLen(3))
				Expect(portnifs).To(HaveEach(HaveField("Attrs().Name", HavePrefix(NetdevsimPrefix))))
			})

			It("creates a one-port netdevsim in a different network namespace", func() {
				netnsfd := netns.NewTransient()

				_, portnifs := NewTransient(
					WithPorts(1),
					WithRxTxQueueCountEach(1),
					InNamespace(netnsfd))
				netns.Execute(netnsfd, func() {
					Expect(netlink.LinkByName(portnifs[0].Attrs().Name)).Error().NotTo(HaveOccurred())
				})
				Expect(netlink.LinkByName(portnifs[0].Attrs().Name)).Error().To(HaveOccurred())
			})

			It("cannot create two netdevsims with the same ID", func() {
				defer netns.EnterTransient()()

				id := lowestUnusedID("/")
				_, portnifs := NewTransient(WithID(id))
				Expect(portnifs).To(HaveLen(1))

				oldfail := fail
				defer func() { fail = oldfail }()
				var msg string
				fail = func(message string, callerSkip ...int) { msg = message; panic(message) }
				Expect(func() {
					_, _ = NewTransient(WithID(id))
				}).To(Panic())
				fail = oldfail
				Expect(msg).To(ContainSubstring(fmt.Sprintf("cannot create a netdevsim with ID %d", id)))
			})

		})

		Context("linking netdevsim interfaces", Ordered, func() {

			BeforeAll(func() {
				_, err := os.Stat(netdevsimRoot + "/link_device")
				if errors.Is(err, os.ErrNotExist) {
					Skip("needs kernel 6.9+")
				}
				Expect(err).To(Succeed())
			})

			It("reject invalid network namespace references", func() {
				Expect(linkFds(&netlink.GenericLink{
					LinkAttrs: netlink.LinkAttrs{
						Namespace: netlink.NsFd(666),
					},
				})).Error().To(HaveOccurred())

				Expect(linkFds(&netlink.GenericLink{
					LinkAttrs: netlink.LinkAttrs{
						Namespace: netlink.NsFd(0),
					},
				})).Error().To(HaveOccurred())
			})

			It("rejects an invalid name", func() {
				Expect(linkFds(&netlink.GenericLink{
					LinkAttrs: netlink.LinkAttrs{
						Name: "%NOT-EXIST%",
					},
				})).Error().To(HaveOccurred())

				netnsfd := netns.NewTransient()
				Expect(linkFds(&netlink.GenericLink{
					LinkAttrs: netlink.LinkAttrs{
						Name:      "%NOT-EXIST%",
						Namespace: netlink.NsFd(netnsfd),
					},
				})).Error().To(HaveOccurred())
			})

			It("links and unlinks two peers in the current netns", func() {
				defer netns.EnterTransient()()

				_, portnifs1 := NewTransient()
				_, portnifs2 := NewTransient()
				Link(portnifs1[0], portnifs2[0])
				Unlink(portnifs2[0])
			})

			It("links and unlinks two peers in two different network namespaces", func() {
				netnsfd1 := netns.NewTransient()
				_, portnifs1 := NewTransient(InNamespace(netnsfd1))
				Expect(netlink.LinkByName(portnifs1[0].Attrs().Name)).Error().To(HaveOccurred())
				nlh1 := netns.NewNetlinkHandle(netnsfd1)
				Expect(nlh1.LinkByName(portnifs1[0].Attrs().Name)).Error().NotTo(HaveOccurred())

				netnsfd2 := netns.NewTransient()
				_, portnifs2 := NewTransient(InNamespace(netnsfd2))
				Expect(netlink.LinkByName(portnifs2[0].Attrs().Name)).Error().To(HaveOccurred())
				nlh2 := netns.NewNetlinkHandle(netnsfd2)
				Expect(nlh2.LinkByName(portnifs2[0].Attrs().Name)).Error().NotTo(HaveOccurred())

				portnifs1[0].Attrs().Namespace = netlink.NsFd(netnsfd1)
				portnifs2[0].Attrs().Namespace = netlink.NsFd(netnsfd2)

				Link(portnifs1[0], portnifs2[0])
				Unlink(portnifs1[0])
			})

		})

	})

})
