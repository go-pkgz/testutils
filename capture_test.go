package testutils

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestCaptureStdout(t *testing.T) {
	tests := []struct {
		name string
		want string
		f    func()
	}{
		{
			name: "simple output",
			want: "hello world\n",
			f: func() {
				fmt.Fprintf(os.Stdout, "hello world\n")
			},
		},
		{
			name: "multiple lines",
			want: "line1\nline2\n",
			f: func() {
				fmt.Fprintln(os.Stdout, "line1")
				fmt.Fprintln(os.Stdout, "line2")
			},
		},
		{
			name: "empty output",
			want: "",
			f:    func() {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CaptureStdout(t, tt.f)
			if got != tt.want {
				t.Errorf("CaptureStdout() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCaptureStderr(t *testing.T) {
	tests := []struct {
		name string
		want string
		f    func()
	}{
		{
			name: "simple output",
			want: "hello world\n",
			f: func() {
				fmt.Fprintf(os.Stderr, "hello world\n")
			},
		},
		{
			name: "multiple lines",
			want: "line1\nline2\n",
			f: func() {
				fmt.Fprintln(os.Stderr, "line1")
				fmt.Fprintln(os.Stderr, "line2")
			},
		},
		{
			name: "empty output",
			want: "",
			f:    func() {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CaptureStderr(t, tt.f)
			if got != tt.want {
				t.Errorf("CaptureStderr() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCaptureStdoutAndStderr(t *testing.T) {
	tests := []struct {
		name    string
		wantOut string
		wantErr string
		f       func()
	}{
		{
			name:    "both outputs",
			wantOut: "stdout\n",
			wantErr: "stderr\n",
			f: func() {
				fmt.Fprintln(os.Stdout, "stdout")
				fmt.Fprintln(os.Stderr, "stderr")
			},
		},
		{
			name:    "only stdout",
			wantOut: "stdout\n",
			wantErr: "",
			f: func() {
				fmt.Fprintln(os.Stdout, "stdout")
			},
		},
		{
			name:    "only stderr",
			wantOut: "",
			wantErr: "stderr\n",
			f: func() {
				fmt.Fprintln(os.Stderr, "stderr")
			},
		},
		{
			name:    "large output",
			wantOut: strings.Repeat("a", 100000) + "\n",
			wantErr: strings.Repeat("b", 100000) + "\n",
			f: func() {
				fmt.Fprintln(os.Stdout, strings.Repeat("a", 100000))
				fmt.Fprintln(os.Stderr, strings.Repeat("b", 100000))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOut, gotErr := CaptureStdoutAndStderr(t, tt.f)
			if gotOut != tt.wantOut {
				t.Errorf("CaptureStdoutAndStderr() stdout = %q, want %q", gotOut, tt.wantOut)
			}
			if gotErr != tt.wantErr {
				t.Errorf("CaptureStdoutAndStderr() stderr = %q, want %q", gotErr, tt.wantErr)
			}
		})
	}
}
