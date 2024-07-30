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

package netdevsim

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MACVLAN configuration options", func() {

	It("configures veth", func() {
		o := &Options{}
		for _, opt := range []Opt{
			InNamespace(42),
			WithID(123),
			WithPorts(10),
			WithRxTxQueueCountEach(666),
		} {
			Expect(opt(o)).To(Succeed())
		}
		Expect(o.NetnsFd).To(Equal(42))
		Expect(o.HasID).To(BeTrue())
		Expect(o.ID).To(Equal(uint(123)))
		Expect(o.Ports).To(Equal(uint(10)))
		Expect(o.QueueCount).To(Equal(uint(666)))
	})

})
