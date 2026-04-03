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

package astx

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"github.com/thediveo/nonstd/sets"
	"github.com/thediveo/nonstd/xstrings"
)

const AnchorPrefix = "//anchor:"

// Anchors gathers source code comments in the form “//anchor:name ...” from a
// set of AST files in combination with a token file set. Afterwards, it can be
// used to check that the position of a certain anchor (text) matches the line
// of the corresponding anchor comment.
//
// Note: the texts of anchors ends at first white space.
type Anchors struct {
	m       map[int]string
	fileset *token.FileSet
}

// CollectAnchors returns a new Anchors object after gathering source code
// comments in the form “//anchor:text ...” from the passed AST files. In case
// an anchor name is encountered twice, an error is returned instead.
//
// Note: anchor names end at first white space.
func CollectAnchors(files []*ast.File, fileset *token.FileSet) (*Anchors, error) {
	a := &Anchors{
		m:       map[int]string{},
		fileset: fileset,
	}

	seen := sets.New[string]()
	for _, file := range files {
		for _, cmtgroup := range file.Comments {
			for _, comment := range cmtgroup.List {
				text, ok := strings.CutPrefix(comment.Text, AnchorPrefix)
				if !ok {
					continue
				}
				anchortext, _, _ := xstrings.CutWhitespace(text)
				if anchortext == "" {
					continue
				}
				pos := fileset.Position(comment.Pos())
				if pos.Line == 0 {
					continue
				}
				if seen.Contains(anchortext) {
					return nil, fmt.Errorf("duplicate //anchor:%s at %s:%d:%d",
						anchortext, pos.Filename, pos.Line, pos.Column)
				}
				seen.Add(anchortext)
				a.m[pos.Line] = anchortext
			}
		}
	}

	return a, nil
}

// Matches returns true if the passed token position is at the same line as the
// anchor with the specified text; otherwise, it returns false.
func (a *Anchors) Matches(pos token.Pos, name string) bool {
	if name == "" {
		return false
	}
	p := a.fileset.Position(pos)
	return a.m[p.Line] == name
}

// MatchesExisting returns true if the passed token position is at the same line
// as the anchor with the specified text, else if the anchor name is known it
// returns false; otherwise if the anchor name isn't known at all, it returns an
// error.
func (a *Anchors) MatchesExisting(pos token.Pos, name string) (ok bool, err error) {
	if name == "" {
		return false, errors.New("anchor name must not be empty")
	}
	// fast path, try a successful match first...
	p := a.fileset.Position(pos)
	if a.m[p.Line] == name {
		return true, nil
	}
	// nope, now check if we know even about this anchor...
	for _, anchorname := range a.m {
		if anchorname == name {
			return false, nil
		}
	}
	return false, fmt.Errorf("unknown anchor %q", name)
}
