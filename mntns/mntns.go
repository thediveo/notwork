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
	Expect(unix.Unshare(unix.CLONE_FS|unix.CLONE_NEWNS), "cannot create new mount namespace")
	// Remount root to ensure that later mount point manipulations do not
	// propagate back into our host, trashing it.
	Expect(unix.Mount("none", "/", "", unix.MS_REC|unix.MS_PRIVATE, "")).To(Succeed())

	return func() { // this cannot be DeferCleanup'ed: we need to restore the current locked go routine
		if err := unix.Setns(mntnsfd, 0); err != nil {
			panic(fmt.Sprintf("cannot restore original mount namespace, reason: %s", err.Error()))
		}
		unix.Close(mntnsfd)
		// do NOT unlock the OS-level thread, as we cannot undo unsharing CLONE_FS
	}
}

// MountSysfsRO mounts a new sysfs instance onto /sys when the caller is in a
// new and transient mount namespace. Otherwise, MountSysfsRO will fail the
// current test.
func MountSysfsRO() {
	GinkgoHelper()
	mountSysfs(Default,
		unix.MS_RDONLY|
			unix.MS_NODEV|unix.MS_NOEXEC|unix.MS_NOSUID|unix.MS_RELATIME,
		"")
}

func mountSysfs(g Gomega, flags uintptr, data string) {
	GinkgoHelper()

	// Ensure that we're not still in the process's original mount namespace, as
	// otherwise we would overmount the host's /sysfs.
	g.Expect(Ino("/proc/thread-self/ns/mnt")).NotTo(Equal(Ino("/proc/self/ns/mnt")),
		"current mount namespace must not be the process's original mount namespace")

	g.Expect(unix.Mount("none", "/sys", "sysfs", flags, data)).To(Succeed())
}

// Ino returns the identification/inode number of the passed mount namespace,
// either referenced by a file descriptor or a VFS path name.
func Ino[R ~int | ~string](mntns R) uint64 {
	GinkgoHelper()

	var netnsStat unix.Stat_t
	switch ref := any(mntns).(type) {
	case int:
		Expect(unix.Fstat(ref, &netnsStat)).To(Succeed(),
			"cannot stat mount namespace reference %v", ref)
	case string:
		Expect(unix.Stat(ref, &netnsStat)).To(Succeed(),
			"cannot stat mount namespace reference %v", ref)
	}
	return netnsStat.Ino
}
