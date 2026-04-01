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

package astiter_test

import (
	"go/ast"
	"go/types"

	"github.com/thediveo/notwork/gofix/astiter"
	"github.com/thediveo/notwork/gofix/sauce"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("iterating ASTs", func() {

	When("iterating over specific node types using AllOf", func() {

		var f *ast.File

		BeforeEach(func() {
			f = sauce.MustParse(`
package main
import "fmt"
var _ = fmt.Printf
var _ = fmt.Sprintf
`)
		})

		It("yields the correct AST nodes", func() {
			Expect(astiter.AllOf[ast.SelectorExpr](f)).To(ConsistOf(
				HaveField("Sel.Name", "Printf"),
				HaveField("Sel.Name", "Sprintf")))
		})

		It("correctly aborts iterations", func() {
			selexprs := []*ast.SelectorExpr{}
			for selexpr := range astiter.AllOf[ast.SelectorExpr](f) {
				selexprs = append(selexprs, selexpr)
				break
			}
			Expect(selexprs).To(HaveLen(1))
		})

	})

	When("iterating over specific object types using AllTypesOf", func() {

		var (
			f  *ast.File
			ti *types.Info
		)

		BeforeEach(func() {
			f, ti = sauce.MustParseAndTypeCheck(`
package main
import "fmt"
import "slices"
var (
	_ = fmt.Printf
	_ = fmt.Sprintf
	_ = slices.Collect[string]
)
`)
		})

		It("yields the correct func objects", func() {
			Expect(astiter.AllTypesOf[types.Func](f, ti)).To(ConsistOf(
				HaveField("FullName()", "fmt.Printf"),
				HaveField("FullName()", "fmt.Sprintf"),
				HaveField("FullName()", "slices.Collect")))
		})

		It("correctly aborts iterations", func() {
			fns := []*types.Func{}
			for fn := range astiter.AllTypesOf[types.Func](f, ti) {
				fns = append(fns, fn)
				break
			}
			Expect(fns).To(HaveLen(1))
		})

	})

	// nota bene: testing AllFuncsInPkg excercises AllInPkgOf and allTypesOf at
	// the same time.
	When("iterating over specific func types using AllFuncsInPkg", func() {

		var (
			f  *ast.File
			ti *types.Info
		)

		BeforeEach(func() {
			f, ti = sauce.MustParseAndTypeCheck(`
package main
import "fmt"
import "slices"
var (
	_ = fmt.Printf
	_ = fmt.Sprintf
	_ = slices.Collect[string]
)
`)
		})

		It("yields the correct funcs", func() {
			Expect(astiter.AllFuncsInPkg("fmt", f, ti)).To(ConsistOf(
				HaveField("FullName()", "fmt.Printf"),
				HaveField("FullName()", "fmt.Sprintf")))
		})

		It("correctly aborts iterations", func() {
			fns := []*types.Func{}
			for fn := range astiter.AllFuncsInPkg("fmt", f, ti) {
				fns = append(fns, fn)
				break
			}
			Expect(fns).To(HaveLen(1))
		})

	})

})
