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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("sanitizing strings for use as Go identifiers", func() {

	DescribeTable("sanitizing",
		func(s string, expected string) {
			Expect(imports.ProposeLocalName(s)).To(Equal(expected))
		},
		Entry(nil, "", "unnamed"),
		Entry(nil, "Foo_bar", "Foo_bar"),
		Entry(nil, "42foo", "42foo"), // sic!
		Entry(nil, "überlichtgeschwindigkeit", "überlichtgeschwindigkeit"),
		Entry(nil, "defer", "defer_"),
		Entry(nil, "foo--bar", "foo"),
		Entry(nil, "foo/v2", "foo"),
		Entry(nil, "foo/v2schneider", "v2schneider"),
		Entry(nil, "v2", "v2"),
	)

})
