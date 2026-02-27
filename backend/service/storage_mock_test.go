package service

import (
	"fmt"
	"io/fs"
	"maps"
	"path/filepath"
	"strings"

	"github.com/tfabritius/plainpage/model"
)

// mockStorage implements model.Storage for testing
type mockStorage struct {
	files map[string][]byte
	dirs  map[string]bool
}

func newMockStorage() model.Storage {
	return &mockStorage{
		files: make(map[string][]byte),
		dirs:  make(map[string]bool),
	}
}

func (m *mockStorage) Exists(fsPath string) bool {
	if _, ok := m.files[fsPath]; ok {
		return true
	}
	if _, ok := m.dirs[fsPath]; ok {
		return true
	}
	return false
}

func (m *mockStorage) ReadFile(fsPath string) ([]byte, error) {
	if data, ok := m.files[fsPath]; ok {
		return data, nil
	}
	return nil, fmt.Errorf("could not read file %s", fsPath)
}

// createParentDirs creates all parent directories for a given path
func (m *mockStorage) createParentDirs(fsPath string) {
	dir := filepath.Dir(fsPath)
	for dir != "" && dir != "." {
		m.dirs[dir] = true
		dir = filepath.Dir(dir)
	}
}

// withPrefix returns the path with a trailing separator for prefix matching
func withPrefix(fsPath string) string {
	sep := string(filepath.Separator)
	if fsPath != "" && !strings.HasSuffix(fsPath, sep) {
		return fsPath + sep
	}
	return fsPath
}

func (m *mockStorage) WriteFile(fsPath string, data []byte) error {
	m.createParentDirs(fsPath)
	m.files[fsPath] = data
	return nil
}

func (m *mockStorage) DeleteFile(fsPath string) error {
	if _, ok := m.files[fsPath]; !ok {
		return fmt.Errorf("could not remove file %s", fsPath)
	}
	delete(m.files, fsPath)
	return nil
}

func (m *mockStorage) CreateDirectory(fsPath string) error {
	// Check if parent directory exists
	dir := filepath.Dir(fsPath)
	if dir != "" && dir != "." && !m.dirs[dir] {
		return fmt.Errorf("parent directory does not exist: %s", dir)
	}
	m.dirs[fsPath] = true
	return nil
}

func (m *mockStorage) ReadDirectory(fsPath string) ([]fs.DirEntry, error) {
	var entries []fs.DirEntry
	seen := make(map[string]bool)
	prefix := withPrefix(fsPath)
	sep := string(filepath.Separator)

	for path := range m.files {
		if strings.HasPrefix(path, prefix) {
			remainder := strings.TrimPrefix(path, prefix)
			parts := strings.SplitN(remainder, sep, 2)
			name := parts[0]
			if !seen[name] {
				seen[name] = true
				isDir := len(parts) > 1
				entries = append(entries, mockDirEntry{name: name, isDir: isDir})
			}
		}
	}

	for path := range m.dirs {
		if strings.HasPrefix(path, prefix) {
			remainder := strings.TrimPrefix(path, prefix)
			parts := strings.SplitN(remainder, sep, 2)
			name := parts[0]
			if name != "" && !seen[name] {
				seen[name] = true
				entries = append(entries, mockDirEntry{name: name, isDir: true})
			}
		}
	}

	return entries, nil
}

func (m *mockStorage) DeleteEmptyDirectory(fsPath string) error {
	delete(m.dirs, fsPath)
	return nil
}

func (m *mockStorage) DeleteDirectory(fsPath string) error {
	prefix := withPrefix(fsPath)

	// Delete all files under this directory
	maps.DeleteFunc(m.files, func(path string, _ []byte) bool {
		return strings.HasPrefix(path, prefix)
	})

	// Delete the directory itself and all subdirectories
	maps.DeleteFunc(m.dirs, func(path string, _ bool) bool {
		return strings.HasPrefix(path, prefix) || path == fsPath
	})

	return nil
}

// renamePaths renames all map entries matching oldPrefix to use newPrefix
func renamePaths[V any](m map[string]V, oldPrefix, newPrefix string) {
	for path, value := range m {
		if strings.HasPrefix(path, oldPrefix) {
			newPath := newPrefix + strings.TrimPrefix(path, oldPrefix)
			m[newPath] = value
			delete(m, path)
		}
	}
}

func (m *mockStorage) Rename(oldPath, newPath string) error {
	// Handle file rename
	if data, ok := m.files[oldPath]; ok {
		m.createParentDirs(newPath)
		m.files[newPath] = data
		delete(m.files, oldPath)
		return nil
	}

	// Handle directory rename
	if m.dirs[oldPath] {
		renamePaths(m.files, withPrefix(oldPath), withPrefix(newPath))
		renamePaths(m.dirs, oldPath, newPath) // Uses oldPath directly to also rename the dir itself

		return nil
	}

	return fmt.Errorf("could not rename %s", oldPath)
}

func (m *mockStorage) ReadConfig() (model.Config, error) {
	panic("not implemented")
}

func (m *mockStorage) WriteConfig(config model.Config) error {
	panic("not implemented")
}

// mockDirEntry implements fs.DirEntry for testing
type mockDirEntry struct {
	name  string
	isDir bool
}

func (m mockDirEntry) Name() string { return m.name }
func (m mockDirEntry) IsDir() bool  { return m.isDir }
func (m mockDirEntry) Info() (fs.FileInfo, error) {
	panic("not implemented")
}
func (m mockDirEntry) Type() fs.FileMode {
	panic("not implemented")
}
