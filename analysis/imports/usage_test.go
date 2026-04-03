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

package imports_test

import (
	"github.com/thediveo/notwork/analysis/imports"
	"github.com/thediveo/notwork/analysis/testsauce"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("import usages", func() {

	It("returns the import usages", func() {
		f, _, ti := testsauce.MustParse(`
package main
import "go/ast"				// to be collected
import abc "fmt"			// to be collected
import _ "fmt"				// must not be collected
import . "testing"			// must not be collected
var _ = ast.ImportSpec{}
var _ = abc.Sprintf
var _ = T{}
`).MustTypeCheck()
		u := imports.CollectPackageUsages(f.File(), ti)
		Expect(u).To(HaveLen(2))
		Expect(u).To(HaveKeyWithValue("go/ast", HaveField("LocalName", "ast")))
		Expect(u).To(HaveKeyWithValue("fmt", HaveField("LocalName", "abc")))
	})

})
