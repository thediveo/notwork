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

	"github.com/thediveo/notwork/link"
)

// InNamespace configures the “primary” or “host” netkit network interface to
// be created in the network namespace referenced by fdref, instead of creating
// it in the current network namespace. The “peer” netkit network interface will
// be created in the current network namespace, use [WithPeerNamespace] to
// create that end in a different network namespace.
func InNamespace(fdref int) Opt {
	return func(l *link.Link) error {
		l.Link.(*Netkit).Namespace = netlink.NsFd(fdref)
		return nil
	}
}

// WithPeerNamespace configures the “peer” netkit network interface to be
// created inside the network namespace referenced by fd.
func WithPeerNamespace(fdref int) Opt {
	return func(v *link.Link) error {
		v.Link.(*Netkit).Peer.Namespace = netlink.NsFd(fdref)
		return nil
	}
}

// WithIPMode configures a netkit pair to carry IP L3 traffic only.
func WithIPMode() Opt {
	return func(l *link.Link) error {
		l.Link.(*Netkit).Mode = netlink.NETKIT_MODE_L3
		return nil
	}
}

// WithL2Mode configures a netkit pair to carry Ethernet L2 traffic.
func WithL2Mode() Opt {
	return func(l *link.Link) error {
		l.Link.(*Netkit).Mode = netlink.NETKIT_MODE_L2
		return nil
	}
}

// WithPolicy configures the policy of the “primary” or “host” netkit network
// interface: either forwarding (default) or black-holing.
func WithPolicy(pol netlink.NetkitPolicy) Opt {
	return func(l *link.Link) error {
		l.Link.(*Netkit).Policy = pol
		return nil
	}
}

// WithPeerPolicy configures the policy of the peer” netkit network interface:
// either forwarding (default) or black-holing.
func WithPeerPolicy(pol netlink.NetkitPolicy) Opt {
	return func(l *link.Link) error {
		l.Link.(*Netkit).PeerPolicy = pol
		return nil
	}
}

// WithScrub configures scrubbing the “primary” or “host” netkit network
// interface: either no scrubbing at all (default), or scrubbing (confusingly
// named netlink.NETKIT_SCRUB_DEFAULT) the mark and priority skb fields.
func WithScrub(scrub netlink.NetkitScrub) Opt {
	return func(l *link.Link) error {
		l.Link.(*Netkit).Scrub = scrub
		return nil
	}
}

// WithPeerScrub configures scrubbing the peer netkit network interface: either
// no scrubbing at all (default), or scrubbing (confusingly named
// netlink.NETKIT_SCRUB_DEFAULT) the mark and priority skb fields.
func WithPeerScrub(scrub netlink.NetkitScrub) Opt {
	return func(l *link.Link) error {
		l.Link.(*Netkit).PeerScrub = scrub
		return nil
	}
}

// WithHeadroom configures the socket buffer headroom for both netkit network
// interfaces.
func WithHeadroom(r uint16) Opt {
	return func(l *link.Link) error {
		l.Link.(*Netkit).DesiredHeadroom = r
		return nil
	}
}

// WithTailroom configures the socket buffer tailroom for both netkit network
// interfaces.
func WithTailroom(r uint16) Opt {
	return func(l *link.Link) error {
		l.Link.(*Netkit).DesiredTailroom = r
		return nil
	}
}

// WithRxQueues configures the number of RX queues for the “primary” or “host”
// netkit network interface.
func WithRxQueues(n int) Opt {
	return func(l *link.Link) error {
		l.Link.(*Netkit).NumRxQueues = n
		return nil
	}
}

// WithTxQueues configures the number of TX queues for the “primary” or “host”
// netkit network interface.
func WithTxQueues(n int) Opt {
	return func(l *link.Link) error {
		l.Link.(*Netkit).NumTxQueues = n
		return nil
	}
}

// WithPeerRxQueues configures the number of RX queues for the “peer” netkit
// network interface.
func WithPeerRxQueues(n int) Opt {
	return func(l *link.Link) error {
		l.Link.(*Netkit).Peer.NumRxQueues = n
		return nil
	}
}

// WithPeerTxQueues configures the number of TX queues for the “peer” netkit
// network interface.
func WithPeerTxQueues(n int) Opt {
	return func(l *link.Link) error {
		l.Link.(*Netkit).Peer.NumTxQueues = n
		return nil
	}
}
