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

package mntns

import (
	"fmt"
	"runtime"

	"golang.org/x/sys/unix"

	. "github.com/onsi/ginkgo/v2" //lint:ignore ST1001 rule does not apply
	. "github.com/onsi/gomega"    //lint:ignore ST1001 rule does not apply
)

// Avoid problems that would happen when we accidentally unshare the initial
// thread, so we lock it here, thus ensuring that other Go routines (and
// especially tests) won't ever get scheduled onto the initial thread anymore.
func init() {
	runtime.LockOSThread()
}

// EnterTransient creates and enters a new mount namespace, returning a function
// that needs to be defer'ed. It additionally remounts “/” in this new mount
// namespace to set propagation of mount points to “private” – otherwise, later
// using [MountSysfsRO] will end in tears as the new “/sys” mount would
// otherwise happily propagate back (and we absolutely don't want that to
// happen).
//
// Note: the current OS-level thread won't be unlocked when the calling unit
// test returns, as we cannot undo unsharing filesystem attributes (using
// CLONE_FS) such as the root directory, current directory, and umask
// attributes.
//
// # Background
//
// [unshare(1)] defaults the mount point propagation to "MS_REC | MS_PRIVATE",
// see [util-linux/unshare.c UNSHARE_PROPAGATION_DEFAULT].
//
// [unshare(1)] is remounting "/" in order to apply its propagation defaults --
// that are to not FUBAR the mount points in the mount namespace we got our
// mount points from during unsharing the mount namespace, see
// [util-linux/unshare.c set_propagation].
//
//	/* C */ mount("none", "/", NULL, flags, NULL
//
// [util-linux/unshare.c set_propagation]: https://github.com/util-linux/util-linux/blob/86b6684e7a215a0608bd130371bd7b3faae67aca/sys-utils/unshare.c#L160
// [unshare(1)]: https://man7.org/linux/man-pages/man1/unshare.1.html
// [util-linux/unshare.c UNSHARE_PROPAGATION_DEFAULT]: https://github.com/util-linux/util-linux/blob/86b6684e7a215a0608bd130371bd7b3faae67aca/sys-utils/unshare.c#L57
func EnterTransient() func() {
	GinkgoHelper()

	runtime.LockOSThread()
	mntnsfd, err := unix.Open("/proc/thread-self/ns/mnt", unix.O_RDONLY, 0)
	Expect(err).NotTo(HaveOccurred(), "cannot determine current mount namespace from procfs")

	// Decouple some filesystem-related attributes of this thread from the ones
	// of our process...
	Expect(unix.Unshare(unix.CLONE_FS|unix.CLONE_NEWNS)).To(Succeed(),
		"cannot create new mount namespace")
	// Remount root to ensure that later mount point manipulations do not
	// propagate back into our host, trashing it.
	Expect(unix.Mount("none", "/", "/", unix.MS_REC|unix.MS_PRIVATE, "")).To(Succeed(),
		"cannot change / mount propagation to private")

	return func() { // this cannot be DeferCleanup'ed: we need to restore the current locked go routine
		if err := unix.Setns(mntnsfd, 0); err != nil {
			panic(fmt.Sprintf("cannot restore original mount namespace, reason: %s", err.Error()))
		}
		unix.Close(mntnsfd)
		// do NOT unlock the OS-level thread, as we cannot undo unsharing CLONE_FS
	}
}

// MountSysfsRO mounts a new sysfs instance read-only onto /sys when the caller
// is in a new and transient mount namespace. Otherwise, MountSysfsRO will fail
// the current test.
func MountSysfsRO() {
	GinkgoHelper()
	mountSysfs(Default,
		unix.MS_RDONLY|
			unix.MS_NODEV|unix.MS_NOEXEC|unix.MS_NOSUID|unix.MS_RELATIME,
		"")
}

// mountSysfs mounts a new sysfs instance onto /sys, using the specified flags
// and making sure that the caller is not in the process's original mount
// namespace anymore.
func mountSysfs(g Gomega, flags uintptr, data string) {
	GinkgoHelper()

	// Ensure that we're not still in the process's original mount namespace, as
	// otherwise we would overmount the host's /sysfs.
	g.Expect(Ino("/proc/thread-self/ns/mnt")).NotTo(Equal(Ino("/proc/self/ns/mnt")),
		"current mount namespace must not be the process's original mount namespace")

	g.Expect(unix.Mount("none", "/sys", "sysfs", flags, data)).To(Succeed(),
		"cannot mount new sysfs instance on /sys")
}

// NewTransient creates a new transient mount namespace that is kept alive by a
// an idle OS-level thread; this thread is automatically terminated upon
// returning from the current test.
func NewTransient() (mntfd int, procfsroot string) {
	GinkgoHelper()
	// closing the done channel tells the Go routine we will kick off next to
	// call it a day and terminate (well, unless the called fn is stuck).
	done := make(chan struct{})
	DeferCleanup(func() { close(done) })

	// Kick off a separate Go routine which we then can lock to its OS-level
	// thread and later dispose off because it is tainted due to unsharing the
	// sharing of file attributes.
	readyCh := make(chan idler)
	go func() {
		defer GinkgoRecover()
		runtime.LockOSThread()

		// Whatever is going to happen to us, make sure to unblock the receiving
		// Go routine, and even if this is the zero value...
		defer func() {
			close(readyCh)
		}()

		// Decouple some filesystem-related attributes of this thread from the ones
		// of our process...
		Expect(unix.Unshare(unix.CLONE_FS|unix.CLONE_NEWNS)).To(Succeed(),
			"cannot create new mount namespace")
		// Remount root to ensure that later mount point manipulations do not
		// propagate back into our host, trashing it.
		Expect(unix.Mount("none", "/", "/", unix.MS_REC|unix.MS_PRIVATE, "")).To(Succeed(),
			"cannot change / mount propagation to private")

		readyCh <- idler{
			mntnsfd: Current(),
			tid:     unix.Gettid(),
		}

		<-done // ...idle around, then fall off the discworld...
	}()
	idlerInfo := <-readyCh
	Expect(idlerInfo.mntnsfd).NotTo(BeZero())
	procfsroot = fmt.Sprintf("/proc/%d/root", idlerInfo.tid)
	return idlerInfo.mntnsfd, procfsroot
}

type idler struct {
	mntnsfd int
	tid     int
}

// Execute a function fn in a separate Go routine in the mount namespace
// referenced by the open file descriptor mntnsfd. In order to avoid race
// issues, the calling Go routine is blocked until the called fn returns. Any
// results of the called fn should be communicated back to the caller using a
// buffered(!) channel.
func Execute(mntnsfd int, fn func()) {
	GinkgoHelper()
	execute(Default, mntnsfd, fn)
}

func execute(g Gomega, mntnsfd int, fn func()) {
	done := make(chan struct{})

	// We need to use a separate Go routine because we next need to unshare
	// sharing of file attributes with other OS-level threads of this process.
	// This unsharing cannot be undone, so the separate Go routine is then
	// locked to an OS-level thread that is thrown away upon returning from fn.
	go func() {
		defer func() {
			close(done)
		}()
		defer GinkgoRecover()
		runtime.LockOSThread()
		g.Expect(unix.Unshare(unix.CLONE_FS)).To(Succeed(), "cannot unshare file attributes of transient execution thread")
		g.Expect(unix.Setns(mntnsfd, unix.CLONE_NEWNS), "cannot switch into mount namespace")
		fn()
	}()

	// We don't "Eventually(done)...." on purpose here: this would force the
	// caller to supply meaningful timeout values. As we make sure to close the
	// done channel on the separate (and tainted) Go routine this will unblock
	// us here.
	<-done
}

// Current returns a file descriptor referencing the current mount namespace.
// In particular, the current mount namespace of the OS-level thread of the
// caller's Go routine (which should ideally be thread-locked).
//
// When not running in the initial mount namesepace you should have the
// calling go routine locked to its OS-level thread.
//
// Additionally, Current schedules a DeferCleanup of the returned file
// descriptor to be closed to avoid leaking it.
func Current() int {
	GinkgoHelper()

	mntnsfd, err := unix.Open("/proc/thread-self/ns/mnt", unix.O_RDONLY, 0)
	Expect(err).NotTo(HaveOccurred(), "cannot determine current mount namespace from procfs")
	DeferCleanup(func() {
		_ = unix.Close(mntnsfd)
	})
	return mntnsfd
}

// Ino returns the identification/inode number of the passed mount namespace,
// either referenced by a file descriptor or a VFS path name.
func Ino[R ~int | ~string](mntns R) uint64 {
	GinkgoHelper()

	var mntnsStat unix.Stat_t
	switch ref := any(mntns).(type) {
	case int:
		Expect(unix.Fstat(ref, &mntnsStat)).To(Succeed(),
			"cannot stat mount namespace reference %v", ref)
	case string:
		Expect(unix.Stat(ref, &mntnsStat)).To(Succeed(),
			"cannot stat mount namespace reference %v", ref)
	}
	return mntnsStat.Ino
}
