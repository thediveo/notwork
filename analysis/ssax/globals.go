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

package ssax

import (
	"iter"

	"golang.org/x/tools/go/ssa"
)

// AllPackageGlobalsAssignments iterates over all *ssa.Store instructions that
// assign to globals of the specified package. The global store instructions are
// also from the specified package.
func AllPackageGlobalsAssignments(pkg *ssa.Package) iter.Seq2[*ssa.Global, ssa.Value] {
	return func(yield func(*ssa.Global, ssa.Value) bool) {
		for fn := range AllMembersOf[*ssa.Function](pkg) {
			if !yieldGlobalsAssignmentsRecursively(fn, yield) {
				return
			}
		}
	}
}

// pushGlobalAssignments yields all *ssa.Store instructions in the specified
// function, as well as in any recursively nested anonymous functions.
func yieldGlobalsAssignmentsRecursively(fn *ssa.Function, yield func(*ssa.Global, ssa.Value) bool) bool {
	for store := range AllInstructionsOf[*ssa.Store](fn) {
		global, ok := store.Addr.(*ssa.Global)
		if !ok {
			continue
		}
		if global.Pkg != fn.Pkg {
			continue
		}
		if !yield(global, store.Val) {
			return false
		}
	}
	for _, anonfn := range fn.AnonFuncs {
		if !yieldGlobalsAssignmentsRecursively(anonfn, yield) {
			return false
		}
	}
	return true
}
