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
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"path/filepath"

	gi "github.com/onsi/ginkgo/v2"
	g "github.com/onsi/gomega"
	"github.com/thediveo/nonstd/xslices"
)

// File wraps an *ast.File as well as an associated token.FileSet.
type File struct {
	// unexported as otherwise *File gets accepted in place of *ast.File and
	// that doesn't end well.
	file    *ast.File
	fileset *token.FileSet
}

// File returns wrapped the *ast.File.
func (f *File) File() *ast.File {
	return f.file
}

// FileSet returns the associated FileSet, otherwise it will fail the current
// test.
func (f *File) FileSet() *token.FileSet {
	gi.GinkgoHelper()

	g.Expect(f.fileset).NotTo(g.BeNil())
	return f.fileset
}

// Position returns a textual representation for the specified position in the
// form of “<filename>:<line>:<col>”, where the filename is the so-called
// basename and thus without any path. If the position is invalid, Position
// returns an empty string.
func (f *File) Position(pos token.Pos) string {
	p := f.fileset.Position(pos)
	if p.Line == 0 {
		return ""
	}
	return fmt.Sprintf("%s:%d:%d", filepath.Base(p.Filename), p.Line, p.Column)
}

// MustParse returns a File (wrapping an *ast.File) for the passed valid Go
// source code; otherwise it fails the current test when the source does not
// parse correctly. The returned File can be queried for its associated FileSet
// using [File.FileSet].
//
// Note: MustParse additionally parses comments.
func MustParse(sauce string) *File {
	gi.GinkgoHelper()

	fileset := token.NewFileSet()
	file, err := parser.ParseFile(fileset, "sauce.go", sauce,
		parser.AllErrors|parser.ParseComments)
	g.Expect(err).NotTo(g.HaveOccurred(), "source has parsing errors")

	return &File{file, fileset}
}

// MustTypeCheck returns an *ast.File, *types.Package details, as well as types
// information related to the file and its package; otherwise it fails the
// current test when the AST file fails type checking.
func (f *File) MustTypeCheck() (*File, *types.Package, *types.Info) {
	gi.GinkgoHelper()

	typesInfo := &types.Info{
		Types: map[ast.Expr]types.TypeAndValue{},
		Defs:  map[*ast.Ident]types.Object{},
		Uses:  map[*ast.Ident]types.Object{},
	}
	conf := types.Config{
		Importer: importer.Default(),
	}
	pkg, err := conf.Check("main",
		f.FileSet(),
		xslices.Slice(f.File()),
		typesInfo)
	g.Expect(err).NotTo(g.HaveOccurred(), "source fails to type-check")

	return f, pkg, typesInfo
}
