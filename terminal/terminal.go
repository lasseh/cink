package terminal

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/creack/pty"
	"github.com/lasseh/cink/highlighter"
	"golang.org/x/term"
)

const (
	readBufferSize = 32 * 1024 // Size of the read buffer from PTY

	// partialWordDelay is how long a word cut off at the end of a PTY read
	// waits for its second half. Typed echo arrives one character per read,
	// so this must stay below what a user notices.
	partialWordDelay = 15 * time.Millisecond
)

var (
	debug   bool
	debugMu sync.RWMutex
)

// SetDebug enables or disables debug output to stderr
func SetDebug(enabled bool) {
	debugMu.Lock()
	defer debugMu.Unlock()
	debug = enabled
}

// IsDebug returns whether debug mode is enabled
func IsDebug() bool {
	debugMu.RLock()
	defer debugMu.RUnlock()
	return debug
}

// Terminal wraps a command in a PTY and applies syntax highlighting to its output.
type Terminal struct {
	cmd         *exec.Cmd
	pty         *os.File
	highlighter *highlighter.Highlighter
}

// New creates a new Terminal for the given command
func New(name string, args ...string) *Terminal {
	cmd := exec.Command(name, args...)
	return &Terminal{
		cmd:         cmd,
		highlighter: highlighter.New(),
	}
}

// SetTheme changes the highlighting theme
func (t *Terminal) SetTheme(theme *highlighter.Theme) {
	t.highlighter.SetTheme(theme)
}

// SetEnabled enables or disables highlighting
func (t *Terminal) SetEnabled(enabled bool) {
	if enabled {
		t.highlighter.Enable()
	} else {
		t.highlighter.Disable()
	}
}

// Run starts the command and processes its output with highlighting.
func (t *Terminal) Run() error {
	// Start the command with a PTY
	ptmx, err := pty.Start(t.cmd)
	if err != nil {
		return fmt.Errorf("starting pty: %w", err)
	}
	t.pty = ptmx
	defer func() {
		if err := ptmx.Close(); err != nil && IsDebug() {
			fmt.Fprintf(os.Stderr, "[DEBUG] Error closing pty: %v\n", err)
		}
	}()

	// Handle terminal resize with proper cleanup
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGWINCH)
	sigDone := make(chan struct{})
	go func() {
		defer close(sigDone)
		for range sigCh {
			if err := pty.InheritSize(os.Stdin, ptmx); err != nil && IsDebug() {
				fmt.Fprintf(os.Stderr, "[DEBUG] Error resizing pty: %v\n", err)
			}
		}
	}()
	// Cleanup signal handler when done
	defer func() {
		signal.Stop(sigCh)
		close(sigCh)
		<-sigDone // Wait for goroutine to exit
	}()

	// Trigger initial resize
	sigCh <- syscall.SIGWINCH

	// Put terminal into raw mode
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("setting raw mode: %w", err)
	}
	defer func() {
		if err := term.Restore(int(os.Stdin.Fd()), oldState); err != nil {
			fmt.Fprintf(os.Stderr, "cink: error restoring terminal: %v\n", err)
		}
	}()

	// Create channel for coordination
	done := make(chan struct{})

	// Copy stdin to PTY
	go func() {
		if _, err := io.Copy(ptmx, os.Stdin); err != nil && IsDebug() {
			fmt.Fprintf(os.Stderr, "[DEBUG] Error copying stdin: %v\n", err)
		}
	}()

	// Copy PTY to stdout with highlighting
	go func() {
		t.processOutput(ptmx, os.Stdout)
		close(done)
	}()

	// Wait for command to finish
	<-done
	if err := t.cmd.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return &ExitError{Code: exitErr.ExitCode()}
		}
		return fmt.Errorf("command finished: %w", err)
	}
	return nil
}

// ExitError represents a child process that exited with a non-zero status.
type ExitError struct {
	Code int
}

func (e *ExitError) Error() string {
	return fmt.Sprintf("exit status %d", e.Code)
}

// processOutput reads from the PTY and writes highlighted output.
// Both complete lines and partial lines (prompts) are highlighted.
// Cursor control characters (like \r) are preserved to allow command-line editing.
func (t *Terminal) processOutput(r io.Reader, w io.Writer) {
	if IsDebug() {
		w = debugWriter{w}
	}
	stream := highlighter.NewStream(w, t.highlighter)
	stream.Force = true
	stream.FlushDelay = partialWordDelay

	buf := make([]byte, readBufferSize)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			if IsDebug() {
				fmt.Fprintf(os.Stderr, "\n[DEBUG] Read %d bytes: %q\n", n, buf[:n])
			}
			if _, err := stream.Write(buf[:n]); err != nil && IsDebug() {
				fmt.Fprintf(os.Stderr, "[DEBUG] Write error: %v\n", err)
			}
		}

		if err != nil {
			if IsDebug() && err != io.EOF {
				fmt.Fprintf(os.Stderr, "[DEBUG] Read error: %v\n", err)
			}
			break
		}
	}

	if err := stream.Flush(); err != nil && IsDebug() {
		fmt.Fprintf(os.Stderr, "[DEBUG] Write error: %v\n", err)
	}
}

// debugWriter logs everything written to the terminal.
type debugWriter struct{ w io.Writer }

func (d debugWriter) Write(p []byte) (int, error) {
	fmt.Fprintf(os.Stderr, "[DEBUG] Write: %q\n", p)
	return d.w.Write(p)
}
