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

package funcs_test

import (
	"go/ast"

	"github.com/thediveo/notwork/analysis/astx"
	"github.com/thediveo/notwork/analysis/testsauce"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("experimental tests", func() {

	It("experiments", func() {
		f, pkg, typeinfo := testsauce.MustParse(
			`package main

func Foo() func() { return func() { } } //anchor:Foo

var F = Foo

func main() {
	defer Foo() //anchor:defer-Foo
	defer Foo() //anchor:defer-Foo-result
}
`).MustTypeCheck()
		nodemap = astx.NewNodesByPosOf[ast.Node]([]*ast.File{f.File()})
		ssapkg := f.MustBuildSSAPkg()

	})

})
