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
	"github.com/thediveo/notwork/netns"
	"github.com/vishvananda/netlink"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/success"
)

var _ = Describe("creates netdevsim network interfaces", Ordered, func() {

	BeforeAll(func() {
		if os.Getuid() != 0 {
			Skip("needs root")
		}
		if !HasNetdevsim() {
			Skip("needs loaded kernel module netdevsim")
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

	It("finds an available netdevsim ID", func() {
		// Prohibit systemdoh-notworkdoh from interfering with our
		// netdevsims.
		defer netns.EnterTransient()()

		By("getting a first available ID")
		id1 := Successful(availableID())
		// Expecting the same ID to be found again...
		Expect(availableID()).To(Equal(id1))

		By("creating a first netdevsim instance")
		Expect(os.WriteFile(netdevsimRoot+"/new_device",
			[]byte(fmt.Sprintf("%d 1 1", id1)), 0)).To(Succeed())
		defer func() {
			Expect(os.WriteFile(netdevsimRoot+"/del_device",
				[]byte(fmt.Sprintf("%d", id1)), 0)).To(Succeed())
		}()

		By("getting a second, different available ID")
		id2 := Successful(availableID())
		Expect(id2).NotTo(Equal(id1))
	})

	Context("listing port nifnames", func() {

		It("returns an empty list for a non-existing netdevsim device", func() {
			defer netns.EnterTransient()()
			cl := Successful(devlink.New())
			defer cl.Close()
			Expect(portNifnames(cl, 666)).To(BeEmpty())
		})

		It("returns a list of network interface names for the ports of a netdevsim device", func() {
			defer netns.EnterTransient()()

			id := Successful(availableID())
			Expect(os.WriteFile(netdevsimRoot+"/new_device",
				[]byte(fmt.Sprintf("%d 2 1", id)), 0)).To(Succeed())
			defer func() {
				Expect(os.WriteFile(netdevsimRoot+"/del_device",
					[]byte(fmt.Sprintf("%d", id)), 0)).To(Succeed())
			}()

			cl := Successful(devlink.New())
			defer cl.Close()
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

	})

	Context("linking netdevsim interfaces", Ordered, func() {

		BeforeAll(func() {
			_, err := os.Stat(netdevsimRoot + "/link_device")
			if errors.Is(err, os.ErrNotExist) {
				Skip("needs kernel 6.9+")
			}
			Expect(err).To(Succeed())
		})

		It("links two peers in the current netns", func() {
			defer netns.EnterTransient()()

			_, portnifs1 := NewTransient()
			_, portnifs2 := NewTransient()
			Link(portnifs1[0], portnifs2[0])
			Unlink(portnifs2[0])
		})

	})

})
