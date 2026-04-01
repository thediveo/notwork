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

package imports

import (
	"github.com/thediveo/notwork/gofix/sauce"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/thediveo/testily/tuples"
)

var _ = Describe("imports", func() {

	Context("determining local names of packages", func() {

		FIt("...", func() {
			Expect(packageLocalName("golang.org/x/tools/go/packages", "..")).To(Equal("packages"))
			Expect(packageLocalName("github.com/thediveo/notwork/netns", "..")).To(Equal("gofix"))
		})

	})

	Context("local names and paths of imports", func() {

		It("returns correct local names and path", func() {
			f := sauce.MustParse(`
package main
import "go/ast"
import abc "fmt"
var _ = ast
var _ = abc
`)
			Expect(PackPair(LocalNameAndPath(f.Imports[0]))).To(
				Equal(PackPair("ast", "go/ast")))
			Expect(PackPair(LocalNameAndPath(f.Imports[1]))).To(
				Equal(PackPair("abc", "fmt")))
		})

	})

	Context("looking up and adding local names for imports", func() {

		var imps *Map

		BeforeEach(func() {
			f := sauce.MustParse(`
package main
import "go/ast"
var _ = ast
`)
			imps = NewMap(f)
		})

		It("returns an existing import", func() {
			Expect(imps.LocalNameOf("go/ast")).To(Equal("ast"))
		})

		It("adds a new import (only once)", func() {
			Expect(imps.LocalNameOf("github.com/thediveo/notwork/netns")).To(Equal("netns"))
			Expect(imps.LocalNameOf("github.com/thediveo/notwork/netns")).To(Equal("netns"))
		})

		It("adds a new import with local name differentiation", func() {
			Expect(imps.LocalNameOf("github.com/thediveo/notwork/netns")).To(Equal("netns"))
			Expect(imps.LocalNameOf("github.com/thediveo/spacetest/netns")).To(Equal("spacetestnetns"))
			Expect(imps.LocalNameOf("github.com/thediveo/otherrepo/spacetest/netns")).To(Equal("spacetestnetns1"))
		})

	})

	It("maps imports", func() {
		f := sauce.MustParse(`
package main
import "go/ast"
import abc "fmt"
var _ = ast
var _ = abc
`)
		imps := NewMap(f)
		Expect(imps.byPath).To(HaveKeyWithValue("fmt", "abc"))
	})

})
