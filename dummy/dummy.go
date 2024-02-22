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

package dummy

import (
	"github.com/thediveo/notwork/link"
	"github.com/vishvananda/netlink"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// DummyPrefix is the name prefix used for transient dummy network interfaces.
const DummyPrefix = "dumy-"

// NewTransient creates a transient network interface of type “[dummy]”. It does
// not configure any IP address(es) though. NewTransient automatically defers
// proper automatic removal of the dummy network interface.
//
// [dummy]: https://tldp.org/LDP/nag/node72.html
func NewTransient() netlink.Link {
	GinkgoHelper()
	return link.NewTransient(&netlink.Dummy{}, DummyPrefix)
}

// NewTransientUp creates a transient network interface of type “[dummy]” and
// additionally brings it up. It does not configure any IP address(es) though.
// NewTransient automatically defers proper automatic removal of the dummy
// network interface.
//
// [dummy]: https://tldp.org/LDP/nag/node72.html
func NewTransientUp() netlink.Link {
	GinkgoHelper()
	dummy := NewTransient()
	Expect(netlink.LinkSetUp(dummy)).To(
		Succeed(), "cannot bring transient interface %q up", dummy.Attrs().Name)
	return dummy
}
