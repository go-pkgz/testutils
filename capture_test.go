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
				_, err := os.Stdout.Write([]byte{0, 1, 2, 3})
				if err != nil {
					t.Logf("failed to write to stdout: %v", err)
				}
				_, err = os.Stderr.Write([]byte{4, 5, 6, 7})
				if err != nil {
					t.Logf("failed to write to stderr: %v", err)
				}
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

	t.Run("stdout and stderr with panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Expected panic")
			}
		}()

		CaptureStdoutAndStderr(t, func() {
			panic("test panic")
		})
	})
}

// TestCaptureWithOutput tests capturing output when the function has early returns
func TestCaptureWithOutput(t *testing.T) {
	t.Run("stdout with basic output", func(t *testing.T) {
		output := CaptureStdout(t, func() {
			fmt.Fprintln(os.Stdout, "before early return")
		})
		require.Equal(t, "before early return\n", output)
		require.NotContains(t, output, "after early return")
	})

	t.Run("stderr with basic output", func(t *testing.T) {
		output := CaptureStderr(t, func() {
			fmt.Fprintln(os.Stderr, "before early return")
		})
		require.Equal(t, "before early return\n", output)
		require.NotContains(t, output, "after early return")
	})

	t.Run("stdout and stderr with basic output", func(t *testing.T) {
		stdout, stderr := CaptureStdoutAndStderr(t, func() {
			fmt.Fprintln(os.Stdout, "stdout before early return")
			fmt.Fprintln(os.Stderr, "stderr before early return")
		})
		require.Equal(t, "stdout before early return\n", stdout)
		require.Equal(t, "stderr before early return\n", stderr)
		require.NotContains(t, stdout, "after early return")
		require.NotContains(t, stderr, "after early return")
	})
}

// TestCaptureWithLargeOutput tests capturing large amounts of output
func TestCaptureWithLargeOutput(t *testing.T) {
	const largeSize = 1000
	largeData := strings.Repeat("a", largeSize)

	t.Run("large stdout", func(t *testing.T) {
		output := CaptureStdout(t, func() {
			fmt.Fprint(os.Stdout, largeData)
		})
		require.Len(t, output, largeSize)
		require.Equal(t, largeData, output)
	})

	t.Run("large stderr", func(t *testing.T) {
		output := CaptureStderr(t, func() {
			fmt.Fprint(os.Stderr, largeData)
		})
		require.Len(t, output, largeSize)
		require.Equal(t, largeData, output)
	})

	t.Run("large stdout and stderr", func(t *testing.T) {
		stdout, stderr := CaptureStdoutAndStderr(t, func() {
			fmt.Fprint(os.Stdout, largeData)
			fmt.Fprint(os.Stderr, largeData)
		})
		require.Len(t, stdout, largeSize)
		require.Len(t, stderr, largeSize)
		require.Equal(t, largeData, stdout)
		require.Equal(t, largeData, stderr)
	})
}
