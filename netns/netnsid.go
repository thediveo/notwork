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
	"math/rand"

	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"

	. "github.com/onsi/ginkgo/v2" //lint:ignore ST1001 rule does not apply
	. "github.com/onsi/gomega"    //lint:ignore ST1001 rule does not apply
)

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
