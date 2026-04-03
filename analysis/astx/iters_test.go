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

package astx_test

import (
	"go/ast"
	"go/types"

	"github.com/thediveo/notwork/analysis/astx"
	"github.com/thediveo/notwork/analysis/testsauce"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("iterating ASTs", func() {

	When("iterating over specific node types using AllOf", func() {

		var f *testsauce.File

		BeforeEach(func() {
			f = testsauce.MustParse(`
package main
import "fmt"
var _ = fmt.Printf
var _ = fmt.Sprintf
`)
		})

		It("yields the correct AST nodes", func() {
			Expect(astx.AllOf[*ast.SelectorExpr](f.File())).To(ConsistOf(
				HaveField("Sel.Name", "Printf"),
				HaveField("Sel.Name", "Sprintf")))
		})

		It("correctly aborts iterations", func() {
			selexprs := []*ast.SelectorExpr{}
			for selexpr := range astx.AllOf[*ast.SelectorExpr](f.File()) {
				selexprs = append(selexprs, selexpr)
				break
			}
			Expect(selexprs).To(HaveLen(1))
		})

	})

	When("iterating over specific object types using AllTypesOf", func() {

		var (
			f  *testsauce.File
			ti *types.Info
		)

		BeforeEach(func() {
			f, _, ti = testsauce.MustParse(`
package main
import "fmt"
import "slices"
var (
	_ = fmt.Printf
	_ = fmt.Sprintf
	_ = slices.Collect[string]
)
`).MustTypeCheck()
		})

		It("yields the correct func objects", func() {
			Expect(astx.AllTypesOf[*types.Func](f.File(), ti)).To(ConsistOf(
				HaveField("FullName()", "fmt.Printf"),
				HaveField("FullName()", "fmt.Sprintf"),
				HaveField("FullName()", "slices.Collect")))
		})

		It("correctly aborts iterations", func() {
			fns := []*types.Func{}
			for fn := range astx.AllTypesOf[*types.Func](f.File(), ti) {
				fns = append(fns, fn)
				break
			}
			Expect(fns).To(HaveLen(1))
		})

	})

	When("iterating over specific func types using AllInPkgOf", func() {

		var (
			f  *testsauce.File
			ti *types.Info
		)

		BeforeEach(func() {
			f, _, ti = testsauce.MustParse(`
package main
import "fmt"
import "slices"
var (
	_ = fmt.Printf
	_ = fmt.Sprintf
	_ = slices.Collect[string]
)
`).MustTypeCheck()
		})

		It("yields the correct funcs", func() {
			Expect(astx.AllInPkgOf[*types.Func]("fmt", f.File(), ti)).To(ConsistOf(
				HaveField("FullName()", "fmt.Printf"),
				HaveField("FullName()", "fmt.Sprintf")))
		})

		It("correctly aborts iterations", func() {
			fns := []*types.Func{}
			for fn := range astx.AllInPkgOf[*types.Func]("fmt", f.File(), ti) {
				fns = append(fns, fn)
				break
			}
			Expect(fns).To(HaveLen(1))
		})

	})

})
