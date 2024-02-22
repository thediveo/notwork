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

package netns

import (
	"os"
	"runtime"
	"syscall"
	"time"

	"github.com/thediveo/notwork/dummy"
	"github.com/vishvananda/netlink"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/success"
)

var _ = Describe("transient network namespaces", Ordered, func() {

	BeforeEach(func() {
		if os.Getuid() != 0 {
			Skip("needs root")
		}
		goodfds := Filedescriptors()
		DeferCleanup(func() {
			Eventually(Filedescriptors).Within(2 * time.Second).ProbeEvery(250 * time.Millisecond).
				ShouldNot(HaveLeakedFds(goodfds))
		})
	})

	It("creates, enters, and leaves a transient network namespace", func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		initialNetnsInfo := Successful(os.Stat("/proc/thread-self/ns/net"))

		By("creating and entering a new network namespace")
		f := EnterTransient()
		currentNetnsInfo := Successful(os.Stat("/proc/thread-self/ns/net"))
		Expect(initialNetnsInfo.Sys().(*syscall.Stat_t).Ino).NotTo(
			Equal(currentNetnsInfo.Sys().(*syscall.Stat_t).Ino))

		By("switching back into the original network namespace")
		Expect(f).NotTo(Panic())
		currentNetnsInfo = Successful(os.Stat("/proc/thread-self/ns/net"))
		Expect(initialNetnsInfo.Sys().(*syscall.Stat_t).Ino).To(
			Equal(currentNetnsInfo.Sys().(*syscall.Stat_t).Ino))
	})

	It("creates a transient network namespace without entering it", func() {
		homeIno := CurrentIno()
		Expect(homeIno).NotTo(BeZero())
		Expect(homeIno).To(Equal(Ino("/proc/self/ns/net")))

		netnsfd := NewTransient()
		netnsIno := Ino(netnsfd)
		Expect(netnsIno).NotTo(BeZero())
		Expect(netnsIno).NotTo(Equal(homeIno))
	})

	It("cannot enter an invalid network namespace", func() {
		var msg string
		g := NewGomega(func(message string, callerSkip ...int) {
			msg = message
		})
		execute(g, 0, func() {})
		Expect(msg).To(ContainSubstring("cannot switch into network namespace"))
	})

	It("executes a function in a different network namespace", func() {
		netnsfd := NewTransient()
		netnsIno := Ino(netnsfd)
		var currentnetnsIno uint64
		Execute(netnsfd, func() { currentnetnsIno = Ino("/proc/thread-self/ns/net") })
		Expect(currentnetnsIno).NotTo(BeZero())
		Expect(currentnetnsIno).To(Equal(netnsIno))
	})

	It("returns a netlink handle for a network namespace fd reference", func() {
		netnsfd := NewTransient()
		var dmy netlink.Link
		Execute(netnsfd, func() {
			dmy = dummy.NewTransient()
		})
		h := NewNetlinkHandle(netnsfd)
		defer h.Close()
		Expect(Successful(h.LinkByName(dmy.Attrs().Name)).Attrs().Name).
			To(Equal(dmy.Attrs().Name))
	})

})
