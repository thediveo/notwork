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

package vxlan

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

var _ = Describe("provides transient VXLAN network interfaces", Ordered, func() {

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

	It("creates a VXLAN with a dummy underlasy and a configuration option", func() {
		defer netns.EnterTransient()()
		_ = NewTransient(dummy.NewTransientUp(), WithID(666))
	})

	It("creates a MACVLAN with its parent in a different network namespace", func() {
		dmyNetnsfd := netns.NewTransient()
		dmy := dummy.NewTransient(dummy.InNamespace(dmyNetnsfd))

		destNetnsfd := netns.NewTransient()
		vxlan := NewTransient(dmy,
			InNamespace(destNetnsfd),
			WithLinkNamespace(dmyNetnsfd))
		Expect(vxlan.Attrs().Index).NotTo(BeZero())

		destnlh := nlhandle.New(destNetnsfd)
		Expect(Successful(destnlh.LinkByName(vxlan.Attrs().Name))).To(
			HaveField("Attrs().Index", vxlan.Attrs().Index))
	})

})
