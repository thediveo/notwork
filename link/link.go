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
// network interface ("link").
//
// The newly created transient link starts in down operational state. Use
// [netlink.LinkSetUp] to bring its operational state "up".
func NewTransient(link netlink.Link, prefix string) netlink.Link {
	GinkgoHelper()

	Expect(link).NotTo(BeNil(), "need a non-nil link description")
	newlink := reflect.New(reflect.ValueOf(link).Elem().Type()).Interface().(netlink.Link)
	copier.CopyWithOption(newlink, link, copier.Option{DeepCopy: true, IgnoreEmpty: true})
	link = newlink

	randbytes := make([]byte, 5)
	for attempt := 1; attempt <= 10; attempt++ {
		rand.Read(randbytes)
		ifname := prefix + hex.EncodeToString(randbytes)
		link.Attrs().Name = ifname
		err := netlink.LinkAdd(link)
		if err == nil {
			// Phew, this worked.
			By(fmt.Sprintf("creating a transient network interface %q", link.Attrs().Name))
			DeferCleanup(func() {
				By(fmt.Sprintf("removing transient network interface %q", link.Attrs().Name))
				Expect(netlink.LinkDel(link)).To(Succeed(), "cannot remove transient network interface %q", link.Attrs().Name)
			})
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
