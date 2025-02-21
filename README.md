# testutils [![Build Status](https://github.com/go-pkgz/testutils/workflows/build/badge.svg)](https://github.com/go-pkgz/testutils/actions) [![Go Report Card](https://goreportcard.com/badge/github.com/go-pkgz/testutils)](https://goreportcard.com/report/github.com/go-pkgz/testutils) [![Coverage Status](https://coveralls.io/repos/github/go-pkgz/testutils/badge.svg?branch=master)](https://coveralls.io/github/go-pkgz/testutils?branch=master)

Package `testutils` provides useful test helpers.

## Details

- `CaptureStdout`, `CaptureSterr` and `CaptureStdoutAndStderr`: capture stdout, stderr or both for testing purposes. All capture functions are not thread-safe if used in parallel tests, and usually it is better to pass a custom io.Writer to the function under test instead.

- `containers`: provides test containers for integration testing:
    - `SSHTestContainer`: SSH server container for testing SSH connections and operations
    - `PostgresTestContainer`: PostgreSQL database container with automatic database creation
    - `MySQLTestContainer`: MySQL database container with automatic database creation
    - `MongoTestContainer`: MongoDB container with support for multiple versions (5, 6, 7)
    - `LocalstackTestContainer`: LocalStack container with S3 service for AWS testing

## Install and update

`go get -u github.com/go-pkgz/testutils`

### Example usage:

```go
// PostgreSQL test container
func TestWithPostgres(t *testing.T) {
    ctx := context.Background()
    pg := containers.NewPostgresTestContainer(ctx, t)
    defer pg.Close(ctx)
    
    db, err := sql.Open("postgres", pg.ConnectionString())
    require.NoError(t, err)
    defer db.Close()
    
    // run your tests with the database
}

// MySQL test container
func TestWithMySQL(t *testing.T) {
    ctx := context.Background()
    mysql := containers.NewMySQLTestContainer(ctx, t)
    defer mysql.Close(ctx)
    
    db, err := sql.Open("mysql", mysql.DSN())
    require.NoError(t, err)
    defer db.Close()
    
    // run your tests with the database
}

// MongoDB test container
func TestWithMongo(t *testing.T) {
    ctx := context.Background()
    mongo := containers.NewMongoTestContainer(ctx, t, 7) // version 7
    defer mongo.Close(ctx)
    
    coll := mongo.Collection("test_db")
    _, err := coll.InsertOne(ctx, bson.M{"test": "value"})
    require.NoError(t, err)
}

// SSH test container
func TestWithSSH(t *testing.T) {
    ctx := context.Background()
    ssh := containers.NewSSHTestContainer(ctx, t)
    defer ssh.Close(ctx)
    
    // use ssh.Address() to get host:port
    // default user is "test"
    sshAddr := ssh.Address()
}

// Localstack (S3) test container
func TestWithS3(t *testing.T) {
    ctx := context.Background()
    ls := containers.NewLocalstackTestContainer(ctx, t)
    defer ls.Close(ctx)
    
    s3Client, bucketName := ls.MakeS3Connection(ctx, t)
    
    // put object example
    _, err := s3Client.PutObject(ctx, &s3.PutObjectInput{
        Bucket: aws.String(bucketName),
        Key:    aws.String("test-key"),
        Body:   strings.NewReader("test content"),
    })
    require.NoError(t, err)
}
```

