package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Keep the installer's existing shortcut-config interface. Only the plugin's
// own [[keys.command]] block is removed; no other user binding is altered.
const shortcutBlock = `# Command Lists shortcut. Change "ctrl+y" to your preferred Herdr key.
[[keys.command]]
key = "ctrl+y"
type = "plugin_action"
command = "herdr.command-lists.toggle"
description = "Command Lists"
`

type keyBlock struct {
	start, end   int
	key, command string
}

var (
	keyProperty     = regexp.MustCompile(`(?m)^\s*key\s*=\s*"([^"]*)"`)
	commandProperty = regexp.MustCompile(`(?m)^\s*command\s*=\s*"([^"]*)"`)
)

func keyBlocks(content string) []keyBlock {
	var result []keyBlock
	start := -1
	pos := 0
	for _, line := range strings.SplitAfter(content, "\n") {
		trimmed := strings.TrimSpace(line)
		isTable := strings.HasPrefix(trimmed, "[")
		if isTable && start >= 0 {
			block := content[start:pos]
			rec := keyBlock{start: start, end: pos}
			if m := keyProperty.FindStringSubmatch(block); len(m) > 1 {
				rec.key = m[1]
			}
			if m := commandProperty.FindStringSubmatch(block); len(m) > 1 {
				rec.command = m[1]
			}
			result = append(result, rec)
			start = -1
		}
		if trimmed == "[[keys.command]]" {
			start = pos
		}
		pos += len(line)
	}
	if start >= 0 {
		block := content[start:]
		rec := keyBlock{start: start, end: len(content)}
		if m := keyProperty.FindStringSubmatch(block); len(m) > 1 {
			rec.key = m[1]
		}
		if m := commandProperty.FindStringSubmatch(block); len(m) > 1 {
			rec.command = m[1]
		}
		result = append(result, rec)
	}
	return result
}

func configureShortcut(action, path string) (string, error) {
	if action != "install" && action != "remove" {
		return "", errors.New("shortcut-config requires install or remove")
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		if action == "remove" {
			return "shortcut not present", nil
		}
		data = nil
	} else if err != nil {
		return "", err
	} else if err := checkPath(path, false); err != nil {
		return "", err
	}
	content := string(data)
	blocks := keyBlocks(content)
	if action == "install" {
		for _, block := range blocks {
			if block.command == "herdr.command-lists.toggle" {
				return "shortcut already configured", nil
			}
		}
		for _, block := range blocks {
			if strings.EqualFold(block.key, "ctrl+y") {
				return "Ctrl+Y is already in use; shortcut not added", nil
			}
		}
		if len(content) > 0 && !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		if content != "" {
			content += "\n"
		}
		content += shortcutBlock
		if err := storeConfig(path, []byte(content), data); err != nil {
			return "", err
		}
		return "Ctrl+Y shortcut added", nil
	}
	changed := false
	for i := len(blocks) - 1; i >= 0; i-- {
		block := blocks[i]
		if block.command != "herdr.command-lists.toggle" {
			continue
		}
		start := block.start
		ownComment := strings.SplitN(shortcutBlock, "\n", 2)[0] + "\n"
		if strings.HasSuffix(content[:start], ownComment) {
			start -= len(ownComment)
		}
		content = content[:start] + content[block.end:]
		changed = true
	}
	if !changed {
		return "shortcut not present", nil
	}
	if err := storeConfig(path, []byte(content), data); err != nil {
		return "", err
	}
	return "shortcut removed", nil
}

func storeConfig(path string, newData, oldData []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	if len(oldData) > 0 {
		// Keep a first-edit backup without overwriting an older backup.
		backup := path + ".command-lists.bak"
		if _, err := os.Lstat(backup); errors.Is(err, fs.ErrNotExist) {
			if err := os.WriteFile(backup, oldData, 0o600); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
	}
	f, err := os.CreateTemp(dir, ".command-lists-config-")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if oldStat, err := os.Stat(path); err == nil {
		if err := f.Chmod(oldStat.Mode().Perm()); err != nil {
			f.Close()
			return err
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		f.Close()
		return err
	}
	if _, err := f.Write(newData); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := os.Rename(f.Name(), path); err != nil {
		return fmt.Errorf("write Herdr config: %w", err)
	}
	return nil
}
