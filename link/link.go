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

package link

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/jinzhu/copier"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var fail = Fail // allow testing Fails without terminally failing the current test.

// NewTransient creates a transient network interface of the specified type (via
// the type of the link value passed in) and with a name that begins with the
// given prefix and a random string of 10 hex digits. The newly created link is
// additionally scheduled for deletion using Ginko's DeferCleanup.
//
// The passed link description is deep-copied first and thus never modified.
// Only the returned link description correctly references the newly created
// network interface (“link”).
//
// The newly created transient link starts in down operational state. Use
// [netlink.LinkSetUp] to bring its operational state “up”.
//
// If VETH link information is passed in, NewTransient will automatically
// populate the [netlink.Veth.PeerName] with a name that also begins with the
// given prefix and a random string of 10 hex digits.
//
// NewTransient remembers the network namespace the network interface was
// created in, so that it can correctly clean up the transient network interface
// later from one of Ginkgo's deferred cleanup handlers. This also covers the
// situation where the passed in link details reference a network namespace (in
// form of an open fd) different from the current network namespace.
func NewTransient(link netlink.Link, prefix string) netlink.Link {
	GinkgoHelper()

	Expect(link).NotTo(BeNil(), "need a non-nil link description")
	if _, ok := link.Attrs().Namespace.(netlink.NsFd); link.Attrs().Namespace != nil && !ok {
		fail("link.Attrs().Namespace reference must be nil or a netlink.NsFd")
	}

	newlink := reflect.New(reflect.ValueOf(link).Elem().Type()).Interface().(netlink.Link)
	Expect(copier.CopyWithOption(newlink, link, copier.Option{DeepCopy: true, IgnoreEmpty: true})).
		To(Succeed())
	link = newlink

	// We want to keep a netlink handle to the network namespace where the
	// network interface is to be created in, in order to later remove it in the
	// deferred cleanup handler. Now, the link information passed in may
	// reference a network namespace different from the current network
	// namespace, so we need to take care to get the netlink handle in the
	// correct network namespace.
	var netnsh *netlink.Handle
	var err error
	if link.Attrs().Namespace == nil {
		netnsh, err = netlink.NewHandle()
		Expect(err).NotTo(HaveOccurred(), "cannot create NETLINK handle")
	} else {
		// Type assertion is guarded by BeAssignableToTypeOf assertion above.
		netnsh, err = netlink.NewHandleAt(netns.NsHandle(link.Attrs().Namespace.(netlink.NsFd)))
		Expect(err).NotTo(HaveOccurred(), "cannot create NETLINK handle")
	}
	defer func() {
		// Only close the netlink handle when it wasn't captured for a deferred
		// cleanup and thus hasn't been set to nil.
		if netnsh != nil {
			netnsh.Close()
		}
	}()

	randbytes := make([]byte, 5)
	for attempt := 1; attempt <= 10; attempt++ {
		// Roll the dice to create a (new) random interface name...
		Expect(rand.Read(randbytes)).Error().NotTo(HaveOccurred())
		ifname := prefix + hex.EncodeToString(randbytes)
		link.Attrs().Name = ifname
		// If this is going to be a VETH peer-to-peer link, then also roll the
		// dice to create a random peer interface name...
		if veth, ok := link.(*netlink.Veth); ok {
			Expect(rand.Read(randbytes)).Error().NotTo(HaveOccurred())
			peername := prefix + hex.EncodeToString(randbytes)
			veth.PeerName = peername
		}
		// Try to create the link a see what happens...
		err := netlink.LinkAdd(link)
		if err == nil {
			// Phew, this worked.
			By(fmt.Sprintf("creating a transient network interface %q", link.Attrs().Name))
			// Note that in case of VETH pairs we only need to remove one end in
			// order to also remove the other end automatically. No dangling
			// virtual wires.
			{
				netnsh := netnsh // the deferred cleanup closure must capture the handle value copy.
				DeferCleanup(func() {
					defer func() {
						netnsh.Close() // finally release the netlink handle
					}()
					By(fmt.Sprintf("removing transient network interface %q", link.Attrs().Name))
					Expect(netnsh.LinkDel(link)).To(Succeed(), "cannot remove transient network interface %q", link.Attrs().Name)
				})
			}
			// tell the deferred handler (this is NOT the DeferCleanup handler)
			// to not close the netlink handle as it is still needed later by
			// the deferred cleanup handler.
			netnsh = nil
			return link
		}
		if !errors.Is(err, os.ErrExist) {
			fail(fmt.Sprintf("cannot create a transient network interface of type %q, reason: %v", link.Type(), err))
			return nil // not reachable
		}
	}
	fail(fmt.Sprintf("too many failed attempts to create a transient network interface of type %q", link.Type()))
	return nil // not reachable
}

// EnsureUp brings the specified network interface up and waits for it to become
// operationally “UP”. The maximum wait duration can be optionally specified; it
// defaults to 2s.
func EnsureUp(link netlink.Link, within ...time.Duration) {
	GinkgoHelper()
	ensureUp(Default, link, within...)
}

// ensureUp takes an additional Gomega in order to allow unit testing it.
func ensureUp(g Gomega, link netlink.Link, within ...time.Duration) {
	var atmost time.Duration
	switch len(within) {
	case 0:
		atmost = 2 * time.Second
	case 1:
		atmost = within[0]
	default:
		panic("only a single optional maximum wait duration allowed")
	}

	g.Eventually(func() netlink.LinkOperState {
		lnk, _ := netlink.LinkByIndex(link.Attrs().Index)
		return lnk.Attrs().OperState
	}).Within(atmost).ProbeEvery(100 * time.Millisecond).
		Should(Equal(netlink.LinkOperState(netlink.OperUp)))
}
