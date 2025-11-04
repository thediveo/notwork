// Copyright 2025 Harald Albrecht.
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

package load

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"pault.ag/go/modprobe"
)

// Try checks if the caller is root and netdevsim is available as a system bus,
// returning true, otherwise attempting to load the required kernel module.
//
// Please note that Try does use a pure-Go kernel module prober and loader
// ([modprobe.Load]), so modprobe(8) doesn't need to be present.
func Try() bool { return TryRoot("/") }

// TryRoot is like [Try], but expects sys/bus/netdevsim to be inside the
// specified sysfsroot, instead of "/".
func TryRoot(sysfsroot string) bool {
	// managing netdevsim devices requires root, because creating and
	// linking/unlinking netdevsim devices goes through the DAC of the
	// filesystem inside /sys/bus/netdevsim. Thus, if we're not root, even with
	// the bus or CAP_SYS_MODULE present, we won't be able to create or
	// otherwise manage any netdevsim devices later, so we don't need to try.
	if os.Getuid() != 0 {
		return false
	}
	// now check if the netsimdev bus already correctly exists inside sysfs.
	info, err := os.Stat(filepath.Join(sysfsroot, "sys/bus/netdevsim"))
	if err == nil && info.Mode().IsDir() {
		return true
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return false // no chance, something broken here.
	}
	// try to modprobe
	return modprobe.Load("netdevsim", "") == nil
}
