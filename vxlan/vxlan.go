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

package vxlan

import (
	"github.com/vishvananda/netlink"

	"github.com/thediveo/notwork/link"

	. "github.com/onsi/ginkgo/v2" //nolint:staticcheck // ST1001 rule does not apply
	. "github.com/onsi/gomega"    //nolint:staticcheck // ST1001 rule does not apply
)

// VxlanPrefix is the name prefix used for transient VXLAN network
// interfaces.
const VxlanPrefix = "vxl-"

// Opt is a configuration option when creating a new VXLAN network interface.
type Opt func(*link.Link) error

// NewTransient creates and returns a new (and transient) VXLAN network
// interface attached to the specified underlay network interface (which must be
// a hardware network interface, including the dummy kind). NewTransient
// automatically defers proper automatic removal of the VXLAN network interface.
func NewTransient(underlay netlink.Link, opts ...Opt) netlink.Link {
	GinkgoHelper()

	vxlan := &link.Link{
		Link: &netlink.Vxlan{
			VtepDevIndex: underlay.Attrs().Index,
		},
	}
	for _, opt := range opts {
		Expect(opt(vxlan)).To(Succeed())
	}
	return link.NewTransient(vxlan, VxlanPrefix)
}
