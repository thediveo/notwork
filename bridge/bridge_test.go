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
	"os"
	"time"

	"github.com/thediveo/spacetest/netns"
	"github.com/vishvananda/netlink"

	"github.com/thediveo/notwork/internal/auxvalues"
	"github.com/thediveo/notwork/internal/neu"
	"github.com/thediveo/notwork/veth"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/success"
)

var _ = Describe("provides transient bridge network interfaces", func() {

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

	It("rejects creating a bridge with invalid configuration", func() {
		Expect(InterceptGomegaFailure(func() {
			_ = NewTransient(WithAgeing(-1 * time.Second))
		})).NotTo(Succeed())
	})

	It("creates a bridge and adds a port", func() {
		defer netns.EnterTransient()()

		By("creating a bridge")
		br := NewTransient()
		brchk := Successful(netlink.LinkByName(br.Attrs().Name))
		Expect(brchk).To(HaveField("AgeingTime", neu.Value(uint32(300*auxvalues.CLKTCK()))))

		By("creating a veth pair and adding one end to the bridge")
		there := netns.NewTransient()
		dupond, _ := veth.NewTransient(veth.WithPeerNamespace(there))
		AddPort(br, dupond)

		dupond = Successful(netlink.LinkByName(dupond.Attrs().Name))
		Expect(dupond.Attrs().MasterIndex).To(Equal(br.Attrs().Index))
	})

})
