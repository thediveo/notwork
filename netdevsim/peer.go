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
	"fmt"
	"os"

	"github.com/vishvananda/netlink"
	vishnetns "github.com/vishvananda/netns"
	"golang.org/x/sys/unix"

	. "github.com/onsi/ginkgo/v2" //nolint:staticcheck // ST1001 rule does not apply
	. "github.com/onsi/gomega"    //nolint:staticcheck // ST1001 rule does not apply
)

// Link to netdevsim “port” interfaces with each other. Please note that the
// passed link descriptions must reference netdevsim network interfaces in the
// current network namespace; either by name or by index. If one or both instead
// refer to netdevsim network interfaces in other network namespaces than the
// current network namespace, then the link descriptions must have their
// [netlink.LinkAttrs.Namespace] fields properly set.
//
// Note: requires Linux kernel 6.9+.
func Link(dupond, dupont netlink.Link) {
	GinkgoHelper()

	Expect(dupond).NotTo(BeNil(), "dupond/first link must be non-nil")
	Expect(dupont).NotTo(BeNil(), "dupond/second link must be non-nil")

	netnsfd1, ifindex1, err := linkFds(dupond)
	Expect(err).NotTo(HaveOccurred(), "invalid dupond/first link information")
	defer func() { _ = unix.Close(netnsfd1) }()
	netnsfd2, ifindex2, err := linkFds(dupont)
	Expect(err).NotTo(HaveOccurred(), "invalid dupont/second link information")
	defer func() { _ = unix.Close(netnsfd2) }()

	Expect(os.WriteFile(netdevsimRoot+"/link_device",
		fmt.Appendf(nil, "%d:%d %d:%d",
			netnsfd1, ifindex1,
			netnsfd2, ifindex2), 0)).To(Succeed(),
		"cannot link two netdevsims '%s' (netns(%d):%d) and '%s' (netns(%d):%d)",
		dupond.Attrs().Name, netnsfd1, ifindex1,
		dupont.Attrs().Name, netnsfd2, ifindex2)
}

// Unlink the specified “port” interface from its peer.
//
// Note: requires Linux kernel 6.9+.
func Unlink(l netlink.Link) {
	GinkgoHelper()

	Expect(l).NotTo(BeNil(), "link must be non-nil")

	netnsfd, ifindex, err := linkFds(l)
	Expect(err).NotTo(HaveOccurred(), "invalid link information")
	defer func() { _ = unix.Close(netnsfd) }()
	Expect(os.WriteFile(netdevsimRoot+"/unlink_device",
		[]byte(fmt.Sprintf("%d:%d", netnsfd, ifindex)), 0)).To(Succeed())
}

// linkFds returns a netns fd as well as the ifindex of the link in question,
// otherwise an error. The caller is responsible to close the netns fd when not
// needing it any longer.
func linkFds(l netlink.Link) (netnsfd int, ifindex int, err error) {
	ifindex = l.Attrs().Index
	// If the Namespace attribute field has been set, then this netdevsim is
	// located in a different network namespace. We clone the network namespace
	// reference as the contract with our callers is that we always return a
	// network namespace fd that the caller needs to close after use.
	if netnsfd, ok := l.Attrs().Namespace.(netlink.NsFd); ok {
		netnsfd, err := unix.Dup(int(netnsfd))
		if err != nil {
			return 0, 0, fmt.Errorf("cannot duplicate network namespace fd reference, reason: %w", err)
		}
		if ifindex == 0 {
			// Get the interface index, when necessary; and in the correct
			// network namespace...
			nlh, err := netlink.NewHandleAt(vishnetns.NsHandle(netnsfd))
			if err != nil {
				_ = unix.Close(netnsfd)
				return 0, 0, fmt.Errorf("invalid network namespace fd reference, reason: %w", err)
			}
			defer nlh.Close()
			l, err := nlh.LinkByName(l.Attrs().Name)
			if err != nil {
				_ = unix.Close(netnsfd)
				return 0, 0, fmt.Errorf("cannot determine link index, reason: %w", err)
			}
			ifindex = l.Attrs().Index
		}
		return netnsfd, ifindex, nil
	}
	// If the network interface index is not known, we need to get it by asking
	// for the details of the network interface by its name.
	if ifindex == 0 {
		l, err := netlink.LinkByName(l.Attrs().Name)
		if err != nil {
			return 0, 0, fmt.Errorf("cannot determine link index, reason: %w", err)
		}
		ifindex = l.Attrs().Index
	}
	// We're not locking the go routine, as either it doesn't matter (all
	// unlocked) or we're already locked correctly.
	netnsfd, err = unix.Open("/proc/thread-self/ns/net", unix.O_RDONLY, 0)
	if err != nil {
		return 0, 0, fmt.Errorf("cannot get current netns fd reference, reason: %w", err)
	}
	return netnsfd, ifindex, nil
}
