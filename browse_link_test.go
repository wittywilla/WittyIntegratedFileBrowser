package wittyintegratedfilebrowser

import (
	"os"
	"path/filepath"
	"testing"
)

// A link is listed as what it points to: a folder that opens, a file with its
// own size. A link that points nowhere stays a file.
func TestListFollowsLinks(t *testing.T) {
	root := t.TempDir()
	films := filepath.Join(root, "films")
	if err := os.MkdirAll(filepath.Join(films, "VIDEO_TS"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "film.mkv"), make([]byte, 1234), 0o644); err != nil {
		t.Fatal(err)
	}
	for link, to := range map[string]string{"Films-link": films, "film-link.mkv": filepath.Join(root, "film.mkv"), "gone": filepath.Join(root, "nowhere")} {
		if err := os.Symlink(to, filepath.Join(root, link)); err != nil {
			t.Skipf("can't make links here: %v", err) // Windows without developer mode
		}
	}
	l, err := List(root, Options{})
	if err != nil {
		t.Fatal(err)
	}
	by := map[string]Entry{}
	for _, e := range l.Entries {
		by[e.Name] = e
	}
	if e := by["Films-link"]; !e.Dir || e.Disc != "dvd" {
		t.Errorf("a link to a folder: %+v, want a folder holding a DVD", e)
	}
	if e := by["film-link.mkv"]; e.Dir || e.Size != 1234 || e.Kind != "video" {
		t.Errorf("a link to a file: %+v, want the file's size and kind", e)
	}
	if e := by["gone"]; e.Dir || e.Name == "" {
		t.Errorf("a link to nothing: %+v, want a file", e)
	}
	if l.Counts["folder"] != 2 {
		t.Errorf("counts = %v, want 2 folders", l.Counts)
	}
	if l.Entries[0].Name != "films" || l.Entries[1].Name != "Films-link" {
		t.Errorf("folders first: %v, %v", l.Entries[0].Name, l.Entries[1].Name)
	}
}
