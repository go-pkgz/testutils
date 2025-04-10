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

	t.Run("Upload", func(t *testing.T) {
		for filename := range testFiles {
			localPath := filepath.Join(tempDir, filename)
			remotePath := filename
			err := ftpContainer.SaveFile(ctx, localPath, remotePath)
			require.NoError(t, err, "Failed to upload file %s", filename)
		}
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
	})
}
