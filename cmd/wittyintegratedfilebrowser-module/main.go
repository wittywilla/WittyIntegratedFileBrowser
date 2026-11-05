// Command wittyintegratedfilebrowser-module is WittyIntegratedFileBrowser as a
// helper process, so a new version can reach a server without a new server. It
// only reads. See WittyModule for the protocol.
package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/wittywilla/wittyintegratedfilebrowser"
	"github.com/wittywilla/wittymodule"
)

func main() { wittymodule.Main(wittyintegratedfilebrowser.ManifestJSON, handle) }

func handle(method string, params json.RawMessage) (any, []byte, error) {
	var p struct {
		Path       string `json:"path"`
		ShowHidden bool   `json:"showHidden"`
		MaxEntries int    `json:"maxEntries"`
		Home       string `json:"home"` // a helper starts with almost no environment, so it is told
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, nil, &wittymodule.CallError{Code: "bad-call", Message: "the call is unreadable"}
	}
	switch method {
	case "places":
		return wittyintegratedfilebrowser.PlacesFor(p.Home), nil, nil
	case "list":
		if p.Path == "" {
			return nil, nil, &wittymodule.CallError{Code: "bad-call", Message: "no path was given"}
		}
		l, err := wittyintegratedfilebrowser.List(p.Path, wittyintegratedfilebrowser.Options{ShowHidden: p.ShowHidden, MaxEntries: p.MaxEntries})
		if err != nil {
			return nil, nil, coded(err)
		}
		// A blob, because a large folder is longer than one protocol line may be.
		b, err := json.Marshal(l)
		if err != nil {
			return nil, nil, err
		}
		return nil, b, nil
	case "access":
		// Why a folder can't be read, for a host to ask after "denied" or an empty
		// listing. null outside a container and when there's nothing to say.
		if p.Path == "" {
			return nil, nil, &wittymodule.CallError{Code: "bad-call", Message: "no path was given"}
		}
		return wittyintegratedfilebrowser.Explain(p.Path), nil, nil
	}
	return nil, nil, &wittymodule.CallError{Code: "unknown-method", Message: fmt.Sprintf("this version has no %q", method)}
}

// coded names the errors the server acts on.
func coded(err error) error {
	code := ""
	switch {
	case errors.Is(err, wittyintegratedfilebrowser.ErrNotFound):
		code = "not-found"
	case errors.Is(err, wittyintegratedfilebrowser.ErrDenied):
		code = "denied"
	}
	return &wittymodule.CallError{Code: code, Message: err.Error()}
}
