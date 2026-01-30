package highlighter

import (
	"bytes"
	"strings"
	"sync"

	"github.com/lasseh/cink/lexer"
)

// Highlight is a convenience function that highlights Cisco config/output using the default theme.
func Highlight(input string) string {
	return New().Highlight(input)
}

// Highlighter applies ANSI color codes to Cisco IOS/IOS-XE configuration and show command output.
// It supports multiple color themes and can be toggled on/off at runtime.
// All methods are safe for concurrent use.
type Highlighter struct {
	theme   *Theme
	enabled bool
	mu      sync.RWMutex
}

// New creates a new Highlighter with the default theme (Tokyo Night).
func New() *Highlighter {
	return &Highlighter{
		theme:   DefaultTheme(),
		enabled: true,
	}
}

// NewWithTheme creates a new Highlighter with a specific theme
func NewWithTheme(theme *Theme) *Highlighter {
	return &Highlighter{
		theme:   theme,
		enabled: true,
	}
}

// SetTheme changes the highlighting theme.
func (h *Highlighter) SetTheme(theme *Theme) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.theme = theme
}

// Enable turns highlighting on.
func (h *Highlighter) Enable() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.enabled = true
}

// Disable turns highlighting off.
func (h *Highlighter) Disable() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.enabled = false
}

// IsEnabled returns whether highlighting is enabled.
func (h *Highlighter) IsEnabled() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.enabled
}

// Toggle switches highlighting on/off and returns the new state.
func (h *Highlighter) Toggle() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.enabled = !h.enabled
	return h.enabled
}

// Highlight applies syntax highlighting to the input text.
// Returns input unchanged if highlighting is disabled, input is empty,
// or input doesn't look like Cisco config/output (uses heuristic detection).
func (h *Highlighter) Highlight(input string) string {
	if !h.IsEnabled() || input == "" {
		return input
	}

	cleaned := StripANSI(input)

	if !h.looksLikeCisco(cleaned) {
		return input
	}

	return h.highlightTokensCleaned(cleaned)
}

// HighlightForced applies syntax highlighting without checking if input looks like Cisco.
func (h *Highlighter) HighlightForced(input string) string {
	if !h.IsEnabled() || input == "" {
		return input
	}
	return h.highlightTokens(input)
}

// highlightTokens tokenizes and colorizes the input while preserving cursor control sequences
func (h *Highlighter) highlightTokens(input string) string {
	segments := extractSegments(input)

	var buf bytes.Buffer
	for _, seg := range segments {
		if seg.isEscape {
			buf.WriteString(seg.text)
		} else {
			highlighted := h.highlightTokensCleaned(seg.text)
			buf.WriteString(highlighted)
		}
	}
	return buf.String()
}

// highlightTokensCleaned tokenizes and colorizes already-cleaned input
func (h *Highlighter) highlightTokensCleaned(cleaned string) string {
	lex := lexer.New(cleaned)
	tokens := lex.Tokenize()
	return h.renderTokens(tokens)
}

// renderTokens applies theme colors to a slice of tokens and returns the colorized string
func (h *Highlighter) renderTokens(tokens []lexer.Token) string {
	h.mu.RLock()
	theme := h.theme
	h.mu.RUnlock()

	var buf bytes.Buffer
	for _, token := range tokens {
		color := theme.GetColor(token.Type)
		if color != "" {
			buf.WriteString(color)
			buf.WriteString(token.Value)
			buf.WriteString(Reset)
		} else {
			buf.WriteString(token.Value)
		}
	}
	return buf.String()
}

// HighlightLines highlights multiple lines preserving line structure
func (h *Highlighter) HighlightLines(lines []string) []string {
	result := make([]string, len(lines))
	for i, line := range lines {
		result[i] = h.Highlight(line)
	}
	return result
}

// Cisco-specific keyword patterns for quick detection
var ciscoSpecificKeywords = []string{
	"switchport mode", "ip address ", "ip route ",
	"router ospf", "router bgp", "router eigrp",
	"transport input", "exec-timeout",
	"channel-group", "spanning-tree portfast",
}

// looksLikeCisco performs a quick check to see if text appears to be Cisco config or show output
func (h *Highlighter) looksLikeCisco(input string) bool {
	// Check for Cisco CLI prompts
	if isPromptLine(input) {
		return true
	}

	lower := strings.ToLower(input)

	if hasConfigIndicators(lower) {
		return true
	}

	if hasShowIndicators(lower) {
		return true
	}

	// Check for ! section separators (lines with just "!")
	if hasCiscoSeparators(input) {
		return true
	}

	// Check absence of JunOS indicators (helps disambiguate)
	// If we see braces or semicolons, it's probably not Cisco
	if hasCiscoKeywords(lower) {
		return true
	}

	return false
}

// isPromptLine checks if the input looks like a Cisco CLI prompt
func isPromptLine(input string) bool {
	if lexer.IsPrompt(input) {
		return true
	}

	trimmed := strings.TrimSpace(input)
	// Quick check for hostname> or hostname# patterns
	if len(trimmed) > 1 {
		last := trimmed[len(trimmed)-1]
		if last == '>' || last == '#' {
			// Check that everything before prompt char is valid hostname chars or mode
			prefix := trimmed[:len(trimmed)-1]
			// Remove mode suffix like (config-if)
			if idx := strings.LastIndex(prefix, ")"); idx >= 0 {
				if pIdx := strings.LastIndex(prefix, "("); pIdx >= 0 {
					prefix = prefix[:pIdx]
				}
			}
			if len(prefix) > 0 && isValidHostname(prefix) {
				return true
			}
		}
	}

	return false
}

// isValidHostname checks if a string looks like a valid hostname
func isValidHostname(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, ch := range s {
		if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') || ch == '-' || ch == '.' || ch == '_') {
			return false
		}
	}
	return true
}

// hasConfigIndicators checks for common Cisco config keywords/patterns
func hasConfigIndicators(lower string) bool {
	for _, indicator := range lexer.ConfigIndicators {
		if strings.Contains(lower, indicator) {
			return true
		}
	}
	return false
}

// hasShowIndicators checks for show command output patterns
func hasShowIndicators(lower string) bool {
	for _, indicator := range lexer.ShowIndicators {
		if strings.Contains(lower, indicator) {
			return true
		}
	}
	return false
}

// hasCiscoSeparators checks for ! section separators
func hasCiscoSeparators(input string) bool {
	bangCount := 0
	i := 0
	for i < len(input) {
		// Find start of line (or beginning of input)
		lineStart := i
		// Scan to end of line
		end := strings.IndexByte(input[i:], '\n')
		var line string
		if end == -1 {
			line = input[lineStart:]
			i = len(input)
		} else {
			line = input[lineStart : lineStart+end]
			i = lineStart + end + 1
		}
		if strings.TrimSpace(line) == "!" {
			bangCount++
			if bangCount >= 2 {
				return true
			}
		}
	}
	return false
}

// hasCiscoKeywords checks for Cisco-specific command patterns
func hasCiscoKeywords(lower string) bool {
	for _, kw := range ciscoSpecificKeywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

// HighlightShowOutput highlights show command output specifically using show mode.
func (h *Highlighter) HighlightShowOutput(input string) string {
	if !h.IsEnabled() || input == "" {
		return input
	}

	lex := lexer.New(input)
	lex.SetParseMode(lexer.ParseModeShow)
	tokens := lex.Tokenize()
	return h.renderTokens(tokens)
}

// segment represents either an escape sequence or text content
type segment struct {
	text     string
	isEscape bool
}

// CSI sequence byte range constants
const (
	csiParamStart = 0x20
	csiParamEnd   = 0x3F
	csiFinalStart = 0x40
	csiFinalEnd   = 0x7E
	csiIntermEnd  = 0x2F
	escapeChar    = '\033'
	csiBracket    = '['
)

func isCSIParamByte(b byte) bool {
	return b >= csiParamStart && b <= csiParamEnd
}

func isCSIFinalByte(b byte) bool {
	return b >= csiFinalStart && b <= csiFinalEnd
}

func isCSIIntermediateByte(b byte) bool {
	return b >= csiParamStart && b <= csiIntermEnd
}

func skipCSISequence(input string, i int) int {
	for i < len(input) && isCSIParamByte(input[i]) {
		i++
	}
	if i < len(input) && isCSIFinalByte(input[i]) {
		i++
	}
	return i
}

func skipOtherEscapeSequence(input string, i int) int {
	for i < len(input) && isCSIIntermediateByte(input[i]) {
		i++
	}
	if i < len(input) {
		i++
	}
	return i
}

// extractSegments splits input into escape sequences and text segments
func extractSegments(input string) []segment {
	var segments []segment
	var textBuf bytes.Buffer
	i := 0

	for i < len(input) {
		if input[i] == escapeChar && i+1 < len(input) {
			if input[i+1] == csiBracket {
				// CSI sequence: \033[...
				if textBuf.Len() > 0 {
					segments = append(segments, segment{text: textBuf.String(), isEscape: false})
					textBuf.Reset()
				}

				start := i
				i = skipCSISequence(input, i+2)
				segments = append(segments, segment{text: input[start:i], isEscape: true})
				continue
			}
			// Non-CSI escape sequence (OSC, charset selection, etc.)
			if textBuf.Len() > 0 {
				segments = append(segments, segment{text: textBuf.String(), isEscape: false})
				textBuf.Reset()
			}

			start := i
			i = skipOtherEscapeSequence(input, i+1)
			segments = append(segments, segment{text: input[start:i], isEscape: true})
			continue
		}
		textBuf.WriteByte(input[i])
		i++
	}

	if textBuf.Len() > 0 {
		segments = append(segments, segment{text: textBuf.String(), isEscape: false})
	}

	return segments
}

// StripANSI removes ANSI escape codes from text.
func StripANSI(input string) string {
	var buf bytes.Buffer
	i := 0

	for i < len(input) {
		if input[i] == escapeChar && i+1 < len(input) && input[i+1] == csiBracket {
			i = skipCSISequence(input, i+2)
			continue
		}
		if input[i] == escapeChar {
			i = skipOtherEscapeSequence(input, i+1)
			continue
		}
		buf.WriteByte(input[i])
		i++
	}

	return buf.String()
}

// HasANSI checks if the input contains ANSI escape codes
func HasANSI(input string) bool {
	return strings.Contains(input, "\033[")
}
