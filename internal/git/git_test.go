package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestOSClientInitAndCommit(t *testing.T) {
	tempDir := t.TempDir()

	client := NewClient()
	err := client.Init(tempDir)
	if err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Configure git user locally for this test repo
	exec.Command("git", "-C", tempDir, "config", "user.name", "Test User").Run()
	exec.Command("git", "-C", tempDir, "config", "user.email", "test@example.com").Run()

	dummyFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(dummyFile, []byte("hello"), 0644)
	if err != nil {
		t.Fatalf("Failed to create dummy file: %v", err)
	}

	err = client.Add(tempDir, "test.txt")
	if err != nil {
		t.Fatalf("Failed to add file: %v", err)
	}

	err = client.Commit(tempDir, "Initial commit")
	if err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}
}
