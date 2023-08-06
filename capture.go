package testutils

import (
	"bytes"
	"io"
	"os"
	"sync"
	"testing"
)

// CaptureStdout captures the output of a function that writes to stdout.
// All Capture functions are not thread-safe if used in parallel tests.
// Usually it is better to pass a custom io.Writer to the function under test instead.
func CaptureStdout(t *testing.T, f func()) string {
	t.Helper()
	return capture(t, os.Stdout, f)
}

// CaptureStderr captures the output of a function that writes to stderr.
func CaptureStderr(t *testing.T, f func()) string {
	t.Helper()
	return capture(t, os.Stderr, f)
}

// CaptureStdoutAndStderr captures the output of a function that writes to
// stdout and stderr.
func CaptureStdoutAndStderr(t *testing.T, f func()) (o, e string) {
	t.Helper()

	oldout, olderr := os.Stdout, os.Stderr
	rOut, wOut, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	rErr, wErr, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout, os.Stderr = wOut, wErr

	outCh, errCh := make(chan string), make(chan string)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		var buf bytes.Buffer
		wg.Done()
		if _, err := io.Copy(&buf, rOut); err != nil {
			t.Fatal(err)
		}
		outCh <- buf.String()
	}()

	go func() {
		var buf bytes.Buffer
		wg.Done()
		if _, err := io.Copy(&buf, rErr); err != nil {
			t.Fatal(err)
		}
		errCh <- buf.String()
	}()

	wg.Wait()
	f()

	if err := wOut.Close(); err != nil {
		t.Fatal(err)
	}
	if err := wErr.Close(); err != nil {
		t.Fatal(err)
	}

	os.Stdout, os.Stderr = oldout, olderr
	stdout, stderr := <-outCh, <-errCh
	return stdout, stderr
}

func capture(t *testing.T, out *os.File, f func()) string {
	old := out
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	*out = *w

	f()

	w.Close()
	*out = *old

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	if err != nil {
		t.Fatal(err)
	}

	return buf.String()
}
