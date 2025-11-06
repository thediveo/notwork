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

package veth

import (
	"github.com/thediveo/notwork/link"
	"github.com/vishvananda/netlink"
	vishnetns "github.com/vishvananda/netns"

	. "github.com/onsi/ginkgo/v2"   //nolint:staticcheck // ST1001 rule does not apply
	. "github.com/onsi/gomega"      //nolint:staticcheck // ST1001 rule does not apply
	. "github.com/thediveo/success" //nolint:staticcheck // ST1001 rule does not apply
)

// VethPrefix is the name prefix used for transient VETH network interfaces.
const VethPrefix = "veth-"

// Opt is a configuration option when creating a new pair of VETH network
// interfaces.
type Opt func(*link.Link) error

// NewTransient creates and returns a new (and transient) VETH pair of network
// interfaces. The one VETH end is created in the current network namespace,
// while the other VETH end can optionally be created in a differend network
// namespace using [WithPeerNamespace].
//
// See also: https://en.wikipedia.org/wiki/Thomson_and_Thompson
func NewTransient(opts ...Opt) (dupond netlink.Link, dupont netlink.Link) {
	GinkgoHelper()
	veth := &link.Link{
		Link: &netlink.Veth{
			LinkAttrs: netlink.LinkAttrs{},
		},
	}
	for _, opt := range opts {
		Expect(opt(veth)).To(Succeed())
	}
	dupond = link.NewTransient(veth, VethPrefix)
	// Now things get tricky as want to return proper link information about the
	// peer; unfortunately, RTNETLINK again acts odd: with the destination
	// network namespace set, if the peer network namespace is unset then the
	// peer will end up in the current(!) network namespace, not in the
	// destination network namespace. Yuck.
	if peerNamespace := veth.Link.(*netlink.Veth).PeerNamespace; peerNamespace != nil {
		nlh := Successful(netlink.NewHandleAt(vishnetns.NsHandle(int(peerNamespace.(netlink.NsFd)))))
		defer nlh.Close()
		dupont = Successful(nlh.LinkByName(dupond.(*netlink.Veth).PeerName))
		return
	}
	dupont = Successful(netlink.LinkByName(dupond.(*netlink.Veth).PeerName))
	return
}
