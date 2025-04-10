package testutils

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWriteTestFile(t *testing.T) {
	t.Run("standard file creation", func(t *testing.T) {
		// test creating a file with content
		content := "test content"
		filePath := WriteTestFile(t, content)

		// check if file exists
		_, err := os.Stat(filePath)
		require.False(t, os.IsNotExist(err), "WriteTestFile did not create file at %s", filePath)

		// check content
		// although ReadFile takes a path from a test-generated file, this is safe
		// because the file is created in a temporary directory controlled by the test
		// this is safe because the file path was created by WriteTestFile in a controlled manner
		data, err := os.ReadFile(filePath) // #nosec G304 -- safe file access in test
		require.NoError(t, err, "Failed to read test file")
		require.Equal(t, content, string(data), "File content doesn't match expected")

		// verify directory structure
		dir := filepath.Dir(filePath)
		require.Contains(t, dir, "testutils-", "Temp directory should contain expected prefix")

		// file should be cleaned up automatically at the end of the test
	})

	t.Run("with empty content", func(t *testing.T) {
		filePath := WriteTestFile(t, "")

		// check empty file was created
		info, err := os.Stat(filePath)
		require.NoError(t, err, "File should exist")
		require.Zero(t, info.Size(), "File should be empty")
	})

	t.Run("with multi-line content", func(t *testing.T) {
		content := "line 1\nline 2\nline 3"
		filePath := WriteTestFile(t, content)

		// although ReadFile takes a path from a test-generated file, this is safe
		// because the file is created in a temporary directory controlled by the test
		// this is safe because the file path was created by WriteTestFile in a controlled manner
		data, err := os.ReadFile(filePath) // #nosec G304 -- safe file access in test
		require.NoError(t, err)
		require.Equal(t, content, string(data))
	})

	t.Run("cleanup by direct call", func(t *testing.T) {
		// create a test file
		content := "test cleanup"
		filePath := WriteTestFile(t, content)

		// get the directory to be cleaned up
		dir := filepath.Dir(filePath)

		// verify directory and file exist
		require.DirExists(t, dir)
		require.FileExists(t, filePath)

		// manually clean up (simulating what t.Cleanup would do)
		err := os.RemoveAll(dir)
		require.NoError(t, err)

		// after manual cleanup, the file should no longer exist
		_, err = os.Stat(filePath)
		require.True(t, os.IsNotExist(err), "File should be removed after cleanup")
	})

	t.Run("with large content", func(t *testing.T) {
		// create a large string (100KB)
		largeContent := strings.Repeat("abcdefghij", 10*1024) // 10 chars * 10K = 100KB
		filePath := WriteTestFile(t, largeContent)

		// verify the file exists and has correct size
		info, err := os.Stat(filePath)
		require.NoError(t, err)
		require.Equal(t, int64(len(largeContent)), info.Size())

		// verify the content
		data, err := os.ReadFile(filePath) // #nosec G304 -- safe file access in test
		require.NoError(t, err)
		require.Equal(t, largeContent, string(data))
	})

	t.Run("with binary content", func(t *testing.T) {
		// create some binary data
		binaryContent := string([]byte{0, 1, 2, 3, 4, 5, 255, 254, 253})
		filePath := WriteTestFile(t, binaryContent)

		// verify the content
		data, err := os.ReadFile(filePath) // #nosec G304 -- safe file access in test
		require.NoError(t, err)
		require.Equal(t, binaryContent, string(data))
	})

	t.Run("file permissions", func(t *testing.T) {
		content := "permission test"
		filePath := WriteTestFile(t, content)

		// check that the file has the expected permissions (0o600)
		info, err := os.Stat(filePath)
		require.NoError(t, err)

		// on Unix systems, we can check the exact permission bits
		if runtime.GOOS != "windows" {
			// 0o600 = rw------- (read/write for owner only)
			expectedPerm := os.FileMode(0o600)
			actualPerm := info.Mode().Perm()
			require.Equal(t, expectedPerm, actualPerm, "File should have 0600 permissions")
		}
	})
}
