package containers

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalstackTestContainer(t *testing.T) {
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
}
