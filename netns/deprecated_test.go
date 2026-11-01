// Copyright 2026 Harald Albrecht.
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

	spacetestnetns "github.com/thediveo/spacetest/netns"

	"github.com/thediveo/notwork/ns"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
)

var _ = Describe("deprecated network namespace functions", func() {

	BeforeEach(func() {
		if os.Getuid() != 0 {
			Skip("needs root")
		}
		goodfds := Filedescriptors()
		goodgos := Goroutines()
		DeferCleanup(func() {
			Eventually(Goroutines).Within(2 * time.Second).
				ProbeEvery(250 * time.Millisecond).
				ShouldNot(HaveLeaked(goodgos))
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
		})
	})

	It("enters and leaves a transient network namespace", func() {
		original := CurrentIno()
		func() {
			defer EnterTransient()()
			Expect(CurrentIno()).NotTo(Equal(original))
		}()
		Expect(CurrentIno()).To(Equal(original))
	})

	It("creates a transient network namespace without entering it", func() {
		original := CurrentIno()
		fd := NewTransient()
		Expect(Ino(fd)).NotTo(Equal(original))
		Expect(CurrentIno()).To(Equal(original))
	})

	It("executes a function in the specified network namespace", func() {
		original := CurrentIno()
		fd := NewTransient()
		target := Ino(fd)
		Expect(target).NotTo(Equal(original))
		var executedIn uint64
		Execute(fd, func() { executedIn = CurrentIno() })
		Expect(executedIn).To(Equal(target))
		Expect(CurrentIno()).To(Equal(original))
	})

	It("returns a file descriptor referencing the current network namespace", func() {
		fd := Current()
		Expect(Ino(fd)).To(Equal(CurrentIno()))
	})

	It("returns the inode of a network namespace", func() {
		fd := Current()
		// Test both supported reference types.
		Expect(Ino(fd)).To(Equal(CurrentIno()))
		Expect(Ino("/proc/thread-self/ns/net")).To(Equal(CurrentIno()))
	})

	It("returns the inode of the current network namespace", func() {
		fd := spacetestnetns.Current()
		Expect(CurrentIno()).To(Equal(spacetestnetns.Ino(fd)))
	})

	It("creates a netlink handle for the specified network namespace", func() {
		fd := NewTransient()
		handle := NewNetlinkHandle(fd)
		Expect(handle).NotTo(BeNil())
		defer func() { _ = handle.Close() }()

		links, err := handle.LinkList()
		Expect(err).NotTo(HaveOccurred())

		// A freshly created network namespace contains its loopback device.
		Expect(links).NotTo(BeEmpty())
		Expect(links[0].Attrs().Name).To(Equal("lo"))
	})

	It("returns the network namespace ID for a file descriptor", func() {
		fd := Current()
		Expect(NsID(fd)).To(Equal(ns.ID(fd)))
	})

	It("returns the network namespace ID for a path", func() {
		const path = "/proc/thread-self/ns/net"
		Expect(NsID(path)).To(Equal(ns.ID(path)))
	})

})
