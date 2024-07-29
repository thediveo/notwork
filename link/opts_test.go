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

var _ = Describe("link configuration options", func() {

	It("configure a new link", func() {
		lnk := &Link{
			Link: &netlink.GenericLink{},
		}
		for _, opt := range []Opt{
			WithLinkNamespace(42),
			InNamespace(666),
		} {
			Expect(opt(lnk)).To(Succeed())
		}
		Expect(lnk.LinkNamespace).To(Equal(netlink.NsFd(42)))
		Expect(lnk.Attrs().Namespace).To(Equal(netlink.NsFd(666)))
	})

})
