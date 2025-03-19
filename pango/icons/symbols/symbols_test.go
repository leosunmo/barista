// Copyright 2018 Google Inc.
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

package symbols

import (
	"testing"

	"github.com/leosunmo/barista/pango"
	"github.com/leosunmo/barista/testing/cron"
	"github.com/leosunmo/barista/testing/githubfs"
	pangoTesting "github.com/leosunmo/barista/testing/pango"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestInvalid(t *testing.T) {
	fs = afero.NewMemMapFs()
	require.Error(t, Load("/src/no-such-directory"))

	err := afero.WriteFile(fs, "/src/material-error-1/variablefont/MaterialSymbolsOutlined[FILL,GRAD,opsz,wght].codepoints", []byte(
		`-- Lines in weird formats --
		 Empty:

		 Valid line:
		 someIcon 61
		 otherIcon 62
		 Invalid codepoint:
		 badIcon xy`,
	), 0644)
	require.NoError(t, err, "afero.WriteFile failed")
	require.Error(t, Load("/src/material-error-1"))

	err = afero.WriteFile(fs, "/src/material-error-2/iconfont/codepoint", nil, 0644)
	require.NoError(t, err, "afero.WriteFile failed")
	require.Error(t, Load("/src/material-error-2"))

	err = afero.WriteFile(fs, "/src/material-error-3/variablefont/MaterialSymbolsOutlined[FILL,GRAD,opsz,wght].codepoints", []byte(
		`someIcon 61
		 otherIcon 62
		 badIcon xy`,
	), 0644)
	require.NoError(t, err, "afero.WriteFile failed")
	require.Error(t, Load("/src/material-error-3"))
}

func TestValid(t *testing.T) {
	fs = afero.NewMemMapFs()
	err := afero.WriteFile(fs, "/src/material/variablefont/MaterialSymbolsOutlined[FILL,GRAD,opsz,wght].codepoints", []byte(
		`someIcon 61
		 otherIcon 62
		 thirdIcon 63`,
	), 0644)
	require.NoError(t, err, "afero.WriteFile failed")
	require.NoError(t, Load("/src/material"))
	t.Logf("pango.Icon(\"symbol-someIcon\").String() = %v\n", pango.Icon("symbol-someIcon").String())
	pangoTesting.AssertText(t, "a", pango.Icon("symbol-someIcon").String())
}

func TestValidFromFile(t *testing.T) {
	fs = afero.NewMemMapFs()
	err := afero.WriteFile(fs, "/src/material/variablefont/MaterialSymbolsOutlined[FILL,GRAD,opsz,wght].codepoints", []byte(
		`someIcon 61
		 otherIcon 62
		 thirdIcon 63`,
	), 0644)
	require.NoError(t, err, "afero.WriteFile failed")
	require.NoError(t, LoadFile("/src/material/variablefont/MaterialSymbolsOutlined[FILL,GRAD,opsz,wght].codepoints"))
	pangoTesting.AssertText(t, "a", pango.Icon("symbol-someIcon").String())
}

// TestLive tests that current master branch of the icon font works with
// this package. This test only runs when CI runs tests in 'cron' mode,
// which provides timely notifications of incompatible changes while
// keeping default tests hermetic.
func TestLive(t *testing.T) {
	fs = githubfs.New()
	cron.Test(t, func() error {
		if err := Load("/google/material-design-icons/master"); err != nil {
			return err
		}
		// At least one of these icons should be loaded.
		testIcons := pango.New(
			pango.Icon("symbol-face"),
			pango.Icon("symbol-room"),
			pango.Icon("symbol-view-agenda"),
			pango.Icon("symbol-link-off"),
		)
		require.NotEmpty(t, testIcons.String(), "No expected icons were loaded")
		return nil
	})
}
