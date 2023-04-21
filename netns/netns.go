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
	"fmt"
	"runtime"

	"github.com/vishvananda/netlink"
	nlnetns "github.com/vishvananda/netns"
	"golang.org/x/sys/unix"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// EnterTransient creates and enters a new (and isolated) network namespace,
// returning a function that needs to be defer'ed in order to correctly switch
// the calling go routine and its locked OS-level thread back when the caller
// itself returns.
//
// In case the caller cannot be switched back correctly, the defer'ed clean up
// will panic with an error description.
func EnterTransient() func() {
	GinkgoHelper()

	runtime.LockOSThread()
	netnsfd, err := unix.Open("/proc/thread-self/ns/net", unix.O_RDONLY, 0)
	Expect(err).NotTo(HaveOccurred(), "cannot determine current network namespace from procfs")
	Expect(unix.Unshare(unix.CLONE_NEWNET)).To(Succeed(), "cannot create new network namespace")
	return func() { // this cannot be DeferCleanup'ed: we need to restore the current locked go routine
		if err := unix.Setns(netnsfd, 0); err != nil {
			panic(fmt.Sprintf("cannot restore original network namespace, reason: %s", err.Error()))
		}
		unix.Close(netnsfd)
		runtime.UnlockOSThread()
	}
}

// NewTransient creates a new network namespace, but doesn't enter it. Instead,
// it returns a file descriptor referencing the new network namespace. It
// additionally schedules a Ginkgo deferred cleanup in order to close the fd
// referencing the newly created network namespace.
func NewTransient() int {
	GinkgoHelper()

	runtime.LockOSThread()
	// no deferred unlock, as we need to throw away the OS-level thread if
	// things go south.
	orignetnsfd := Current()
	defer unix.Close(orignetnsfd)
	Expect(unix.Unshare(unix.CLONE_NEWNET)).To(Succeed(), "cannot create new network namespace")
	netnsfd, err := unix.Open("/proc/thread-self/ns/net", unix.O_RDONLY, 0)
	Expect(err).NotTo(HaveOccurred(), "cannot determine new network namespace from procfs")
	Expect(unix.Setns(orignetnsfd, unix.CLONE_NEWNET)).To(Succeed(), "cannot switch back into original network namespace")
	DeferCleanup(func() {
		unix.Close(netnsfd)
	})
	runtime.UnlockOSThread()
	return netnsfd
}

// Execute a function fn in the specified network namespace, referenced by the
// open file descriptor netnsfd.
func Execute(netnsfd int, fn func()) {
	execute(Default, netnsfd, fn)
}

func execute(g Gomega, netnsfd int, fn func()) {
	runtime.LockOSThread()
	// no deferred unlock, as we need to throw away the OS-level thread if
	// things go south.
	orignetnsfd := Current()
	defer unix.Close(orignetnsfd)
	g.Expect(unix.Setns(netnsfd, unix.CLONE_NEWNET)).To(Succeed(), "cannot switch into network namespace")
	defer func() {
		g.Expect(unix.Setns(orignetnsfd, unix.CLONE_NEWNET)).To(Succeed(), "cannot switch back into original network namespace")
		runtime.UnlockOSThread()
	}()
	fn()
}

// Current returns a file descriptor referencing the current network namespace.
// In particular, the current network namespace of the OS-level thread of the
// caller's Go routine (which should ideally be thread-locked).
func Current() int {
	GinkgoHelper()

	netnsfd, err := unix.Open("/proc/thread-self/ns/net", unix.O_RDONLY, 0)
	Expect(err).NotTo(HaveOccurred(), "cannot determine current network namespace from procfs")
	return netnsfd
}

// Ino returns the identification/inode number of the passed network namespace.
func Ino[R ~int | ~string](netns R) uint64 {
	GinkgoHelper()

	var netnsStat unix.Stat_t
	switch ref := any(netns).(type) {
	case int:
		Expect(unix.Fstat(ref, &netnsStat)).To(Succeed(),
			"cannot stat network namespace reference %v", ref)
	case string:
		Expect(unix.Stat(ref, &netnsStat)).To(Succeed(),
			"cannot stat network namespace reference %v", ref)
	}
	return netnsStat.Ino
}

// CurrentIno returns the identification/inode number of the network namespace
// for the current thread.
func CurrentIno() uint64 {
	return Ino("/proc/thread-self/ns/net")
}

// NewNetlinkHandle returns a netlink handle connected to the network namespace
// referenced by the specified fd (file descriptor). For instance, this file
// descriptor might be one returned by [NewTransient] or [Current].
func NewNetlinkHandle(netnsfd int) *netlink.Handle {
	GinkgoHelper()

	nlh, err := netlink.NewHandleAt(nlnetns.NsHandle(netnsfd))
	Expect(err).NotTo(HaveOccurred(), "cannot create netlink handle for network namespace")
	return nlh
}
