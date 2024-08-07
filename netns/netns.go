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
	"math/rand"
	"runtime"

	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"

	. "github.com/onsi/ginkgo/v2" //lint:ignore ST1001 rule does not apply
	. "github.com/onsi/gomega"    //lint:ignore ST1001 rule does not apply
)

// EnterTransient creates and enters a new (and isolated) network namespace,
// returning a function that needs to be defer'ed in order to correctly switch
// the calling go routine and its locked OS-level thread back when the caller
// itself returns.
//
//	defer netns.EnterTransient()()
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
// it returns a file descriptor referencing the new network namespace.
// NewTransient also schedules a Ginkgo deferred cleanup in order to close the
// fd referencing the newly created network namespace. The caller thus must not
// close the file descriptor returned.
func NewTransient() int {
	GinkgoHelper()

	runtime.LockOSThread()
	// no deferred unlock, as we need to throw away the OS-level thread if
	// things go south.
	orignetnsfd := current()
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

// Execute a function fn in the network namespace referenced by the open file
// descriptor netnsfd.
func Execute(netnsfd int, fn func()) {
	GinkgoHelper()
	execute(Default, netnsfd, fn)
}

func execute(g Gomega, netnsfd int, fn func()) {
	runtime.LockOSThread()
	// no deferred unlock, as we need to throw away the OS-level thread if
	// things go south. Nota bene: Ginkgo runs its tests on fresh go routines.
	orignetnsfd := current()
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
//
// When not running in the initial network namesepace you should have the
// calling go routine locked to its OS-level thread.
//
// Additionally, Current schedules a DeferCleanup of the returned file
// descriptor to be closed to avoid leaking it.
func Current() int {
	GinkgoHelper()

	netnsfd := current()
	DeferCleanup(func() {
		_ = unix.Close(netnsfd)
	})
	return netnsfd
}

// Package-internal convenience helper for DRY. We don't schedule any
// DeferCleanup at this level.
func current() int {
	GinkgoHelper()

	netnsfd, err := unix.Open("/proc/thread-self/ns/net", unix.O_RDONLY, 0)
	Expect(err).NotTo(HaveOccurred(), "cannot determine current network namespace from procfs")
	return netnsfd
}

// Ino returns the identification/inode number of the passed network namespace,
// either referenced by a file descriptor or a VFS path name.
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
	GinkgoHelper()

	return Ino("/proc/thread-self/ns/net")
}

// NsID returns the so-called network namespace ID for the passed network
// namespace, either referenced by a file descriptor or a VFS path name. The
// nsid identifies the passed network namespace from the perspective of the
// current network namespace.
//
// If no nsid has been assigned yet to the passed network namespace from the
// perspective of the current network namespace, NsID will assign a random nsid
// and return it.
func NsID[R ~int | ~string](netns R) int {
	GinkgoHelper()

	var netnsfd int
	switch ref := any(netns).(type) {
	case int:
		netnsfd = ref
	case string:
		var err error
		netnsfd, err = unix.Open(ref, unix.O_RDONLY, 0)
		Expect(err).NotTo(HaveOccurred(), "cannot open network namespace reference %v", ref)
		defer unix.Close(netnsfd)
	}
	netnsid, err := netlink.GetNetNsIdByFd(netnsfd)
	Expect(err).NotTo(HaveOccurred(), "cannot retrieve netnsid")
	// netnsid might be -1, signalling that no netnsid has been assigned yet ...
	// which begs the question why RTM_GETNSID simply isn't allocating a free
	// one...?!
	if netnsid != -1 {
		return netnsid
	}
	for attempt := 1; attempt <= 10; attempt++ {
		// as per https://elixir.bootlin.com/linux/v6.9.4/source/lib/idr.c#L87,
		// netnsid's are uint32 (to use Go's data type terminology).
		netnsid := int(rand.Int31())
		if err := netlink.SetNetNsIdByFd(netnsfd, netnsid); err != nil {
			continue
		}
		return netnsid
	}
	Fail("too many failed attempts to assign a new netnsid first")
	return -1 // unreachable
}
