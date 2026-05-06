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

package auxvalues

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("auxiliary vector", func() {

	It("returns the default value for a nil aux vector", func() {
		Expect(auxvMap(nil).Get(AT_NULL, 1234567)).To(Equal(ulong(1234567)))
	})

	It("returns the default value when the type can't be found in the aux vector", func() {
		Expect(auxvMap{}.Get(AT_NULL, 1234567)).To(Equal(ulong(1234567)))
	})

	It("returns the typed value", func() {
		Expect(auxvMap{AT_CLKTCK: 100}.Get(AT_CLKTCK, 123)).To(Equal(ulong(100)))
	})

	It("returns an empty mapping when the aux vector data cannot be read", func() {
		Expect(readauxv("/dev/nullllll")).To(BeEmpty())
	})

	It("reads /proc/self/auxv successfully", func() {
		Expect(readauxv(procSelfAuxv)).NotTo(BeEmpty())
	})

	It("returns the correct EUID", func() {
		const AT_EUID = 12
		expected := ulong(os.Geteuid())
		if expected == 0 {
			// eh? user rude gets -1 AT_EUID?
			expected = ^ulong(0)
		}
		Expect(readauxv(procSelfAuxv).Get(AT_EUID, ^ulong(0))).To(
			Equal(expected))
	})

	It("returns CLKTCK", func() {
		Expect(typedValues).NotTo(BeNil())
		Expect(typedValues).To(HaveKey(ulong(AT_CLKTCK)))
		Expect(CLKTCK).NotTo(BeZero())
	})

})
