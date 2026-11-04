# Changelog

What changed in WittyIntegratedFileBrowser, newest first. What changed before it had a repository of its own is in the
[WittyMediaServer changelog](https://github.com/wittywilla/WittyMediaServer/blob/master/CHANGELOG.md).

## Unreleased

- **Connect to a share:** when the host can connect to SMB and NFS shares itself, the window asks for the address (and for SMB a user name and password), and keeps them under Network shares. An address holding a password is refused, and the window never keeps one (#16).
- Enter and Space press the folder window's buttons and links, and the sort menu works with the keyboard; the arrows, Enter and Space work the list only while it has focus (#3).

## 0.2.0

- Module API 3 and WittyMediaServer 0.6.0-alpha: this is how a server gets it now, with no copy compiled in.
- Knows many more file types when it runs on its own, and playlists have a name and an icon.
- WittyIntegratedFileBrowser is now a project of its own, developed and released from this repository.
