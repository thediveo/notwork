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
	"maps"
	"slices"

	"golang.org/x/tools/go/ssa"
)

// AllPkgInstructionsOf iterates over all instructions of type T in all
// functions of the specified package, including anonymous functions.
func AllPkgInstructionsOf[T ssa.Instruction](pkg *ssa.Package) iter.Seq[T] {
	return func(yield func(T) bool) {
		for fn := range MembersOf[*ssa.Function](pkg) {
			if !allFuncsInstructions(fn, yield) {
				return
			}
		}
	}
}

// AllFuncsInstructions iterates over all instructions of type T in this
// function as well as all its direct and indirect anonymous functions.
func AllFuncsInstructions[T ssa.Instruction](fn *ssa.Function) iter.Seq[T] {
	return func(yield func(T) bool) {
		allFuncsInstructions(fn, yield) // kick off the hunt...
	}
}

// allFuncsInstructions iterates over all instructions of type T in this
// function as well as all its direct and indirect anonymous functions.
func allFuncsInstructions[T ssa.Instruction](fn *ssa.Function, yield func(T) bool) bool {
	for instr := range InstructionsOf[T](fn) {
		if !yield(instr) {
			return false
		}
	}
	for _, fn := range fn.AnonFuncs {
		if !allFuncsInstructions(fn, yield) {
			return false
		}
	}
	return true
}

// InstructionsOf iterates over all instructions of type T of the specified
// function.
func InstructionsOf[T ssa.Instruction](fn *ssa.Function) iter.Seq[T] {
	return func(yield func(T) bool) {
		if fn.Blocks == nil {
			return
		}
		for _, block := range fn.Blocks {
			for _, instr := range block.Instrs {
				instr, ok := instr.(T)
				if !ok {
					continue
				}
				if !yield(instr) {
					return
				}
			}
		}
	}
}

// MembersOf iterates over members of type T in the specified package.
func MembersOf[T ssa.Member](pkg *ssa.Package) iter.Seq[T] {
	return func(yield func(T) bool) {
		if pkg == nil {
			return
		}
		for _, member := range pkg.Members {
			member, ok := member.(T)
			if !ok {
				continue
			}
			if !yield(member) {
				return
			}
		}
	}
}

// AllSortedMembersOf iterates over members of type T in the specified package
// in sorted order.
func AllSortedMembersOf[T ssa.Member](pkg *ssa.Package) iter.Seq[T] {
	return func(yield func(T) bool) {
		if pkg == nil {
			return
		}
		for _, name := range slices.Sorted(maps.Keys(pkg.Members)) {
			member, ok := pkg.Members[name].(T)
			if !ok {
				continue
			}
			if !yield(member) {
				return
			}
		}
	}
}
