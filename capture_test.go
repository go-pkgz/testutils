package testutils

import (
	"fmt"
	"os"
	"testing"
)

func TestCaptureStdout(t *testing.T) {
	want := "hello world\n"
	got := CaptureStdout(t, func() {
		fmt.Fprintf(os.Stdout, want)
	})
	if want != got {
		t.Errorf("want %q, got %q", want, got)
	}
}

func TestCaptureStderr(t *testing.T) {
	want := "hello world\n"
	got := CaptureStderr(t, func() {
		fmt.Fprintf(os.Stderr, want)
	})
	if want != got {
		t.Errorf("want %q, got %q", want, got)
	}
}

func TestCaptureStdoutAndStderr(t *testing.T) {
	wantOut := "hello world\n"
	wantErr := "hello world\n"
	gotOut, gotErr := CaptureStdoutAndStderr(t, func() {
		fmt.Fprintf(os.Stdout, wantOut)
		fmt.Fprintf(os.Stderr, wantErr)
	})
	if wantOut != gotOut {
		t.Errorf("want %q, got %q", wantOut, gotOut)
	}
	if wantErr != gotErr {
		t.Errorf("want %q, got %q", wantErr, gotErr)
	}
}
