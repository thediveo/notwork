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

package idents

import "go/types"

type Map map[string]struct{}

// Collect returns a map/dictionary of identifiers from the definitions and uses
// of the passed type information.
func Collect(ti types.Info) Map {
	m := Map{}

	for ident := range ti.Defs {
		if ident == nil || ident.Name == "_" {
			continue
		}
		m[ident.Name] = struct{}{}
	}
	for ident := range ti.Uses {
		if ident == nil || ident.Name == "_" {
			continue
		}
		m[ident.Name] = struct{}{}
	}

	return m
}
