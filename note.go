// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 WittyWilla and contributors

package wittyintegratedfilebrowser

// Note says why a folder may not be readable when the server runs in a
// container: the user and group the server runs as, and who owns the folder in
// the way, so the two can be compared. It holds numbers only; the page words it
// and links to the fix.
type Note struct {
	// Why: "denied" when the folder can't be opened, "empty-mount" when it is a
	// mount point that looks empty to the server.
	Why string `json:"why"`
	// Folder is the one in the way: the folder asked for, or a folder above it
	// that the server isn't allowed to go through.
	Folder   string `json:"folder"`
	UID      int    `json:"uid"`
	GID      int    `json:"gid"`
	Groups   []int  `json:"groups,omitempty"` // the server's other groups
	OwnerUID int    `json:"ownerUid"`
	OwnerGID int    `json:"ownerGid"`
}

// Explain says why a folder can't be read, or nil when there is nothing worth
// saying: outside a container, for a folder that opens and has something in it,
// or one that isn't there. It reads at most one name from the folder, so it is
// quick on a large or slow one.
func Explain(path string) *Note { return explain(path) }
