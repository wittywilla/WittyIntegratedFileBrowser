// SPDX-License-Identifier: AGPL-3.0-or-later
//
// WittyIntegratedFileBrowser - Choosing a server's folders from any device. Read-only.
// Copyright (C) 2026 WittyWilla and contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

// WittyIntegratedFileBrowser: a folder browser in a window, for choosing
// folders on another computer (a media server) from any device. Read-only.
// The host lists folders; the server half is the Go module of the same name.
export { default } from './FileBrowser'
export type { Entry, FileBrowserProps, Listing, Note, Place } from './FileBrowser'
