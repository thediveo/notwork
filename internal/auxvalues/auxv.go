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

package auxvalues

import (
	"encoding/binary"
	"fmt"
	"os"
	"unsafe"
)

const (
	AT_NULL   = 0
	AT_IGNORE = 1
	// AT_CLKTCK specifies the frequency at thich times() increment, see also:
	// https://elixir.bootlin.com/linux/v7.0.1/source/include/uapi/linux/auxvec.h#L26
	AT_CLKTCK = 17
)

// CLKTCK returns the kernel's USER_HZ (AT_CLKTCK) configuration.
func CLKTCK() uint64 {
	return uint64(typedValues.Get(AT_CLKTCK, 100))
}

// https://www.man7.org/linux/man-pages/man5/proc_pid_auxv.5.html
const procSelfAuxv = "/proc/self/auxv"

type ulong = uint

type auxvMap map[ulong]ulong

// Get looks up the specified auxiliary vector type and returns its value; if
// the type isn't found, it returns the specified default value instead.
func (m auxvMap) Get(typ ulong, defvalue ulong) ulong {
	if m == nil {
		return defvalue
	}
	val := m[typ]
	if val == 0 {
		return defvalue
	}
	return val
}

// typedValues maps “AT_” auxv types to their values that have been read from
// “/proc/self/auxv”.
var typedValues auxvMap

func init() {
	typedValues = readauxv(procSelfAuxv)
}

// readauxv reads the auxiliary vector typed values and returns the
// type-to-value mapping. If the auxiliary vector data cannot be read, readauxv
// returns an empty map.
//
// For background information, see:
// https://www.man7.org/linux/man-pages/man5/proc_pid_auxv.5.html as well as
// https://www.man7.org/linux/man-pages/man3/getauxval.3.html
func readauxv(path string) auxvMap {
	auxv, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var readUlong func([]byte) ulong
	wordSize := unsafe.Sizeof(uintptr(0))
	switch wordSize {
	case 8:
		readUlong = func(b []byte) ulong {
			return ulong(binary.NativeEndian.Uint64(b))
		}
	case 4:
		readUlong = func(b []byte) ulong {
			return ulong(binary.NativeEndian.Uint32(b))
		}
	default:
		panic(fmt.Sprintf("unsupported auxv word size of %d", wordSize))
	}

	m := auxvMap{}
	for len(auxv) >= 2*int(wordSize) {
		typ := readUlong(auxv[0:wordSize])
		if typ == AT_NULL {
			break
		}
		value := readUlong(auxv[wordSize : 2*wordSize])
		m[typ] = value
		auxv = auxv[2*wordSize:]
	}
	return m
}
