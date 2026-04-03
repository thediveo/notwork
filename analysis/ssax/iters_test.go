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
	"slices"

	"github.com/thediveo/nonstd/xiter"
	"github.com/thediveo/notwork/analysis/ssax"
	"github.com/thediveo/notwork/analysis/testsauce"
	"golang.org/x/tools/go/ssa"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("iterating SSA", func() {

	var fn *ssa.Function

	BeforeEach(func() {
		f := testsauce.MustParse(`
package main

func foo(x int) int {
	y := x + 42
	return y * 666
}
`)
		fn = f.MustBuildSSAPkg().Func("foo")
		Expect(fn).NotTo(BeNil())
		Expect(fn.Blocks).NotTo(BeEmpty())
	})

	It("iterates correctly all elements of type *ssa.BinOp", func() {
		Expect(slices.Collect(xiter.Map(
			ssax.AllInstructionsOf[*ssa.BinOp](fn),
			func(e *ssa.BinOp) string { return e.Op.String() },
		))).To(Equal([]string{"+", "*"}))
	})

	It("correctly aborts the iterator", func() {
		count := 0
		for range ssax.AllInstructionsOf[*ssa.BinOp](fn) {
			count++
			break
		}
		Expect(count).To(Equal(1))
	})

})
