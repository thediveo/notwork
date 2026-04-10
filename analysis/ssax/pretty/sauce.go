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

package pretty

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"

	"github.com/thediveo/nonstd/xslices"
	"golang.org/x/tools/go/ssa"
)

// MustBuildSSA returns the *ssa.Package and *token.FileSet parsed and build
// from the passed source code, where the source is assumed to be in a file
// called “main.go”. MustBuildSSA panics if there are any errors in parsing,
// type-checking, and building the SSA steps.
func MustBuildSSA(sauce string) (*ssa.Package, *token.FileSet) {
	fileset := token.NewFileSet()
	file, err := parser.ParseFile(fileset, "main.go", sauce, parser.AllErrors)
	if err != nil {
		panic(err)
	}

	typesInfo := &types.Info{
		Types:     map[ast.Expr]types.TypeAndValue{},
		Instances: map[*ast.Ident]types.Instance{},
		Defs:      map[*ast.Ident]types.Object{},
		Uses:      map[*ast.Ident]types.Object{},
	}
	conf := types.Config{
		Importer: importer.Default(),
	}
	pkg, err := conf.Check("main",
		fileset,
		xslices.Slice(file),
		typesInfo)
	if err != nil {
		panic(err)
	}

	prog := ssa.NewProgram(fileset, ssa.SanityCheckFunctions)
	ssaPkg := prog.CreatePackage(pkg, xslices.Slice(file), typesInfo, false)
	prog.Build()
	return ssaPkg, fileset
}
