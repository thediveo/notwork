// Copyright 2025 Harald Albrecht.
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

package load

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"pault.ag/go/modprobe"
)

var _ = Describe("loading and unloading the netsimdev bus", Ordered, func() {

	Context("when rootless", func() {

		BeforeAll(func() {
			if os.Getuid() == 0 {
				Skip("don't be root")
			}
		})

		It("always returns false", func() {
			Expect(TryRoot("/")).To(BeFalse())
		})

	})

	Context("when in command", Ordered, func() {

		BeforeAll(func() {
			if os.Getuid() != 0 {
				Skip("needs root")
			}
		})

		It("loads when needed", func() {
			info, err := os.Stat("/sys/bus/netdevsim")
			netdevsimPreloaded := err == nil && info.Mode().IsDir()
			defer func() {
				if netdevsimPreloaded {
					return // keep it loaded
				}
				Expect(modprobe.Remove("netdevsim")).To(Succeed(), "cannot unload netdevsim")
			}()

			_ = modprobe.Remove("netdevsim")
			Expect("/sys/bus/netdevsim").NotTo(BeAnExistingFile(), "netdevsim could not be unloaded for test")
			Expect(Try()).To(BeTrue())
			Expect("/sys/bus/netdevsim").To(BeADirectory())
			Expect(Try()).To(BeTrue())
			Expect("/sys/bus/netdevsim").To(BeADirectory())
		})

	})

})
