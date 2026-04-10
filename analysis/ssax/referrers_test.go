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

package ssax_test

import (
	"go/ast"
	"slices"

	"github.com/thediveo/nonstd/xiter"
	astxm "github.com/thediveo/notwork/analysis/astx/matchers"
	"github.com/thediveo/notwork/analysis/ssax"
	"github.com/thediveo/notwork/analysis/ssax/pretty"
	"github.com/thediveo/notwork/analysis/testsauce"
	"golang.org/x/tools/go/ssa"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("iterating value referrers", Ordered, func() {

	var pkg *ssa.Package
	var ankmap *astxm.Anchors

	BeforeAll(func() {
		f, _, _ := testsauce.MustParse(
			`package main

func Foo() { } //anchor:foo-def

func main() {
	f1 := Foo //anchor:f1
	f2 := Foo //anchor:f2
	defer f1()
	defer f2()
	x := 2
	x = x * 2
	y := 42
	z := x * y
	z += z
	_ = z
}
`).MustTypeCheck()
		ankmap = astxm.MustCollectAnchors([]*ast.File{f.File()}, f.FileSet())
		pkg = f.MustBuildSSAPkg()
		pretty.NewPrinter(GinkgoWriter, f.FileSet()).Package(pkg, 0)
	})

	FIt("iterates correctly", func() {
		stores := slices.Collect(xiter.Filter(ssax.AllPkgInstructionsOf[*ssa.Store](pkg),
			func(store *ssa.Store) bool {
				println("**** store")
				fnval, ok := store.Val.(*ssa.Function)
				if !ok {
					return false
				}
				println("****", fnval.Name())
				return fnval.Name() == "main.Foo"
			}))
		Expect(stores).To(HaveLen(2))
		Expect(stores[0].Pos()).To(ankmap.IsAnchor("f1"))
		Expect(stores[1].Pos()).To(ankmap.IsAnchor("f2"))
	})

})
