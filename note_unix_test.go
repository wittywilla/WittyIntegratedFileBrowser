//go:build !windows

package wittyintegratedfilebrowser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// inContainer makes InContainer say yes, or no, for one test.
func inContainer(t *testing.T, yes bool) {
	was := InContainer
	InContainer = func() bool { return yes }
	t.Cleanup(func() { InContainer = was })
}

// locked is a folder nobody but root can open, with a folder inside it.
func locked(t *testing.T) (dir, inner string) {
	if os.Geteuid() == 0 {
		t.Skip("root opens any folder; run as another user (the Docker check does)")
	}
	dir = filepath.Join(t.TempDir(), "locked")
	inner = filepath.Join(dir, "inner")
	if err := os.MkdirAll(inner, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(dir, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(dir, 0o755) })
	return dir, inner
}

func TestExplainDenied(t *testing.T) {
	dir, inner := locked(t)
	inContainer(t, false)
	if n := Explain(dir); n != nil {
		t.Fatalf("outside a container: %+v, want nothing", n)
	}
	inContainer(t, true)
	n := Explain(dir)
	if n == nil || n.Why != "denied" || n.Folder != dir || n.UID != os.Getuid() || n.GID != os.Getgid() || n.OwnerUID != os.Getuid() {
		t.Fatalf("a folder that can't be opened: %+v", n)
	}
	// One inside it can't even be seen: the folder in the way is the one above.
	if _, err := List(inner, Options{}); err != ErrDenied {
		t.Errorf("List under a locked folder: %v, want ErrDenied (it is there, just not reachable)", err)
	}
	n = Explain(inner)
	if n == nil || n.Why != "denied" || n.Folder != dir {
		t.Fatalf("a folder under a locked one: %+v, want the locked one named", n)
	}
}

func TestExplainEmptyMount(t *testing.T) {
	root := t.TempDir()
	mount := filepath.Join(root, "nas media")
	full := filepath.Join(root, "full")
	plain := filepath.Join(root, "plain")
	for _, d := range []string{mount, full, plain} {
		if err := os.Mkdir(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(full, "film.mkv"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	table := filepath.Join(root, "mountinfo")
	esc := func(p string) string { return strings.ReplaceAll(p, " ", `\040`) }
	lines := "22 1 0:21 / / rw - overlay overlay rw\n" +
		"40 22 8:1 /srv/media " + esc(mount) + " rw,relatime - ext4 /dev/sda1 rw\n" +
		"41 22 8:1 /srv/full " + esc(full) + " rw,relatime - ext4 /dev/sda1 rw\n"
	if err := os.WriteFile(table, []byte(lines), 0o644); err != nil {
		t.Fatal(err)
	}
	was := mountTable
	mountTable = table
	t.Cleanup(func() { mountTable = was })

	inContainer(t, false)
	if n := Explain(mount); n != nil {
		t.Errorf("outside a container: %+v, want nothing", n)
	}
	inContainer(t, true)
	if n := Explain(mount); n == nil || n.Why != "empty-mount" || n.Folder != mount || n.OwnerUID != os.Getuid() {
		t.Errorf("an empty mount point: %+v", n)
	}
	if n := Explain(full); n != nil {
		t.Errorf("a mount point with something in it: %+v, want nothing", n)
	}
	if n := Explain(plain); n != nil {
		t.Errorf("an empty folder that isn't a mount point: %+v, want nothing", n)
	}
	if n := Explain(filepath.Join(root, "missing")); n != nil {
		t.Errorf("a folder that isn't there: %+v, want nothing", n)
	}
	if n := Explain("relative"); n != nil {
		t.Errorf("a relative path: %+v, want nothing", n)
	}
}

// In a container run as one user, with a folder made by another: the check in
// docs/TESTING.md sets these.
func TestExplainInDocker(t *testing.T) {
	foreign, empty := os.Getenv("WIFB_TEST_FOREIGN"), os.Getenv("WIFB_TEST_EMPTY_MOUNT")
	if foreign == "" && empty == "" {
		t.Skip("set WIFB_TEST_FOREIGN and WIFB_TEST_EMPTY_MOUNT in a container (docs/TESTING.md)")
	}
	if !InContainer() {
		t.Fatal("InContainer says no inside a container")
	}
	if foreign != "" {
		if _, err := List(foreign, Options{}); err != ErrDenied {
			t.Errorf("List of another user's folder: %v, want ErrDenied", err)
		}
		n := Explain(foreign)
		if n == nil || n.Why != "denied" || n.UID != os.Getuid() || n.OwnerUID == os.Getuid() {
			t.Errorf("another user's folder: %+v", n)
		}
		t.Logf("denied: runs as %d:%d, folder belongs to %d:%d", n.UID, n.GID, n.OwnerUID, n.OwnerGID)
		if n, err := List(filepath.Join(foreign, "inner"), Options{}); err != ErrDenied {
			t.Errorf("List under another user's folder: %v %v, want ErrDenied", n, err)
		}
	}
	if empty != "" {
		n := Explain(empty)
		if n == nil || n.Why != "empty-mount" {
			t.Errorf("an empty mount: %+v", n)
		}
	}
}

func TestUnescapeMount(t *testing.T) {
	for in, want := range map[string]string{`/mnt/a`: "/mnt/a", `/mnt/my\040films`: "/mnt/my films", `/mnt/tab\011x`: "/mnt/tab\tx", `/mnt/x\04`: `/mnt/x\04`, `/mnt/back\134slash`: `/mnt/back\slash`} {
		if got := unescapeMount(in); got != want {
			t.Errorf("unescapeMount(%q) = %q, want %q", in, got, want)
		}
	}
}
