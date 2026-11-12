// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 WittyWilla and contributors

package wittyintegratedfilebrowser

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// The module's data source (sdui.json, "sources"): the server's drives and usual folders, for
// a page an administrator builds. The server asks with the helper method sdui.items
// (WittyMediaServer docs/SDUI-MODULES.md, "How a source is answered").

// PlacesSource is the source's type.
const PlacesSource = "wittyintegratedfilebrowser.places"

// ItemsRequest is the part of the server's sdui.items call this module reads.
type ItemsRequest struct {
	Source map[string]any `json:"source"`
	Offset int            `json:"offset"`
	Limit  int            `json:"limit"`
	Viewer struct {
		Admin bool `json:"admin"`
	} `json:"viewer"`
}

// Item is one card, in the server's SDUI item shape.
type Item struct {
	ID       string `json:"id"`
	Kind     string `json:"kind"`
	Title    string `json:"title"`
	Subtitle string `json:"subtitle,omitempty"`
	Icon     string `json:"icon,omitempty"`
}

// ItemsAnswer is the answer to sdui.items.
type ItemsAnswer struct {
	Items []Item `json:"items"`
	More  bool   `json:"more"`
	Total int    `json:"total"`
}

// ParseItemsRequest reads an sdui.items call's parameters.
func ParseItemsRequest(raw json.RawMessage) (ItemsRequest, error) {
	var r ItemsRequest
	if err := json.Unmarshal(raw, &r); err != nil {
		return r, err
	}
	return r, nil
}

// PlaceItems answers the places source. Where folders are on the server is for its
// administrators only, so anyone else is answered with nothing.
func PlaceItems(r ItemsRequest, places []Place) (ItemsAnswer, error) {
	if t, _ := r.Source["type"].(string); t != PlacesSource {
		return ItemsAnswer{}, fmt.Errorf("this module has no source %q", t)
	}
	if !r.Viewer.Admin {
		return ItemsAnswer{Items: []Item{}}, nil
	}
	all := make([]Item, 0, len(places))
	for _, p := range places {
		sum := sha256.Sum256([]byte(p.Path))
		it := Item{ID: "place-" + hex.EncodeToString(sum[:8]), Kind: "library", Title: p.Name, Icon: "folder"}
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
	start := min(max(r.Offset, 0), len(all))
	end := len(all)
	if r.Limit > 0 && start+r.Limit < end {
		end = start + r.Limit
	}
	return ItemsAnswer{Items: all[start:end], More: end < len(all), Total: len(all)}, nil
}

func size(b uint64) string {
	const k = 1000
	switch {
	case b >= k*k*k*k:
		return fmt.Sprintf("%.1f TB", float64(b)/(k*k*k*k))
	case b >= k*k*k:
		return fmt.Sprintf("%.0f GB", float64(b)/(k*k*k))
	default:
		return fmt.Sprintf("%.0f MB", float64(b)/(k*k))
	}
}
