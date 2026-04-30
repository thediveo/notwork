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

package vlan

import (
	"github.com/vishvananda/netlink"

	"github.com/thediveo/notwork/link"

	. "github.com/onsi/ginkgo/v2" //nolint:staticcheck // ST1001 rule does not apply
	. "github.com/onsi/gomega"    //nolint:staticcheck // ST1001 rule does not apply
)

// VlanPrefix is the name prefix used for transient VLAN network
// interfaces.
const VlanPrefix = "vln-"

// Opt is a configuration option when creating a new VLAN network interface.
type Opt func(*link.Link) error

// NewTransient creates and returns a new (and transient) VLAN network interface
// with VLAN ID passed in vid, and attached to the specified network interface.
// NewTransient automatically defers proper automatic removal of the VLAN
// network interface.
func NewTransient(vid uint16, lnk netlink.Link, opts ...Opt) netlink.Link {
	GinkgoHelper()

	vlan := &link.Link{
		Link: &netlink.Vlan{
			LinkAttrs: netlink.LinkAttrs{
				ParentIndex: lnk.Attrs().Index,
			},
			VlanId: int(vid),
		},
	}
	for _, opt := range opts {
		Expect(opt(vlan)).To(Succeed())
	}
	return link.NewTransient(vlan, VlanPrefix)
}
