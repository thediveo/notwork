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

package testsauce_test

import (
	"go/ast"

	"github.com/thediveo/notwork/analysis/testsauce"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("looking for specific types of AST nodes", func() {

	It("returns the correct node", func() {
		f := testsauce.MustParse(`
package main

import "fmt"

func main() {
fmt.Sprintf("%s", "hooray!")
}
`)
		var node *ast.SelectorExpr
		Expect(InterceptGomegaFailure(func() {
			node = testsauce.MustFirstOf[*ast.SelectorExpr](f.File())
		})).To(Succeed())
		Expect(node).NotTo(BeNil())
	})

	It("fails if no node of this type can be found", func() {
		f := testsauce.MustParse(`
package main
`)
		Expect(InterceptGomegaFailure(func() {
			_ = testsauce.MustFirstOf[*ast.SelectorExpr](f.File())
		})).To(MatchError(ContainSubstring("no node of type *ast.SelectorExpr found")))
	})

})
