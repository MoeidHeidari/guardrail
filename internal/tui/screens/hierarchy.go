package screens

import (
	"path/filepath"
	"sort"

	"github.com/MoeidHeidari/guardrail/internal/watcher"
)

func (m *Main) rebuildRows() {
	m.entryPositions = make(map[string]int, len(m.entries))
	m.treeChildren = make(map[string][]string)
	for i, entry := range m.entries {
		m.entryPositions[entry.Path] = i
		if entry.Path != "." {
			parent := filepath.Dir(entry.Path)
			m.treeChildren[parent] = append(m.treeChildren[parent], entry.Path)
		}
	}
	for parent := range m.treeChildren {
		m.sortChildren(parent)
	}
	m.flattenRows()
}

func (m *Main) sortChildren(parent string) {
	nodes := m.treeChildren[parent]
	sort.Slice(nodes, func(i, j int) bool {
		a, b := m.entries[m.entryPositions[nodes[i]]], m.entries[m.entryPositions[nodes[j]]]
		if a.IsDir() != b.IsDir() {
			return a.IsDir()
		}
		return nodes[i] < nodes[j]
	})
}

func (m *Main) updateHierarchy(update watcher.Update) {
	changed := make(map[string]bool, len(update.Removed)+len(update.Upsert))
	added := make(map[string][]string)
	parents := make(map[string]bool)
	for _, path := range update.Removed {
		changed[path] = true
		parents[filepath.Dir(path)] = true
		delete(m.treeChildren, path)
	}
	for _, entry := range update.Upsert {
		if entry.Path == "." {
			continue
		}
		path := entry.Path
		changed[path] = true
		parent := filepath.Dir(path)
		parents[parent] = true
		added[parent] = append(added[parent], path)
	}
	for parent := range parents {
		if _, exists := m.entryPositions[parent]; !exists {
			delete(m.treeChildren, parent)
			continue
		}
		children := m.treeChildren[parent]
		kept := children[:0]
		for _, path := range children {
			if !changed[path] {
				kept = append(kept, path)
			}
		}
		clear(children[len(kept):])
		m.treeChildren[parent] = append(kept, added[parent]...)
		m.sortChildren(parent)
	}
}

func (m *Main) flattenRows() {
	previous := m.rows
	m.rows = m.rows[:0]
	clear(m.rowPositions)
	if len(m.entries) == 0 {
		clear(previous)
		return
	}
	var visit func(string, int)
	visit = func(path string, depth int) {
		index, exists := m.entryPositions[path]
		if !exists {
			return
		}
		entry := m.entries[index]
		m.rows = append(m.rows, treeRow{entry: entry, depth: depth})
		if m.collapsed[entry.Path] {
			return
		}
		for _, child := range m.treeChildren[path] {
			visit(child, depth+1)
		}
	}
	visit(".", 0)
	clear(previous[min(len(previous), len(m.rows)):])
	if m.rowPositions == nil {
		m.rowPositions = make(map[string]int, len(m.rows))
	}
	for i, row := range m.rows {
		m.rowPositions[row.entry.Path] = i
	}
}
