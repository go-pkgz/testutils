package testutils

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
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
			require.Equal(t, tt.want, got, "CaptureStdout() returned unexpected output")
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
			require.Equal(t, tt.want, got, "CaptureStderr() returned unexpected output")
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
			name:    "empty function",
			wantOut: "",
			wantErr: "",
			f:       func() {},
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
		{
			name:    "binary data",
			wantOut: string([]byte{0, 1, 2, 3}),
			wantErr: string([]byte{4, 5, 6, 7}),
			f: func() {
				os.Stdout.Write([]byte{0, 1, 2, 3})
				os.Stderr.Write([]byte{4, 5, 6, 7})
			},
		},
		{
			name:    "concurrent writes",
			wantOut: "out1\nout2\nout3\n",
			wantErr: "err1\nerr2\nerr3\n",
			f: func() {
				var wg sync.WaitGroup
				wg.Add(6)

				go func() { defer wg.Done(); fmt.Fprintln(os.Stdout, "out1") }()
				go func() { defer wg.Done(); fmt.Fprintln(os.Stderr, "err1") }()
				go func() { defer wg.Done(); fmt.Fprintln(os.Stdout, "out2") }()
				go func() { defer wg.Done(); fmt.Fprintln(os.Stderr, "err2") }()
				go func() { defer wg.Done(); fmt.Fprintln(os.Stdout, "out3") }()
				go func() { defer wg.Done(); fmt.Fprintln(os.Stderr, "err3") }()

				wg.Wait()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOut, gotErr := CaptureStdoutAndStderr(t, tt.f)

			// for concurrent writes, we can't guarantee order
			if tt.name == "concurrent writes" {
				// check that all expected lines are present, ignoring order
				for _, line := range []string{"out1", "out2", "out3"} {
					require.Contains(t, gotOut, line)
				}
				for _, line := range []string{"err1", "err2", "err3"} {
					require.Contains(t, gotErr, line)
				}
			} else {
				require.Equal(t, tt.wantOut, gotOut, "CaptureStdoutAndStderr() stdout returned unexpected output")
				require.Equal(t, tt.wantErr, gotErr, "CaptureStdoutAndStderr() stderr returned unexpected output")
			}
		})
	}
}

// TestCaptureFunctionErrors tests edge cases for the capture functions
func TestCaptureFunctionErrors(t *testing.T) {
	t.Run("panic in function", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Expected panic")
			}
		}()

		CaptureStdout(t, func() {
			panic("test panic")
		})
	})

	t.Run("stderr with panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Expected panic")
			}
		}()

		CaptureStderr(t, func() {
			panic("test panic")
		})
	})
}
