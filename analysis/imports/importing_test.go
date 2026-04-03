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

var _ = Describe("checking for the presence of import paths", func() {

	It("correctly indicates presence of any of the specified import paths", func() {
		_, pkg, _ := testsauce.MustParse(`
package main
import _ "fmt"
import _ "go/ast"
`).MustTypeCheck()
		Expect(imports.IsImporting(pkg, []string{"foo", "go/ast"})).To(BeTrue())
		Expect(imports.IsImporting(pkg, []string{"foo", "abc"})).To(BeFalse())
	})

})
