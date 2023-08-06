# testutils [![Build Status](https://github.com/go-pkgz/testutils/workflows/build/badge.svg)](https://github.com/go-pkgz/testutils/actions) [![Go Report Card](https://goreportcard.com/badge/github.com/go-pkgz/testutils)](https://goreportcard.com/report/github.com/go-pkgz/testutils) [![Coverage Status](https://coveralls.io/repos/github/go-pkgz/testutils/badge.svg?branch=master)](https://coveralls.io/github/go-pkgz/testutils?branch=master)

Package `testutils` provides useful test helpers.

## Details

- `CaptureStdout`, `CaptureSterr` and `CaptureStdoutAndStderr`: capture stdout, stderr or both for testing purposes. All capture functions are not thread-safe if used in parallel tests, and usually it is better to pass a custom io.Writer to the function under test instead.

## Install and update

`go get -u github.com/go-pkgz/testutils`
