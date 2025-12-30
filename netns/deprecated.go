// Copyright 2023 Harald Albrecht.
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

package netns

import (
	"github.com/thediveo/spacetest/netns"

	gi "github.com/onsi/ginkgo/v2"
)

// EnterTransient creates and enters a new (and isolated) network namespace,
// returning a function that needs to be defer'ed in order to correctly switch
// the calling go routine and its locked OS-level thread back when the caller
// itself returns.
//
// Deprecated: use [netns.EnterTransient] from “thediveo/spacetest” instead.
func EnterTransient() func() {
	gi.GinkgoHelper()
	return netns.EnterTransient()
}

// NewTransient creates a new network namespace, but doesn't enter it, returning
// a file descriptor referencing the new network namespace.
//
// Deprecated: use [netns.NewTransient] from “thediveo/spacetest” instead.
func NewTransient() int {
	gi.GinkgoHelper()
	return netns.NewTransient()
}

// Execute a function fn in the network namespace referenced by the open file
// descriptor netnsfd.
//
// Deprecated: use [netns.Execute] from “thediveo/spacetest” instead.
func Execute(netnsfd int, fn func()) {
	gi.GinkgoHelper()
	netns.Execute(netnsfd, fn)
}

// Current returns a file descriptor referencing the current network namespace.
//
// Deprecated: use [netns.Current] from “thediveo/spacetest” instead.
func Current() int {
	gi.GinkgoHelper()
	return netns.Current()
}

// Ino returns the identification/inode number of the passed network namespace,
// either referenced by a file descriptor or a VFS path name.
//
// Deprecated: use [netns.Ino] from “thediveo/spacetest” instead.
func Ino[R ~int | ~string](netnsref R) uint64 {
	gi.GinkgoHelper()
	return netns.Ino(netnsref)
}

// CurrentIno returns the identification/inode number of the network namespace
// for the current thread.
//
// Deprecated: use [netns.CurrentIno] from “thediveo/spacetest” instead.
func CurrentIno() uint64 {
	gi.GinkgoHelper()
	return netns.CurrentIno()
}
