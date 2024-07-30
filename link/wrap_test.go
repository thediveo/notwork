// Copyright 2024 Harald Albrecht.
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

package link

import (
	"github.com/vishvananda/netlink"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("wrapped links", func() {

	It("handles nil", func() {
		lnk, netns := Unwrap(nil)
		Expect(lnk).To(BeNil())
		Expect(netns).To(BeNil())
	})

	It("returns an unwrapped, native netlink.Link unmodified", func() {
		lnk, netns := Unwrap(&netlink.GenericLink{})
		Expect(lnk).To(BeAssignableToTypeOf(&netlink.GenericLink{}))
		Expect(netns).To(BeNil())
	})

	It("unwraps a namespaced netlink.Link", func() {
		expectedFd := 42
		lnk, netns := Unwrap(WrapWithLinkNamespace(
			&netlink.Dummy{},
			expectedFd))
		Expect(lnk).To(BeAssignableToTypeOf(&netlink.Dummy{}))
		Expect(netns).To(Equal(netlink.NsFd(expectedFd)))
	})

	It("wraps if unwrapped", func() {
		w := EnsureWrap(&netlink.GenericLink{})
		Expect(w.(*Link)).NotTo(BeNil())
		var gen *netlink.GenericLink
		Expect(EnsureWrap(w).(*Link).Link).To(BeAssignableToTypeOf(gen))
	})

})
