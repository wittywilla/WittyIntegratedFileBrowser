// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 WittyWilla and contributors

package ui

import (
	"slices"
	"testing"

	"github.com/wittywilla/wittymodule/sdui"
	core "github.com/wittywilla/wittymodule/sdui/core"
	"github.com/wittywilla/wittymodule/sdui/sduitest"
)

// sdui.json is this file's UI(), checked as a server checks it.
func TestSDUIFile(t *testing.T) {
	sduitest.Golden(t, UI(), "../sdui.json")
}

func TestServerReadsIt(t *testing.T) {
	b, err := sdui.JSON(UI())
	if err != nil {
		t.Fatal(err)
	}
	m, err := core.ParseModuleUI(b)
	if err != nil {
		t.Fatal(err)
	}
	for _, is := range core.ValidateModuleUI(m, Module, core.ModuleChecks{ColorToken: func(s string) bool { return slices.Contains(sdui.ColorTokens, s) }}) {
		if is.Severity == "error" {
			t.Errorf("%s: %s", is.Path, is.Message)
		}
	}
	if len(m.Pages) != 1 || m.Pages[0].ID != PageFolders || len(m.Navigation) != 1 {
		t.Fatalf("the folders page and its menu entry: %+v", m.Pages)
	}
}
