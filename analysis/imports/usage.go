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
	"go/ast"
	"go/types"
	"strings"
)

// PackageUsage of a single, specific package import and identifier within a
// specific file: most notably the actual local name and the specific uses
// within this file. Use [CollectPackageUsages] to create a file-specific map from package
// import paths to the import uses in this file.
type PackageUsage /* not Usagi scnr */ struct {
	LocalName string          // might be "" in case this import is unused
	Import    *ast.ImportSpec // the originating AST import spec
	Uses      []*ast.Ident    // all uses of this package ident we've seen in the AST
	PkgName   *types.PkgName  // the imported package
}

// PackageUsagesByPath maps import paths to import usages.
type PackageUsagesByPath map[string]*PackageUsage

// CollectPackageUsages returns a new map from import paths to the local name
// (and thus package ident) of the corresponding packages, as well as their uses
// in the passed *ast.File.
//
// Please note that dot imports as well as underline imports never included.
func CollectPackageUsages(file *ast.File, typesinfo *types.Info) PackageUsagesByPath {
	m := PackageUsagesByPath{}

	// Phase I: scan the import specifications in the given file, initializing
	// their corresponding Usage entries.
	for _, imp := range file.Imports {
		// Luckily (for us), the local import name is correctly set for both dot
		// imports as well as underscore imports. Phew. So we can easily
		// identify them and then skip over them.
		if imp.Path.Value == `"C"` ||
			(imp.Name != nil && (imp.Name.Name == "." || imp.Name.Name == "_")) {
			continue
		}
		imppath := strings.Trim(imp.Path.Value, `"`)
		m[imppath] = &PackageUsage{
			Import: imp,
		}
	}

	// Phase II: scan all identifiers, getting their type information and see if
	// they are belonging to one of the imports. Please note that Uses doesn't
	// include idents that are in Defs and never used. Fun fact: package idents
	// actually appear in Defs for alias imports, but not for default imports
	// ...bummer.
	for ident, obj := range typesinfo.Uses {
		pkgname, ok := obj.(*types.PkgName)
		if !ok {
			continue
		}
		usage, ok := m[pkgname.Imported().Path()]
		if !ok {
			continue
		}
		usage.Uses = append(usage.Uses, ident)
		if usage.LocalName == "" {
			usage.LocalName = ident.Name
		}
		if usage.PkgName == nil {
			usage.PkgName = pkgname
		}
	}

	return m
}
