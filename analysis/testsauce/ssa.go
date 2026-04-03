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
	"go/importer"
	"go/types"

	"golang.org/x/tools/go/ssa"

	gi "github.com/onsi/ginkgo/v2"
	g "github.com/onsi/gomega"
)

// MustBuildSSAPkg returns an *ssa.Package for this *File; otherwise, it fails
// the current test when the AST file fails type checking before going through
// the SSA.
//
// Note bene: use [ssa.Package.Prog] to navigate to the owning program.
func (f *File) MustBuildSSAPkg() *ssa.Package {
	gi.GinkgoHelper()

	// Nota bene: the probLLMs hallucinate slob code (except when the planets
	// are randomly arranged in good orbit) that crashes CreatePackage because
	// our types.Info must have at least its Types and Defs maps initialized and
	// thus non-nil. So much for "understanding" and "reasoning".
	typesInfo := &types.Info{
		Types:     map[ast.Expr]types.TypeAndValue{},
		Defs:      map[*ast.Ident]types.Object{},
		Instances: map[*ast.Ident]types.Instance{},
		Uses:      map[*ast.Ident]types.Object{},
	}
	conf := types.Config{
		Importer: importer.Default(),
	}
	pkg, err := conf.Check("main",
		f.FileSet(),
		[]*ast.File{f.file},
		typesInfo)
	g.Expect(err).NotTo(g.HaveOccurred(), "source fails to type-check")

	prog := ssa.NewProgram(f.fileset, ssa.SanityCheckFunctions)
	// Nota bene: contrary to probLLM hallucinations (elaborate BS), importable
	// doesn't need to be true in order for a "main" module to get its functions
	// build. This code base proves the slob generators wrong. Yeah, they
	// "understand" code ... for a sufficient definition of "no clue".
	ssaPkg := prog.CreatePackage(pkg, []*ast.File{f.file}, typesInfo, false)
	prog.Build()
	return ssaPkg
}
