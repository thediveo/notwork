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
	"fmt"
	"runtime"

	"golang.org/x/sys/unix"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// EnterTransientNetns creates and enters a new (and isolated) network
// namespace, returning a function that needs to be defer'ed in order to
// correctly switch the calling go routine and its locked OS-level thread back
// when the caller itself returns.
//
// In case the caller cannot be switched back correctly, the defer'ed clean up
// will panic with an error description.
func EnterTransientNetns() func() {
	GinkgoHelper()

	runtime.LockOSThread()
	netnsfd, err := unix.Open("/proc/thread-self/ns/net", unix.O_RDONLY, 0)
	Expect(err).NotTo(HaveOccurred(), "cannot determine current network namespace from procfs")
	unix.Unshare(unix.CLONE_NEWNET)
	return func() { // this cannot be DeferCleanup'ed: we need to restore the current locked go routine
		if err := unix.Setns(netnsfd, 0); err != nil {
			panic(fmt.Sprintf("cannot restore original network namespace, reason: %s", err.Error()))
		}
		unix.Close(netnsfd)
		runtime.UnlockOSThread()
	}
}
