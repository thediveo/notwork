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

package mntns

import (
	"os"
	"time"

	"github.com/thediveo/notwork/netns"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/success"
)

var _ = Describe("transient network namespaces", Ordered, func() {

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

	It("rejects mounting sysfs in the original mount namespace", func() {
		var r any
		func() {
			defer func() { r = recover() }()
			g := NewGomega(func(message string, callerSkip ...int) {
				panic(message)
			})
			mountSysfs(g, 0, "")
		}()
		Expect(r).To(ContainSubstring("current mount namespace must not be the process's original mount namespace"))
	})

	It("mounts a fresh sysfs (RO) in a transient mount namespace", func() {
		defer netns.EnterTransient()()
		Expect(len(Successful(os.ReadDir("/sys/class/net")))).To(BeNumerically(">", 1))

		defer EnterTransient()() // well, only for symmetry
		MountSysfsRO()

		Expect(Successful(os.ReadDir("/sys/class/net"))).To(
			ConsistOf(HaveField("Name()", "lo")))
	})

	It("creates a transient mount namespace and then mounts a new sysfs into it", func() {
		netnsfd := netns.NewTransient()

		hostMntnsID := Ino("/proc/thread-self/ns/mnt")

		mntnsfd, procfsroot := NewTransient()
		Expect(mntnsfd).NotTo(BeZero())
		Expect(procfsroot).NotTo(BeEmpty())

		Execute(mntnsfd, func() {
			defer GinkgoRecover()
			Expect(Ino("/proc/thread-self/ns/mnt")).NotTo(Equal(hostMntnsID))
			netns.Execute(netnsfd, func() {
				MountSysfsRO()
			})
		})

		Expect(Successful(os.ReadDir(procfsroot + "/sys/class/net"))).To(
			ConsistOf(HaveField("Name()", "lo")))
	})

})
