// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 WittyWilla and contributors

// Package ui is WittyIntegratedFileBrowser's server-driven UI (sdui.json), written from Go so
// the file and the code answering its sources can't drift: `go test ./ui -update-sdui`
// writes it. Every app draws its page natively, a TV's included (tvOS has no web view, so the
// web part's folder window never reaches one).
package ui

import (
	"github.com/wittywilla/wittymodule/sdui"
)

// Module is the module's ID.
const Module = "wittyintegratedfilebrowser"

// PageFolders is the server's folders: its drives and usual folders, and, opened with
// ?name=<a folder's path>, what that folder holds.
const PageFolders = Module

// The module's data sources.
const (
	// SourcePlaces is the server's drives and usual folders, each opening its folder.
	SourcePlaces = Module + ".places"
	// SourceFolder is one folder's folders and files (path: the folder; empty: the places).
	SourceFolder = Module + ".folder"
)

// FolderRoute is the page of one folder, for a navigate action.
func FolderRoute(query string) string { return "/page/" + PageFolders + "?name=" + query }

func sources() []sdui.SourceSpec {
	return []sdui.SourceSpec{
		{Type: SourcePlaces, Label: "The server's drives and folders", Icon: "folder", Title: "Server folders",
			Description: "The drives and usual folders on the server, each with the space left, opening the folder. Answered for administrators only; empty for anyone else.",
			Yields:      []string{"library"}, CardStyle: "tile", Paged: true},
		{Type: SourceFolder, Label: "A folder on the server", Icon: "folder", Title: "",
			Description: "One folder on the server: the way up, then its folders, each opening its own page, then its files with their sizes. Read-only. With no path, the drives and usual folders. Answered for administrators only; empty for anyone else.",
			Params:      []sdui.PropSpec{{Name: "path", Type: "string", Label: "Folder", Help: "The folder's full path on the server, $name on the folder page. Empty: the drives and usual folders."}},
			Yields:      []string{"view"}, CardStyle: "tile", Paged: true},
	}
}

// UI is the module's sdui.json.
func UI() *sdui.Module {
	m := sdui.New(Module, "Folder browser")
	for _, s := range sources() {
		m.AddSource(s)
	}
	yes, no := true, false
	admins := &sdui.Condition{Admin: &yes}

	folder := sdui.Sec("folder", "List", sdui.Props("title", "", "focusable", true, "empty", "message",
		"emptyText", "Nothing here, or the server isn't allowed to open this folder."), sdui.Src(SourceFolder, "path", "$name"))
	folder.When = admins
	others := sdui.Notice("admins", "info", "Only an administrator can look through the server's folders.")
	others.When = &sdui.Condition{Admin: &no}
	page := sdui.Page{ID: PageFolders, Kind: "custom", Title: "Server folders", Icon: "folder", Scope: Module,
		Sections: []sdui.Section{folder, others,
			sdui.Notice("readonly", "info", "Read-only: nothing on the server can be changed from here. Add a folder to a library in Settings, Libraries.")}}
	page.Description = "The server's drives and usual folders, and what each folder holds, opened with ?name=<a folder's path>: folders open their own page, Up goes to the folder above. For administrators, on every app and TV. Read-only."
	m.AddPage(page)

	// Under More on every app, for administrators only.
	m.AddNav(sdui.NavItem{ID: PageFolders, Type: "page", Page: PageFolders, Label: "Server folders", Icon: "folder", When: admins,
		Placement: map[string]string{"phone": "more", "tablet": "more", "tv": "more", "desktop": "more", "web": "more"}})

	m.AddPreset(sdui.ModulePreset{ID: Module + "-places", Label: "Server drives and space",
		Description: "A row of the server's drives and usual folders, with the space left on each. Only administrators see it.",
		Section:     sdui.Sec("server-places", "Shelf", sdui.Props("title", "Server drives", "cardStyle", "tile", "seeAll", "never"), sdui.Src(SourcePlaces))})
	return m
}
