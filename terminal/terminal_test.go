package terminal

import (
	"bytes"
	"io"
	"strings"
	"sync"
	"testing"

	"github.com/lasseh/cink/highlighter"
)

func TestSetDebug(t *testing.T) {
	// Reset to known state
	SetDebug(false)

	if IsDebug() {
		t.Error("expected debug to be false after SetDebug(false)")
	}

	SetDebug(true)
	if !IsDebug() {
		t.Error("expected debug to be true after SetDebug(true)")
	}

	SetDebug(false)
	if IsDebug() {
		t.Error("expected debug to be false after SetDebug(false)")
	}
}

func TestDebugConcurrency(t *testing.T) {
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			SetDebug(true)
			SetDebug(false)
		}()
		go func() {
			defer wg.Done()
			_ = IsDebug()
		}()
	}

	wg.Wait()
}

func TestNew(t *testing.T) {
	term := New("echo", "hello")
	if term == nil {
		t.Fatal("New() returned nil")
	}
	if term.cmd == nil {
		t.Error("cmd should not be nil")
	}
	if term.highlighter == nil {
		t.Error("highlighter should not be nil")
	}
	if !term.highlighter.IsEnabled() {
		t.Error("highlighting should be enabled by default")
	}
}

func TestSetTheme(t *testing.T) {
	term := New("echo", "test")
	theme := highlighter.MonokaiTheme()
	term.SetTheme(theme)
}

func TestSetEnabled(t *testing.T) {
	term := New("echo", "test")

	if !term.highlighter.IsEnabled() {
		t.Error("should be enabled by default")
	}

	term.SetEnabled(false)
	if term.highlighter.IsEnabled() {
		t.Error("should be disabled after SetEnabled(false)")
	}

	term.SetEnabled(true)
	if !term.highlighter.IsEnabled() {
		t.Error("should be enabled after SetEnabled(true)")
	}
}

func TestProcessOutputBasic(t *testing.T) {
	term := New("echo", "test")
	term.SetEnabled(false)

	input := "line1\nline2\nline3\n"
	reader := strings.NewReader(input)
	var output bytes.Buffer

	term.processOutput(reader, &output)

	if output.String() != input {
		t.Errorf("expected %q, got %q", input, output.String())
	}
}

func TestProcessOutputPartialLine(t *testing.T) {
	term := New("echo", "test")
	term.SetEnabled(false)

	input := "Router> "
	reader := strings.NewReader(input)
	var output bytes.Buffer

	term.processOutput(reader, &output)

	if output.String() != input {
		t.Errorf("expected %q, got %q", input, output.String())
	}
}

func TestProcessOutputLargeBuffer(t *testing.T) {
	term := New("echo", "test")
	term.SetEnabled(false)

	largeLine := strings.Repeat("x", 5000)
	reader := strings.NewReader(largeLine)
	var output bytes.Buffer

	term.processOutput(reader, &output)

	if output.String() != largeLine {
		t.Errorf("expected %d chars, got %d chars", len(largeLine), len(output.String()))
	}
}

func TestProcessOutputWithHighlighting(t *testing.T) {
	term := New("echo", "test")
	term.SetEnabled(true)

	input := "interface GigabitEthernet0/0/0\n"
	reader := strings.NewReader(input)
	var output bytes.Buffer

	term.processOutput(reader, &output)

	if !strings.Contains(output.String(), "\033[") {
		t.Error("output should contain ANSI escape codes when highlighting is enabled")
	}

	stripped := highlighter.StripANSI(output.String())
	if stripped != input {
		t.Errorf("stripped output %q should equal input %q", stripped, input)
	}
}

type mockReader struct {
	data    []byte
	pos     int
	errAt   int
	errOnce error
}

func (m *mockReader) Read(p []byte) (n int, err error) {
	if m.pos >= len(m.data) {
		return 0, io.EOF
	}
	if m.errAt >= 0 && m.pos >= m.errAt && m.errOnce != nil {
		err := m.errOnce
		m.errOnce = nil
		return 0, err
	}
	n = copy(p, m.data[m.pos:])
	m.pos += n
	return n, nil
}

func TestProcessOutputHandlesErrors(t *testing.T) {
	term := New("echo", "test")
	term.SetEnabled(false)

	reader := &mockReader{data: []byte("test\n"), errAt: -1}
	var output bytes.Buffer

	term.processOutput(reader, &output)

	if output.String() != "test\n" {
		t.Errorf("expected 'test\\n', got %q", output.String())
	}
}
