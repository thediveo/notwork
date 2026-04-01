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

package astiter

import (
	"go/ast"
	"go/types"
	"iter"
)

// ASTNodePointer constrains T to a pointer to a type fulfilling the [ast.Node]
// interface.
type ASTNodePointer[T any] interface {
	// For background details about pointer constraints, see also:
	// https://stackoverflow.com/a/72091526 and
	// https://stackoverflow.com/a/71444968
	ast.Node
	*T
}

// AllOf iterates over AST nodes of the specified type T that must satisfy
// [ast.Node].
func AllOf[T any, PT ASTNodePointer[T]](root ast.Node) iter.Seq[*T] {
	return func(yield func(*T) bool) {
		for node := range ast.Preorder(root) {
			node, ok := node.(PT)
			if !ok {
				continue
			}
			if !yield(node) {
				break
			}
		}
	}
}

// TypesObjectPointer constrains T to a pointer to a type fulfilling the
// [types.Object] interface.
type TypesObjectPointer[T any] interface {
	// For background details about pointer constraints, see also:
	// https://stackoverflow.com/a/72091526 and
	// https://stackoverflow.com/a/71444968
	types.Object
	*T
}

// AllTypesOf iterates over [types.Object]s  of the specified type T.
func AllTypesOf[T any, PT TypesObjectPointer[T]](root ast.Node, typesinfo *types.Info) iter.Seq[*T] {
	return func(yield func(*T) bool) {
		for typedobj := range allTypesOf[T, PT](root, typesinfo) {
			if !yield(typedobj) {
				break
			}
		}
	}
}

func allTypesOf[T any, PT TypesObjectPointer[T]](root ast.Node, typesinfo *types.Info) iter.Seq[PT] {
	return func(yield func(PT) bool) {
		for ident := range AllOf[ast.Ident](root) {
			obj, ok := typesinfo.Uses[ident]
			if !ok {
				continue
			}
			typedobj, ok := obj.(PT)
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
func AllInPkgOf[T any, PT TypesObjectPointer[T]](importpath string, root ast.Node, typesinfo *types.Info) iter.Seq[*T] {
	return func(yield func(*T) bool) {
		for typedobj := range allTypesOf[T, PT](root, typesinfo) {
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

// AllFuncsInPkg is a convenience wrapper that iterates over the
// [types.Func]tions with the specified import path in the passed file.
func AllFuncsInPkg(importpath string, root ast.Node, typesinfo *types.Info) iter.Seq[*types.Func] {
	return AllInPkgOf[types.Func](importpath, root, typesinfo)
}
