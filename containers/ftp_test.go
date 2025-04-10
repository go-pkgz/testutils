// file: testutils/containers/ftp_test.go
package containers

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFTPContainer uses Testcontainers to manage the FTP server lifecycle.
func TestFTPContainer(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping FTP container test in short mode")
	}
	if os.Getenv("CI") != "" && os.Getenv("RUN_FTP_TESTS_ON_CI") == "" {
		t.Skip("skipping FTP container test in CI environment unless RUN_FTP_TESTS_ON_CI is set")
	}

	ctx := context.Background()
	ftpContainer := NewFTPTestContainer(ctx, t) // uses the final working logic from ftp.go
	defer func() { assert.NoError(t, ftpContainer.Close(context.Background())) }()

	require.NotEmpty(t, ftpContainer.GetIP())
	require.Positive(t, ftpContainer.GetPort())
	require.Equal(t, "ftpuser", ftpContainer.GetUser())
	require.Equal(t, "ftppass", ftpContainer.GetPassword())
	connectionString := ftpContainer.ConnectionString()
	require.Equal(t, fmt.Sprintf("%s:%d", ftpContainer.GetIP(), ftpContainer.GetPort()), connectionString)
	t.Logf("connection string: %s", connectionString)

	t.Logf("FTP Connection: %s", ftpContainer.ConnectionString())
	t.Logf("FTP Username: %s", ftpContainer.GetUser())
	t.Logf("FTP Password: %s", ftpContainer.GetPassword())
	logs, err := ftpContainer.Container.Logs(ctx)
	if err == nil {
		defer logs.Close()
		logBytes, _ := io.ReadAll(logs)
		t.Logf("FTP Container logs: %s", string(logBytes))
	} else {
		t.Logf("Could not get container logs: %v", err)
	}

	tempDir := t.TempDir()
	testFiles := map[string][]byte{
		"test1.txt":          []byte("Hello world from test1"),
		"test2.txt":          []byte("This is test2 content: " + fmt.Sprint(time.Now().UnixNano())),
		"testdir/nested.txt": []byte("Nested file"),
		"testdir/more.txt":   []byte("Another nested: " + time.Now().Format(time.RFC3339)),
	}
	for filename, content := range testFiles {
		localPath := filepath.Join(tempDir, filename)
		if dir := filepath.Dir(localPath); dir != tempDir {
			require.NoError(t, os.MkdirAll(dir, 0o750))
		}
		require.NoError(t, os.WriteFile(localPath, content, 0o600))
	}

	t.Run("Connect", func(t *testing.T) {
		// test the connect function directly
		conn, err := ftpContainer.connect(ctx)
		require.NoError(t, err)
		require.NotNil(t, conn)

		// test connection state
		pwd, err := conn.CurrentDir()
		require.NoError(t, err)
		require.NotEmpty(t, pwd)

		// close properly
		err = conn.Quit()
		require.NoError(t, err)
	})

	t.Run("SaveAndRestoreCurrentDirectory", func(t *testing.T) {
		// first we need to make sure the testdir exists
		testdirPath := filepath.Join(tempDir, "testdir")
		moreFilePath := filepath.Join(testdirPath, "more.txt")
		require.NoError(t, os.MkdirAll(testdirPath, 0o750))
		require.NoError(t, os.WriteFile(moreFilePath, []byte("Test content"), 0o600))
		err := ftpContainer.SaveFile(ctx, moreFilePath, "testdir/more.txt")
		require.NoError(t, err, "Failed to create directory for test")

		conn, err := ftpContainer.connect(ctx)
		require.NoError(t, err)
		defer func() {
			err := conn.Quit()
			require.NoError(t, err)
		}()

		// test saving current directory
		originalDir, err := ftpContainer.saveCurrentDirectory(conn)
		require.NoError(t, err)
		require.NotEmpty(t, originalDir)

		// change directory - now testdir should exist
		err = conn.ChangeDir("testdir")
		require.NoError(t, err)

		// verify we're in a different directory
		currentDir, err := conn.CurrentDir()
		require.NoError(t, err)
		require.NotEqual(t, originalDir, currentDir)

		// restore to original directory
		ftpContainer.restoreWorkingDirectory(conn, originalDir)

		// verify we're back in the original directory
		restoredDir, err := conn.CurrentDir()
		require.NoError(t, err)
		require.Equal(t, originalDir, restoredDir)

		// test with empty original directory (should be a no-op)
		ftpContainer.restoreWorkingDirectory(conn, "")
	})

	t.Run("Upload", func(t *testing.T) {
		for filename := range testFiles {
			localPath := filepath.Join(tempDir, filename)
			remotePath := filename
			err := ftpContainer.SaveFile(ctx, localPath, remotePath)
			require.NoError(t, err, "Failed to upload file %s", filename)
		}

		// test with invalid local path
		err := ftpContainer.SaveFile(ctx, "/path/does/not/exist.txt", "remote.txt")
		require.Error(t, err)

		// test with invalid path traversal attempt
		err = ftpContainer.SaveFile(ctx, "../../../etc/passwd", "remote.txt")
		require.Error(t, err)
	})

	t.Run("ListHome", func(t *testing.T) {
		// *** List the explicit home directory path ***
		// using "." might be less reliable if connection state resets
		homeDir := "/ftp/ftpuser" // based on PWD logs
		homeEntries, err := ftpContainer.ListFiles(ctx, homeDir)
		require.NoError(t, err, "Failed to list home directory %s", homeDir)
		require.NotEmpty(t, homeEntries, "Home directory %s should not be empty", homeDir)

		foundFiles := make(map[string]bool)
		for _, entry := range homeEntries {
			t.Logf("Found in home dir (%s): %s", homeDir, entry.Name)
			// assuming 0 is file and 1 is folder
			if (entry.Name == "test1.txt" || entry.Name == "test2.txt") && entry.Type == 0 {
				foundFiles[entry.Name] = true
			}
			if entry.Name == "testdir" && entry.Type == 1 {
				foundFiles[entry.Name] = true
			}
		}
		assert.True(t, foundFiles["test1.txt"], "test1.txt should be in home directory")
		assert.True(t, foundFiles["test2.txt"], "test2.txt should be in home directory")
		assert.True(t, foundFiles["testdir"], "testdir should be in home directory")

		// test empty path and "." path
		emptyEntries, err := ftpContainer.ListFiles(ctx, "")
		require.NoError(t, err)
		require.NotEmpty(t, emptyEntries)

		dotEntries, err := ftpContainer.ListFiles(ctx, ".")
		require.NoError(t, err)
		require.NotEmpty(t, dotEntries)
	})

	t.Run("ListSubdir", func(t *testing.T) {
		// list relative path "testdir" or absolute path "/ftp/ftpuser/testdir"
		subdirPath := "testdir" // relative path should work fine here
		subdirEntries, err := ftpContainer.ListFiles(ctx, subdirPath)
		require.NoError(t, err, "Failed to list subdirectory '%s'", subdirPath)
		require.NotEmpty(t, subdirEntries, "Subdirectory '%s' should not be empty", subdirPath)

		foundSubdirFiles := make(map[string]bool)
		for _, entry := range subdirEntries {
			t.Logf("Found in %s: %s", subdirPath, entry.Name)
			if (entry.Name == "nested.txt" || entry.Name == "more.txt") && entry.Type == 0 {
				foundSubdirFiles[entry.Name] = true
			}
		}
		assert.True(t, foundSubdirFiles["nested.txt"], "nested.txt should be in subdirectory '%s'", subdirPath)
		assert.True(t, foundSubdirFiles["more.txt"], "more.txt should be in subdirectory '%s'", subdirPath)
	})

	t.Run("DownloadAndVerify", func(t *testing.T) {
		downloadDir := filepath.Join(tempDir, "download")
		require.NoError(t, os.MkdirAll(downloadDir, 0o750))
		downloadFiles := []string{"test1.txt", "testdir/nested.txt"}
		for _, filename := range downloadFiles {
			originalContent, ok := testFiles[filename]
			require.True(t, ok)
			remotePath := filename
			localRelPath := strings.ReplaceAll(filename, "/", string(filepath.Separator))
			downloadPath := filepath.Join(downloadDir, localRelPath)
			if dir := filepath.Dir(downloadPath); dir != downloadDir {
				require.NoError(t, os.MkdirAll(dir, 0o750))
			}
			t.Logf("Downloading remote '%s' to local '%s'", remotePath, downloadPath)
			require.NoError(t, ftpContainer.GetFile(ctx, remotePath, downloadPath), "Failed to download file %s", filename)
			// read file is safe here since we control the downloadPath in tests
			downloadedContent, err := os.ReadFile(downloadPath) // #nosec G304 -- safe file access in test
			require.NoError(t, err)
			require.Equal(t, originalContent, downloadedContent, "Content mismatch for file %s", filename)
			t.Logf("Verified content for downloaded file %s", filename)
		}

		// test with non-existent remote file
		err := ftpContainer.GetFile(ctx, "non-existent-file.txt", filepath.Join(downloadDir, "non-existent.txt"))
		require.Error(t, err)

		// test with path traversal attempt
		err = ftpContainer.GetFile(ctx, "file.txt", "../../../etc/malicious.txt")
		require.Error(t, err)
	})

	t.Run("SplitPath", func(t *testing.T) {
		testCases := []struct {
			path     string
			expected []string
		}{
			{path: "foo/bar/baz", expected: []string{"foo", "bar", "baz"}},
			{path: "/foo/bar/baz", expected: []string{"foo", "bar", "baz"}},
			{path: "foo/bar/baz/", expected: []string{"foo", "bar", "baz"}},
			{path: "/foo/bar/baz/", expected: []string{"foo", "bar", "baz"}},
			{path: "", expected: []string{}},
			{path: "/", expected: []string{}},
		}

		for _, tc := range testCases {
			t.Run(tc.path, func(t *testing.T) {
				result := splitPath(tc.path)
				require.Equal(t, tc.expected, result)
			})
		}
	})

	t.Run("CreateDirRecursive", func(t *testing.T) {
		conn, err := ftpContainer.connect(ctx)
		require.NoError(t, err)
		defer func() {
			err := conn.Quit()
			require.NoError(t, err)
		}()

		// create a deep directory structure
		err = ftpContainer.createDirRecursive(conn, "deep/nested/directory/structure")
		require.NoError(t, err)

		// verify it exists by listing files
		entries, err := ftpContainer.ListFiles(ctx, "deep/nested/directory")
		require.NoError(t, err)

		foundStructure := false
		for _, entry := range entries {
			if entry.Name == "structure" && entry.Type == 1 {
				foundStructure = true
				break
			}
		}
		assert.True(t, foundStructure, "deep/nested/directory/structure should exist")

		// test with empty path (should be a no-op)
		err = ftpContainer.createDirRecursive(conn, "")
		require.NoError(t, err)
	})
}

// TestFTPContainerErrorHandling tests error handling in FTP container
func TestFTPContainerErrorHandling(t *testing.T) {
	t.Run("TestHandleMakeDirFailure", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping FTP error handling test in short mode")
		}

		// create a test container
		ctx := context.Background()
		ftpContainer := NewFTPTestContainer(ctx, t)
		defer func() {
			err := ftpContainer.Close(context.Background())
			require.NoError(t, err)
		}()

		// connect to the container
		conn, err := ftpContainer.connect(ctx)
		require.NoError(t, err)
		defer conn.Quit()

		// test the handleMakeDirFailure function with a directory that already exists
		// first create the directory normally
		err = conn.MakeDir("testdir2")
		require.NoError(t, err)

		// now simulate a failure but where the directory actually exists
		err = ftpContainer.handleMakeDirFailure(conn, "testdir2", fmt.Errorf("simulated error"))
		require.NoError(t, err, "Should handle the case where directory exists but MakeDir failed")

		// test with a non-existent directory
		err = ftpContainer.handleMakeDirFailure(conn, "definitely_not_exists_dir", fmt.Errorf("simulated error"))
		require.Error(t, err, "Should fail when directory doesn't exist")
	})
}

// Test utility methods separately
func TestSplitPath(t *testing.T) {
	testCases := []struct {
		path     string
		expected []string
	}{
		{path: "foo/bar/baz", expected: []string{"foo", "bar", "baz"}},
		{path: "/foo/bar/baz", expected: []string{"foo", "bar", "baz"}},
		{path: "foo/bar/baz/", expected: []string{"foo", "bar", "baz"}},
		{path: "/foo/bar/baz/", expected: []string{"foo", "bar", "baz"}},
		{path: "", expected: []string{}},
		{path: "/", expected: []string{}},
	}

	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			result := splitPath(tc.path)
			require.Equal(t, tc.expected, result)
		})
	}
}
