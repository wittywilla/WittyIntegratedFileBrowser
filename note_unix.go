// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 WittyWilla and contributors

//go:build !windows

package wittyintegratedfilebrowser

import (
	"bufio"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

func explain(path string) *Note {
	if !InContainer() {
		return nil
	}
	p := filepath.Clean(strings.TrimSpace(path))
	if !filepath.IsAbs(p) {
		return nil
	}
	// The effective IDs, which are the ones the kernel checks.
	n := &Note{Folder: p, UID: os.Geteuid(), GID: os.Getegid()}
	if gs, err := os.Getgroups(); err == nil {
		for _, g := range gs {
			if g != n.GID {
				n.Groups = append(n.Groups, g)
			}
		}
	}
	st, err := os.Stat(p)
	switch {
	case errors.Is(err, fs.ErrPermission):
		// A folder above it can't be gone through: the nearest one that can be
		// seen is that folder.
		n.Why = "denied"
		for d := filepath.Dir(p); ; d = filepath.Dir(d) {
			if st, err = os.Stat(d); err == nil {
				n.Folder = d
				break
			}
			if d == filepath.Dir(d) {
				return nil
			}
		}
	case err != nil || !st.IsDir():
		return nil
	default:
		f, err := os.Open(p)
		if err != nil {
			if !errors.Is(err, fs.ErrPermission) {
				return nil
			}
			n.Why = "denied"
			break
		}
		_, err = f.Readdirnames(1)
		f.Close()
		switch {
		case errors.Is(err, fs.ErrPermission):
			n.Why = "denied"
		case errors.Is(err, io.EOF) && isMount(p, st):
			n.Why = "empty-mount"
		default:
			return nil
		}
	}
	sys, ok := st.Sys().(*syscall.Stat_t)
	if !ok {
		return nil
	}
	n.OwnerUID, n.OwnerGID = int(sys.Uid), int(sys.Gid)
	return n
}

// isMount says whether a folder is where something is mounted that could hold
// media: on another device than its parent, or named in the mount table (a
// folder bound from the same disk as the one above it is only found there).
// Memory and system file systems (an empty tmpfs at /tmp, /run, /dev/shm) are
// empty by design, so they don't count.
func isMount(p string, st fs.FileInfo) bool {
	parent := filepath.Dir(p)
	if parent == p {
		return false
	}
	found, fstype := inMountTable(p)
	if found {
		return !systemFS[fstype]
	}
	if pst, err := os.Stat(parent); err == nil {
		a, aok := st.Sys().(*syscall.Stat_t)
		b, bok := pst.Sys().(*syscall.Stat_t)
		return aok && bok && a.Dev != b.Dev
	}
	return false
}

// systemFS are file systems that hold no one's media.
var systemFS = map[string]bool{"tmpfs": true, "ramfs": true, "proc": true, "sysfs": true, "devtmpfs": true, "devpts": true, "mqueue": true, "cgroup": true, "cgroup2": true, "securityfs": true, "debugfs": true, "tracefs": true, "pstore": true, "bpf": true, "configfs": true, "fusectl": true, "hugetlbfs": true}

// mountTable is Linux's list of this process's mounts; other systems have none
// to read, and the device check is all there is.
var mountTable = "/proc/self/mountinfo"

// inMountTable says whether p is a mount point in the table, and what file
// system is mounted there. The table names folders with links resolved.
func inMountTable(p string) (found bool, fstype string) {
	if r, err := filepath.EvalSymlinks(p); err == nil {
		p = r
	}
	f, err := os.Open(mountTable)
	if err != nil {
		return false, ""
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 4096), 1<<20)
	for sc.Scan() {
		// id parent major:minor root mount-point options [optional...] - fstype source options
		fields := strings.Fields(sc.Text())
		if len(fields) < 5 || unescapeMount(fields[4]) != p {
			continue
		}
		found, fstype = true, ""
		for i, x := range fields {
			if x == "-" && i+1 < len(fields) {
				fstype = fields[i+1]
				break
			}
		}
		// The last line for a folder is what is mounted on top, and is what shows.
	}
	return found, fstype
}

// unescapeMount undoes the table's octal escapes (\040 for a space).
func unescapeMount(s string) string {
	if !strings.Contains(s, `\`) {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+3 < len(s) {
			if v, err := strconv.ParseUint(s[i+1:i+4], 8, 8); err == nil {
				b.WriteByte(byte(v))
				i += 3
				continue
			}
		}
		b.WriteByte(s[i])
	}
	return b.String()
}
