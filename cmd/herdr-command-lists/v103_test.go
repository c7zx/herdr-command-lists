package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func listNames(lists []commandList) []string {
	names := make([]string, len(lists))
	for i, list := range lists {
		names[i] = list.name
	}
	return names
}

func TestTabOrderCustomAndAdditionalLists(t *testing.T) {
	configDir := t.TempDir()
	listDir := filepath.Join(configDir, "lists")
	if err := os.Mkdir(listDir, 0o700); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"Firewall", "System", "SSH", "Alpha", "Zoo"} {
		mustWrite(t, filepath.Join(listDir, name+".md"), "whoami\n")
	}
	if err := ensureOrderPath(configDir); err != nil {
		t.Fatal(err)
	}
	assertMode(t, filepath.Join(configDir, orderFile), 0o600)
	got, err := loadLists(listDir)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"System", "SSH", "Firewall", "Alpha", "Zoo"}; !reflect.DeepEqual(listNames(got), want) {
		t.Fatalf("default order = %v, want %v", listNames(got), want)
	}
	order := "# My preferred tabs\nFirewall.md\nZoo.md\nFirewall.md\nNotPresent.md\n"
	mustWrite(t, filepath.Join(configDir, orderFile), order)
	got, err = loadLists(listDir)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"Firewall", "Zoo", "Alpha", "SSH", "System"}; !reflect.DeepEqual(listNames(got), want) {
		t.Fatalf("custom order = %v, want %v", listNames(got), want)
	}
	// Updating in place must not reset the user's manually edited order.
	t.Setenv("HERDR_PLUGIN_CONFIG_DIR", configDir)
	if _, err := ensureListsDir(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(configDir, orderFile))
	if err != nil || string(data) != order {
		t.Fatalf("order file overwritten on reinstall: %q (%v)", data, err)
	}
}

func TestMoveTabPersistsAndPreservesListContent(t *testing.T) {
	configDir := t.TempDir()
	listDir := filepath.Join(configDir, "lists")
	if err := os.Mkdir(listDir, 0o700); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"System", "SSH", "Firewall"} {
		mustWrite(t, filepath.Join(listDir, name+".md"), "# Heading\nwhoami\n")
	}
	if err := ensureOrderPath(configDir); err != nil {
		t.Fatal(err)
	}
	lists, err := loadLists(listDir)
	if err != nil {
		t.Fatal(err)
	}
	model := &uiModel{lists: lists, listsDir: listDir, activeTab: 2, selected: -1}
	model.entries = lists[model.activeTab].entries
	model.refresh(false)
	if err := model.moveTab(-1); err != nil {
		t.Fatal(err)
	}
	if err := model.moveTab(-1); err != nil {
		t.Fatal(err)
	}
	if err := model.moveTab(-1); err != nil {
		t.Fatal(err)
	}
	if model.activeTab != 0 {
		t.Fatalf("active tab = %d, want 0", model.activeTab)
	}
	if want := []string{"Firewall", "System", "SSH"}; !reflect.DeepEqual(listNames(model.lists), want) {
		t.Fatalf("reordered tabs = %v, want %v", listNames(model.lists), want)
	}
	reloaded, err := loadLists(listDir)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(listNames(reloaded), listNames(model.lists)) {
		t.Fatalf("order not persisted: %v", listNames(reloaded))
	}
	if err := model.moveTab(1); err != nil {
		t.Fatal(err)
	}
	if want := []string{"System", "Firewall", "SSH"}; !reflect.DeepEqual(listNames(model.lists), want) {
		t.Fatalf("move right = %v, want %v", listNames(model.lists), want)
	}
	for _, list := range model.lists {
		data, err := os.ReadFile(list.path)
		if err != nil || string(data) != "# Heading\nwhoami\n" {
			t.Fatalf("command list modified by sort: %s, %q, %v", list.name, data, err)
		}
	}
}

func TestTabOrderRejectsInvalidAndSymlink(t *testing.T) {
	for _, content := range []string{"../etc/passwd.md\n", "bad\x1b[31m.md\n", "not-a-list.txt\n"} {
		t.Run(content, func(t *testing.T) {
			dir := t.TempDir()
			mustWrite(t, filepath.Join(dir, orderFile), content)
			if _, err := loadTabOrder(dir); err == nil {
				t.Fatalf("accepted unsafe order: %q", content)
			}
		})
	}
	dir := t.TempDir()
	if err := os.Symlink(filepath.Join(t.TempDir(), "somefile"), filepath.Join(dir, orderFile)); err != nil {
		t.Fatal(err)
	}
	if err := ensureOrderPath(dir); err == nil {
		t.Fatal("symlinked order file accepted")
	}
}

func TestHighlightAlwaysWhiteBackgroundBlackText(t *testing.T) {
	m := &uiModel{selected: -1}
	m.entries = []listEntry{{kind: entryHeader, text: "# Heading"}, {kind: entryCommand, text: "    echo hi"}}
	m.lists = []commandList{{name: "One", entries: m.entries}, {name: "Two"}}
	m.refresh(false)
	m.down()
	screen := render(m, 12, 80)
	const whiteBlack = "\x1b[48;2;255;255;255;38;2;0;0;0m"
	if n := strings.Count(screen, whiteBlack); n != 2 {
		t.Fatalf("white/black selection style used %d times, want 2: %q", n, screen)
	}
	if strings.Contains(screen, "\x1b[7m") || strings.Contains(screen, "\x1b[1;7m") {
		t.Fatal("theme-dependent reverse video still used for selection")
	}
	if !strings.Contains(screen, whiteBlack+"›     echo hi\x1b[0m") {
		t.Fatal("selected indented command changes color or loses highlight")
	}
}

func TestKeyReaderAllowsTabMovesAndIgnoresBracketedPaste(t *testing.T) {
	input := []byte("\x1b[A\x1b[C\x10\x0e\x1b[1;5D\x1b[1;5C\x1b[200~date\n\x1b[201~\x1b[B")
	pos := 0
	k := &keyReader{nextByte: func() (byte, bool, error) {
		if pos >= len(input) {
			return 0, false, nil
		}
		b := input[pos]
		pos++
		return b, true, nil
	}}
	var events []string
	for pos < len(input) {
		e, _, err := k.next()
		if err != nil {
			t.Fatal(err)
		}
		if e != "" {
			events = append(events, e)
		}
	}
	want := []string{"up", "right", "tab-left", "tab-right", "tab-left", "tab-right", "down"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("key events %v, want %v", events, want)
	}
}

func TestShortcutConfigAddsAndRemovesOnlyOwnKey(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	const existing = "onboarding = false\n\n[[keys.command]]\nkey = \"prefix+j\"\ntype = \"action\"\ncommand = \"other.toggle\"\n"
	mustWrite(t, path, existing)
	status, err := configureShortcut("install", path)
	if err != nil || !strings.Contains(status, "added") {
		t.Fatalf("install shortcut = %q, %v", status, err)
	}
	if status, err := configureShortcut("install", path); err != nil || !strings.Contains(status, "already") {
		t.Fatalf("idempotency = %q, %v", status, err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(got), "herdr.command-lists.toggle") != 1 {
		t.Fatalf("duplicate command bindings: %q", got)
	}
	status, err = configureShortcut("remove", path)
	if err != nil || !strings.Contains(status, "removed") {
		t.Fatalf("remove shortcut = %q, %v", status, err)
	}
	got, err = os.ReadFile(path)
	if err != nil || !strings.Contains(string(got), "other.toggle") || strings.Contains(string(got), "herdr.command-lists.toggle") || strings.Contains(string(got), "# Command Lists shortcut.") {
		t.Fatalf("removed another shortcut, or left own binding: %q, %v", got, err)
	}
}
