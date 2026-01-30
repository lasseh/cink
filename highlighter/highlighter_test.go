package highlighter

import (
	"strings"
	"testing"

	"github.com/lasseh/cink/lexer"
)

func TestNew(t *testing.T) {
	h := New()
	if h == nil {
		t.Fatal("New() returned nil")
	}
	if !h.IsEnabled() {
		t.Error("new highlighter should be enabled by default")
	}
}

func TestNewWithTheme(t *testing.T) {
	theme := MonokaiTheme()
	h := NewWithTheme(theme)
	if h == nil {
		t.Fatal("NewWithTheme() returned nil")
	}
	if !h.IsEnabled() {
		t.Error("new highlighter should be enabled by default")
	}
}

func TestSetTheme(t *testing.T) {
	h := New()
	h.SetTheme(NordTheme())
}

func TestEnableDisable(t *testing.T) {
	h := New()

	if !h.IsEnabled() {
		t.Error("should be enabled by default")
	}

	h.Disable()
	if h.IsEnabled() {
		t.Error("should be disabled after Disable()")
	}

	h.Enable()
	if !h.IsEnabled() {
		t.Error("should be enabled after Enable()")
	}
}

func TestToggle(t *testing.T) {
	h := New()

	if !h.IsEnabled() {
		t.Error("should be enabled by default")
	}

	result := h.Toggle()
	if result {
		t.Error("Toggle() should return false after disabling")
	}

	result = h.Toggle()
	if !result {
		t.Error("Toggle() should return true after enabling")
	}
}

func TestHighlightEmpty(t *testing.T) {
	h := New()
	result := h.Highlight("")
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestHighlightDisabled(t *testing.T) {
	h := New()
	h.Disable()

	input := "interface GigabitEthernet0/0/0"
	result := h.Highlight(input)
	if result != input {
		t.Errorf("disabled highlighter should return input unchanged, got %q", result)
	}
}

func TestHighlightNonCisco(t *testing.T) {
	h := New()

	input := "Hello, this is just some random text"
	result := h.Highlight(input)

	if result != input {
		t.Errorf("non-Cisco text should be returned unchanged")
	}
}

func TestHighlightBasic(t *testing.T) {
	h := New()

	input := "interface GigabitEthernet0/0/0"
	result := h.Highlight(input)

	if !strings.Contains(result, "\033[") {
		t.Error("highlighted output should contain ANSI escape codes")
	}

	stripped := StripANSI(result)
	if stripped != input {
		t.Errorf("stripped output should match input, got %q", stripped)
	}
}

func TestHighlightLines(t *testing.T) {
	h := New()

	input := []string{
		"hostname core-router-01",
		"interface GigabitEthernet0/0/0",
	}
	result := h.HighlightLines(input)

	if len(result) != len(input) {
		t.Fatalf("expected %d lines, got %d", len(input), len(result))
	}

	for i, line := range result {
		if !HasANSI(line) {
			t.Errorf("line %d should have ANSI codes", i)
		}
	}
}

func TestLooksLikeCisco(t *testing.T) {
	h := New()

	positives := []string{
		"hostname core-router-01",
		"interface GigabitEthernet0/0/0",
		"router ospf 1",
		"ip address 10.0.0.1 255.255.255.0",
		"switchport mode access",
		"no shutdown",
		"line vty 0 15",
		"access-list 100 permit ip any any",
		"Router>",
		"Router#",
		"Router(config)#",
	}

	for _, input := range positives {
		if !h.looksLikeCisco(input) {
			t.Errorf("should recognize %q as Cisco config", input)
		}
	}

	negatives := []string{
		"Hello world",
		"This is plain text",
		"SELECT * FROM users",
		"function main() {}",
		"import os",
	}

	for _, input := range negatives {
		if h.looksLikeCisco(input) {
			t.Errorf("should NOT recognize %q as Cisco config", input)
		}
	}
}

func TestStripANSI(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"plain text", "plain text"},
		{"\033[31mred\033[0m", "red"},
		{"\033[1m\033[94mblue\033[0m", "blue"},
		{"\033[38;5;208morange\033[0m", "orange"},
		{"no codes here", "no codes here"},
		{"", ""},
		{"\033[Khello", "hello"},
		{"\033[2Khello", "hello"},
		{"\033[Ahello", "hello"},
		{"\033[1;1Hhello", "hello"},
		{"before\033[Kafter", "beforeafter"},
	}

	for _, tt := range tests {
		result := StripANSI(tt.input)
		if result != tt.expected {
			t.Errorf("StripANSI(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestExtractSegments(t *testing.T) {
	tests := []struct {
		input         string
		expectedCount int
		description   string
	}{
		{"plain text", 1, "plain text has one segment"},
		{"\033[Khello", 2, "clear + text"},
		{"\033[31mred\033[0m", 3, "color + text + reset"},
		{"hello\033[Kworld", 3, "text + clear + text"},
		{"", 0, "empty string"},
	}

	for _, tt := range tests {
		result := extractSegments(tt.input)
		if len(result) != tt.expectedCount {
			t.Errorf("extractSegments(%q): expected %d segments, got %d (%s)",
				tt.input, tt.expectedCount, len(result), tt.description)
		}
	}
}

func TestHighlightForcedPreservesEscapeSequences(t *testing.T) {
	h := New()

	input := "\033[KRouter> show ip route"

	result := h.HighlightForced(input)

	if !strings.HasPrefix(result, "\033[K") {
		t.Errorf("HighlightForced should preserve cursor control sequences, got: %q", result)
	}

	if !strings.Contains(result, "\033[") {
		t.Error("result should contain ANSI codes")
	}
}

func TestHasANSI(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"plain text", false},
		{"\033[31mred\033[0m", true},
		{"\033[1mtext", true},
		{"no escape", false},
		{"", false},
	}

	for _, tt := range tests {
		result := HasANSI(tt.input)
		if result != tt.expected {
			t.Errorf("HasANSI(%q) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestDefaultTheme(t *testing.T) {
	theme := DefaultTheme()
	if theme == nil {
		t.Fatal("DefaultTheme() returned nil")
	}

	tokenTypes := []lexer.TokenType{
		lexer.TokenCommand,
		lexer.TokenSection,
		lexer.TokenProtocol,
		lexer.TokenAction,
		lexer.TokenInterface,
		lexer.TokenIPv4,
		lexer.TokenIPv4Prefix,
		lexer.TokenString,
		lexer.TokenComment,
		lexer.TokenNegation,
	}

	for _, tt := range tokenTypes {
		color := theme.GetColor(tt)
		if color == "" {
			t.Errorf("DefaultTheme should have color for %v", tt)
		}
	}
}

func TestAllThemes(t *testing.T) {
	themes := []struct {
		name  string
		theme *Theme
	}{
		{"TokyoNight", TokyoNightTheme()},
		{"Vibrant", VibrantTheme()},
		{"Solarized", SolarizedDarkTheme()},
		{"Monokai", MonokaiTheme()},
		{"Nord", NordTheme()},
		{"Catppuccin", CatppuccinMochaTheme()},
		{"Dracula", DraculaTheme()},
		{"Gruvbox", GruvboxDarkTheme()},
		{"OneDark", OneDarkTheme()},
	}

	for _, tt := range themes {
		t.Run(tt.name, func(t *testing.T) {
			if tt.theme == nil {
				t.Fatalf("%s returned nil", tt.name)
			}
			if tt.theme.GetColor(lexer.TokenCommand) == "" {
				t.Error("should have command color")
			}
			if tt.theme.GetColor(lexer.TokenNegation) == "" {
				t.Error("should have negation color")
			}
		})
	}
}

func TestDefaultThemeIsTokyoNight(t *testing.T) {
	defaultTheme := DefaultTheme()
	tokyoTheme := TokyoNightTheme()

	if defaultTheme.GetColor(lexer.TokenCommand) != tokyoTheme.GetColor(lexer.TokenCommand) {
		t.Error("DefaultTheme should match TokyoNightTheme")
	}
}

func TestThemeSetColor(t *testing.T) {
	theme := DefaultTheme()

	newColor := "\033[35m"
	theme.SetColor(lexer.TokenCommand, newColor)

	if theme.GetColor(lexer.TokenCommand) != newColor {
		t.Error("SetColor should update the color")
	}
}

func TestThemeGetColorUnknown(t *testing.T) {
	theme := DefaultTheme()

	color := theme.GetColor(lexer.TokenType(999))
	if color != "" {
		t.Error("unknown token type should return empty string")
	}
}

func TestColor256(t *testing.T) {
	result := Color256(208)
	expected := "\033[38;5;208m"
	if result != expected {
		t.Errorf("Color256(208) = %q, want %q", result, expected)
	}
}

func TestRGB(t *testing.T) {
	result := RGB(255, 128, 0)
	expected := "\033[38;2;255;128;0m"
	if result != expected {
		t.Errorf("RGB(255,128,0) = %q, want %q", result, expected)
	}
}

func TestHighlightPreservesContent(t *testing.T) {
	h := New()

	configs := []string{
		"hostname core-router-01",
		"interface GigabitEthernet0/0/0",
		"ip address 192.168.1.1 255.255.255.0",
		"router bgp 65001",
		"no shutdown",
	}

	for _, config := range configs {
		result := h.Highlight(config)
		stripped := StripANSI(result)
		if stripped != config {
			t.Errorf("content not preserved:\ninput:    %q\nstripped: %q", config, stripped)
		}
	}
}

func TestHighlightCiscoConfig(t *testing.T) {
	h := New()

	input := `!
hostname router
!
interface GigabitEthernet0/0/0
 ip address 10.0.0.1 255.255.255.0
 no shutdown
!`

	result := h.Highlight(input)

	if !HasANSI(result) {
		t.Error("Cisco config should be highlighted")
	}

	stripped := StripANSI(result)
	if stripped != input {
		t.Errorf("content not preserved")
	}
}

func TestThemeByName(t *testing.T) {
	names := ThemeNames()
	for _, name := range names {
		theme := ThemeByName(name)
		if theme == nil {
			t.Errorf("ThemeByName(%q) returned nil", name)
		}
	}

	// Unknown name should return default
	theme := ThemeByName("nonexistent")
	if theme == nil {
		t.Error("ThemeByName with unknown name should return default, not nil")
	}
}
