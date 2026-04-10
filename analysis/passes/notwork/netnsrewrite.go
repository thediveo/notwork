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

package notwork

import (
	"fmt"

	"github.com/thediveo/notwork/analysis/funcs"
	"golang.org/x/tools/go/analysis"
)

const (
	netnsOldImportPath    = "github.com/thediveo/notwork/netns"
	netnsNewImportPath    = "github.com/thediveo/spacetest/netns"
	nlhandleNewImportPath = "github.com/thediveo/notwork/nlhandle"
	nsNewImportPath       = "github.com/thediveo/notwork/ns"
)

var refactoring = funcs.Relocate(netnsOldImportPath, netnsNewImportPath,
	[]string{
		"EnterTransient",
		"NewTransient",
		"Execute",
		"Current",
		"Ino",
		"CurrentIno",
	}).Remap(netnsOldImportPath, nlhandleNewImportPath,
	funcs.NameMapping{
		"NewNetlinkHandle": "New",
	}).Remap(netnsOldImportPath, nsNewImportPath,
	funcs.NameMapping{
		"NsID": "ID",
	})

var Analyzer = &analysis.Analyzer{
	Name: "notworknetns",
	Doc:  "rewrite deprecated imports of github.com/thediveo/netns",
	Run:  run,
}

// run our analysis on a specific package and over all its individual files.
func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		for pfnmapping, pos := range refactoring.AllFuncs(file, pass.TypesInfo) {
			pass.Report(analysis.Diagnostic{
				Pos:     pos,
				Message: fmt.Sprintf("deprecated usage of %s", pfnmapping.Before.String()),
			})
		}
	}
	return nil, nil
}
