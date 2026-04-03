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
	"go/token"
	"maps"
	"slices"

	"github.com/thediveo/nonstd/xslices"
	"github.com/thediveo/notwork/analysis/astx"
	"github.com/thediveo/notwork/analysis/testsauce"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func orderPos(a, b token.Pos) int { return int(a - b) }

var _ = Describe("mapping token positions to AST nodes", func() {

	It("correctly maps defer statement positions", func() {
		f := testsauce.MustParse(`
package main
func foo() { defer func(){}() }
func init() { defer func(){}(); defer func(){}() }
`)
		fset := f.FileSet()
		m := astx.NewNodesByPosOf[*ast.DeferStmt]([]*ast.File{f.File()})
		posers := xslices.Map(
			slices.SortedFunc(maps.Keys(m), orderPos),
			func(p token.Pos) string { return fset.Position(p).String() })
		Expect(posers).To(ConsistOf(
			"sauce.go:3:14",
			"sauce.go:4:15",
			"sauce.go:4:33"))
	})

})
