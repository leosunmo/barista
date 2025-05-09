// Copyright 2017 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

/*
Package material provides support for Google's Material Design Symbols.
from https://github.com/google/material-design-icons

It uses variablefont/MaterialSymbolsOutlined[FILL,GRAD,opsz,wght].codepoints
to get the list of icons,
and requires variablefont/MaterialSymbolsOutlined[FILL,GRAD,opsz,wght].ttf to be installed.
*/
package symbols

import (
	"bufio"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/leosunmo/barista/pango"
	"github.com/leosunmo/barista/pango/icons"

	"github.com/spf13/afero"
)

var fs = afero.NewOsFs()

func LoadFile(filePath string) error {
	f, err := fs.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	symbol := icons.NewProvider("symbol")
	symbol.Font("Material Symbols Outlined")
	symbol.AddStyle(func(n *pango.Node) { n.Rise(-3000) })
	s := bufio.NewScanner(f)
	s.Split(bufio.ScanLines)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		components := strings.Split(line, " ")
		if len(components) != 2 {
			return fmt.Errorf("unexpected line '%s' in file", line)
		}
		// Material Design Symbols uses '_', but all other fonts use '-',
		// so we'll normalise it here.
		name := strings.Replace(components[0], "_", "-", -1)
		value := components[1]
		err = symbol.Hex(name, value)
		if err != nil {
			return err
		}
	}
	return nil
}

// Load initialises the material design symbols provider from the given repo.
func Load(repoPath string) error {
	f, err := fs.Open(filepath.Join(repoPath, "variablefont/MaterialSymbolsOutlined[FILL,GRAD,opsz,wght].codepoints"))
	if err != nil {
		return err
	}
	defer f.Close()
	material := icons.NewProvider("symbol")
	material.Font("Material Symbols Outlined")
	material.AddStyle(func(n *pango.Node) { n.UltraLight().Rise(-4000) })
	s := bufio.NewScanner(f)
	s.Split(bufio.ScanLines)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		components := strings.Split(line, " ")
		if len(components) != 2 {
			return fmt.Errorf("unexpected line '%s' in 'variablefont/MaterialSymbolsOutlined[FILL,GRAD,opsz,wght].codepoints'", line)
		}
		// Material Design Symbols uses '_', but all other fonts use '-',
		// so we'll normalise it here.
		name := strings.Replace(components[0], "_", "-", -1)
		value := components[1]
		err = material.Hex(name, value)
		if err != nil {
			return err
		}
	}
	return nil
}
