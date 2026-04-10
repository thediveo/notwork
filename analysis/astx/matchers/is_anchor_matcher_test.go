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

package matchers_test

import (
	"go/ast"

	"github.com/thediveo/nonstd/xiter"
	"github.com/thediveo/nonstd/xslices"
	"github.com/thediveo/notwork/analysis/astx"
	"github.com/thediveo/notwork/analysis/astx/matchers"
	"github.com/thediveo/notwork/analysis/testsauce"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	//. "github.com/thediveo/success"
)

var _ = Describe("IsAnchor matcher", func() {

	It("matches correctly", func() {
		f := testsauce.MustParse(
			`
package main

const Foo = 42 //anchor:foo
`)
		anks := matchers.MustCollectAnchors(xslices.Slice(f.File()), f.FileSet())
		fooident, ok := xiter.FirstOk(
			xiter.Filter(astx.AllOf[*ast.Ident](f.File()),
				func(id *ast.Ident) bool { return id.Name == "Foo" }))
		Expect(ok).To(BeTrue())
		Expect(fooident.Pos()).To(anks.IsAnchor("foo"))
	})

})
