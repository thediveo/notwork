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
	"golang.org/x/sys/unix"

	. "github.com/onsi/ginkgo/v2"   //lint:ignore ST1001 rule does not apply
	. "github.com/onsi/gomega"      //lint:ignore ST1001 rule does not apply
	. "github.com/thediveo/success" //lint:ignore ST1001 rule does not apply
)

// Link to netdevsim “port” interfaces with each other.
//
// Note: requires Linux kernel 6.9+.
func Link(dupond, dupont netlink.Link) {
	GinkgoHelper()

	netnsfd1, ifindex1 := Successful2R(linkFds(dupond))
	defer unix.Close(netnsfd1)
	netnsfd2, ifindex2 := Successful2R(linkFds(dupont))
	defer unix.Close(netnsfd2)

	Expect(os.WriteFile(netdevsimRoot+"/link_device",
		[]byte(fmt.Sprintf("%d:%d %d:%d",
			netnsfd1, ifindex1,
			netnsfd2, ifindex2)), 0)).To(Succeed())
}

// Unlink the specified “port” interface from its peer.
//
// Note: requires Linux kernel 6.9+.
func Unlink(l netlink.Link) {
	GinkgoHelper()

	netnsfd, ifindex := Successful2R(linkFds(l))
	Expect(os.WriteFile(netdevsimRoot+"/unlink_device",
		[]byte(fmt.Sprintf("%d:%d", netnsfd, ifindex)), 0)).To(Succeed())

}

// linkFds returns a netns fd as well as the ifindex of the link in question,
// otherwise an error. The caller is responsible to close the netns fd when not
// needing it any longer.
func linkFds(l netlink.Link) (netnsfd int, ifindex int, err error) {
	ifindex = l.Attrs().Index
	if ifindex == 0 {
		l := Successful(netlink.LinkByName(l.Attrs().Name))
		ifindex = l.Attrs().Index
	}

	if netnsfd, ok := l.Attrs().Namespace.(netlink.NsFd); ok {
		netnsfd, err := unix.Dup(int(netnsfd))
		if err != nil {
			return 0, 0, fmt.Errorf("cannot duplicate netns fd reference, reason: %w", err)
		}
		return ifindex, netnsfd, nil
	}
	// We're not locking the go routine, as either it doesn't matter (all
	// unlocked) or we're already locked correctly.
	netnsfd, err = unix.Open("/proc/thread-self/ns/net", unix.O_RDONLY, 0)
	if err != nil {
		return 0, 0, fmt.Errorf("cannot get current netns fd reference, reason: %w", err)
	}
	return ifindex, netnsfd, nil
}
