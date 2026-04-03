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
	"go/types"
	"slices"
)

// IsImporting returns true if the specified package imports one or more of the
// specified packages; otherwise, it returns false.
func IsImporting(pkg *types.Package, importpaths []string) bool {
	for _, imprt := range pkg.Imports() {
		if slices.Contains(importpaths, imprt.Path()) {
			return true
		}
	}
	return false
}
