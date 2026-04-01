package plan

import (
	"os"
	"path/filepath"
	"strings"
)

// FindRootModules walks rootDir recursively and returns directories that contain
// a Terraform backend block (the canonical marker for a root module).
// It skips .terraform directories to avoid scanning downloaded providers/modules.
func FindRootModules(rootDir string) ([]string, error) {
	var modules []string

	err := filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden / vendor / provider cache directories
		if d.IsDir() {
			name := d.Name()
			if name == ".terraform" || name == ".git" || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			return nil
		}

		// Only inspect .tf files
		if filepath.Ext(path) != ".tf" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil // skip unreadable files
		}

		if hasBackendBlock(string(data)) {
			modules = append(modules, filepath.Dir(path))
			// Don't descend further into this directory's sub-dirs for the same module
			return filepath.SkipDir
		}

		return nil
	})

	return modules, err
}

// hasBackendBlock reports whether a .tf file body contains a terraform {} block
// with a nested backend sub-block, which is the canonical marker of a root module.
func hasBackendBlock(content string) bool {
	// Simple heuristic: look for both "terraform" and "backend" keywords.
	// This is intentionally lightweight to avoid a full HCL parse at discovery time.
	lower := strings.ToLower(content)
	return strings.Contains(lower, "terraform") && strings.Contains(lower, "backend")
}
