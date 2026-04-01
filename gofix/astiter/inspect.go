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

import "go/ast"

// Inspect traverses the AST starting at the root node in depth-first order,
// skipping children of nodes for which the called fn returns false. In contrast
// to Go's [iter.Seq] and [iter.Seq2] returning false does not abort the whole
// iteration, just descending further into children of a particular node.
//
// Please note that in contrast to [ast.Inspect], fn won't be ever called with a
// nil node.
func Inspect(root ast.Node, fn func(ast.Node) bool) {
	ast.Inspect(root, func(node ast.Node) bool {
		if node == nil {
			return true
		}
		return fn(node)
	})
}
