package highlighter

import (
	"bytes"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/lasseh/cink/lexer"
)

// renderAs highlights each line of text on its own in the given mode.
func renderAs(h *Highlighter, text string, mode lexer.ParseMode) string {
	var out strings.Builder
	for _, line := range strings.SplitAfter(text, "\n") {
		if line == "" {
			continue
		}
		lex := lexer.New(line)
		lex.SetParseMode(mode)
		out.WriteString(h.renderTokens(lex.Tokenize()))
	}
	return out.String()
}

// feed writes chunks to a new stream and flushes it.
func feed(t *testing.T, h *Highlighter, force bool, chunks ...string) string {
	t.Helper()
	var buf bytes.Buffer
	s := NewStream(&buf, h)
	s.Force = force
	for _, c := range chunks {
		if _, err := s.Write([]byte(c)); err != nil {
			t.Fatalf("Write: %v", err)
		}
	}
	if err := s.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}
	return buf.String()
}

const intBrief = "Interface              IP-Address      OK? Method Status                Protocol\n" +
	"GigabitEthernet0/0     10.0.0.1        YES manual up                    up\n" +
	"GigabitEthernet0/1     unassigned      YES unset  administratively down down\n"

func TestStreamDetectsShowBlockFromFollowingLines(t *testing.T) {
	h := New()
	got := feed(t, h, false, intBrief)
	if want := renderAs(h, intBrief, lexer.ParseModeShow); got != want {
		t.Errorf("header should be lexed as show output\ngot  %q\nwant %q", got, want)
	}
}

func TestStreamKeepsModeAcrossWrites(t *testing.T) {
	h := New()
	// "Loopback0 is down" alone does not look like show output.
	tail := "Loopback0 is down\n"
	got := feed(t, h, false, intBrief, tail)
	want := renderAs(h, intBrief+tail, lexer.ParseModeShow)
	if got != want {
		t.Errorf("show mode should last until a prompt\ngot  %q\nwant %q", got, want)
	}
}

func TestStreamPromptResetsMode(t *testing.T) {
	h := New()
	config := "interface Loopback0\n shutdown\n"
	got := feed(t, h, true, intBrief, "Router#\n", config)
	want := renderAs(h, intBrief, lexer.ParseModeShow) +
		renderAs(h, "Router#\n", lexer.ParseModeConfig) +
		renderAs(h, config, lexer.ParseModeConfig)
	if got != want {
		t.Errorf("prompt should reset mode\ngot  %q\nwant %q", got, want)
	}
}

func TestStreamShowPromptStartsShowBlock(t *testing.T) {
	h := New()
	for _, cmd := range []string{"show", "sh", "SHOW"} {
		t.Run(cmd, func(t *testing.T) {
			prompt := "Router#" + cmd + " interfaces\n"
			got := feed(t, h, true, prompt, "Loopback0 is down\n")
			want := renderAs(h, prompt, lexer.ParseModeConfig) +
				renderAs(h, "Loopback0 is down\n", lexer.ParseModeShow)
			if got != want {
				t.Errorf("got  %q\nwant %q", got, want)
			}
		})
	}
}

func TestStreamJoinsWordSplitAcrossWrites(t *testing.T) {
	h := New()
	input := "interface GigabitEthernet0/1\n"
	whole := feed(t, h, true, input)
	split := feed(t, h, true, "interface GigabitEth", "ernet0/1\n")
	if split != whole {
		t.Errorf("split word lexed differently\ngot  %q\nwant %q", split, whole)
	}
}

type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func TestStreamFlushDelay(t *testing.T) {
	h := New()
	var out syncBuffer
	s := NewStream(&out, h)
	s.Force = true
	s.FlushDelay = 10 * time.Millisecond

	// A partial line ending in whitespace is written at once.
	if _, err := s.Write([]byte("Router# ")); err != nil {
		t.Fatal(err)
	}
	if StripANSI(out.String()) != "Router# " {
		t.Fatalf("partial line ending in space should be written at once, got %q", out.String())
	}

	// A partial line ending mid-word is written after FlushDelay.
	if _, err := s.Write([]byte("sh")); err != nil {
		t.Fatal(err)
	}
	if StripANSI(out.String()) != "Router# " {
		t.Fatalf("partial word should be held back, got %q", out.String())
	}
	deadline := time.Now().Add(time.Second)
	for StripANSI(out.String()) != "Router# sh" {
		if time.Now().After(deadline) {
			t.Fatalf("held-back word never flushed, got %q", out.String())
		}
		time.Sleep(time.Millisecond)
	}
}

func TestStreamPassesNonCiscoThrough(t *testing.T) {
	h := New()
	input := "there is no problem with the build\nwe are connected now\n"
	if got := feed(t, h, false, input); got != input {
		t.Errorf("non-Cisco text should pass through, got %q", got)
	}
}

func TestStreamStaysHighlightedOnceCisco(t *testing.T) {
	h := New()
	got := feed(t, h, false, "just a note\n", "hostname R1\n", "just a note\n")
	want := "just a note\n" + renderAs(h, "hostname R1\njust a note\n", lexer.ParseModeConfig)
	if got != want {
		t.Errorf("got  %q\nwant %q", got, want)
	}
}

func TestStreamDisabledPassesThrough(t *testing.T) {
	h := New()
	h.Disable()
	input := "interface GigabitEthernet0/1\nRouter#"
	if got := feed(t, h, true, input); got != input {
		t.Errorf("disabled stream should pass through, got %q", got)
	}
}

func TestStreamPreservesEscapeSequences(t *testing.T) {
	h := New()
	got := feed(t, h, true, "\033[KRouter> show ip route")
	if !strings.HasPrefix(got, "\033[K") {
		t.Errorf("escape sequence should be kept, got %q", got)
	}
	if StripANSI(got) != "Router> show ip route" {
		t.Errorf("text changed: %q", StripANSI(got))
	}
}
