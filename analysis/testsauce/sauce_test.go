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
	"go/types"

	"github.com/thediveo/notwork/analysis/testsauce"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("working with Go sources", func() {

	When("only parsing a Go source", func() {

		It("succeeds for a valid sauce", func() {
			var f *testsauce.File
			Expect(InterceptGomegaFailure(func() {
				f = testsauce.MustParse(`
package main

func main() {
	_ = unknown.Symbol
}
`)
			})).To(Succeed())
			Expect(f).NotTo(BeNil())
			Expect(f.FileSet()).NotTo(BeNil())
		})

	})

	When("parsing and type-checking a Go source", func() {

		It("succeeds for a valid sauce", func() {
			var f *testsauce.File
			var ti *types.Info
			Expect(InterceptGomegaFailure(func() {
				f, _, ti = testsauce.MustParse(`
package main

import "fmt"

func main() {
	println(fmt.Sprintf("%s", "hooray!"))
}
`).MustTypeCheck()
			})).To(Succeed())
			Expect(f).NotTo(BeNil())
			Expect(ti).NotTo(BeNil())
		})

		It("rejects invalid sauce", func() {
			Expect(InterceptGomegaFailure(func() {
				_, _, _ = testsauce.MustParse(`
package main

package main
`).MustTypeCheck()
			})).To(MatchError(ContainSubstring("has parsing errors")))
		})

		It("rejects invalid types", func() {
			Expect(InterceptGomegaFailure(func() {
				_, _, _ = testsauce.MustParse(`
package main

var _ = thisDoesNotExist
`).MustTypeCheck()
			})).To(MatchError(ContainSubstring("fails to type-check")))
		})

	})

})
