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
)

// InNamespace configures a VXLAN network interface to be created in the
// network namespace referenced by fdref, instead of creating it in the current
// network namespace.
func InNamespace(fdref int) Opt {
	return func(l *link.Link) error {
		l.Attrs().Namespace = netlink.NsFd(fdref)
		return nil
	}
}

// WithLinkNamespace specifies the “reference” or “link” network namespace other
// than the current network namespace when creating a new network interface.
func WithLinkNamespace(fdref int) Opt {
	return func(l *link.Link) error {
		l.LinkNamespace = netlink.NsFd(fdref)
		return nil
	}
}

// With8021Q configures the IEEE 802.1Q VLAN protocol.
func With8021Q() Opt {
	return func(l *link.Link) error {
		l.Link.(*netlink.Vlan).VlanProtocol = netlink.VLAN_PROTOCOL_8021Q
		return nil
	}
}

// With8021AD configures the IEEE 802.1AD VLAN protocol.
func With8021AD() Opt {
	return func(l *link.Link) error {
		l.Link.(*netlink.Vlan).VlanProtocol = netlink.VLAN_PROTOCOL_8021AD
		return nil
	}
}

// WithReorderingHeaders configures reordering of Ethernet headers.
func WithReorderingHeaders() Opt {
	return func(l *link.Link) error {
		l.Link.(*netlink.Vlan).ReorderHdr = neu(true)
		return nil
	}
}

// WithoutReorderingHeaders configures to not reorder Ethernet headers.
func WithoutReorderingHeaders() Opt {
	return func(l *link.Link) error {
		l.Link.(*netlink.Vlan).ReorderHdr = neu(false)
		return nil
	}
}

// WithGVRP configures that this VLAN is to be registered using the GARP VLAN
// Registration Protocol.
func WithGVRP() Opt {
	return func(l *link.Link) error {
		l.Link.(*netlink.Vlan).Gvrp = neu(true)
		return nil
	}
}

// WithoutGVRP configures that this VLAN must not be registered using the GARP
// VLAN Registration Protocol.
func WithoutGVRP() Opt {
	return func(l *link.Link) error {
		l.Link.(*netlink.Vlan).Gvrp = neu(false)
		return nil
	}
}

// WithMVRP configures that this VLAN is to be registered using the Multiple
// VLAN Registration Protocol.
func WithMVRP() Opt {
	return func(l *link.Link) error {
		l.Link.(*netlink.Vlan).Mvrp = neu(true)
		return nil
	}
}

// WithoutMVRP configures that this VLAN must not be registered using the
// Multiple VLAN Registration Protocol.
func WithoutMVRP() Opt {
	return func(l *link.Link) error {
		l.Link.(*netlink.Vlan).Mvrp = neu(false)
		return nil
	}
}

// WithLooseBinding configures the VLAN device state to be independent from the
// state of the bound network interface.
func WithLooseBinding() Opt {
	return func(l *link.Link) error {
		l.Link.(*netlink.Vlan).LooseBinding = neu(true)
		return nil
	}
}

// WithTightBinding configures the VLAN device state to be tightly coupled to
// the state of the bound network interface.
func WithTightBinding() Opt {
	return func(l *link.Link) error {
		l.Link.(*netlink.Vlan).LooseBinding = neu(false)
		return nil
	}
}

// neu provides pre-1.26 Go compatibility.
func neu[V any](v V) *V { return &v }
