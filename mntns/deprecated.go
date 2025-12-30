// Copyright 2024 Harald Albrecht.
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

package mntns

import (
	"github.com/thediveo/spacetest/mntns"

	gi "github.com/onsi/ginkgo/v2" //lint:ignore ST1001 rule does not apply
)

// EnterTransient creates and enters a new mount namespace, returning a function
// that needs to be defer'ed.
//
// Deprecated: use [mntns.EnterTransient] from “thediveo/spacetest” instead.
func EnterTransient() func() {
	gi.GinkgoHelper()
	return mntns.EnterTransient()
}

// MountSysfsRO mounts a new sysfs instance read-only onto /sys when the caller
// is in a new and transient mount namespace.
//
// Deprecated: use [mntns.EnterTransient] from “thediveo/spacetest” instead.
func MountSysfsRO() {
	gi.GinkgoHelper()
	mntns.MountSysfsRO()
}

// NewTransient creates a new transient mount namespace that is kept alive by a
// an idle OS-level thread; this idle thread is automatically terminated upon
// returning from the current test.
//
// Deprecated: use [mntns.EnterTransient] from “thediveo/spacetest” instead.
func NewTransient() (mntfd int, procfsroot string) {
	gi.GinkgoHelper()
	return mntns.NewTransient()
}

// Execute a function fn in a separate(!) Go routine in the mount namespace
// referenced by the open file descriptor mntnsfd.
//
// Deprecated: use [mntns.EnterTransient] from “thediveo/spacetest” instead.
func Execute(mntnsfd int, fn func()) {
	gi.GinkgoHelper()
	mntns.Execute(mntnsfd, fn)
}

// Current returns a file descriptor referencing the current mount namespace.
//
// Deprecated: use [mntns.EnterTransient] from “thediveo/spacetest” instead.
func Current() int {
	gi.GinkgoHelper()
	return mntns.Current()
}

// Ino returns the identification/inode number of the passed mount namespace,
// either referenced by a file descriptor or a VFS path name.
//
// Deprecated: use [mntns.EnterTransient] from “thediveo/spacetest” instead.
func Ino[R ~int | ~string](mntnsref R) uint64 {
	gi.GinkgoHelper()
	return mntns.Ino(mntnsref)
}
