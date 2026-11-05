# Changelog

What changed in WittyIntegratedFileBrowser, newest first. What changed before it had a repository of its own is in the
[WittyMediaServer changelog](https://github.com/wittywilla/WittyMediaServer/blob/master/CHANGELOG.md).

## Unreleased

- A folder reached through a link (or a Windows junction) is a folder that opens, and a link to a file has the file's size (#10).
- The folder window takes keyboard focus when it opens, keeps Tab inside it, and gives focus back when it closes (#8).
- On a long path, the path bar shows its end, the folder you're in, and scrolls back to the start (#6).
- Screen readers say which row is highlighted as the arrows move through the folder list, and what it is: a folder, a DVD or Blu-ray folder, or a file with its kind and size (#7).
- In a folder with more than 3,000 things, folders are no longer left out: they come first, and files fill what's left (#9).
- A folder that is slow to list no longer replaces the one opened after it (#4).
- In a container (Docker and the like), a folder the server can't open, or a mount point that looks empty, says which user and group the server runs as, who owns the folder, and links to the fix for each kind of NAS. The helper answers a new `access` call for hosts to ask (#2).
- Enter and Space press the folder window's buttons and links, and the sort menu works with the keyboard; the arrows, Enter and Space work the list only while it has focus (#3).

## 0.2.0

- Module API 3 and WittyMediaServer 0.6.0-alpha: this is how a server gets it now, with no copy compiled in.
- Knows many more file types when it runs on its own, and playlists have a name and an icon.
- WittyIntegratedFileBrowser is now a project of its own, developed and released from this repository.
