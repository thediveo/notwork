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

package sauce_test

import (
	"go/ast"
	"go/types"

	"github.com/thediveo/notwork/gofix/sauce"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("working with Go sources", func() {

	When("only parsing a Go source", func() {

		It("succeeds for a valid sauce", func() {
			var f *ast.File
			Expect(InterceptGomegaFailure(func() {
				f = sauce.MustParse(`
package main

func main() {
	_ = unknown.Symbol
}
`)
			})).To(Succeed())
			Expect(f).NotTo(BeNil())
		})

	})

	When("parsing and type-checking a Go source", func() {

		It("succeeds for a valid sauce", func() {
			var f *ast.File
			var ti *types.Info
			Expect(InterceptGomegaFailure(func() {
				f, ti = sauce.MustParseAndTypeCheck(`
package main

import "fmt"

func main() {
	println(fmt.Sprintf("%s", "hooray!"))
}
`)
			})).To(Succeed())
			Expect(f).NotTo(BeNil())
			Expect(ti).NotTo(BeNil())
		})

		It("rejects invalid sauce", func() {
			Expect(InterceptGomegaFailure(func() {
				_, _ = sauce.MustParseAndTypeCheck(`
package main

package main
`)
			})).To(MatchError(ContainSubstring("has parsing errors")))
		})

		It("rejects invalid types", func() {
			Expect(InterceptGomegaFailure(func() {
				_, _ = sauce.MustParseAndTypeCheck(`
package main

var _ = thisDoesNotExist
`)
			})).To(MatchError(ContainSubstring("fails to type-check")))
		})

	})

	When("looking for specific types of AST nodes", func() {

		It("returns the correct node", func() {
			var f *ast.File
			f, _ = sauce.MustParseAndTypeCheck(`
package main

import "fmt"

func main() {
fmt.Sprintf("%s", "hooray!")
}
`)
			var node *ast.SelectorExpr
			Expect(InterceptGomegaFailure(func() {
				node = sauce.FirstOf[ast.SelectorExpr](f)
			})).To(Succeed())
			Expect(node).NotTo(BeNil())
		})

		It("fails if no node of this type can be found", func() {
			var f *ast.File
			f, _ = sauce.MustParseAndTypeCheck(`
package main
`)
			Expect(InterceptGomegaFailure(func() {
				_ = sauce.FirstOf[ast.SelectorExpr](f)
			})).To(MatchError(ContainSubstring("no node of type *ast.SelectorExpr found")))
		})

	})

})
