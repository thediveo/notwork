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

package nkwrap

import (
	"reflect"
	"unsafe"

	"github.com/vishvananda/netlink"
)

// Netkit wraps netlink's slightly broken Netkit link type that has no way to
// retrieve the peer attributes after setting them. We work around this API
// quirk by using our own exported peer link attributes and copying them over
// using SetPeerAttrs only just before the link creation.
type Netkit struct {
	netlink.Netkit
	Peer netlink.LinkAttrs
}

// UpdatePrivatePeer updates netlink.Netkit's private peer link attributes from
// the wrapper's Peer link attributes.
func (n *Netkit) UpdatePrivatePeer() {
	n.SetPeerAttrs(&n.Peer)
}

// RetrievePrivatePeer updates this wrapper's Peer link attributes from the
// wrapped netlink.Netkit's private peer link attributes.
func (n *Netkit) RetrievePrivatePeer() {
	peerLinkAttrs := reflect.ValueOf(n.Netkit).Elem().FieldByName("peerLinkAttrs")
	n.Peer = *(*netlink.LinkAttrs)(unsafe.Pointer(peerLinkAttrs.UnsafeAddr()))
}
