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

package link

import "github.com/vishvananda/netlink"

// Link wraps a netlink.Link and adds an (optional) network namespace reference,
// understood by link.NewTransient to mean that the link described is to be
// created in the referenced network namespace, and not necessarily in the
// current network namespace.
//
// This slight cludge allows to keep the existing API intact, while shimming in
// the ability to tell link.NewTransient to start from a different network
// namespace and not from the current one, so link references such as a MACVLAN
// parent can be properly resolved.
type Link struct {
	netlink.Link
	LinkNamespace any // nil | NsPid | NsFd ... we follow the netns reference pattern used in the netlink package
}

var _ (netlink.Link) = (*Link)(nil)

// WrapWithLinkNamespace returns a wrapped netlink.Link namespaced to the
// network namespace referenced by the passed netnsfd.
func WrapWithLinkNamespace(link netlink.Link, netnsfd int) netlink.Link {
	return &Link{
		Link:          link,
		LinkNamespace: netlink.NsFd(netnsfd),
	}
}

// EnsureWrap always returns a wrapper Link, wrapping the passed link where
// necessary.
func EnsureWrap(link netlink.Link) netlink.Link {
	if _, ok := link.(*Link); ok {
		return link
	}
	return &Link{Link: link}
}

// Unwrap takes a potentally wrapped netlink.Link, unwraps the original
// netlink.Link if necessary, and returns it, together with an optional network
// namespace reference where the link should be created in.
func Unwrap(link netlink.Link) (l netlink.Link, namespace any) {
	if wl, ok := link.(*Link); ok {
		return wl.Link, wl.LinkNamespace
	}
	return link, nil
}
