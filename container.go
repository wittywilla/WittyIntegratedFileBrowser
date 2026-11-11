// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 WittyWilla and contributors

package wittyintegratedfilebrowser

import "os"

// InContainer says whether this runs in a container (Docker, Podman and the
// like), where a folder the server can't read is most often a user and group
// mapping that doesn't match the folder's owner. The host may say so
// (WITTYMODULE_CONTAINER, or WMS_DOCKER from WittyMediaServer's image); a helper
// isn't given those today, so the files container engines leave at the root
// tell it instead. A variable, so tests can say either.
var InContainer = func() bool {
	if os.Getenv("WITTYMODULE_CONTAINER") != "" || os.Getenv("WMS_DOCKER") == "1" {
		return true
	}
	return exists("/.dockerenv") || exists("/run/.containerenv")
}
