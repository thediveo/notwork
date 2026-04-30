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

// WithID configures the VXLAN ID.
func WithID(id int) Opt {
	return func(l *link.Link) error {
		l.Link.(*netlink.Vxlan).VxlanId = id
		return nil
	}
}

// WithDestinationPort configures the UDP destination port for communicating to
// the remote VXLAN tunnel endpoint.
func WithDestinationPort(port uint16) Opt {
	return func(l *link.Link) error {
		l.Link.(*netlink.Vxlan).Port = int(port)
		return nil
	}
}

// WithTTL configures the TTL value to use in outgoing packets.
func WithTTL(ttl uint8) Opt {
	return func(l *link.Link) error {
		l.Link.(*netlink.Vxlan).TTL = int(ttl)
		return nil
	}
}

// WithSourcePorts configures the range of port numbers to use as UDP source
// ports when communicating to the remote VXLAN tunnel endpoint.
func WithSourcePorts(minport, maxport uint16) Opt {
	return func(l *link.Link) error {
		minport, maxport = min(minport, maxport), max(minport, maxport)
		vxlan := l.Link.(*netlink.Vxlan)
		vxlan.PortLow = int(minport)
		vxlan.PortHigh = int(maxport)
		return nil
	}
}
