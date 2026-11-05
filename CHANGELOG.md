# Changelog

What changed in WittyIntegratedFileBrowser, newest first. What changed before it had a repository of its own is in the
[WittyMediaServer changelog](https://github.com/wittywilla/WittyMediaServer/blob/master/CHANGELOG.md).

## Unreleased

- In a folder with more than 3,000 things, folders are no longer left out: they come first, and files fill what's left (#9).
- In a container (Docker and the like), a folder the server can't open, or a mount point that looks empty, says which user and group the server runs as, who owns the folder, and links to the fix for each kind of NAS. The helper answers a new `access` call for hosts to ask (#2).
- Enter and Space press the folder window's buttons and links, and the sort menu works with the keyboard; the arrows, Enter and Space work the list only while it has focus (#3).

## 0.2.0

- Module API 3 and WittyMediaServer 0.6.0-alpha: this is how a server gets it now, with no copy compiled in.
- Knows many more file types when it runs on its own, and playlists have a name and an icon.
- WittyIntegratedFileBrowser is now a project of its own, developed and released from this repository.
