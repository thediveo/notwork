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

package matchers

import (
	"go/ast"
	"go/token"

	gi "github.com/onsi/ginkgo/v2"
	g "github.com/onsi/gomega"
	"github.com/onsi/gomega/gcustom"
	"github.com/onsi/gomega/types"
	"github.com/thediveo/notwork/analysis/astx"
)

// Anchors is the Gomega-supporting variant of astx.Anchors that allows
// asserting token positions to be on the same line as named anchors, specified
// by “//anchor:name” source comments.
type Anchors struct {
	*astx.Anchors
}

// MustCollectAnchors returns a new Gomega-supporting Anchors object after
// gathering source code comments in the form “//anchor:text ...” from the
// passed AST files.
//
// Note: anchor names end at first white space.
func MustCollectAnchors(files []*ast.File, fileset *token.FileSet) *Anchors {
	gi.GinkgoHelper()

	anks, err := astx.CollectAnchors(files, fileset)
	g.Expect(err).NotTo(g.HaveOccurred(), "invalid anchor found")
	return &Anchors{
		Anchors: anks,
	}
}

// IsAnchor succeeds if the actual token position is on the same line as the
// anchor with the specified name.
func (a *Anchors) IsAnchor(name string) types.GomegaMatcher {
	return gcustom.MakeMatcher(func(actual token.Pos) (bool, error) {
		return a.Matches(actual, name), nil
	}).WithTemplate("Expected:\n{{.FormattedActual}}\n{{.To}} match anchor \"{{.Data}}\"", name)
}
