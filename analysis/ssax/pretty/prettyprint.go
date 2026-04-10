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
	"bytes"
	"fmt"
	"go/token"
	"go/types"
	"io"
	"reflect"
	"slices"

	"github.com/thediveo/nonstd/xmaps"
	"github.com/thediveo/notwork/analysis/ssax"
	"golang.org/x/tools/go/ssa"
)

type Printer struct {
	w       io.Writer
	fileset *token.FileSet
}

func NewPrinter(w io.Writer, fileset *token.FileSet) *Printer {
	return &Printer{
		w:       w,
		fileset: fileset,
	}
}

func (p Printer) padbefore(first bool) bool {
	if first {
		return false
	}
	fmt.Fprint(p.w, "\n")
	return false
}

func (p Printer) Package(pkg *ssa.Package, level uint) {
	first := p.Globals(pkg, true, level)
	for fn := range ssax.AllSortedMembersOf[*ssa.Function](pkg) {
		first = p.padbefore(first)
		p.Func(fn, " member", level)
	}
}

// Globals prints all package members that are globals. Where applicable, it
// additionally shows where assignments to these globals are to be found.
func (p Printer) Globals(pkg *ssa.Package, first bool, level uint) bool {
	// scan the package's functions for assignments to package globals.
	globassigns := xmaps.Collect(ssax.AllPackageGlobalsAssignments(pkg))

	heading := true
	for global := range ssax.AllSortedMembersOf[*ssa.Global](pkg) {
		if heading {
			// ensures that if we show globals then we'll pad before the next
			// section; also, show the heading only for the first iteration.
			first = p.padbefore(first)
			Iprintf(p.w, level, " global members\n")
			heading = false
		}
		p.Value(global, level+1)
		// in case we know where the assignments are done, print them...
		if values, ok := globassigns[global]; ok {
			for _, val := range values {
				if pos := p.PosString(val.Pos()); pos != "" {
					Iprintf(p.w, level+2, "󰞓 value %s %s\n", val.String(), pos)
				}
			}
		}
	}
	return first
}

// PosString renders the textual representation of the specified token.Pos in
// the form of “<file>:<line>:<column>” if valid; otherwise, it returns an empty
// string.
func (p Printer) PosString(pos token.Pos) string {
	position := p.fileset.Position(pos)
	if position.Line == 0 || position.Column == 0 {
		return ""
	}
	return fmt.Sprintf("%s:%d:%d", position.Filename, position.Line, position.Column)
}

// Func pretty-prints details about the specified *ssa.Function. In particular:
// - an optional role, such as “member” (func) or “anon” (func).
// - the name of the function as well as its position in the source.
// - type parameters, if any.
// - parameters, if any.
// - results, if any.
// - blocks with their instructions.
func (p Printer) Func(fn *ssa.Function, role string, level uint) {
	if role != "" {
		role += " "
	}
	Iprintf(p.w, level, "%s 󰊕 func %s %s\n", role, fn.String(), p.PosString(fn.Pos()))
	p.TypeParamsList(fn.Signature.TypeParams(), level+1)
	p.ParamsList(fn.Params, level+1)
	p.TupleList(fn.Signature.Results(), " results", level+1)
	//p.Signature(fn.Signature, level+1)
	for _, block := range fn.Blocks {
		Iprintf(p.w, level+1, "󱃖 block %d (%s)\n", block.Index, block.Comment)
		for _, instr := range block.Instrs {
			p.Instr(instr, level+2)
		}
	}
	// recursively print any anonymous functions directly beneath this function,
	// further indented.
	for _, fn := range fn.AnonFuncs {
		Iprintf(p.w, 0, "\n")
		p.Func(fn, "󱀣 anon", level+1)
	}
}

func (p Printer) TypeParamsList(tparams *types.TypeParamList, level uint) {
	if tparams == nil || tparams.Len() == 0 {
		return
	}
	Iprintf(p.w, level, "󰰦 type parameters\n")
	for tparam := range tparams.TypeParams() {
		obj := tparam.Obj()
		Iprintf(p.w, level+1, "%s %s\n", obj.Name(), obj.Type().Underlying().String())
	}
}

func (p Printer) ParamsList(params []*ssa.Parameter, level uint) {
	if len(params) == 0 {
		return
	}
	Iprintf(p.w, level, " parameters\n")
	for _, param := range params {
		obj := param.Object()
		Iprintf(p.w, level+1, "%s %s\n", obj.Name(), obj.Type())
	}
}

func (p Printer) TupleList(params *types.Tuple, role string, level uint) {
	if params == nil || params.Len() == 0 {
		return
	}
	Iprintf(p.w, level, "%s\n", role)
	for idx := range params.Len() {
		Iprintf(p.w, level+1, "%d %s\n", idx, params.At(idx).String())
	}
}

func (p Printer) Value(val ssa.Value, level uint) {
	if instr, ok := val.(ssa.Instruction); ok {
		p.Instr(instr, level)
		return
	}
	Iprintf(p.w, level, "󱄑 %s %T %s %s\n",
		val.Name(), val, val.String(), p.PosString(val.Pos()))
}

// InstrPseudoLocation renders a pseudo instruction location in the form of
// “[<block>:<instr-num>]”. It appends a single space to the location string,
// unless it cannot determine the location.
func (p Printer) InstrPseudoLocation(instr ssa.Instruction) string {
	idx := slices.Index(instr.Block().Instrs, instr)
	if idx < 0 {
		return ""
	}
	return fmt.Sprintf("󰜘[%d:%d] ", instr.Block().Index, idx)
}

// Referrers renders a list of instructions that use the specified value, where
// the referencing instructions are shown in “[<block>:<instr-num>]” format.
func (p Printer) Referrers(val ssa.Value, level uint) {
	var out bytes.Buffer
	header := true
	for instr := range ssax.AllReferrers(val) {
		if header {
			header = false
			out.WriteString(" ")
		}
		out.WriteString(p.InstrPseudoLocation(instr))
	}
	if header {
		return
	}
	Iprintf(p.w, level, "%s\n", out.String())
}

func (p Printer) Instr(instr ssa.Instruction, level uint) {
	var prefix string
	if val, ok := instr.(ssa.Value); ok {
		// probLLMs: so much for "thinking" and "reasoning" ... when asking
		// these klankers if there is an existing function for the
		// variable/register names it all goes downhill. No no no, there's no
		// such function for assign variable signs (but I'm great in writing
		// slop code that is just terribly and sometimes might do the job by
		// sheer accident). Or, no no no, there's nothing like register names,
		// as we're in the analysis graph and there are no machine(!) registers.
		// It doesn't matter that the Go doc comments for both ssa.Value.Name()
		// as well as ssa.register spell out the beans very plainly and
		// eloquently.
		prefix = "󱄑 " + val.Name() + " "
	}

	if icon := instrIcon(instr); icon != "" {
		if storeinstr, ok := instr.(*ssa.Store); ok {
			if _, ok := storeinstr.Addr.(*ssa.Global); ok {
				icon += " "
			}
		}
		prefix = icon + " " + prefix
	}

	Iprintf(p.w, level, "%s%s%T %s %s\n",
		p.InstrPseudoLocation(instr), prefix, instr, instr.String(), p.PosString(instr.Pos()))
	// Optionally more details...
	switch instr := instr.(type) {
	case ssa.CallInstruction:
		// In case of a call instructions, that is, a function call, defer or go,
		// print also the common parts for these different types of calls.
		// Nota bene: the common value of call instructions is the preceeding
		// instruction.
		p.printCallCommon(instr.Common(), level+1)
		// Nota bene: the callinstr.Value() is only non-nil for calls and then
		// it refers to this call itself. No use in trying to print it here, as
		// we already did it just a few lines of code above.
	case *ssa.Store:
		p.Value(instr.Val, level+1)
	case *ssa.Return:
		for idx, result := range instr.Results {
			Iprintf(p.w, level+1, "%d %s\n", idx, result.String())
		}
	}
	if val, ok := instr.(ssa.Value); ok {
		p.Referrers(val, level+1)
	}
}

// instrIcon returns an icon for the passed concrete ssa.Instruction type; or an
// empty string if no icon has been assigned.
func instrIcon(instr ssa.Instruction) string {
	return iconsByInstr[reflect.TypeOf(instr).String()]
}

var iconsByInstr = map[string]string{
	"*ssa.Call":      "󰃷",
	"*ssa.Defer":     "󰃢",
	"*ssa.If":        "󰙁",
	"*ssa.Jump":      "󰞁",
	"*ssa.Store":     "",
	"*ssa.Panic":     "",
	"*ssa.Return":    "󰩈",
	"*ssa.RunDefers": "󰑮",
}

func (p Printer) printCallCommon(cc *ssa.CallCommon, level uint) {
	// nota bene: the method, if present, is always an *ssa.Function.
	if cc.Method == nil {
		p.Value(cc.Value, level)
		p.Referrers(cc.Value, level+1)
		return
	}
	Iprintf(p.w, level, "%T %s . %s\n", cc.Value, cc.Value.String(), cc.Method.String())
	p.Referrers(cc.Value, level+1)
}
