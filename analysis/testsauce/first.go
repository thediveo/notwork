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

package testsauce

import (
	"go/ast"

	"github.com/thediveo/nonstd/xiter"
	"github.com/thediveo/notwork/analysis/astx"

	gi "github.com/onsi/ginkgo/v2"
	g "github.com/onsi/gomega"
	"github.com/thediveo/testily/zero"
)

// MustFirstOf returns a pointer to the first AST node of the given specific
// node type that must implement the ast.Node interface; otherwise, it fails the
// current test if there is no node of type T in the AST.
func MustFirstOf[T ast.Node](file *ast.File) T {
	gi.GinkgoHelper()

	node, ok := xiter.FirstOk(astx.AllOf[T](file))
	g.Expect(ok).To(g.BeTrue(), "no node of type %T found", zero.Value[T])
	return node
}
