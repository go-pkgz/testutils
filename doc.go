// Package testutils provides a collection of utilities for testing Go applications
// including stdout/stderr capture, file handling, HTTP test helpers, and test containers.
//
// It consists of several main components:
//
// Capture Utilities:
//   - CaptureStdout, CaptureStderr, CaptureStdoutAndStderr: Functions to capture output from standard streams
//     during test execution. These are useful for testing functions that write directly to stdout/stderr.
//     Note: These functions are not thread-safe for parallel tests.
//
// File Utilities:
//   - WriteTestFile: Creates a temporary file with specified content for testing purposes,
//     with automatic cleanup after the test completes.
//
// HTTP Utilities:
//   - MockHTTPServer: Creates a test HTTP server with the provided handler
//   - HTTPRequestCaptor: Captures and records HTTP requests for later inspection
//
// Test Containers:
// The 'containers' subpackage provides Docker containers for integration testing:
//   - SSHTestContainer: SSH server container with file operation support (upload, download, list, delete)
//   - FTPTestContainer: FTP server container with file operation support
//   - PostgresTestContainer: PostgreSQL database container with automatic DB creation
//   - MySQLTestContainer: MySQL database container with automatic DB creation
//   - MongoTestContainer: MongoDB container with support for multiple versions
//   - LocalstackTestContainer: LocalStack container with S3 service for AWS testing
//
// All container implementations support a common pattern:
//   - Container creation with NewXXXTestContainer
//   - Automatic port mapping and connection configuration
//   - Graceful shutdown with the Close method
//   - File operations where applicable (SaveFile, GetFile, ListFiles, DeleteFile)
//
// These utilities help simplify test setup, improve test reliability, and reduce
// boilerplate code in test suites, especially for integration tests.
package testutils
