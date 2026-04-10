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
	"github.com/thediveo/nonstd/xslices"
	"github.com/thediveo/notwork/analysis/astx"
	"github.com/thediveo/notwork/analysis/testsauce"

	"github.com/onsi/gomega/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/thediveo/opt"
	. "github.com/thediveo/success"
)

type AssertionMethod func(matcher types.GomegaMatcher, optionalDescription ...any) bool

var _ = Describe("anchoring lines in ASTs", Ordered, func() {

	var f *testsauce.File
	var anchorwhat *astx.Anchors // bad pun, very bad

	BeforeAll(func() {
		f = testsauce.MustParse(
			`package main

const Foo = 42 		//anchor:42
const Bar = "abc" 	// anchor:42
const Baz = "" 		//anchor: <-- nada, nüscht.

func foo() { } //anchor:foo
`)
		anchorwhat = Successful(astx.CollectAnchors(xslices.Slice(f.File()), f.FileSet()))
	})

	DescribeTable("checks anchors",
		func(anchor string, line int, shouldmatch bool) {
			tfile := f.FileSet().File(f.File().Pos())
			pos := tfile.LineStart(line)
			Expect(anchorwhat.Matches(pos, anchor)).To(Equal(shouldmatch))
		},
		Entry(nil, "12345", 1, false),
		Entry(nil, "42", 3, true),
		Entry(nil, "42", 4, false),
		Entry(nil, "", 3, false),
		Entry(nil, "foo", 7, true),
	)

	It("raises an error for duplicate anchors", func() {
		f := testsauce.MustParse(`
package main //anchor:1
const Foo = 42 //anchor:1
`)
		Expect(astx.CollectAnchors(xslices.Slice(f.File()), f.FileSet())).
			Error().To(HaveOccurred())
	})

	DescribeTable("exact anchor checking",
		func(anchor string, line int, shouldmatch bool, expectederr bool) {
			tfile := f.FileSet().File(f.File().Pos())
			pos := tfile.LineStart(line)
			match, err := anchorwhat.MatchesExisting(pos, anchor)
			Expect(match).To(Equal(shouldmatch))
			ass := Expect(err)
			If[AssertionMethod](expectederr).Then(ass.To).Else(ass.NotTo)(HaveOccurred())
		},
		Entry(nil, "", 1, false, true),
		Entry(nil, "42", 3, true, false),
		Entry(nil, "42", 4, false, false),
		Entry(nil, "bar", 1, false, true),
	)

})
