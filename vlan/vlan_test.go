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

package vlan

import (
	"os"
	"time"

	"github.com/thediveo/spacetest/netns"

	"github.com/thediveo/notwork/dummy"
	"github.com/thediveo/notwork/nlhandle"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/success"
)

var _ = Describe("provides transient VLAN network interfaces", Ordered, func() {

	BeforeEach(func() {
		if os.Getuid() != 0 {
			Skip("needs root")
		}
		goodfds := Filedescriptors()
		goodgos := Goroutines()
		DeferCleanup(func() {
			Eventually(Goroutines).Within(2 * time.Second).ProbeEvery(250 * time.Millisecond).
				ShouldNot(HaveLeaked(goodgos))
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
		})
	})

	It("creates a VLAN with a dummy parent and a configuration option", func() {
		defer netns.EnterTransient()()
		_ = NewTransient(999, dummy.NewTransientUp(), WithLooseBinding())
	})

	It("creates a VLAN with its parent in a different network namespace", func() {
		dmyNetnsfd := netns.NewTransient()
		dmy := dummy.NewTransient(dummy.InNamespace(dmyNetnsfd))

		destNetnsfd := netns.NewTransient()
		vlan := NewTransient(999, dmy,
			InNamespace(destNetnsfd),
			WithLinkNamespace(dmyNetnsfd))
		Expect(vlan.Attrs().Index).NotTo(BeZero())

		destnlh := nlhandle.New(destNetnsfd)
		Expect(Successful(destnlh.LinkByName(vlan.Attrs().Name))).To(
			HaveField("Attrs().Index", vlan.Attrs().Index))
	})

})
