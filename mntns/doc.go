/*
Package mntns supports running unit tests in separated transient mount
namespaces.

Wait ... what do mount(!) namespaces have to do with virtual networks and
testing?

The “/sys/class/net” branch of [sysfs(5)].

# Some Sysfs Background

According to [sysfs(5)], “[e]ach of the entries in this directory is a symbolic
link representing one of the real or virtual networking devices [...] Each of
these symbolic links refers to entries in the /sys/devices directory.” (omission
ours)

Unfortunately, the man page is wrong in the omitted part of the description,
about “real or virtual networking devices that are visible in the network
namespace of the process that is accessing the directory”. According to this
[answer to Switching into a network namespace does not change /sys/class/net?] –
which can be easily verified, not least in the self-test units of notwork – the
sysfs locks the “sys/class/net“ view to the network namespace of the (OS-level)
thread that mounted that particular sysfs instance.

Thus, unit tests working with the “sys/class/net” branch need to create and
enter a transient mount namespace after they've created and entered a transient
network namespace, and then also mount a new sysfs instance onto “/sys” to get
a consistent view.

# Usage

To create transient mount and network namespaces and mount a correct sysfs
instance:

	import (
		"github.com/notwork/mntns"
		"github.com/notwork/netns"
	)

	It("creates transient mount and network namespaces, mounts a sysfs", func() {
		defer netns.EnterTransient()()
		defer mntns.EnterTransient()()
		mntns.MountSysfsRO()
	})

Or without normally entering the mount namespace:

	import (
		"github.com/notwork/mntns"
		"github.com/notwork/netns"
	)

	It("creates transient mount and network namespaces, mounts a sysfs", func() {
		defer netns.EnterTransient()()
		mntnsfd, procfsroot := mntns.NewTransient()
		mntns.Execute(mntnsfd, func() {
			mntns.MountSysfsRO()
		}
	})

Here, the returned “procfsroot” path is in the form of “/proc/$TID/root” and
allows accessing directories and files in the transient mount namespace without
the need to enter it. In case you need to work with absolute symbolic links,
[procfsroot] will help by resolving absolute symbolic links inside a different
mount namespace correctly; please refer to the procfsroot package for details.

[sysfs(5)]: https://man7.org/linux/man-pages/man5/sysfs.5.html
[answer to Switching into a network namespace does not change /sys/class/net?]: https://unix.stackexchange.com/a/457384/288012
[procfsroot]: https://github.com/thediveo/procfsroot
*/
package mntns
