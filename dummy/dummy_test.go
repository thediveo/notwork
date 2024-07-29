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
	"errors"
	"os"
	"time"

	"github.com/thediveo/notwork/link"
	"github.com/vishvananda/netlink"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/success"
)

var canaryErr = errors.New("cheep cheep")

func withFailingOpt() Opt {
	return func(l *link.Link) error {
		return canaryErr
	}
}

var _ = Describe("creating transient dummy network interfaces", func() {

	BeforeEach(func() {
		if os.Getuid() != 0 {
			Skip("needs root")
		}

		goodfds := Filedescriptors()
		goodgos := Goroutines()
		DeferCleanup(func() {
			Eventually(Goroutines).Within(2 * time.Second).ProbeEvery(100 * time.Millisecond).
				ShouldNot(HaveLeaked(goodgos))
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
		})
	})

	It("throws a tantrum when an option fails", func() {
		Expect(InterceptGomegaFailure(func() { _ = NewTransient(withFailingOpt()) })).
			To(MatchError(ContainSubstring(canaryErr.Error())))
	})

	It("creates a transient dummy network interface and brings it up", func() {
		dl := NewTransientUp()
		Expect(dl.Attrs().Name).To(And(
			Not(Equal(DummyPrefix)),
			HavePrefix(DummyPrefix),
		))
		// Check that the network interface was in fact created.
		ql := Successful(netlink.LinkByIndex(dl.Attrs().Index))
		Expect(ql.Attrs().OperState).NotTo(Equal(netlink.OperDown))
	})

	When("using options", func() {

		It("configures a different destination network namespace", func() {
			lnk := &link.Link{
				Link: &netlink.Macvlan{},
			}
			Expect(InNamespace(-42)(lnk)).To(Succeed())
			Expect(lnk).To(HaveField("Attrs().Namespace", netlink.NsFd(-42)))
		})

	})

})
