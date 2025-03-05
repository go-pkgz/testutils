package testutils

import (
	"os"
	"testing"
	
	"github.com/stretchr/testify/require"
)

func TestWriteTestFile(t *testing.T) {
	// test creating a file with content
	content := "test content"
	filePath := WriteTestFile(t, content)

	// check if file exists
	_, err := os.Stat(filePath)
	require.False(t, os.IsNotExist(err), "WriteTestFile did not create file at %s", filePath)
	
	// check content
	data, err := os.ReadFile(filePath)
	require.NoError(t, err, "Failed to read test file")
	require.Equal(t, content, string(data), "File content doesn't match expected")
	
	// file should be cleaned up automatically at the end of the test
}
