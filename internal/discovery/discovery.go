package discovery

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var ignoredDirs = map[string]bool{
	".git":          true,
	"node_modules":  true,
	".idea":         true,
	".vscode":       true,
	"vendor":        true,
	"build":         true,
	"dist":          true,
	".next":         true,
	"coverage":      true,
	"__pycache__":   true,
	".pytest_cache": true,
	"bin":           true,
}

// FindAPIMockFiles takes a list of paths (files or directories) and returns
// all .apimock files found. For directories, it recursively searches for .apimock files.
func FindAPIMockFiles(paths ...string) ([]string, error) {
	var files []string
	seen := make(map[string]bool)

	for _, path := range paths {
		if path == "" {
			continue
		}

		absPath, err := filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("failed to get absolute path for '%s': %w", path, err)
		}

		info, err := os.Stat(absPath)
		if err != nil {
			return nil, fmt.Errorf("failed to access path '%s': %w", absPath, err)
		}

		if info.IsDir() {
			dirFiles, err := findInDirectory(absPath)
			if err != nil {
				return nil, fmt.Errorf("failed to search directory '%s': %w", absPath, err)
			}
			for _, f := range dirFiles {
				if !seen[f] {
					files = append(files, f)
					seen[f] = true
				}
			}

			continue
		}

		if !strings.HasSuffix(absPath, ".apimock") {
			return nil, fmt.Errorf("file '%s' is not an .apimock file", absPath)
		}

		if !seen[absPath] {
			files = append(files, absPath)
			seen[absPath] = true
		}

	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no .apimock files found in the specified paths")
	}

	return files, nil
}

// findInDirectory recursively finds all .apimock files in a directory
func findInDirectory(dirPath string) ([]string, error) {
	var files []string

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip ignored directories
		if info.IsDir() {
			dirName := filepath.Base(path)
			if ignoredDirs[dirName] && path != dirPath {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if it's an .apimock file
		if strings.HasSuffix(path, ".apimock") {
			files = append(files, path)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}
