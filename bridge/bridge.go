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

package bridge

import (
	"github.com/vishvananda/netlink"

	"github.com/thediveo/notwork/link"

	. "github.com/onsi/ginkgo/v2" //nolint:staticcheck // ST1001 rule does not apply
	. "github.com/onsi/gomega"    //nolint:staticcheck // ST1001 rule does not apply
)

// BridgePrefix is the name prefix used for transient bridge (network
// interfaces). Note that we don't want to interfere with Docker's “br-”
// namespace.
const BridgePrefix = "bri-"

// Opt is a configuration option when creating a new bridge (network interface).
type Opt func(*link.Link) error

// NewTransient creates and returns a new (and transient) VLAN network interface
// with VLAN ID passed in vid, and attached to the specified network interface.
// NewTransient automatically defers proper automatic removal of the VLAN
// network interface.
func NewTransient(opts ...Opt) netlink.Link {
	GinkgoHelper()

	br := &link.Link{
		Link: &netlink.Bridge{},
	}
	for _, opt := range opts {
		Expect(opt(br)).To(Succeed())
	}
	return link.NewTransient(br, BridgePrefix)
}

// AddPort adds the specified “port” link to the passed “br” bridge, failing the
// current test in case the link cannot be added as a port to the bridge.
func AddPort(br netlink.Link, port netlink.Link) {
	Expect(br.Type()).To(Equal("bridge"), "expected br to be a bridge")
	Expect(netlink.LinkSetMaster(port, br)).To(Succeed(), "cannot add link as port to bridge")
}
