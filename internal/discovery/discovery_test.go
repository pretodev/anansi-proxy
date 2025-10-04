package discovery

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindAPIMockFiles_SingleFile(t *testing.T) {
	// Create a temporary directory with test files
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.apimock")
	if err := os.WriteFile(testFile, []byte("test content"), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	files, err := FindAPIMockFiles(testFile)
	if err != nil {
		t.Fatalf("FindAPIMockFiles() error = %v", err)
	}

	if len(files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(files))
	}

	if files[0] != testFile {
		t.Errorf("Expected file path %s, got %s", testFile, files[0])
	}
}

func TestFindAPIMockFiles_MultipleFiles(t *testing.T) {
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "test1.apimock")
	file2 := filepath.Join(tmpDir, "test2.apimock")

	if err := os.WriteFile(file1, []byte("test1"), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if err := os.WriteFile(file2, []byte("test2"), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	files, err := FindAPIMockFiles(file1, file2)
	if err != nil {
		t.Fatalf("FindAPIMockFiles() error = %v", err)
	}

	if len(files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(files))
	}
}

func TestFindAPIMockFiles_Directory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files in root
	file1 := filepath.Join(tmpDir, "test1.apimock")
	if err := os.WriteFile(file1, []byte("test1"), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a subdirectory with a file
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	file2 := filepath.Join(subDir, "test2.apimock")
	if err := os.WriteFile(file2, []byte("test2"), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a non-.apimock file that should be ignored
	otherFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(otherFile, []byte("other"), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	files, err := FindAPIMockFiles(tmpDir)
	if err != nil {
		t.Fatalf("FindAPIMockFiles() error = %v", err)
	}

	if len(files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(files))
	}
}

func TestFindAPIMockFiles_IgnoredDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a .git directory with a file
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0o755); err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

	gitFile := filepath.Join(gitDir, "test.apimock")
	if err := os.WriteFile(gitFile, []byte("git"), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a regular file in root
	regularFile := filepath.Join(tmpDir, "regular.apimock")
	if err := os.WriteFile(regularFile, []byte("regular"), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	files, err := FindAPIMockFiles(tmpDir)
	if err != nil {
		t.Fatalf("FindAPIMockFiles() error = %v", err)
	}

	if len(files) != 1 {
		t.Errorf("Expected 1 file (should ignore .git), got %d", len(files))
	}

	if files[0] != regularFile {
		t.Errorf("Expected file %s, got %s", regularFile, files[0])
	}
}

func TestFindAPIMockFiles_NonApimockFile(t *testing.T) {
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err := FindAPIMockFiles(testFile)
	if err == nil {
		t.Error("Expected error for non-.apimock file, got nil")
	}
}

func TestFindAPIMockFiles_NonExistentPath(t *testing.T) {
	_, err := FindAPIMockFiles("/nonexistent/path")
	if err == nil {
		t.Error("Expected error for non-existent path, got nil")
	}
}

func TestFindAPIMockFiles_EmptyPaths(t *testing.T) {
	_, err := FindAPIMockFiles()
	if err == nil {
		t.Error("Expected error for empty paths, got nil")
	}
}

func TestFindAPIMockFiles_NoDuplicates(t *testing.T) {
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.apimock")
	if err := os.WriteFile(testFile, []byte("test"), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Pass the same file twice
	files, err := FindAPIMockFiles(testFile, testFile)
	if err != nil {
		t.Fatalf("FindAPIMockFiles() error = %v", err)
	}

	if len(files) != 1 {
		t.Errorf("Expected 1 file (no duplicates), got %d", len(files))
	}
}
