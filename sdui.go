// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 WittyWilla and contributors

package wittyintegratedfilebrowser

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/wittywilla/wittyintegratedfilebrowser/ui"
	"github.com/wittywilla/wittymodule/sdui"
	core "github.com/wittywilla/wittymodule/sdui/core"
)

// The module's data sources (sdui.json, written by package ui): the server's drives and usual
// folders, and one folder's contents, for the Server folders page every app draws, a TV's
// included. The server asks with the helper method sdui.items (WittyMediaServer
// docs/SDUI-MODULES.md, "How a source is answered"). Where folders are on the server is for its
// administrators only, so anyone else is answered with nothing.

// maxFolderItems is as many rows as a folder's page shows: enough to choose by, never a page a
// TV can't scroll.
const maxFolderItems = 500

// Answers answers the server's SDUI calls; places are where to start (PlacesFor).
func Answers(places func() []Place) sdui.Answers {
	return sdui.Answers{Items: func(r sdui.ItemsRequest) (sdui.ItemsAnswer, error) { return Items(r, places) }}
}

// Items answers one of the module's sources.
func Items(r sdui.ItemsRequest, places func() []Place) (sdui.ItemsAnswer, error) {
	t, _ := r.Source["type"].(string)
	switch t {
	case ui.SourcePlaces, ui.SourceFolder:
	default:
		return sdui.ItemsAnswer{}, fmt.Errorf("this module has no source %q", t)
	}
	if !r.Viewer.Admin {
		return sdui.ItemsAnswer{Items: []sdui.Item{}}, nil
	}
	path, _ := r.Source["path"].(string)
	path = strings.TrimSpace(path)
	if t == ui.SourcePlaces || path == "" || strings.HasPrefix(path, "$") {
		all := placeItems(places())
		// The folder source yields views (sdui.json): a server keeps only items of the kinds a
		// source says it yields, so its places are views too.
		if t == ui.SourceFolder {
			for i := range all {
				all[i].Kind = "view"
			}
		}
		return page(r, all), nil
	}
	return page(r, folderItems(path)), nil
}

func page(r sdui.ItemsRequest, all []sdui.Item) sdui.ItemsAnswer {
	start := min(max(r.Offset, 0), len(all))
	end := len(all)
	// Compared as what's left, not start+limit: a limit near the largest int would wrap round.
	if r.Limit > 0 && r.Limit < end-start {
		end = start + r.Limit
	}
	return sdui.ItemsAnswer{Items: all[start:end], More: end < len(all), Total: len(all)}
}

func id(prefix, path string) string {
	sum := sha256.Sum256([]byte(path))
	return prefix + hex.EncodeToString(sum[:8])
}

func open(path string) *sdui.Action {
	return &sdui.Action{Type: core.ActNavigate, To: ui.FolderRoute(url.QueryEscape(path))}
}

func placeItems(places []Place) []sdui.Item {
	all := make([]sdui.Item, 0, len(places))
	for _, p := range places {
		it := sdui.Item{ID: id("place-", p.Path), Kind: "library", Title: p.Name, Icon: "folder", Action: open(p.Path)}
		if p.Group == "drives" {
			it.Icon = "devices"
		}
		switch {
		case p.Total > 0:
			it.Subtitle = fmt.Sprintf("%s free of %s", size(p.Free), size(p.Total))
		case p.Label != "":
			it.Subtitle = p.Label
		default:
			it.Subtitle = p.Path
		}
		all = append(all, it)
	}
	return all
}

// networkPath is a path naming another machine (\\nas\share, //nas/share): opening one would
// have the server sign in there, so the page never does.
func networkPath(p string) bool {
	return strings.HasPrefix(p, `\\`) || strings.HasPrefix(p, "//")
}

// folderItems is one folder: where it is, the way up, its folders, then its files.
func folderItems(path string) []sdui.Item {
	note := func(text string) []sdui.Item {
		return []sdui.Item{{ID: "note", Kind: "view", Title: text, Icon: "info"}}
	}
	if networkPath(path) {
		return note("A folder on another machine can't be opened here. Use a drive or a folder mounted on the server.")
	}
	if !filepath.IsAbs(path) {
		return note("That isn't a folder's full path.")
	}
	l, err := List(path, Options{MaxEntries: maxFolderItems})
	if err != nil {
		return note("The server can't open this folder: " + err.Error() + ".")
	}
	here := sdui.Item{ID: "here", Kind: "view", Title: l.Path, Icon: "folder",
		Subtitle: fmt.Sprintf("%d folders, %d files", l.Counts["folder"], len(l.Entries)-l.Counts["folder"])}
	if l.Truncated {
		here.Subtitle += fmt.Sprintf(", the first %d shown", maxFolderItems)
	}
	out := []sdui.Item{here}
	if l.Parent != nil {
		out = append(out, sdui.Item{ID: "up", Kind: "view", Title: "Up", Subtitle: *l.Parent, Icon: "folder", Action: open(*l.Parent)})
	} else {
		out = append(out, sdui.Item{ID: "up", Kind: "view", Title: "All drives and folders", Icon: "home",
			Action: &sdui.Action{Type: core.ActNavigate, To: "/page/" + ui.PageFolders}})
	}
	for _, e := range l.Entries {
		it := sdui.Item{ID: id("entry-", e.Path), Kind: "view", Title: e.Name}
		if e.Dir {
			it.Icon, it.Action = "folder", open(e.Path)
			switch e.Disc {
			case "dvd":
				it.Subtitle = "A DVD"
			case "bluray":
				it.Subtitle = "A Blu-ray"
			}
		} else {
			it.Icon, it.Trailing = fileIcons[e.Kind], size(uint64(max(e.Size, 0)))
			if it.Icon == "" {
				it.Icon = "list"
			}
		}
		out = append(out, it)
	}
	return out
}

// fileIcons is the catalogue's icon for each kind of file FileKind names.
var fileIcons = map[string]string{"video": "video", "audio": "music", "book": "book", "disc": "disc", "image": "photo", "subtitle": "list"}

// size is a number of bytes as the server and the file browser's own window write it: 1024 to a
// KB, as the operating systems show a disk, so a drive reads the same everywhere.
func size(b uint64) string {
	const k = 1024
	switch {
	case b >= k*k*k*k:
		return fmt.Sprintf("%.1f TB", float64(b)/(k*k*k*k))
	case b >= k*k*k:
		return fmt.Sprintf("%.0f GB", float64(b)/(k*k*k))
	case b >= k*k:
		return fmt.Sprintf("%.0f MB", float64(b)/(k*k))
	case b >= k:
		return fmt.Sprintf("%.0f KB", float64(b)/k)
	default:
		return fmt.Sprintf("%d bytes", b)
	}
}
