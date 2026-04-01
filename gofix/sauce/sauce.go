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

package sauce

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"

	gi "github.com/onsi/ginkgo/v2"
	g "github.com/onsi/gomega"
)

// MustParse returns an *ast.File for the passed valid Go source code; otherwise
// it fails the current test when the source does not parse correctly.
func MustParse(sauce string) *ast.File {
	gi.GinkgoHelper()

	file, _ := mustParse(sauce)
	return file
}

func mustParse(sauce string) (*ast.File, *token.FileSet) {
	gi.GinkgoHelper()

	fileset := token.NewFileSet()
	file, err := parser.ParseFile(fileset, "sauce.go", sauce, parser.AllErrors)
	g.Expect(err).NotTo(g.HaveOccurred(), "source has parsing errors")
	return file, fileset
}

// MustParseAndTypeCheck returns an *ast.File as well as types information for
// the passed valid Go source code; otherwise it fails the current test when the
// source either does not parse correctly or doesn't pass type-checking.
func MustParseAndTypeCheck(sauce string) (*ast.File, *types.Info) {
	gi.GinkgoHelper()

	file, fileset := mustParse(sauce)
	typesInfo := &types.Info{
		Uses: map[*ast.Ident]types.Object{},
	}
	conf := types.Config{
		Importer: importer.Default(),
	}
	_, err := conf.Check("main",
		fileset,
		[]*ast.File{file},
		typesInfo)
	g.Expect(err).NotTo(g.HaveOccurred(), "source fails to type-check")

	return file, typesInfo
}

// FirstOf returns a pointer to the first AST node of the given specific node
// type that must implement the ast.Node interface; otherwise, it fails the
// current test.
func FirstOf[T any, PT interface {
	// https://stackoverflow.com/a/72091526,
	// https://stackoverflow.com/a/71444968
	ast.Node
	*T
}](file *ast.File) *T {
	gi.GinkgoHelper()

	var zero *T
	var node ast.Node
	g.Expect(ast.Preorder(file)).To(g.ContainElement(
		g.BeAssignableToTypeOf(zero), &node),
		"no node of type %T found", zero)
	return node.(PT)
}
