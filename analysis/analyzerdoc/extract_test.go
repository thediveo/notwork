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

package analyzerdoc_test

import (
	"github.com/thediveo/notwork/analysis/analyzerdoc"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const testcontents = `
/*
Package main

# Ignorant heading

This should be ignored.

# Analyzer good

good: always reports a sycophantic result

This is the documentation for the "good" analyzer.

# Analyzer bad

This one hasn't a summary line.

# Analyzer ugly

ugly: reports our negative coding style verdict

This is the documentation for the "ugly" analyzer.


*/
package main
`

var _ = Describe("extracing Analyzer documentation", func() {

	DescribeTable("Analyzer documentation",
		func(contents, analyzername, want string) {
			contents, err := analyzerdoc.Extract(contents, analyzername)
			if err != nil {
				Expect(err).To(MatchError(want))
				return
			}
			Expect(contents).To(Equal(want))
		},
		Entry(nil, "", "gopher", "empty Go source file"),
		Entry(nil, "?!$*@!!", "gopher", "invalid Go source file"),
		Entry(nil, "package main", "gopher", "Go source file without package doc comment"),
		Entry(nil, testcontents, "foo", "package doc comment missing the \"Analyzer foo\" heading"),
		Entry(nil, testcontents, "bad", "\"Analyzer bad\" heading missing following \"bad: summary...\" line"),
		Entry(nil, testcontents, "good", "always reports a sycophantic result\n\nThis is the documentation for the \"good\" analyzer."),
		Entry(nil, testcontents, "ugly", "reports our negative coding style verdict\n\nThis is the documentation for the \"ugly\" analyzer."),
	)

	It("returns documentation", func() {
		var contents string
		Expect(func() {
			contents = analyzerdoc.MustExtract(testcontents, "good")
		}).NotTo(Panic())
		Expect(contents).To(HavePrefix("always reports"))
	})

	It("returns documentation", func() {
		Expect(func() {
			_ = analyzerdoc.MustExtract(testcontents, "foo")
		}).To(Panic())
	})

})
