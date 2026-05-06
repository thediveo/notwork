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
	"errors"
	"math"
	"time"

	"github.com/vishvananda/netlink"

	"github.com/thediveo/notwork/internal/auxvalues"
	"github.com/thediveo/notwork/internal/neu"
	"github.com/thediveo/notwork/link"
)

// InNamespace configures a bridge network interface to be created in the
// network namespace referenced by fdref, instead of creating it in the current
// network namespace.
func InNamespace(fdref int) Opt {
	return func(br *link.Link) error {
		br.Attrs().Namespace = netlink.NsFd(fdref)
		return nil
	}
}

// WithLinkNamespace specifies the “reference” or “link” network namespace other
// than the current network namespace when creating a new network interface.
func WithLinkNamespace(fdref int) Opt {
	return func(br *link.Link) error {
		br.LinkNamespace = netlink.NsFd(fdref)
		return nil
	}
}

// WithVLANFiltering configures the bridge with VLAN filtering.
func WithVLANFiltering() Opt {
	return func(br *link.Link) error {
		br.Link.(*netlink.Bridge).VlanFiltering = neu.Value(true)
		return nil
	}
}

// WithoutVLANFiltering configures the bridge to not use VLAN filtering.
func WithoutVLANFiltering() Opt {
	return func(br *link.Link) error {
		br.Link.(*netlink.Bridge).VlanFiltering = neu.Value(false)
		return nil
	}
}

const microsecondsPerSecond = uint64(time.Second / time.Microsecond)

// WithAgeing configures the bridge's FDB entries ageing time (duration). Please
// note that this duration is rounded down to the corresponding USER_HZ-granular
// duration. The [default duration] is 300s.
//
// [default duration]: https://www.kernel.org/doc/html/latest/networking/bridge.html#bridge-netlink-attributes
func WithAgeing(d time.Duration) Opt {
	return func(br *link.Link) error {
		mus := d.Microseconds()
		if mus < 0 || mus > math.MaxUint32 {
			return errors.New("ageing duration out of range")
		}
		br.Link.(*netlink.Bridge).AgeingTime = neu.Value(uint32(
			uint64(mus) / (microsecondsPerSecond / auxvalues.CLKTCK())))
		return nil
	}
}

// WithHello configures the time (duration) between hello packets sent by the
// bridge. Please note that this duration is rounded down to the corresponding
// USER_HZ-granular duration. The [default duration] is 2s.
//
// [default duration]: https://www.kernel.org/doc/html/latest/networking/bridge.html#bridge-netlink-attributes
func WithHello(d time.Duration) Opt {
	return func(br *link.Link) error {
		mus := d.Microseconds()
		if mus < 0 || mus > math.MaxUint32 {
			return errors.New("hello duration out of range")
		}
		br.Link.(*netlink.Bridge).HelloTime = neu.Value(uint32(
			uint64(mus) / (microsecondsPerSecond / auxvalues.CLKTCK())))
		return nil
	}
}

// WithMulticastSnooping configures multicast snooping. The default is for
// multicast snooping to be turned on.
func WithMulticastSnooping() Opt {
	return func(br *link.Link) error {
		br.Link.(*netlink.Bridge).MulticastSnooping = neu.Value(true)
		return nil
	}
}

// WithoutMulticastSnooping configures multicast snooping to be turned off.
func WithoutMulticastSnooping() Opt {
	return func(br *link.Link) error {
		br.Link.(*netlink.Bridge).MulticastSnooping = neu.Value(false)
		return nil
	}
}
