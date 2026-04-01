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

package imports

import (
	"errors"
	"fmt"
	"go/ast"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/tools/go/packages"
)

func packageLocalName(path string, dir string) (string, error) {
	pkgs, err := packages.Load(
		&packages.Config{
			Mode: packages.NeedName,
			Dir:  dir,
		}, path)
	if err != nil {
		return "", err
	}
	// Try to find out if there are any errors...
	var msgs strings.Builder
	mods := map[*packages.Module]bool{}
	for pkg := range packages.Postorder(pkgs) {
		for _, err := range pkg.Errors {
			msgs.WriteString(err.Error())
			msgs.WriteRune('\n')
		}
		mod := pkg.Module
		if mod == nil || mod.Error == nil || mods[mod] {
			continue
		}
		mods[mod] = true
		msgs.WriteString(mod.Error.Err)
		msgs.WriteRune('\n')
	}
	if msgs.Len() != 0 {
		return "", fmt.Errorf("failed to load package, reason(s):\n%s", msgs.String())
	}
	//
	if len(pkgs) != 1 {
		return "", errors.New("failed to load package")
	}
	return pkgs[0].Name, nil
}

// LocalNameAndPath returns the local name for the package as well as its
// (unquoted) import path. Contrary to a plain ast.ImportSpec, the local name
// returned is always a non-zero string, deriving it from the import path where
// necessary.
func LocalNameAndPath(i *ast.ImportSpec) (localname string, path string) {
	path = strings.Trim(i.Path.Value, "\"")
	if i.Name != nil {
		return i.Name.Name, path
	}
	return filepath.Base(path), path
}

type Map struct {
	byPath     map[string]string
	localNames map[string]struct{}
}

func NewMap(file *ast.File) *Map {
	m := &Map{
		byPath:     map[string]string{},
		localNames: map[string]struct{}{},
	}

	for _, importSpec := range file.Imports {
		name, path := LocalNameAndPath(importSpec)
		m.byPath[path] = name
		m.localNames[name] = struct{}{}
	}

	return m
}

// LocalNameOf returns the local name for the passed import path. If the path
// hasn't been seen yet, a local name will be derived from it
func (m *Map) LocalNameOf(path string) string {
	name, ok := m.byPath[path]
	if ok {
		return name
	}
	// we don't know about this import path yet; but in order to correctly add
	// it we must ensure to use a non-colliding local name for this new import.
	name = filepath.Base(path)
	if _, ok := m.localNames[name]; ok {
		// ...collision! Let's try to get collision-free, so that for
		// "github.com/example/repo" we now try "examplerepo".
		name = filepath.Base(filepath.Dir(path)) + name
	}
	if _, ok := m.localNames[name]; ok {
		for n := range 100 {
			suffix := strconv.FormatInt(int64(n+1), 10)
			if _, ok := m.localNames[name+suffix]; !ok {
				name += suffix
				break
			}
		}
	}
	m.byPath[path] = name
	m.localNames[name] = struct{}{}
	return name
}
