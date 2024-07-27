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
	"errors"
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"time"

	"github.com/jinzhu/copier"
	"github.com/thediveo/notwork/link/namespaced"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"golang.org/x/sys/unix"

	. "github.com/onsi/ginkgo/v2" //lint:ignore ST1001 rule does not apply
	. "github.com/onsi/gomega"    //lint:ignore ST1001 rule does not apply
)

var fail = Fail // allow testing Fails without terminally failing the current test.

// NewTransient creates a transient network interface of the specified type (via
// the type of the link value passed in) and with a name that begins with the
// given prefix and a random string of digits and uppercase and lowercase ASCII
// letters. These random network interface names will always be the maximum
// allowed length of 15 ASCII characters by the Linux kernel.
//
// The newly created link is automatically scheduled for deletion using Ginko's
// DeferCleanup. (See also notes below.)
//
// For typical use cases, you might want to look at these convenience functions
// instead:
//   - [github.com/thediveo/notwork/dummy.CreateTransient]
//   - [github.com/thediveo/notwork/macvlan.CreateTransient]
//   - [github.com/thediveo/notwork/veth.CreateTransient]
//
// The passed-in link description is deep-copied first and thus taken as a
// template, but never modified. On success, the returned link description then
// correctly references the newly created virtual network interface (“link”).
//
// The newly created transient link starts in down operational state, unless
// [netlink.LinkAttrs.Flags] has [net.FlagUp]. Alternatively, use
// [netlink.LinkSetUp] to bring the interface's operational state “up” in a
// guaranteed manner.
//
// If VETH link information is passed in, NewTransient will automatically
// populate the [netlink.Veth.PeerName] with a name that also begins with the
// given prefix and a random string of digits and lowercase and uppercase ASCII
// letters filling up the remaining part up to the maximum allowed interface
// name length in Linux.
//
// NewTransient remembers the network namespace the network interface was
// created in, so that it can correctly clean up the transient network interface
// later from one of Ginkgo's deferred cleanup handlers. This also covers the
// situation where the passed in link details reference a network namespace (in
// form of an open fd) different from the current network namespace.
//
// By setting the passed-in [netlink.Attrs.Namespace] and/or
// [netlink.Veth.PeerNamespace] it is possible to create the new virtual network
// in a different network namespace than the caller's current network namespace.
// The current network namespace still can play a role, such as when creating a
// MACVLAN network interface: then, the MACVLAN's parent network interface
// reference (in form of an interface index) must be in the scope of the current
// network namespace.
//
// Do not move a link to a different network namespace, as this interferes with
// the automated cleanup.
func NewTransient(link netlink.Link, prefix string) netlink.Link {
	GinkgoHelper()

	Expect(link).NotTo(BeNil(), "need a non-nil link description")
	if _, ok := link.Attrs().Namespace.(netlink.NsFd); link.Attrs().Namespace != nil && !ok {
		fail("link.Attrs().Namespace reference must be nil or a netlink.NsFd")
	}

	// Callers might pass in a wrapped.Link in order to transport network
	// namespace information, or they might not (especially external API
	// callers). So unwrap when necessary, keeping the piggy-backed link
	// namespace reference, if any.
	link, linkNamespace := namespaced.Unwrap(link)
	// Create a deep copy of the (unwrapped) link description.
	newlink := reflect.New(reflect.ValueOf(link).Elem().Type()).Interface().(netlink.Link)
	Expect(copier.CopyWithOption(newlink, link, copier.Option{DeepCopy: true, IgnoreEmpty: true})).
		To(Succeed())
	link = newlink

	// The caller might pass us an additional "link" network namespace, to use
	// Linux kernel terminology. This "link" network namespace is not to be
	// confused with netlink.LinkAttrs.Namespace, but instead specifies the
	// network namespace in which to start creation from in order to correctly
	// resolve parent/master link ifindex references.
	var linknetnsh *netlink.Handle // ...only needed temporarily
	if linkNamespace == nil {
		linknetnsh = &netlink.Handle{} // ...use the current network namespace
	} else {
		linknetnsfd, ok := linkNamespace.(netlink.NsFd)
		if !ok {
			fail("wrapped namespace.LinkNamespace must be nil or a netlink.NsFd")
		}
		var err error
		linknetnsh, err = netlink.NewHandleAt(netns.NsHandle(linknetnsfd))
		Expect(err).NotTo(HaveOccurred(), "cannot create NETLINK handle for link network namespace")
		defer linknetnsh.Close() // only needed momentarily
	}

	// We want to keep a netlink handle to the network namespace where the
	// network interface is to be created in (or more precise, to end up in), in
	// order to later remove it in the deferred cleanup handler. Now, the link
	// information passed in may reference a network namespace different from
	// the current network namespace, so we need to take care to get the netlink
	// handle in the correct network namespace.
	var netnsh *netlink.Handle // ...that should be needed till the end.
	var err error
	if link.Attrs().Namespace == nil {
		// Avoid promoting a potential circular dependency, so we get the
		// reference to the current network namespace by hand instead of using
		// the convenience function from the netns package; furthermore,
		// netns.Current arranges for a DeferCleanup that we don't want to be
		// done yet.
		netnsfd, err := unix.Open("/proc/thread-self/ns/net", unix.O_RDONLY, 0)
		defer unix.Close(netnsfd)
		Expect(err).NotTo(HaveOccurred(), "cannot determine current network namespace from procfs")
		netnsh, err = netlink.NewHandleAt(netns.NsHandle(netnsfd))
		Expect(err).NotTo(HaveOccurred(), "cannot create NETLINK handle for network namespace")
	} else {
		// Type assertion is guarded by BeAssignableToTypeOf assertion above.
		netnsh, err = netlink.NewHandleAt(netns.NsHandle(link.Attrs().Namespace.(netlink.NsFd)))
		Expect(err).NotTo(HaveOccurred(), "cannot create NETLINK handle")
	}

	defer func() {
		// Only close the netlink handle when it wasn't captured for a (Ginkgo)
		// deferred cleanup and thus hasn't been set to nil.
		if netnsh != nil {
			netnsh.Close()
		}
	}()

	for attempt := 1; attempt <= 10; attempt++ {
		// Roll the dice to create a (new) random interface name...
		ifname := base62Nifname(prefix)
		link.Attrs().Name = ifname
		// If this is going to be a VETH peer-to-peer link, then also roll the
		// dice to create a random peer interface name...
		if veth, ok := link.(*netlink.Veth); ok {
			peername := base62Nifname(prefix)
			veth.PeerName = peername
		}
		// Try to create the link and let's see what happens...
		err := linknetnsh.LinkAdd(link)
		if err != nil {
			// did we run just run into an accidentally duplicate random name,
			// or into a general error instead?
			if errors.Is(err, os.ErrExist) {
				continue
			}
			fail(fmt.Sprintf("cannot create a transient network interface of type %q, reason: %v", link.Type(), err))
		}
		// Phew, this worked.
		By(fmt.Sprintf("creating a transient network interface %q", link.Attrs().Name))
		// Work around a bug in vishvananda/netlink where the Index attribute
		// isn't updated correctly or even wrongly when
		// netlink.LinkAttrs.Namespace has been set.
		targetLink, err := netnsh.LinkByName(link.Attrs().Name)
		Expect(err).NotTo(HaveOccurred(), "cannot determine network interface index after creation")
		Expect(targetLink).NotTo(BeNil(), "cannot determine network interface index after creation")
		link.Attrs().Index = targetLink.Attrs().Index
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
	fail(fmt.Sprintf("too many failed attempts to create a transient network interface of type %q", link.Type()))
	return nil // not reachable
}

// EnsureUp brings the specified network interface up and waits for it to become
// operationally “UP” or “UNKNOWN”. The maximum wait duration can be optionally
// specified; it defaults to 2s.
func EnsureUp(link netlink.Link, within ...time.Duration) {
	GinkgoHelper()
	ensureUp(Default, link, false, within...)
}

// ensureUp takes an additional Gomega in order to allow unit testing it.
func ensureUp(g Gomega, link netlink.Link, skipup bool, within ...time.Duration) {
	GinkgoHelper()

	g.Expect(link).NotTo(BeNil(), "need a non-nil link description")

	var atmost time.Duration
	switch len(within) {
	case 0:
		atmost = 2 * time.Second
	case 1:
		atmost = within[0]
	default:
		panic("only a single optional maximum wait duration allowed")
	}

	if !skipup {
		g.Expect(netlink.LinkSetUp(link)).To(Succeed())
	}
	g.Eventually(func() bool {
		lnk, err := netlink.LinkByIndex(link.Attrs().Index)
		if err != nil {
			StopTrying("link cannot come up").Wrap(err).Now()
		}
		switch lnk.Attrs().OperState {
		case netlink.LinkOperState(netlink.OperUp):
			return true
		case netlink.LinkOperState(netlink.OperUnknown):
			return true
		}
		return false
	}).Within(atmost).ProbeEvery(20 * time.Millisecond).
		Should(BeTrue())
}

// RandomNifname returns a network interface name consisting of the specified
// prefix and a random string, and of the maximum length allowed for network
// interface names. The random string part consists of only digits as well as
// lowercase and uppercase ASCII letters.
func RandomNifname(prefix string) string {
	GinkgoHelper()
	return base62Nifname(prefix)
}

// Maximum allowed length for Linux network interface names.
const maxNifnameLen = 15

// Minimum of random base63 characters required.
const minRandomLen = 4

// The set of characters to create a random string from.
const base62chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// base62Nifname returns a random network interface name consisting of the
// specified prefix and a random string, and of the maximum length allowed for
// network interface names. The random string part consists of only digits as
// well as lowercase and uppercase ASCII letters.
func base62Nifname(prefix string) string {
	GinkgoHelper()
	if len(prefix) > maxNifnameLen-minRandomLen {
		fail(fmt.Sprintf("cannot create random network interface name, because prefix %q is longer than %d characters",
			prefix, maxNifnameLen-4))
	}
	name := make([]byte, maxNifnameLen)
	copy(name, prefix)
	for idx := len(prefix); idx < maxNifnameLen; idx++ {
		name[idx] = base62chars[rand.Intn(len(base62chars))]
	}
	return string(name)
}
