// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 WittyWilla and contributors

package wittyintegratedfilebrowser

import (
	"encoding/json"
	"testing"

	"github.com/wittywilla/wittymodule/sdui"
)

// A page of places is asked for with any offset and limit a caller sends: none may crash the
// helper, and a page never holds more than was asked for or than there is.
func FuzzItemsPaging(f *testing.F) {
	f.Add(`{"source":{"type":"wittyintegratedfilebrowser.places"},"offset":1,"limit":9223372036854775807,"viewer":{"admin":true}}`)
	f.Add(`{"source":{"type":"wittyintegratedfilebrowser.places"},"offset":-5,"limit":-1,"viewer":{"admin":true}}`)
	f.Add(`{"source":{"type":"wittyintegratedfilebrowser.folder","path":"../.."},"offset":0,"limit":2}`)
	places := func() []Place {
		return []Place{{Name: "Media", Path: "/mnt/media", Group: "drives", Free: 1, Total: 2}, {Name: "Movies", Path: "/home/me/Movies", Group: "folders"}, {Name: "NAS", Path: "/mnt/nas", Group: "drives"}}
	}
	f.Fuzz(func(t *testing.T, raw string) {
		var r sdui.ItemsRequest
		if json.Unmarshal([]byte(raw), &r) != nil {
			return
		}
		a, err := Items(r, places)
		if err != nil {
			return
		}
		if len(a.Items) > a.Total || (r.Limit > 0 && len(a.Items) > r.Limit) {
			t.Errorf("offset %d limit %d: %d items of %d", r.Offset, r.Limit, len(a.Items), a.Total)
		}
	})
}
