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

package macvlan

import (
	"github.com/thediveo/notwork/link"
	"github.com/vishvananda/netlink"

	. "github.com/onsi/ginkgo/v2" //lint:ignore ST1001 rule does not apply
	. "github.com/onsi/gomega"    //lint:ignore ST1001 rule does not apply
)

// MacvlanPrefix is the name prefix used for transient MACVLAN network
// interfaces.
const MacvlanPrefix = "mcvl-"

// Opt is a configuration option when creating a new MACVLAN network interface.
type Opt func(*link.Link) error

// LocateHWParent locates a “hardware” network interface in the current network
// namespace that is operationally up and returns it. If no suitable network
// interface can be found, then the current test is failed. If multiple suitable
// network interfaces are found, a random one of them is returned.
//
// Please consider using a “dummy” network interface instead as a MACVLAN parent
// unless it's absolutely necessary to use a hardware network interface. Dummy
// network interfaces can be created using [dummy.NewTransient].
func LocateHWParent() netlink.Link {
	GinkgoHelper()

	var parents []netlink.Link
	links, err := netlink.LinkList()
	Expect(err).NotTo(HaveOccurred(), "cannot retrieve list of netdevs")
	Expect(links).To(ContainElement(
		And(
			HaveField("Type()", "device"),
			HaveField("Attrs().Name", Not(Equal("lo"))),
			HaveField("Attrs().OperState", netlink.LinkOperState(netlink.OperUp))),
		&parents), "could not find any hardware netdev in up state")
	// ContainElement guarantees when in filter result mode that there were
	// one or more matches and fail otherwise in case of no matches at all.
	// We just pick "randomly" (obligatory XKCD ref here) the parent to work
	// with further.
	return parents[0]
}

// NewTransient creates and returns a new (and transient) MACVLAN network
// interface attached to the specified parent network interface (which must be a
// hardware network interface, including the dummy kind). CreateTransient
// automatically defers proper automatic removal of the MACVLAN network
// interface.
func NewTransient(parent netlink.Link, opts ...Opt) netlink.Link {
	GinkgoHelper()

	mcvlan := &link.Link{
		Link: &netlink.Macvlan{
			LinkAttrs: netlink.LinkAttrs{
				ParentIndex: parent.Attrs().Index,
			},
			Mode: netlink.MACVLAN_MODE_BRIDGE,
		},
	}
	for _, opt := range opts {
		Expect(opt(mcvlan)).To(Succeed())
	}
	return link.NewTransient(mcvlan, MacvlanPrefix)
}

// CreateTransient creates and returns a new (and transient) MACVLAN network
// interface attached to the specified parent network interface (which must be a
// hardware network interface, including the dummy kind).
//
// Deprecated: use [NewTransient] instead.
func CreateTransient(parent netlink.Link) netlink.Link {
	GinkgoHelper()
	return NewTransient(parent)
}
