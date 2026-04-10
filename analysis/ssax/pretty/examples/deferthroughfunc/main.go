//usr/bin/true; exec /usr/bin/env go run "$0" "$@"

// simply execute this file as a script to see its analysis SSA output.

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

package main

import (
	"os"

	"github.com/thediveo/notwork/analysis/ssax/pretty"
)

func main() {
	pkg, fileset := pretty.MustBuildSSA(
		`package main

func Foo() func() {
	return func() { }
}

var F = Foo

func s[A any, B any](a A, b B) A { return a }

func main() {
	f := func(i int, fn func() func()) (func() func(), int) {
		return fn, i+1
	}
	defer s(f(42, F))
	defer s(f(123, F))()
	_ = s("foo", 123)
}
`)
	prt := pretty.NewPrinter(os.Stdout, fileset)
	prt.Package(pkg, 0)
}
