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

package netns

import (
	"github.com/vishvananda/netlink"
	vishnetns "github.com/vishvananda/netns"

	"github.com/onsi/gomega/types"

	. "github.com/onsi/ginkgo/v2" //nolint:staticcheck // ST1001 rule does not apply
	. "github.com/onsi/gomega"    //nolint:staticcheck // ST1001 rule does not apply
)

// NewNetlinkHandle returns a *netlink.Handle that is connected to the specified
// network namespace (in form of a file descriptor). Such a file descriptor can
// be obtained from especially [netns.NewTransient] or [netns.Current].
//
//	 import (
//	     "github.com/notwork/netns"
//	 )
//
//	 It("lists links in a transient network namespace", func() {
//		netnsfd := netns.NewTransient() // ...no explicit close needed
//		nlh := netns.NewNetlinkHandle(netnsfd) // ...also no explicit close needed
//		links := Successful(nlh.LinkList())
//		Expect(links).NotTo(BeEmpty())
//	 })
//
// The caller doesn't need to close the returned handle, as NewHandle
// automatically schedules for the netlink.Handle to be closed when the calling
// test node terminates (for whichever reason).
func NewNetlinkHandle(netnsfd int) *netlink.Handle {
	GinkgoHelper()
	return newNetlinkHandle(Default, netnsfd)
}

func newNetlinkHandle(g types.Gomega, netnsfd int) *netlink.Handle {
	GinkgoHelper()

	h, err := netlink.NewHandleAt(vishnetns.NsHandle(netnsfd))
	g.Expect(err).NotTo(HaveOccurred(), "cannot create netlink handle for network namespace")
	DeferCleanup(func() {
		h.Close()
	})
	return h
}
