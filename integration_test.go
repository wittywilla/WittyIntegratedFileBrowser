// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 WittyWilla and contributors

package wittyintegratedfilebrowser

import (
	"os"
	"testing"

	"github.com/wittywilla/wittymodule"
)

// integration.json is what a server reads strictly at install: it passes WittyModule's own check.
func TestIntegrationFile(t *testing.T) {
	m, err := wittymodule.ParseManifest(ManifestJSON)
	if err != nil {
		t.Fatal(err)
	}
	if !m.NeedsIntegration() {
		t.Fatal("a module of API 5 with a helper carries integration.json")
	}
	b, err := os.ReadFile("integration.json")
	if err != nil {
		t.Fatal(err)
	}
	in, err := wittymodule.ParseIntegration(b)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range in.Check(m) {
		t.Errorf("%v", p)
	}
	// The server keeps the folder route (GET /api/fs) and calls the helper itself: no routes,
	// data or lifecycle of the module's own.
	if len(in.HTTP) > 0 || in.Data != nil || in.Lifecycle != nil {
		t.Errorf("routes, data or lifecycle: %+v", in)
	}
}
