// Copyright 2026 Harald Albrecht.
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

package netkit

import (
	"github.com/vishvananda/netlink"
	vishnetns "github.com/vishvananda/netns"

	"github.com/thediveo/notwork/internal/nkwrap"
	"github.com/thediveo/notwork/link"

	. "github.com/onsi/ginkgo/v2"   //nolint:staticcheck // ST1001 rule does not apply
	. "github.com/onsi/gomega"      //nolint:staticcheck // ST1001 rule does not apply
	. "github.com/thediveo/success" //nolint:staticcheck // ST1001 rule does not apply
)

// NetkitPrefix is the name prefix used for transient netkit-type network
// interfaces.
const NetkitPrefix = "nkt-"

// Opt is a configuration option when creating a new netkit network interface.
type Opt func(*link.Link) error

// Netkit is an alias for wrapping the netlink's module glitched Netkit type so
// that we can even lateron make the peer link attribute configuration
// available. Thus, [NewTransient] returns [*Netkit], not plain
// [*netlink.Netkit].
type Netkit = nkwrap.Netkit

// NewTransient creates and returns a new (and transient) pair of [netkit]
// network interfaces. The one (“primary”) netkit end is created in the current
// network namespace, while the other (“peer”) netkit end can optionally be
// created in a differend network namespace using [WithPeerNamespace].
//
// Note: contrary to VETH peers, the two netkit network interfaces created as a
// combo are not symmetrical but instead have different roles termed
// “primary”/“host” as well as “peer”.
//
// [netkit]: https://blog.yadutaf.fr/2025/07/01/introduction-to-linux-netkit-interfaces-with-a-grain-of-ebpf/
func NewTransient(opts ...Opt) (primary, peer netlink.Link) {
	GinkgoHelper()

	netkit := &link.Link{
		Link: &nkwrap.Netkit{
			Netkit: netlink.Netkit{ // \o/
				LinkAttrs: netlink.LinkAttrs{},
			},
			Peer: netlink.LinkAttrs{},
		},
	}
	for _, opt := range opts {
		Expect(opt(netkit)).To(Succeed())
	}
	// netlink's Netkit link type has made the peer link attributes private,
	// only with a SetPeersAttrs setter, but no getter. (╥_╥) So we have to pull
	// some tricks with wrapping of which link.NewTransient() is aware.
	netkit.Link.(*nkwrap.Netkit).UpdatePrivatePeer()
	primary = link.NewTransient(netkit, NetkitPrefix)
	wrappedHostnk := primary.(*Netkit)
	// Now things get tricky as want to return proper link information about the
	// peer; unfortunately, RTNETLINK again acts odd: with the destination
	// network namespace set, if the peer network namespace is unset then the
	// peer will end up in the current(!) network namespace, not in the
	// destination network namespace. Yuck.
	if wrappedHostnk.Peer.Namespace != nil {
		nlh := Successful(netlink.NewHandleAt(
			vishnetns.NsHandle(int(wrappedHostnk.Peer.Namespace.(netlink.NsFd)))))
		defer func() { _ = nlh.Close() }()
		peer = Successful(nlh.LinkByName(wrappedHostnk.Peer.Name))
		return
	}
	peer = Successful(netlink.LinkByName(wrappedHostnk.Peer.Name))
	return
}
