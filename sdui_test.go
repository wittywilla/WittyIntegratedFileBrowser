// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 WittyWilla and contributors

package wittyintegratedfilebrowser

import (
	"encoding/json"
	"testing"
)

func TestPlaceItems(t *testing.T) {
	places := []Place{
		{Name: "Media", Path: "/mnt/media", Group: "drives", Free: 1_500_000_000_000, Total: 4_000_000_000_000},
		{Name: "Movies", Path: "/home/me/Movies", Group: "folders"},
		{Name: "NAS", Path: "/mnt/nas", Group: "drives", Label: "Network"},
	}
	req := func(admin bool, offset, limit int) ItemsRequest {
		r, err := ParseItemsRequest(json.RawMessage(`{"source":{"type":"wittyintegratedfilebrowser.places"},"offset":` + itoa(offset) + `,"limit":` + itoa(limit) + `,"viewer":{"admin":` + map[bool]string{true: "true", false: "false"}[admin] + `}}`))
		if err != nil {
			t.Fatal(err)
		}
		return r
	}
	a, err := PlaceItems(req(true, 0, 2), places)
	if err != nil || len(a.Items) != 2 || !a.More || a.Total != 3 {
		t.Fatalf("first page: %+v %v", a, err)
	}
	if a.Items[0].Title != "Media" || a.Items[0].Subtitle != "1.5 TB free of 4.0 TB" || a.Items[0].Icon != "devices" || a.Items[0].Kind != "library" {
		t.Fatalf("a drive: %+v", a.Items[0])
	}
	if a.Items[1].Subtitle != "/home/me/Movies" || a.Items[1].Icon != "folder" {
		t.Fatalf("a folder: %+v", a.Items[1])
	}
	b, _ := PlaceItems(req(true, 2, 2), places)
	if len(b.Items) != 1 || b.More || b.Items[0].Subtitle != "Network" {
		t.Fatalf("second page: %+v", b)
	}
	if b.Items[0].ID == a.Items[0].ID || len(b.Items[0].ID) > 200 {
		t.Fatalf("ids must be unique and short: %q", b.Items[0].ID)
	}
	if n, _ := PlaceItems(req(false, 0, 10), places); len(n.Items) != 0 {
		t.Fatalf("someone who isn't an administrator sees folders: %+v", n)
	}
	if _, err := PlaceItems(ItemsRequest{Source: map[string]any{"type": "other"}}, places); err == nil {
		t.Fatal("an unknown source was answered")
	}
}

func itoa(n int) string { b, _ := json.Marshal(n); return string(b) }
