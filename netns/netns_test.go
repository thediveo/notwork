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

package netns

import (
	"os"
	"runtime"
	"syscall"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/success"
)

var _ = Describe("transient network namespaces", Ordered, func() {

	BeforeEach(func() {
		if os.Getuid() != 0 {
			Skip("needs root")
		}
		goodfds := Filedescriptors()
		DeferCleanup(func() {
			Eventually(Filedescriptors).Within(2 * time.Second).ProbeEvery(250 * time.Millisecond).
				ShouldNot(HaveLeakedFds(goodfds))
		})
	})

	It("creates, enters, and leaves a transient network namespace", func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		initialNetnsInfo := Successful(os.Stat("/proc/thread-self/ns/net"))

		By("creating and entering a new network namespace")
		f := EnterTransientNetns()
		currentNetnsInfo := Successful(os.Stat("/proc/thread-self/ns/net"))
		Expect(initialNetnsInfo.Sys().(*syscall.Stat_t).Ino).NotTo(
			Equal(currentNetnsInfo.Sys().(*syscall.Stat_t).Ino))

		By("switching back into the original network namespace")
		Expect(f).NotTo(Panic())
		currentNetnsInfo = Successful(os.Stat("/proc/thread-self/ns/net"))
		Expect(initialNetnsInfo.Sys().(*syscall.Stat_t).Ino).To(
			Equal(currentNetnsInfo.Sys().(*syscall.Stat_t).Ino))
	})

})
