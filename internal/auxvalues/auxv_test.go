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
	"encoding/binary"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("auxiliary vector", func() {

	It("returns a non-zero USER_HZ", func() {
		Expect(CLKTCK()).To(BeNumerically(">", 0))
	})

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

	It("returns an empty mapping for an empty auxiliary vector", func() {
		Expect(parseauxv(nil, 8)).To(BeEmpty())
	})

	It("stops reading when AT_NULL is encountered", func() {
		auxv := make([]byte, 32)
		binary.NativeEndian.PutUint64(auxv[0:8], 17)
		binary.NativeEndian.PutUint64(auxv[8:16], 100)
		binary.NativeEndian.PutUint64(auxv[16:24], 0)
		binary.NativeEndian.PutUint64(auxv[24:32], 999)

		result := parseauxv(auxv, 8)
		Expect(result).To(HaveLen(1))
		Expect(result).To(HaveKeyWithValue(ulong(17), ulong(100)))
	})

	It("ignores incomplete auxiliary vector entries", func() {
		auxv := make([]byte, 12)
		binary.NativeEndian.PutUint64(auxv[0:8], 17)

		result := parseauxv(auxv, 8)
		Expect(result).To(BeEmpty())
	})

	It("reads 32-bit auxiliary vector entries", func() {
		auxv := make([]byte, 12)
		binary.NativeEndian.PutUint32(auxv[0:4], 17)
		binary.NativeEndian.PutUint32(auxv[4:8], 100)
		binary.NativeEndian.PutUint32(auxv[8:12], 0)

		result := parseauxv(auxv, 4)
		Expect(result).To(HaveKeyWithValue(ulong(17), ulong(100)))
	})

	It("panics when the word size is unsupported", func() {
		Expect(func() {
			parseauxv(nil, 16)
		}).To(PanicWith("unsupported auxv word size of 16"))
	})

	It("reads 64-bit auxiliary vector entries", func() {
		auxv := make([]byte, 16)
		binary.NativeEndian.PutUint64(auxv[0:8], 17)
		binary.NativeEndian.PutUint64(auxv[8:16], 100)

		result := parseauxv(auxv, 8)
		Expect(result).To(HaveKeyWithValue(ulong(17), ulong(100)))
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
