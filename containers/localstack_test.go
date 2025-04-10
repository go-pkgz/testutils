package containers

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalstackTestContainer(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping Localstack container test in short mode")
	}

	ctx := context.Background()

	t.Run("create and cleanup container", func(t *testing.T) {
		ls := NewLocalstackTestContainer(ctx, t)
		defer func() { require.NoError(t, ls.Close(ctx)) }()

		assert.NotEmpty(t, ls.Endpoint)
		assert.Contains(t, ls.Endpoint, "http://")
	})

	t.Run("make s3 connection", func(t *testing.T) {
		ls := NewLocalstackTestContainer(ctx, t)
		defer func() { require.NoError(t, ls.Close(ctx)) }()

		client, bucketName := ls.MakeS3Connection(ctx, t)

		// verify we can use the connection
		buckets, err := client.ListBuckets(ctx, &s3.ListBucketsInput{})
		require.NoError(t, err)
		require.Len(t, buckets.Buckets, 1)
		assert.Equal(t, bucketName, *buckets.Buckets[0].Name)
	})

	t.Run("s3 operations", func(t *testing.T) {
		ls := NewLocalstackTestContainer(ctx, t)
		defer func() { require.NoError(t, ls.Close(ctx)) }()

		client, bucketName := ls.MakeS3Connection(ctx, t)

		// test put object
		_, err := client.PutObject(ctx, &s3.PutObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String("test-key"),
			Body:   strings.NewReader("test content"),
		})
		require.NoError(t, err)

		// test get object
		result, err := client.GetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String("test-key"),
		})
		require.NoError(t, err)

		content, err := io.ReadAll(result.Body)
		require.NoError(t, err)
		assert.Equal(t, "test content", string(content))
	})

	t.Run("multiple connections to same container", func(t *testing.T) {
		ls := NewLocalstackTestContainer(ctx, t)
		defer func() { require.NoError(t, ls.Close(ctx)) }()

		client1, bucket1 := ls.MakeS3Connection(ctx, t)
		client2, bucket2 := ls.MakeS3Connection(ctx, t)

		// test operations with first client
		_, err := client1.PutObject(ctx, &s3.PutObjectInput{
			Bucket: aws.String(bucket1),
			Key:    aws.String("test1"),
			Body:   strings.NewReader("content1"),
		})
		require.NoError(t, err)

		// test operations with second client
		_, err = client2.PutObject(ctx, &s3.PutObjectInput{
			Bucket: aws.String(bucket2),
			Key:    aws.String("test2"),
			Body:   strings.NewReader("content2"),
		})
		require.NoError(t, err)

		// verify buckets are different
		assert.NotEqual(t, bucket1, bucket2)

		// verify first client can access its content
		result1, err := client1.GetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(bucket1),
			Key:    aws.String("test1"),
		})
		require.NoError(t, err)
		content1, err := io.ReadAll(result1.Body)
		require.NoError(t, err)
		assert.Equal(t, "content1", string(content1))

		// verify second client can access its content
		result2, err := client2.GetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(bucket2),
			Key:    aws.String("test2"),
		})
		require.NoError(t, err)
		content2, err := io.ReadAll(result2.Body)
		require.NoError(t, err)
		assert.Equal(t, "content2", string(content2))
	})

	t.Run("file operations", func(t *testing.T) {
		ls := NewLocalstackTestContainer(ctx, t)
		defer func() { require.NoError(t, ls.Close(ctx)) }()

		_, bucketName := ls.MakeS3Connection(ctx, t)

		// create a temp directory and test file
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "test-s3-file.txt")
		testContent := "Hello S3 world!"
		require.NoError(t, os.WriteFile(testFile, []byte(testContent), 0o600))

		// test SaveFile - upload to S3
		objectKey := "test-object.txt"
		err := ls.SaveFile(ctx, testFile, bucketName, objectKey)
		require.NoError(t, err, "Failed to upload file to S3")

		// test ListFiles - check if file exists
		objects, err := ls.ListFiles(ctx, bucketName, "")
		require.NoError(t, err, "Failed to list objects in S3 bucket")

		found := false
		for _, obj := range objects {
			if *obj.Key == objectKey {
				found = true
				break
			}
		}
		require.True(t, found, "Uploaded object not found in S3 bucket")

		// test GetFile - download from S3
		downloadedFile := filepath.Join(tempDir, "downloaded-s3-file.txt")
		err = ls.GetFile(ctx, bucketName, objectKey, downloadedFile)
		require.NoError(t, err, "Failed to download file from S3")

		// verify content
		content, err := os.ReadFile(downloadedFile) // #nosec G304 -- Safe file access, path is controlled in test
		require.NoError(t, err)
		assert.Equal(t, testContent, string(content), "Downloaded content doesn't match original")

		// test DeleteFile - delete from S3
		err = ls.DeleteFile(ctx, bucketName, objectKey)
		require.NoError(t, err, "Failed to delete object from S3")

		// verify object was deleted
		objects, err = ls.ListFiles(ctx, bucketName, "")
		require.NoError(t, err)

		found = false
		for _, obj := range objects {
			if *obj.Key == objectKey {
				found = true
				break
			}
		}
		require.False(t, found, "Object should have been deleted from S3")

		// test with prefix
		// create multiple test files in different "directories"
		prefixPaths := []string{
			"prefix1/file1.txt",
			"prefix1/file2.txt",
			"prefix2/file1.txt",
		}

		for _, path := range prefixPaths {
			testFilePath := filepath.Join(tempDir, filepath.Base(path))
			require.NoError(t, os.WriteFile(testFilePath, []byte("Content for "+path), 0o600))
			err := ls.SaveFile(ctx, testFilePath, bucketName, path)
			require.NoError(t, err)
		}

		// list objects with prefix1
		objects, err = ls.ListFiles(ctx, bucketName, "prefix1/")
		require.NoError(t, err)
		assert.Len(t, objects, 2, "Should find 2 objects with prefix1/")

		// list objects with prefix2
		objects, err = ls.ListFiles(ctx, bucketName, "prefix2/")
		require.NoError(t, err)
		assert.Len(t, objects, 1, "Should find 1 object with prefix2/")

		// delete prefix1/file1.txt
		err = ls.DeleteFile(ctx, bucketName, "prefix1/file1.txt")
		require.NoError(t, err)

		// verify it was deleted
		objects, err = ls.ListFiles(ctx, bucketName, "prefix1/")
		require.NoError(t, err)
		assert.Len(t, objects, 1, "Should find 1 object with prefix1/ after deletion")
	})
}
