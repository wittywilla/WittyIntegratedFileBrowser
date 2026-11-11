// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 WittyWilla and contributors

package wittyintegratedfilebrowser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestList(t *testing.T) {
	root := t.TempDir()
	mk := func(p string, dir bool) {
		full := filepath.Join(root, p)
		if dir {
			if err := os.MkdirAll(full, 0o755); err != nil {
				t.Fatal(err)
			}
			return
		}
		if err := os.WriteFile(full, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mk("Swamptale/VIDEO_TS", true)
	mk("Treeline/BDMV", true)
	mk("alpha", true)
	mk(".hidden", true)
	mk("film.mkv", false)
	mk("Film.srt", false)
	mk("disc.iso", false)
	mk("notes.txt", false)

	l, err := List(root, Options{})
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range l.Entries {
		names = append(names, e.Name)
	}
	want := []string{"alpha", "Swamptale", "Treeline", "disc.iso", "film.mkv", "Film.srt", "notes.txt"}
	if len(names) != len(want) {
		t.Fatalf("entries = %v, want %v", names, want)
	}
	for i := range want {
		if names[i] != want[i] {
			t.Fatalf("entries = %v, want %v (folders first, by name)", names, want)
		}
	}
	byName := map[string]Entry{}
	for _, e := range l.Entries {
		byName[e.Name] = e
	}
	if byName["Swamptale"].Disc != "dvd" || byName["Treeline"].Disc != "bluray" || byName["alpha"].Disc != "" {
		t.Errorf("disc folders: %+v %+v %+v", byName["Swamptale"], byName["Treeline"], byName["alpha"])
	}
	if byName["film.mkv"].Kind != "video" || byName["disc.iso"].Kind != "disc" || byName["Film.srt"].Kind != "subtitle" || byName["notes.txt"].Kind != "book" {
		t.Errorf("kinds wrong: %+v", l.Entries)
	}
	if l.Counts["folder"] != 3 || l.Counts["video"] != 1 || l.Counts["disc"] != 1 {
		t.Errorf("counts = %v", l.Counts)
	}
	if l.Parent == nil {
		t.Error("a temp folder has a parent")
	}

	h, _ := List(root, Options{ShowHidden: true})
	if len(h.Entries) != len(want)+1 {
		t.Errorf("ShowHidden: %d entries, want %d", len(h.Entries), len(want)+1)
	}
	small, _ := List(root, Options{MaxEntries: 2})
	if len(small.Entries) != 2 || !small.Truncated {
		t.Errorf("MaxEntries: %d entries, truncated=%v", len(small.Entries), small.Truncated)
	}
	viaFile, _ := List(filepath.Join(root, "film.mkv"), Options{})
	if viaFile == nil || viaFile.Path != l.Path {
		t.Error("a file's path lists its folder")
	}
	if _, err := List(filepath.Join(root, "missing"), Options{}); err != ErrNotFound {
		t.Errorf("missing folder: err = %v", err)
	}
	custom, _ := List(root, Options{Kind: func(string) string { return "x" }})
	if custom.Counts["x"] != 4 {
		t.Errorf("Options.Kind not used: %v", custom.Counts)
	}
}

func TestPlaces(t *testing.T) {
	extra := t.TempDir()
	ps := Places(Place{Name: "Mine", Path: extra, Group: "libraries"}, Place{Name: "Gone", Path: filepath.Join(extra, "nope"), Group: "libraries"})
	var drives, mine, gone int
	for _, p := range ps {
		switch {
		case p.Group == "drives":
			drives++
		case p.Name == "Mine":
			mine++
		case p.Name == "Gone":
			gone++
		}
	}
	if drives == 0 || mine != 1 || gone != 0 {
		t.Errorf("Places: drives=%d mine=%d gone=%d (%+v)", drives, mine, gone, ps)
	}
}

// Past the limit, folders are kept and files fill what's left, whatever the
// names: the browser is for choosing folders.
func TestListKeepsFoldersPastTheLimit(t *testing.T) {
	root := t.TempDir()
	for i := 0; i < 30; i++ {
		if err := os.WriteFile(filepath.Join(root, fmt.Sprintf("A%04d.jpg", i)), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, d := range []string{"extras", "zz", "Behind"} {
		if err := os.Mkdir(filepath.Join(root, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	l, err := List(root, Options{MaxEntries: 10})
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range l.Entries {
		names = append(names, e.Name)
	}
	want := []string{"Behind", "extras", "zz", "A0000.jpg", "A0001.jpg", "A0002.jpg", "A0003.jpg", "A0004.jpg", "A0005.jpg", "A0006.jpg"}
	if !l.Truncated || strings.Join(names, " ") != strings.Join(want, " ") {
		t.Errorf("entries = %v (truncated %v), want %v", names, l.Truncated, want)
	}
	if l.Counts["folder"] != 3 {
		t.Errorf("counts = %v, want 3 folders", l.Counts)
	}
	few, _ := List(root, Options{MaxEntries: 2})
	if len(few.Entries) != 2 || few.Entries[0].Name != "Behind" || few.Entries[1].Name != "extras" || !few.Truncated {
		t.Errorf("more folders than the limit: %+v", few.Entries)
	}
}
