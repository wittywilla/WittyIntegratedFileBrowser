// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 WittyWilla and contributors

package wittyintegratedfilebrowser

import (
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/wittywilla/wittyintegratedfilebrowser/ui"
	"github.com/wittywilla/wittymodule/sdui"
	core "github.com/wittywilla/wittymodule/sdui/core"
)

func ask(t *testing.T, raw string) sdui.ItemsRequest {
	t.Helper()
	var r sdui.ItemsRequest
	if err := json.Unmarshal([]byte(raw), &r); err != nil {
		t.Fatal(err)
	}
	return r
}

func TestPlaceItems(t *testing.T) {
	places := func() []Place {
		return []Place{
			{Name: "Media", Path: "/mnt/media", Group: "drives", Free: 1_649_267_441_664, Total: 4_398_046_511_104},
			{Name: "Movies", Path: "/home/me/Movies", Group: "folders"},
			{Name: "NAS", Path: "/mnt/nas", Group: "drives", Label: "Network"},
		}
	}
	req := func(admin bool, offset, limit int) sdui.ItemsRequest {
		return ask(t, `{"source":{"type":"wittyintegratedfilebrowser.places"},"offset":`+itoa(offset)+`,"limit":`+itoa(limit)+`,"viewer":{"profileKind":"adult","admin":`+map[bool]string{true: "true", false: "false"}[admin]+`}}`)
	}
	a, err := Items(req(true, 0, 2), places)
	if err != nil || len(a.Items) != 2 || !a.More || a.Total != 3 {
		t.Fatalf("first page: %+v %v", a, err)
	}
	if a.Items[0].Title != "Media" || a.Items[0].Subtitle != "1.5 TB free of 4.0 TB" || a.Items[0].Icon != "devices" || a.Items[0].Kind != "library" {
		t.Fatalf("a drive: %+v", a.Items[0])
	}
	if a.Items[1].Subtitle != "/home/me/Movies" || a.Items[1].Icon != "folder" {
		t.Fatalf("a folder: %+v", a.Items[1])
	}
	if act := a.Items[1].Action; act == nil || act.Type != core.ActNavigate || act.To != "/page/wittyintegratedfilebrowser?name="+url.QueryEscape("/home/me/Movies") {
		t.Fatalf("a place opens its folder's page: %+v", act)
	}
	b, _ := Items(req(true, 2, 2), places)
	if len(b.Items) != 1 || b.More || b.Items[0].Subtitle != "Network" {
		t.Fatalf("second page: %+v", b)
	}
	if b.Items[0].ID == a.Items[0].ID || len(b.Items[0].ID) > 200 {
		t.Fatalf("ids must be unique and short: %q", b.Items[0].ID)
	}
	if n, _ := Items(req(false, 0, 10), places); len(n.Items) != 0 {
		t.Fatalf("someone who isn't an administrator sees folders: %+v", n)
	}
	if _, err := Items(sdui.ItemsRequest{Source: sdui.Source{"type": "other"}}, places); err == nil {
		t.Fatal("an unknown source was answered")
	}
}

func TestFolderItems(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "Albums"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{"song.flac", ".hidden"} {
		if err := os.WriteFile(filepath.Join(dir, f), []byte("12345"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	none := func() []Place { return nil }
	folder := func(admin bool, path string) sdui.ItemsAnswer {
		t.Helper()
		src, _ := json.Marshal(map[string]any{"type": ui.SourceFolder, "path": path})
		a, err := Items(ask(t, `{"source":`+string(src)+`,"viewer":{"profileKind":"adult","admin":`+map[bool]string{true: "true", false: "false"}[admin]+`}}`), none)
		if err != nil {
			t.Fatal(err)
		}
		return a
	}
	a := folder(true, dir)
	titles := []string{}
	for _, it := range a.Items {
		titles = append(titles, it.Title)
	}
	if len(a.Items) != 4 || a.Items[0].ID != "here" || a.Items[1].Title != "Up" || a.Items[2].Title != "Albums" || a.Items[3].Title != "song.flac" {
		t.Fatalf("where it is, up, folders, then files, nothing hidden: %q", titles)
	}
	if act := a.Items[2].Action; act == nil || act.Type != core.ActNavigate || act.To != ui.FolderRoute(url.QueryEscape(filepath.Join(dir, "Albums"))) {
		t.Fatalf("a folder opens its own page: %+v", act)
	}
	if act := a.Items[1].Action; act == nil || act.To != ui.FolderRoute(url.QueryEscape(filepath.Dir(dir))) {
		t.Fatalf("up opens the folder above: %+v", act)
	}
	if f := a.Items[3]; f.Action != nil || f.Trailing != "5 bytes" || f.Icon != "music" {
		t.Fatalf("a file is only named: %+v", f)
	}
	if n := folder(false, dir); len(n.Items) != 0 {
		t.Fatalf("someone who isn't an administrator sees a folder: %+v", n)
	}
	for _, p := range []string{`\\nas\share`, "//nas/share", "relative/path"} {
		n := folder(true, p)
		if len(n.Items) != 1 || n.Items[0].ID != "note" || n.Items[0].Action != nil {
			t.Fatalf("%s was opened: %+v", p, n)
		}
	}
	if n := folder(true, filepath.Join(dir, "missing")); len(n.Items) != 1 || !strings.Contains(n.Items[0].Title, "can't open") {
		t.Fatalf("a missing folder: %+v", n)
	}
	// With no path, the folder page shows where to start.
	if n, _ := Items(ask(t, `{"source":{"type":"wittyintegratedfilebrowser.folder","path":""},"viewer":{"admin":true}}`),
		func() []Place { return []Place{{Name: "Here", Path: dir, Group: "folders"}} }); len(n.Items) != 1 || n.Items[0].Title != "Here" {
		t.Fatalf("no path: %+v", n)
	}
}

func itoa(n int) string { b, _ := json.Marshal(n); return string(b) }

// A server keeps only the items of the kinds a source's spec yields, so every item each source
// answers, places and folders alike, has to be one of them.
func TestItemsAreWhatEachSourceYields(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "Albums"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "song.mp3"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	places := func() []Place { return []Place{{Name: "Media", Path: dir, Group: "drives"}} }
	yields := map[string][]string{}
	for _, s := range ui.UI().Sources {
		yields[s.Type] = s.Yields
	}
	for _, src := range []sdui.Source{
		{"type": ui.SourcePlaces},
		{"type": ui.SourceFolder},
		{"type": ui.SourceFolder, "path": dir},
	} {
		a, err := Items(sdui.ItemsRequest{Source: src, Limit: 50, Viewer: sdui.Viewer{ProfileKind: "adult", Admin: true}}, places)
		if err != nil || len(a.Items) == 0 {
			t.Fatalf("%v: %+v %v", src, a, err)
		}
		for _, it := range a.Items {
			if !slices.Contains(yields[src.Type()], it.Kind) {
				t.Errorf("%v: %q is a %q, and the source yields %v", src, it.Title, it.Kind, yields[src.Type()])
			}
		}
	}
}

// A drive's size reads as the server and the file browser's window write it: 1024 to a KB.
func TestSizeCountsLikeTheServer(t *testing.T) {
	for b, want := range map[uint64]string{
		512:               "512 bytes",
		2048:              "2 KB",
		5 << 20:           "5 MB",
		548_682_072_064:   "511 GB",
		2_000_398_934_016: "1.8 TB",
	} {
		if got := size(b); got != want {
			t.Errorf("size(%d) = %q, want %q", b, got, want)
		}
	}
}
