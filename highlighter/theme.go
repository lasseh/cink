package highlighter

import (
	"strconv"
	"sync"

	"github.com/lasseh/cink/lexer"
)

// ANSI color codes
const (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Dim       = "\033[2m"
	Italic    = "\033[3m"
	Underline = "\033[4m"

	// Foreground colors
	Black   = "\033[30m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"

	// Bright foreground colors
	BrightBlack   = "\033[90m"
	BrightRed     = "\033[91m"
	BrightGreen   = "\033[92m"
	BrightYellow  = "\033[93m"
	BrightBlue    = "\033[94m"
	BrightMagenta = "\033[95m"
	BrightCyan    = "\033[96m"
	BrightWhite   = "\033[97m"

	// 256-color mode
	Color256Prefix = "\033[38;5;"
	Color256Suffix = "m"
)

// Color256 returns an ANSI escape for 256-color mode
func Color256(n int) string {
	return Color256Prefix + strconv.Itoa(n) + Color256Suffix
}

// RGB returns an ANSI escape for true color mode
func RGB(r, g, b int) string {
	return "\033[38;2;" + strconv.Itoa(r) + ";" + strconv.Itoa(g) + ";" + strconv.Itoa(b) + "m"
}

// Palette defines the semantic colors used to build a theme.
type Palette struct {
	// Base colors
	Foreground string // default text, identifiers
	Comment    string // comments (! lines)

	// Accent colors (semantic mapping to Cisco elements)
	Command   string // interface, router, ip, show (bold)
	Section   string // interface, router, line (bold)
	Protocol  string // ospf, bgp, tcp
	Action    string // permit, deny (bold)
	Interface string // GigabitEthernet0/0/0, Gi0/0/0 (bold)
	IP        string // IP addresses
	Number    string // numbers
	String    string // quoted strings
	Keyword   string // other keywords
	Operator  string // eq, gt, lt, any, host
	ASN       string // AS numbers
	Community string // BGP communities
	Value     string // values after keywords
	MAC       string // MAC addresses
	Negation  string // "no" prefix (typically red/warning)

	// State colors (for show output)
	StateGood    string // up, connected, established (bold green)
	StateBad     string // down, err-disabled (bold red)
	StateWarning string // init, exstart (bold yellow)

	// Show output extras
	Duration      string // time durations
	RouteProtocol string // [BGP/170] (bold)

	// Prompt colors
	PromptHost string // hostname in prompt
	PromptMode string // (config), (config-if) mode indicator
	PromptOper string // > prompt (user EXEC)
	PromptConf string // # prompt (privileged EXEC / config)
}

// buildTheme creates a Theme from a Palette by mapping semantic colors to token types.
func buildTheme(p Palette) *Theme {
	return &Theme{
		colors: map[lexer.TokenType]string{
			// Config tokens
			lexer.TokenCommand:    Bold + p.Command,
			lexer.TokenSection:    Bold + p.Section,
			lexer.TokenProtocol:   p.Protocol,
			lexer.TokenAction:     Bold + p.Action,
			lexer.TokenInterface:  Bold + p.Interface,
			lexer.TokenIPv4:       p.IP,
			lexer.TokenIPv4Prefix: p.IP,
			lexer.TokenIPv6:       p.IP,
			lexer.TokenIPv6Prefix: p.IP,
			lexer.TokenMAC:        p.MAC,
			lexer.TokenNumber:     p.Number,
			lexer.TokenString:     p.String,
			lexer.TokenComment:    Italic + p.Comment,
			lexer.TokenIdentifier: p.Foreground,
			lexer.TokenKeyword:    p.Keyword,
			lexer.TokenOperator:   p.Operator,
			lexer.TokenASN:        p.ASN,
			lexer.TokenCommunity:  p.Community,
			lexer.TokenValue:      p.Value,
			lexer.TokenNegation:   Bold + p.Negation,
			lexer.TokenText:       "",

			// Show output tokens
			lexer.TokenStateGood:     Bold + p.StateGood,
			lexer.TokenStateBad:      Bold + p.StateBad,
			lexer.TokenStateWarning:  Bold + p.StateWarning,
			lexer.TokenStateNeutral:  Dim + p.Comment,
			lexer.TokenColumnHeader:  Bold + p.Foreground,
			lexer.TokenStatusSymbol:  Bold + p.Protocol,
			lexer.TokenTimeDuration:  p.Duration,
			lexer.TokenPercentage:    p.StateGood,
			lexer.TokenByteSize:      p.Protocol,
			lexer.TokenRouteProtocol: Bold + p.RouteProtocol,

			// Cisco prompt tokens
			lexer.TokenPromptHost: Bold + p.PromptHost,
			lexer.TokenPromptMode: p.PromptMode,
			lexer.TokenPromptOper: Bold + p.PromptOper,
			lexer.TokenPromptConf: Bold + p.PromptConf,
		},
	}
}

// Theme defines ANSI color mappings for each token type.
// All methods are safe for concurrent use.
type Theme struct {
	mu     sync.RWMutex
	colors map[lexer.TokenType]string
}

// DefaultTheme returns the default theme (Tokyo Night)
func DefaultTheme() *Theme {
	return TokyoNightTheme()
}

// TokyoNightTheme returns a Tokyo Night inspired theme
func TokyoNightTheme() *Theme {
	foreground := RGB(192, 202, 245) // #c0caf5
	comment := RGB(86, 95, 137)      // #565f89
	red := RGB(247, 118, 142)        // #f7768e
	green := RGB(158, 206, 106)      // #9ece6a
	yellow := RGB(224, 175, 104)     // #e0af68
	blue := RGB(122, 162, 247)       // #7aa2f7
	magenta := RGB(187, 154, 247)    // #bb9af7
	cyan := RGB(125, 207, 255)       // #7dcfff
	orange := RGB(255, 158, 100)     // #ff9e64
	purple := RGB(157, 124, 216)     // #9d7cd8
	teal := RGB(115, 218, 202)       // #73daca

	return buildTheme(Palette{
		Foreground:     foreground,
		Comment:        comment,
		Command:        magenta,
		Section:        blue,
		Protocol:       cyan,
		Action:         green,
		Interface:      orange,
		IP:             teal,
		Number:         purple,
		String:         green,
		Keyword:        yellow,
		Operator:       blue,
		ASN:            orange,
		Community:      magenta,
		Value:          cyan,
		MAC:            cyan,
		Negation:       red,
		StateGood:      green,
		StateBad:       red,
		StateWarning:   yellow,
		Duration:       orange,
		RouteProtocol:  purple,
		PromptHost:     teal,
		PromptMode:     yellow,
		PromptOper:     green,
		PromptConf:     red,
	})
}

// VibrantTheme returns a vibrant color theme
func VibrantTheme() *Theme {
	return buildTheme(Palette{
		Foreground:     White,
		Comment:        Dim + BrightBlack,
		Command:        BrightYellow,
		Section:        BrightBlue,
		Protocol:       BrightCyan,
		Action:         BrightGreen,
		Interface:      BrightMagenta,
		IP:             BrightGreen,
		Number:         BrightCyan,
		String:         BrightYellow,
		Keyword:        Yellow,
		Operator:       BrightWhite,
		ASN:            BrightMagenta,
		Community:      Magenta,
		Value:          BrightCyan,
		MAC:            Cyan,
		Negation:       BrightRed,
		StateGood:      BrightGreen,
		StateBad:       BrightRed,
		StateWarning:   BrightYellow,
		Duration:       BrightMagenta,
		RouteProtocol:  Magenta,
		PromptHost:     Bold + BrightCyan,
		PromptMode:     BrightYellow,
		PromptOper:     Bold + BrightGreen,
		PromptConf:     Bold + BrightRed,
	})
}

// SolarizedDarkTheme returns a Solarized Dark theme
func SolarizedDarkTheme() *Theme {
	base01 := Color256(240)
	base0 := Color256(244)
	yellow := Color256(136)
	orange := Color256(166)
	red := Color256(160)
	magenta := Color256(125)
	violet := Color256(61)
	blue := Color256(33)
	cyan := Color256(37)
	green := Color256(64)

	return buildTheme(Palette{
		Foreground:     base0,
		Comment:        base01,
		Command:        yellow,
		Section:        blue,
		Protocol:       cyan,
		Action:         green,
		Interface:      magenta,
		IP:             green,
		Number:         cyan,
		String:         yellow,
		Keyword:        orange,
		Operator:       base0,
		ASN:            magenta,
		Community:      violet,
		Value:          cyan,
		MAC:            cyan,
		Negation:       red,
		StateGood:      green,
		StateBad:       red,
		StateWarning:   yellow,
		Duration:       orange,
		RouteProtocol:  violet,
		PromptHost:     Bold + cyan,
		PromptMode:     yellow,
		PromptOper:     Bold + green,
		PromptConf:     Bold + red,
	})
}

// MonokaiTheme returns a Monokai-inspired theme
func MonokaiTheme() *Theme {
	pink := Color256(197)
	green := Color256(148)
	orange := Color256(208)
	purple := Color256(141)
	cyan := Color256(81)
	yellow := Color256(186)
	gray := Color256(242)
	white := Color256(231)
	red := Color256(196)

	return buildTheme(Palette{
		Foreground:     white,
		Comment:        gray,
		Command:        pink,
		Section:        cyan,
		Protocol:       purple,
		Action:         green,
		Interface:      orange,
		IP:             green,
		Number:         purple,
		String:         yellow,
		Keyword:        orange,
		Operator:       pink,
		ASN:            orange,
		Community:      purple,
		Value:          cyan,
		MAC:            cyan,
		Negation:       red,
		StateGood:      green,
		StateBad:       red,
		StateWarning:   yellow,
		Duration:       orange,
		RouteProtocol:  purple,
		PromptHost:     Bold + cyan,
		PromptMode:     yellow,
		PromptOper:     Bold + green,
		PromptConf:     Bold + pink,
	})
}

// NordTheme returns a Nord theme
func NordTheme() *Theme {
	nord4 := Color256(252)
	nord7 := Color256(109)
	nord8 := Color256(110)
	nord9 := Color256(68)
	nord11 := Color256(167)
	nord12 := Color256(173)
	nord13 := Color256(179)
	nord14 := Color256(108)
	nord15 := Color256(139)
	nordComment := Color256(60)

	return buildTheme(Palette{
		Foreground:     nord4,
		Comment:        nordComment,
		Command:        nord13,
		Section:        nord9,
		Protocol:       nord8,
		Action:         nord14,
		Interface:      nord15,
		IP:             nord14,
		Number:         nord15,
		String:         nord13,
		Keyword:        nord12,
		Operator:       nord9,
		ASN:            nord12,
		Community:      nord15,
		Value:          nord8,
		MAC:            nord7,
		Negation:       nord11,
		StateGood:      nord14,
		StateBad:       nord11,
		StateWarning:   nord13,
		Duration:       nord12,
		RouteProtocol:  nord15,
		PromptHost:     Bold + nord7,
		PromptMode:     nord13,
		PromptOper:     Bold + nord14,
		PromptConf:     Bold + nord11,
	})
}

// CatppuccinMochaTheme returns a Catppuccin Mocha theme
func CatppuccinMochaTheme() *Theme {
	text := RGB(205, 214, 244)
	overlay0 := RGB(108, 112, 134)
	red := RGB(243, 139, 168)
	peach := RGB(250, 179, 135)
	yellow := RGB(249, 226, 175)
	green := RGB(166, 227, 161)
	teal := RGB(148, 226, 213)
	sky := RGB(137, 220, 235)
	sapphire := RGB(116, 199, 236)
	blue := RGB(137, 180, 250)
	lavender := RGB(180, 190, 254)
	mauve := RGB(203, 166, 247)
	pink := RGB(245, 194, 231)

	return buildTheme(Palette{
		Foreground:     text,
		Comment:        overlay0,
		Command:        mauve,
		Section:        blue,
		Protocol:       sapphire,
		Action:         green,
		Interface:      peach,
		IP:             teal,
		Number:         lavender,
		String:         green,
		Keyword:        yellow,
		Operator:       sky,
		ASN:            peach,
		Community:      pink,
		Value:          sky,
		MAC:            sky,
		Negation:       red,
		StateGood:      green,
		StateBad:       red,
		StateWarning:   yellow,
		Duration:       peach,
		RouteProtocol:  mauve,
		PromptHost:     Bold + sapphire,
		PromptMode:     yellow,
		PromptOper:     Bold + green,
		PromptConf:     Bold + red,
	})
}

// DraculaTheme returns the popular Dracula color scheme
func DraculaTheme() *Theme {
	foreground := RGB(248, 248, 242)
	comment := RGB(98, 114, 164)
	cyan := RGB(139, 233, 253)
	green := RGB(80, 250, 123)
	orange := RGB(255, 184, 108)
	pink := RGB(255, 121, 198)
	purple := RGB(189, 147, 249)
	red := RGB(255, 85, 85)
	yellow := RGB(241, 250, 140)

	return buildTheme(Palette{
		Foreground:     foreground,
		Comment:        comment,
		Command:        pink,
		Section:        purple,
		Protocol:       cyan,
		Action:         green,
		Interface:      orange,
		IP:             green,
		Number:         purple,
		String:         yellow,
		Keyword:        orange,
		Operator:       pink,
		ASN:            orange,
		Community:      purple,
		Value:          cyan,
		MAC:            cyan,
		Negation:       red,
		StateGood:      green,
		StateBad:       red,
		StateWarning:   yellow,
		Duration:       orange,
		RouteProtocol:  purple,
		PromptHost:     Bold + cyan,
		PromptMode:     yellow,
		PromptOper:     Bold + green,
		PromptConf:     Bold + red,
	})
}

// GruvboxDarkTheme returns the Gruvbox Dark color scheme
func GruvboxDarkTheme() *Theme {
	foreground := RGB(235, 219, 178)
	comment := RGB(146, 131, 116)
	red := RGB(251, 73, 52)
	green := RGB(184, 187, 38)
	yellow := RGB(250, 189, 47)
	blue := RGB(131, 165, 152)
	purple := RGB(211, 134, 155)
	aqua := RGB(142, 192, 124)
	orange := RGB(254, 128, 25)

	return buildTheme(Palette{
		Foreground:     foreground,
		Comment:        comment,
		Command:        yellow,
		Section:        blue,
		Protocol:       aqua,
		Action:         green,
		Interface:      orange,
		IP:             aqua,
		Number:         purple,
		String:         green,
		Keyword:        orange,
		Operator:       foreground,
		ASN:            orange,
		Community:      purple,
		Value:          aqua,
		MAC:            aqua,
		Negation:       red,
		StateGood:      green,
		StateBad:       red,
		StateWarning:   yellow,
		Duration:       orange,
		RouteProtocol:  purple,
		PromptHost:     Bold + aqua,
		PromptMode:     yellow,
		PromptOper:     Bold + green,
		PromptConf:     Bold + red,
	})
}

// OneDarkTheme returns the Atom One Dark color scheme
func OneDarkTheme() *Theme {
	foreground := RGB(171, 178, 191)
	comment := RGB(92, 99, 112)
	red := RGB(224, 108, 117)
	green := RGB(152, 195, 121)
	yellow := RGB(229, 192, 123)
	blue := RGB(97, 175, 239)
	purple := RGB(198, 120, 221)
	cyan := RGB(86, 182, 194)
	orange := RGB(209, 154, 102)

	return buildTheme(Palette{
		Foreground:     foreground,
		Comment:        comment,
		Command:        purple,
		Section:        blue,
		Protocol:       cyan,
		Action:         green,
		Interface:      orange,
		IP:             green,
		Number:         orange,
		String:         green,
		Keyword:        yellow,
		Operator:       foreground,
		ASN:            orange,
		Community:      purple,
		Value:          cyan,
		MAC:            cyan,
		Negation:       red,
		StateGood:      green,
		StateBad:       red,
		StateWarning:   yellow,
		Duration:       orange,
		RouteProtocol:  purple,
		PromptHost:     Bold + cyan,
		PromptMode:     yellow,
		PromptOper:     Bold + green,
		PromptConf:     Bold + red,
	})
}

// GetColor returns the color string for a token type
func (t *Theme) GetColor(tokenType lexer.TokenType) string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if color, ok := t.colors[tokenType]; ok {
		return color
	}
	return ""
}

// ThemeNames returns a list of available theme names.
func ThemeNames() []string {
	return []string{"tokyonight", "vibrant", "solarized", "monokai", "nord", "catppuccin", "dracula", "gruvbox", "onedark"}
}

// ThemeByName returns a theme by its name. Returns DefaultTheme for unknown names.
func ThemeByName(name string) *Theme {
	switch name {
	case "tokyonight", "tokyo-night", "tokyo":
		return TokyoNightTheme()
	case "vibrant":
		return VibrantTheme()
	case "solarized":
		return SolarizedDarkTheme()
	case "monokai":
		return MonokaiTheme()
	case "nord":
		return NordTheme()
	case "catppuccin", "catppuccin-mocha", "mocha":
		return CatppuccinMochaTheme()
	case "dracula":
		return DraculaTheme()
	case "gruvbox", "gruvbox-dark":
		return GruvboxDarkTheme()
	case "onedark", "one-dark":
		return OneDarkTheme()
	default:
		return DefaultTheme()
	}
}

// SetColor allows customizing a color for a token type.
// Safe for concurrent use with GetColor.
func (t *Theme) SetColor(tokenType lexer.TokenType, color string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.colors[tokenType] = color
}
