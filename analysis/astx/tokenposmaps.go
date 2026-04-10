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
	"go/ast"
	"go/token"
)

// NodesByPos maps token positions to AST nodes of type T.
type NodesByPos[T ast.Node] map[token.Pos]T

// NewNodesByPosOf returns a new map from token positions to AST nodes of
// type T, from the passed set of files.
func NewNodesByPosOf[T ast.Node](files []*ast.File) NodesByPos[T] {
	m := NodesByPos[T]{}
	for _, file := range files {
		for stmt := range AllOf[T](file) {
			m[stmt.Pos()] = stmt
		}
	}
	return m
}
