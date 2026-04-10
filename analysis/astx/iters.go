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
	"go/types"
	"iter"
)

// AllOf iterates over AST nodes of the specified type T, where T must satisfy
// [ast.Node].
func AllOf[T ast.Node](root ast.Node) iter.Seq[T] {
	return func(yield func(T) bool) {
		if root == nil {
			return
		}
		for node := range ast.Preorder(root) {
			node, ok := node.(T)
			if !ok {
				continue
			}
			if !yield(node) {
				break
			}
		}
	}
}

// AllTypesOf iterates over [types.Object]s  of the specified type T.
func AllTypesOf[T types.Object](root ast.Node, typesinfo *types.Info) iter.Seq[T] {
	return func(yield func(T) bool) {
		for ident := range AllOf[*ast.Ident](root) {
			obj, ok := typesinfo.Uses[ident]
			if !ok {
				continue
			}
			typedobj, ok := obj.(T)
			if !ok {
				continue
			}
			if !yield(typedobj) {
				break
			}
		}
	}
}

// AllInPkgOf iterates over [types.Object]s of the specified type T and with the
// specified import path in the passed path.
func AllInPkgOf[T types.Object](importpath string, root ast.Node, typesinfo *types.Info) iter.Seq[T] {
	return func(yield func(T) bool) {
		for typedobj := range AllTypesOf[T](root, typesinfo) {
			pkg := typedobj.Pkg()
			if pkg == nil || pkg.Path() != importpath {
				continue
			}
			if !yield(typedobj) {
				break
			}
		}
	}
}
