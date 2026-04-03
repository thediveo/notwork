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

// DotMap is a map/dictionary of package import paths.
type DotMap map[string]struct{}

// CollectDots returns a new map/dictionary of package import paths that have
// been dot-imported.
func CollectDots(file *ast.File, typeinfo *types.Info) DotMap {
	m := DotMap{}

	for _, imp := range file.Imports {
		if imp.Name == nil || imp.Name.Name != "." {
			continue
		}
		imppath := strings.Trim(imp.Path.Value, `"`)
		m[imppath] = struct{}{}
	}

	return m
}
