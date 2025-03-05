package testutils

import (
	"os"
	"testing"
)

func TestWriteTestFile(t *testing.T) {
	// test creating a file with content
	content := "test content"
	filePath := WriteTestFile(t, content)

	// check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("WriteTestFile did not create file at %s", filePath)
	}

	// check content
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Errorf("Failed to read test file: %v", err)
	}

	if string(data) != content {
		t.Errorf("Expected content %q, got %q", content, string(data))
	}

	// file should be cleaned up automatically at the end of the test
}
