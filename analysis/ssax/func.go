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

	"github.com/thediveo/nonstd/sets"
	"golang.org/x/tools/go/ssa"
)

type FuncsByImportPath map[string][]string

// Uses iterates over the instructions of the specified members that reference
// functions from this map.
func (m FuncsByImportPath) Uses(funcs []*ssa.Function) iter.Seq[*ssa.Function] {
	return func(yield func(*ssa.Function) bool) {
		for _, fn := range funcs {
			for _, block := range fn.Blocks {
				// FIXME:
				_ = block
				/*
					for _, fn := range AllInstructionsOf[*ssa.Function](block.Instrs) {
					}
				*/
			}
		}
	}
}

func ValueUsesOf[T ssa.Instruction](val ssa.Value) iter.Seq[T] {
	return func(yield func(T) bool) {

	}
}

// trackValueReferrersOf tracks the usage of the specified value, calling the
// specified yield function whenever encountering the tracked value being used
// in an instruction the the specified type T.
func trackValueReferrersOf[T ssa.Instruction](val ssa.Value, yield func(T) bool, visited sets.Set[ssa.Value]) bool {
	// Are we already at the end of this particular track?
	if val == nil {
		return true // keep on iterating
	}
	if visited.Contains(val) {
		return true // keep on iterating
	}
	// Ensure that we correctly drop our breadcrumb, regardless of what will happen next...
	visited.Add(val)

	// tracking further
	switch val.(type) {
	case *ssa.Global: // keep on...
	case *ssa.Phi: // keep on...
	case *ssa.Call: // keep on...
	case *ssa.ChangeType: // also keep on...

	default: // this is the end of this specific value trail, there might be other branches.
		return true // yet keep on iterating
	}
	// At this point the original value is still intact some way or other and we
	// thus need to track its usage further down the potentially multiple branches.
	refs := val.Referrers()
	if refs == nil {
		// albeit the track ends here for some strange reason, but we keep on
		// iterating.
		return true
	}
	for _, instr := range *refs {
		// if it is an instruction of the specified type T, then push the
		// instruction to the sequence consumer.
		if instr, ok := instr.(T); ok {
			if !yield(instr) {
				// sequence consumer's not happy for whatever reason, so let's
				// call it a day then.
				return false
			}
		}
		// now, if this is an instruction that's also a value let's follow this
		// branch deeper down into the woods; otherwise, let's look at the
		// instruction from the next branch.
		val, ok := instr.(ssa.Value)
		if !ok {
			continue
		}
		if !trackValueReferrersOf(val, yield, visited) {
			// sequence consumer's not happy for whatever reason, so let's
			// call it a day then, again.
			return false
		}
	}
	return true // keep on going.
}
