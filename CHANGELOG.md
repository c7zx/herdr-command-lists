# Changelog

## 1.0.3 — 2026-11-21

Changes from **1.0.2**:

### Added

- Three practical starter lists: `System`, `SSH`, and `Firewall`.
- Persistent tab ordering via `tab-order.txt`; move the active tab with `Ctrl+P` / `Ctrl+N` or edit the order file directly.
- A fixed white background with black text for the selected command and active tab, independent of the terminal color theme.

### Changed

- Refined the popup layout with clearer tabs, section headings, search styling, separators, and selection marker.
- Unlisted and newly added lists remain visible after explicitly ordered tabs, sorted alphabetically.
- Updated the bundled help, installer output, and README for the new lists, tab ordering, and controls.
- Preserve existing command lists and custom tab order during reinstall.

## 1.0.2 — 2026-10-05

Changes from **1.0.1**:

### Added

- Multiple named command lists stored as `.md` files in a persistent `lists/` directory.
- Tab bar with `Left` / `Right` switching and wraparound.
- Three starter lists: `Main`, `Systems`, and `More`.
- Automatic reload of lists whenever the popup opens.
- Safe uninstall script that removes the plugin while preserving personal command lists.
- Bounded list loading and bracketed-paste protection.

### Changed

- Plugin name and ID changed from **Command List** / `herdr.command-list` to **Command Lists** / `herdr.command-lists`.
- `Up` / `Down` navigation now wraps and scrolls naturally instead of keeping the selection pinned to an edge.
- Running a selected command first clears unfinished shell input with `Ctrl+E` and `Ctrl+U`.
- Existing `.md` lists are preserved during reinstall; starter lists are created only when the list directory is missing or empty.
- Help and examples were updated for the multi-list interface.
