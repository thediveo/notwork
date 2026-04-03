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
	"github.com/thediveo/notwork/analysis/testsauce"
	"golang.org/x/tools/go/ssa"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("deriving SSA information", func() {

	It("returns SSA information", func() {
		var f *testsauce.File
		Expect(InterceptGomegaFailure(func() {
			f = testsauce.MustParse(`
package main

func main() {
	_ = 42+666
}
`)
		})).To(Succeed())
		var pkg *ssa.Package
		Expect(InterceptGomegaFailure(func() {
			pkg = f.MustBuildSSAPkg()
		})).To(Succeed())
		Expect(pkg).NotTo(BeNil())
		Expect(pkg.Func("main")).NotTo(BeNil())
	})

})
