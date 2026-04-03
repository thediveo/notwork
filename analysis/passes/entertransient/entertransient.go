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

package entertransient

import (
	"go/ast"
	"os"
	"slices"

	"github.com/thediveo/notwork/analysis/analyzerdoc"
	"github.com/thediveo/notwork/analysis/astx"
	"github.com/thediveo/notwork/analysis/imports"
	"github.com/thediveo/notwork/analysis/ssax"
	"github.com/thediveo/notwork/analysis/ssax/pretty"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/ssa"

	_ "embed"
)

//go:embed doc.go
var doc string

var importpaths = []string{
	"github.com/thediveo/notwork/netns",
	"github.com/thediveo/notwork/mntns",
}

var Analyzer = &analysis.Analyzer{
	Name:     "notworknetns",
	Doc:      analyzerdoc.MustExtract(doc, "entertransient"),
	Requires: []*analysis.Analyzer{buildssa.Analyzer},
	Run:      run,
}

// run our analysis on a specific package and over all its individual files.
func run(pass *analysis.Pass) (any, error) {
	if !imports.IsImporting(pass.Pkg, importpaths) {
		return nil, nil
	}
	// at least one of the "interesting" packages is imported, so we now need to
	// dig deeper.
	ssaResult := pass.ResultOf[buildssa.Analyzer].(*buildssa.SSA)

	for name, member := range ssaResult.Pkg.Members {
		println("****", name, member.String(), member.Type().String())
	}
	pretty.NewPrinter(os.Stdout, pass.Fset).Package(ssaResult.Pkg, 0)

	deferStmtsByPos := astx.NewNodesByPosOf[*ast.DeferStmt](pass.Files)

	for _, fn := range ssaResult.SrcFuncs {
		for defr := range ssax.AllInstructionsOf[*ssa.Defer](fn) {
			if !isFromEnterTransient(defr.Call.Value) {
				continue
			}
			// We now need to go back to the AST...
			deferstmt := deferStmtsByPos[defr.Pos()]
			if deferstmt == nil || deferstmt.Call == nil {
				continue
			}
			call := deferstmt.Call
			pass.Report(analysis.Diagnostic{
				Pos:     call.Pos(),
				End:     call.End(),
				Message: "incorrect defer of EnterTransient: defer the cleanup func returned by EnterTransient, not EnterTransient itself",
				SuggestedFixes: []analysis.SuggestedFix{
					{
						Message: "add ()",
						TextEdits: []analysis.TextEdit{
							{
								Pos:     call.End(),
								End:     call.End(),
								NewText: []byte("()"),
							},
						},
					},
				},
			})
		}
	}
	return nil, nil
}

func isFromEnterTransient(val ssa.Value) bool {
	seen := map[ssa.Value]struct{}{}
	return track(val, seen)
}

// isEnterTransient returns true if the passed function is one of the
// EnterTransient functions from the (deprecated) netns and mntns packages.
func isEnterTransient(fn *ssa.Function) bool {
	obj := fn.Object()
	if obj == nil {
		return false
	}
	// The called fn must be "EnterTransient" from one of the packages
	// that define EnterTransient functions.
	if fn.Name() != "EnterTransient" {
		return false
	}
	pkg := fn.Object().Pkg()
	return pkg != nil && slices.Contains(importpaths, pkg.Path())
}

func track(val ssa.Value, seen map[ssa.Value]struct{}) bool {
	if val == nil {
		return false
	}
	if _, ok := seen[val]; ok {
		return false
	}
	seen[val] = struct{}{}

	switch val := val.(type) {
	case *ssa.Function:
		return isEnterTransient(val)
	case *ssa.Call:
		return false
		//return track(val.Call.Value, seen)
	case *ssa.Phi:
		for _, edge := range val.Edges {
			if track(edge, seen) {
				return true
			}
		}
	case *ssa.UnOp:
		return track(val.X, seen)
	case *ssa.ChangeType:
		return track(val.X, seen)
	case *ssa.Convert:
		return track(val.X, seen)
	case *ssa.MakeClosure:
		return track(val.Fn, seen)
	case *ssa.Extract:
		return track(val.Tuple, seen)
	case *ssa.Parameter:
		return false // TODO: track across functions?
	case *ssa.FreeVar:
		return track(val, seen)
	case *ssa.Global:
		return trackGlobal(val, seen)
	}
	return false
}

func trackGlobal(glob *ssa.Global, seen map[ssa.Value]struct{}) bool {
	if glob.Pkg == nil {
		return false
	}
	for _, member := range glob.Pkg.Members {
		fn, ok := member.(*ssa.Function)
		if !ok || fn.Name() != "init" {
			continue
		}
		for store := range ssax.AllInstructionsOf[*ssa.Store](fn) {
			if store.Addr == glob && track(store.Val, seen) {
				return true
			}
		}
	}
	return false
}
