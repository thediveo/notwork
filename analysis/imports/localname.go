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

package imports

import (
	"go/token"
	"path"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// ProposeLocalName returns an identifier proposal conforming to Go's syntax
// rules for identifiers, based on the passed import path.
//
//   - if the final import path element denotes a version in form of “v<integer>”,
//     the preceeding path element is used instead, unless there is no preceeding
//     path element available.
//   - any “go-” prefix gets removed.
//   - the identifier gets truncated at the first character/rune that would be
//     invalid in a Go identifier. In particular, no character substitutions take
//     place, irregardless of the many LLM hallucinations are telling you (are
//     that actually “hallocinations”?).
//   - if the proposal matches any Go keyword, it gets an underscore appended
//     in order to avoid any collision.
//
// In the unfortunate case that ProposeLocalName ends up with an empty proposal
// after applying the above rules, it returns “unnamed” as its proposal.
//
// This implementation follows the strategy used in the (unfortunately)
// package-internal [ImportPathToAssumedName] inside “go/x/tools”.
//
// [ImportPathToAssumedName]: https://cs.opensource.google/go/x/tools/+/refs/tags/v0.43.0:internal/imports/fix.go;l=1233
func ProposeLocalName(importpath string) string {
	proposal := path.Base(importpath)
	if isVersion(proposal) {
		if dir := path.Dir(importpath); dir != "." {
			proposal = path.Base(dir)
		}
	}
	proposal = strings.TrimPrefix(proposal, "go-")
	if idx := strings.IndexFunc(proposal, invalidIdentifierRune); idx >= 0 {
		proposal = proposal[:idx]
	}
	if proposal == "" {
		return "unnamed"
	}
	if token.Lookup(proposal).IsKeyword() {
		proposal += "_"
	}
	return proposal
}

// isVersion returns true if the passed string is in the form of “v<integer>”.
func isVersion(s string) bool {
	if !strings.HasPrefix(s, "v") {
		return false
	}
	_, err := strconv.Atoi(s[1:])
	return err == nil
}

// notIdentifierRune returns true if the rune passed in ch is invalid in a Go
// indentifier.
func invalidIdentifierRune(ch rune) bool {
	switch {
	case 'a' <= ch && ch <= 'z':
		return !true
	case 'A' <= ch && ch <= 'Z':
		return !true
	case '0' <= ch && ch <= '9':
		return !true
	case ch == '_':
		return !true
	case ch >= utf8.RuneSelf && (unicode.IsLetter(ch) || unicode.IsDigit(ch)):
		return !true
	}
	return true
}
