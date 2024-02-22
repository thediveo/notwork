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
	"github.com/thediveo/notwork/netns"
	"github.com/vishvananda/netlink"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/thediveo/success"
)

// VethPrefix is the name prefix used for transient VETH network interfaces.
const VethPrefix = "veth-"

type Opt func(*netlink.Veth) error

// NewTransient creates and returns a new (and transient) VETH pair of network
// interfaces. The one VETH end is created in the current network namespace,
// while the other VETH end can optionally be created in a differend network
// namespace using [WithPeerNamespace].
//
// See also: https://en.wikipedia.org/wiki/Thomson_and_Thompson
func NewTransient(opts ...Opt) (dupond netlink.Link, dupont netlink.Link) {
	GinkgoHelper()
	veth := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{},
	}
	for _, opt := range opts {
		Expect(opt(veth)).To(Succeed())
	}
	dupond = link.NewTransient(veth, VethPrefix)
	if veth.PeerNamespace != nil {
		netnsfd, _ := veth.PeerNamespace.(netlink.NsFd)
		netns.Execute(int(netnsfd), func() {
			dupont = Successful(netlink.LinkByName(dupond.(*netlink.Veth).PeerName))
		})
		return
	}
	dupont = Successful(netlink.LinkByName(dupond.(*netlink.Veth).PeerName))
	return
}

// WithPeerNamespace configures the VETH peer end to be created inside the
// network namespace referenced by fd.
func WithPeerNamespace(fd int) Opt {
	return func(v *netlink.Veth) error {
		v.PeerNamespace = netlink.NsFd(fd)
		return nil
	}
}
