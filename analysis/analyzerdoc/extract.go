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

package analyzerdoc

import (
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"
)

var _ = analysis.Analyzer{}

// MustExtract must successfully extract the analyzer section for the specified
// analyzername; otherwise, it will panic.
func MustExtract(contents, analyzername string) string {
	doc, err := Extract(contents, analyzername)
	if err != nil {
		panic(err)
	}
	return doc
}

// Extract a section of a package doc comment from the provided contents of a
// doc.go file of an analyzer. The section is that part of the doc comment
// between one heading and the next that conform to the following form:
//
//	# Analyzer NAME
//
//	NAME: SUMMARY
//
//	FULL DESCRIPTION...
//
// ...where NAME must match the passed analyzername. SUMMARY should be a brief
// verb-phrase describing the analyzer. Extract returns the portion following
// the colon, that is, starting with SUMMARY, which is the form expected by
// [analysis.Analyzer.Doc].
func Extract(contents, analyzername string) (string, error) {
	if contents == "" {
		return "", errors.New("empty Go source file")
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", contents,
		parser.ParseComments|parser.PackageClauseOnly)
	if err != nil {
		return "", errors.New("invalid Go source file")
	}

	if file.Doc == nil {
		return "", errors.New("Go source file without package doc comment")
	}

	for section := range strings.SplitSeq(file.Doc.Text(), "\n# ") {
		body, ok := strings.CutPrefix(section, "Analyzer "+analyzername)
		if !ok || body == "" || (body[0] != '\r' && body[0] != '\n') {
			continue
		}
		body = strings.TrimSpace(body)
		bodycontents, ok := strings.CutPrefix(body, analyzername+":")
		if !ok {
			return "", fmt.Errorf("\"Analyzer %s\" heading missing following \"%s: summary...\" line",
				analyzername, analyzername)
		}
		return strings.TrimSpace(bodycontents), nil
	}
	return "", fmt.Errorf("package doc comment missing the \"Analyzer %s\" heading",
		analyzername)
}
