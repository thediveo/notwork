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

package funcs

import (
	"go/ast"
	"go/token"
	"go/types"
	"iter"
	"maps"

	"github.com/thediveo/notwork/analysis/astx"
)

// FuncID unambiguously identifies a particular function by its package import
// path in combination with its function name.
type FuncID struct {
	ImportPath string // package import path, such as "example.org/bar/foo".
	FuncName   string // function name, such as "Foo".
}

// String returns the unambiguous ID of this function in the form of
// “<import-path>.<func-name>”.
func (i FuncID) String() string {
	return i.ImportPath + "." + i.FuncName
}

// FuncRenovation map a particular “deprecated” old function to its “modern” new
// counterpart.
type FuncRenovation struct {
	Before, After FuncID
}

// NameMapping maps “deprecated” old function names to their “modern” new
// counterpart names based on external context about the deprecated and modern
// import paths.
type NameMapping map[string]string

// Refactoring maps old function full names to their new function full names
// counterparts. Refactoring maps can be easily created using [Remap] and
// [Relocate] and then stepwise augmented using chaining with
// [Refactoring.Remap] and [Refactoring.Translate].
type Refactoring map[FuncID]FuncID

// Remap returns a new Refactoring that maps functions from their old package to
// other functions in the new package.
func Remap(oldimportpath string, newimportpath string, fnmapping NameMapping) Refactoring {
	return maps.Collect(Remappings(oldimportpath, newimportpath, fnmapping))
}

// Remap returns an updated Refactoring that maps functions from their old
// packages to other functions in new packages.
func (r Refactoring) Remap(oldimportpath string, newimportpath string, fnmapping NameMapping) Refactoring {
	r = maps.Clone(r)
	maps.Insert(r, Remappings(oldimportpath, newimportpath, fnmapping))
	return r
}

// Relocate returns a new Refactoring that maps functions from their old package
// to the same functions in the new package.
func Relocate(oldimportpath string, newimportpath string, fns []string) Refactoring {
	return maps.Collect(Relocations(oldimportpath, newimportpath, fns))
}

// Relocate returns an updated Refactoring that maps functions from their old
// packages to other functions in the new package.
func (r Refactoring) Relocate(oldimportpath string, newimportpath string, fns []string) Refactoring {
	r = maps.Clone(r)
	maps.Insert(r, Relocations(oldimportpath, newimportpath, fns))
	return r
}

// AllFuncs iterates over the functions to transform in pairs of [*types.Func],
// [FuncID]. The types.Func is the actual place that needs transformation and
// the PkgFunc describes the new function in its new import path.
func (r Refactoring) AllFuncs(root ast.Node, typesinfo *types.Info) iter.Seq2[FuncRenovation, token.Pos] {
	return func(yield func(FuncRenovation, token.Pos) bool) {
		cont := true
		astx.Inspect(root, func(node ast.Node) bool {
			if !cont {
				return false // unwind as quick as possible
			}

			var obj types.Object
			var pos token.Pos
			descend := true
			switch node := node.(type) {
			case *ast.Ident:
				obj = typesinfo.ObjectOf(node)
				pos = node.Pos()
			case *ast.SelectorExpr:
				obj = typesinfo.ObjectOf(node.Sel)
				pos = node.Sel.Pos()
				descend = false
			default:
				return true
			}
			fnobj, ok := obj.(*types.Func)
			if fnobj == nil {
				return descend
			}
			pkg := fnobj.Pkg()
			if pkg == nil {
				return descend
			}
			oldpkgfn := FuncID{pkg.Path(), fnobj.Name()}
			newpkgfn, ok := r[oldpkgfn]
			if !ok {
				return descend
			}
			if !yield(FuncRenovation{oldpkgfn, newpkgfn}, pos) {
				cont = false
			}
			return cont && descend
		})
	}
}

// Remappings iterates over a mapping of old-to-new function names, producing
// [FuncID] pairs that map each function name in the old package to a new
// function name in the new package.
func Remappings(oldimportpath string, newimportpath string, fnmapping NameMapping) iter.Seq2[FuncID, FuncID] {
	return func(yield func(FuncID, FuncID) bool) {
		for oldfn, newfn := range fnmapping {
			if newfn == "" {
				newfn = oldfn
			}
			if !yield(FuncID{
				ImportPath: oldimportpath,
				FuncName:   oldfn,
			}, FuncID{
				ImportPath: newimportpath,
				FuncName:   newfn,
			}) {
				break
			}
		}
	}
}

// Relocations iterates over a list of function names, producing [FuncID] pairs
// that map a function from the old package to the same function in the new
// package.
func Relocations(oldimportpath string, newimportpath string, fns []string) iter.Seq2[FuncID, FuncID] {
	return func(yield func(FuncID, FuncID) bool) {
		for _, fn := range fns {
			if !yield(FuncID{
				ImportPath: oldimportpath,
				FuncName:   fn,
			}, FuncID{
				ImportPath: newimportpath,
				FuncName:   fn,
			}) {
				break
			}
		}
	}
}
