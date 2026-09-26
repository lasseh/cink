package highlighter

import (
	"bytes"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/lasseh/cink/lexer"
)

const (
	// maxPending is the longest partial line held back before it is written anyway.
	maxPending = 4096
	// detectSample caps the text passed to lexer.Detect, which samples less than this.
	detectSample = 512
)

// Stream highlights text written to it and forwards the result to an
// underlying writer. It keeps context across writes and lines:
//
//   - Unless Force is set, output passes through unchanged until a line
//     looks like Cisco. From then on every line is highlighted.
//   - The parse mode (config or show output) is detected at the start of a
//     block, looking ahead at the text already written, and lasts until the
//     next prompt line.
//   - A partial line is held back so a word split across writes is lexed
//     as one word. See FlushDelay.
//
// Escape sequences pass through untouched. Call Flush after the last Write.
// Set Force and FlushDelay before the first Write.
type Stream struct {
	// Force highlights every line without checking whether it looks like Cisco.
	Force bool

	// FlushDelay is how long a partial line that ends mid-word waits for the
	// rest of the word before it is written anyway. Zero holds partial lines
	// until the next Write or Flush.
	FlushDelay time.Duration

	w  io.Writer
	h  *Highlighter
	mu sync.Mutex

	pending []byte
	cisco   bool
	mode    lexer.ParseMode // ParseModeAuto until a block's mode is known
	timer   *time.Timer
	err     error
}

// NewStream returns a Stream that writes highlighted text to w using h.
func NewStream(w io.Writer, h *Highlighter) *Stream {
	return &Stream{w: w, h: h}
}

// Write highlights the complete lines in p and holds back a trailing
// partial line according to FlushDelay. It returns the first error from
// the underlying writer.
func (s *Stream) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.timer != nil {
		s.timer.Stop()
		s.timer = nil
	}

	if !s.h.IsEnabled() {
		s.write(string(s.pending))
		s.write(string(p))
		s.pending = s.pending[:0]
		return len(p), s.err
	}

	s.pending = append(s.pending, p...)
	start := 0
	for {
		i := bytes.IndexByte(s.pending[start:], '\n')
		if i < 0 {
			break
		}
		end := start + i + 1
		s.emit(string(s.pending[start:end]), s.pending[start:])
		start = end
	}
	s.pending = s.pending[:copy(s.pending, s.pending[start:])]

	switch {
	case len(s.pending) == 0:
	case len(s.pending) > maxPending, s.FlushDelay > 0 && !endsMidWord(s.pending):
		s.emitPending()
	case s.FlushDelay > 0:
		s.timer = time.AfterFunc(s.FlushDelay, s.flushTimer)
	}

	return len(p), s.err
}

// Flush writes any held-back partial line.
func (s *Stream) Flush() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.timer != nil {
		s.timer.Stop()
		s.timer = nil
	}
	s.emitPending()
	return s.err
}

func (s *Stream) flushTimer() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.timer = nil
	s.emitPending()
}

func (s *Stream) emitPending() {
	if len(s.pending) == 0 {
		return
	}
	s.emit(string(s.pending), s.pending)
	s.pending = s.pending[:0]
}

// emit writes one line, highlighted if the stream has seen Cisco text.
// sample is the text available from this line on, used for detection.
func (s *Stream) emit(line string, sample []byte) {
	if !s.Force && !s.cisco {
		if _, ok := lexer.Detect(StripANSI(line)); !ok {
			s.write(line)
			return
		}
		s.cisco = true
	}

	mode := s.mode
	if mode == lexer.ParseModeAuto {
		if len(sample) > detectSample {
			sample = sample[:detectSample]
		}
		detected, ok := lexer.Detect(StripANSI(string(sample)))
		mode = detected
		if ok {
			s.mode = detected
		}
	}

	var buf strings.Builder
	prompt := false
	for _, seg := range extractSegments(line) {
		if seg.isEscape {
			buf.WriteString(seg.text)
			continue
		}
		lex := lexer.New(seg.text)
		lex.SetParseMode(mode)
		tokens := lex.Tokenize()
		prompt = prompt || hasPrompt(tokens)
		buf.WriteString(s.h.renderTokens(tokens))
	}
	if prompt {
		s.mode = lexer.ParseModeAuto
	}
	s.write(buf.String())
}

func (s *Stream) write(text string) {
	if s.err != nil || text == "" {
		return
	}
	_, s.err = io.WriteString(s.w, text)
}

func hasPrompt(tokens []lexer.Token) bool {
	for _, tok := range tokens {
		if tok.Type == lexer.TokenPromptOper || tok.Type == lexer.TokenPromptConf {
			return true
		}
	}
	return false
}

func endsMidWord(b []byte) bool {
	switch b[len(b)-1] {
	case ' ', '\t', '\r', '\n':
		return false
	}
	return true
}
