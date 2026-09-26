package lexer

import (
	"regexp"
	"strings"
)

// Constants for lexer configuration
const (
	// parseModeDetectionSampleSize is the number of characters sampled for auto-detection
	parseModeDetectionSampleSize = 500
)

// Lexer tokenizes Cisco IOS/IOS-XE configuration text and show command output
type Lexer struct {
	input          string
	pos            int
	line           int
	col            int
	parseMode      ParseMode
	detectedMode   bool
	expectingValue bool   // true after keywords like "description" that consume rest of line
	lastToken      string // tracks the last non-whitespace token value for context
}

// ParseMode determines which classification rules to use for tokenization.
type ParseMode int

const (
	// ParseModeAuto automatically detects whether input is configuration
	// syntax or show command output based on content heuristics.
	ParseModeAuto ParseMode = iota

	// ParseModeConfig uses configuration syntax classification rules.
	ParseModeConfig

	// ParseModeShow uses show command output classification rules.
	ParseModeShow
)

// String returns a human-readable name for the parse mode.
func (m ParseMode) String() string {
	switch m {
	case ParseModeAuto:
		return "Auto"
	case ParseModeConfig:
		return "Config"
	case ParseModeShow:
		return "Show"
	default:
		return "Unknown"
	}
}

// Keyword sets for Cisco IOS/IOS-XE classification
var (
	// configWords maps each config-mode word to its token type. A word has
	// exactly one type: the compiler rejects a duplicate key.
	configWords = map[string]TokenType{
		// Commands
		"interface": TokenCommand, "router": TokenCommand, "ip": TokenCommand, "ipv6": TokenCommand,
		"show": TokenCommand, "configure": TokenCommand, "hostname": TokenCommand, "username": TokenCommand,
		"enable": TokenCommand, "service": TokenCommand, "line": TokenCommand, "logging": TokenCommand,
		"ntp": TokenCommand, "snmp-server": TokenCommand, "crypto": TokenCommand, "aaa": TokenCommand,
		"spanning-tree": TokenCommand, "vlan": TokenCommand, "banner": TokenCommand,
		"shutdown": TokenCommand, "write": TokenCommand, "copy": TokenCommand, "reload": TokenCommand,
		"ping": TokenCommand, "traceroute": TokenCommand, "clock": TokenCommand, "boot": TokenCommand,
		"archive": TokenCommand, "errdisable": TokenCommand, "default-gateway": TokenCommand,
		"do": TokenCommand, "exit": TokenCommand, "end": TokenCommand,

		// Section headers
		"access-list": TokenSection, "route-map": TokenSection, "prefix-list": TokenSection,
		"class-map": TokenSection, "policy-map": TokenSection,
		"redundancy": TokenSection, "controller": TokenSection,
		"key": TokenSection, "track": TokenSection,
		"monitor": TokenSection, "event": TokenSection, "applet": TokenSection,

		// Protocols
		"ospf": TokenProtocol, "bgp": TokenProtocol, "eigrp": TokenProtocol, "rip": TokenProtocol,
		"isis": TokenProtocol, "mpls": TokenProtocol, "hsrp": TokenProtocol, "vrrp": TokenProtocol,
		"stp": TokenProtocol, "rstp": TokenProtocol, "lacp": TokenProtocol, "dot1q": TokenProtocol,
		"ipsec": TokenProtocol, "gre": TokenProtocol, "tcp": TokenProtocol, "udp": TokenProtocol,
		"icmp": TokenProtocol, "ssh": TokenProtocol, "dhcp": TokenProtocol, "bfd": TokenProtocol,
		"cdp": TokenProtocol, "lldp": TokenProtocol, "evpn": TokenProtocol, "vxlan": TokenProtocol,
		"isakmp": TokenProtocol, "nhrp": TokenProtocol, "pim": TokenProtocol, "igmp": TokenProtocol,
		"msdp": TokenProtocol, "lisp": TokenProtocol, "omp": TokenProtocol, "snmp": TokenProtocol,
		"radius": TokenProtocol, "tacacs": TokenProtocol, "tacacs+": TokenProtocol,
		"telnet": TokenProtocol, "ftp": TokenProtocol, "tftp": TokenProtocol, "http": TokenProtocol,
		"https": TokenProtocol, "dns": TokenProtocol, "syslog": TokenProtocol,
		"netflow": TokenProtocol, "sflow": TokenProtocol, "ipfix": TokenProtocol,

		// Actions
		"permit": TokenAction, "deny": TokenAction, "log": TokenAction, "log-input": TokenAction,
		"established": TokenAction, "match": TokenAction, "set": TokenAction,
		"remark": TokenAction, "evaluate": TokenAction, "reflect": TokenAction,

		// Operators
		"eq": TokenOperator, "gt": TokenOperator, "lt": TokenOperator, "neq": TokenOperator,
		"range": TokenOperator, "ge": TokenOperator, "le": TokenOperator, "any": TokenOperator,
		"host": TokenOperator,

		// Other keywords
		// Interface keywords
		"description": TokenKeyword, "address": TokenKeyword, "switchport": TokenKeyword,
		"speed": TokenKeyword, "duplex": TokenKeyword, "mtu": TokenKeyword, "bandwidth": TokenKeyword,
		"encapsulation": TokenKeyword, "channel-group": TokenKeyword, "channel-protocol": TokenKeyword,
		"standby":      TokenKeyword,
		"no-autostate": TokenKeyword, "autostate": TokenKeyword,

		// Routing keywords
		"network": TokenKeyword, "neighbor": TokenKeyword, "redistribute": TokenKeyword,
		"area": TokenKeyword, "remote-as": TokenKeyword, "update-source": TokenKeyword,
		"access-group": TokenKeyword, "nat": TokenKeyword,
		"inside": TokenKeyword, "outside": TokenKeyword, "overload": TokenKeyword,
		"default-information": TokenKeyword, "originate": TokenKeyword,
		"summary-address": TokenKeyword, "passive-interface": TokenKeyword,
		"distance": TokenKeyword, "metric": TokenKeyword, "weight": TokenKeyword,
		"local-preference": TokenKeyword, "next-hop-self": TokenKeyword,
		"soft-reconfiguration": TokenKeyword, "inbound": TokenKeyword,
		"distribute-list": TokenKeyword,
		"maximum-paths":   TokenKeyword, "auto-summary": TokenKeyword,
		"synchronization": TokenKeyword, "log-neighbor-changes": TokenKeyword,
		"address-family": TokenKeyword, "unicast": TokenKeyword, "multicast": TokenKeyword,
		"vpnv4": TokenKeyword, "vpnv6": TokenKeyword,

		// Security keywords
		"access-class": TokenKeyword, "transport": TokenKeyword, "input": TokenKeyword,
		"output": TokenKeyword, "login": TokenKeyword, "password": TokenKeyword,
		"secret": TokenKeyword, "privilege": TokenKeyword, "authentication": TokenKeyword,
		"authorization": TokenKeyword, "accounting": TokenKeyword, "group": TokenKeyword,
		"method": TokenKeyword, "local": TokenKeyword,

		// System keywords
		"version": TokenKeyword, "source": TokenKeyword, "trap": TokenKeyword,
		"community": TokenKeyword, "location": TokenKeyword, "contact": TokenKeyword,
		"default": TokenKeyword, "timeout": TokenKeyword, "exec-timeout": TokenKeyword,
		"mask": TokenKeyword, "wildcard": TokenKeyword, "inverse-mask": TokenKeyword,

		// Spanning-tree keywords
		"mode": TokenKeyword, "priority": TokenKeyword,
		"portfast": TokenKeyword, "bpduguard": TokenKeyword, "bpdufilter": TokenKeyword,
		"guard": TokenKeyword, "root": TokenKeyword,

		// VLAN keywords
		"name": TokenKeyword, "state": TokenKeyword, "active": TokenKeyword, "suspend": TokenKeyword,

		// QoS keywords
		"class": TokenKeyword, "police": TokenKeyword, "shape": TokenKeyword,
		"queue": TokenKeyword, "dscp": TokenKeyword, "cos": TokenKeyword,
		"service-policy": TokenKeyword,

		// AAA keywords
		"new-model": TokenKeyword, "server": TokenKeyword,

		// Other
		"trunk":  TokenKeyword,
		"native": TokenKeyword, "allowed": TokenKeyword, "tagging": TokenKeyword,
		"nonegotiate": TokenKeyword, "negotiation": TokenKeyword, "auto": TokenKeyword,
		"half": TokenKeyword, "flow-control": TokenKeyword,
		"send": TokenKeyword, "both": TokenKeyword,
		"storm-control": TokenKeyword, "level": TokenKeyword,
	}

	// Keywords that consume the rest of the line as a value
	valueKeywords = map[string]bool{
		"description": true,
		"remark":      true,
	}

	// Cisco interface naming patterns
	// Matches: GigabitEthernet0/0/0, Gi0/0/0, FastEthernet0/0, Fa0/0,
	//          TenGigabitEthernet1/0/0, Te1/0/0, Loopback0, Lo0,
	//          Vlan100, Vl100, Port-channel1, Po1, Tunnel0, Tu0,
	//          Serial0/0/0, Se0/0/0, Null0, BDI1, mgmt0, nve1
	interfacePattern = regexp.MustCompile(`^(?i)(GigabitEthernet|Gi|FastEthernet|Fa|TenGigabitEthernet|TenGigE|Te|TwentyFiveGigE|TwentyFiveGigabitEthernet|FortyGigabitEthernet|Fo|HundredGigE|Hu|Ethernet|Eth|Loopback|Lo|Vlan|Vl|Port-channel|Po|Tunnel|Tu|Serial|Se|Null|BDI|mgmt|nve|NVE|Dialer|Di|Virtual-Template|Vt|Virtual-Access|Va|Multilink|Mu|ATM|Cellular|Async)\d+(/\d+)*(\.\d+)?$`)

	ipv4Pattern       = regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}$`)
	ipv4PrefixPattern = regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}/\d{1,2}$`)
	// IPv6: require either "::" (compressed) or at least 3 colon-separated groups
	ipv6Pattern       = regexp.MustCompile(`^([0-9a-fA-F]{0,4}:){2,7}[0-9a-fA-F]{0,4}$|^::([0-9a-fA-F]{1,4}:)*[0-9a-fA-F]{0,4}$|^[0-9a-fA-F]{1,4}::([0-9a-fA-F]{1,4}:)*[0-9a-fA-F]{0,4}$`)
	ipv6PrefixPattern = regexp.MustCompile(`^(([0-9a-fA-F]{0,4}:){2,7}[0-9a-fA-F]{0,4}|::([0-9a-fA-F]{1,4}:)*[0-9a-fA-F]{0,4}|[0-9a-fA-F]{1,4}::([0-9a-fA-F]{1,4}:)*[0-9a-fA-F]{0,4})/\d{1,3}$`)

	// Cisco MAC format: 0011.2233.4455 (dotted) and also colon format
	macPatternCisco = regexp.MustCompile(`^[0-9a-fA-F]{4}\.[0-9a-fA-F]{4}\.[0-9a-fA-F]{4}$`)
	macPatternColon = regexp.MustCompile(`^([0-9a-fA-F]{2}:){5}[0-9a-fA-F]{2}$`)

	communityPattern = regexp.MustCompile(`^\d+:\d+$`)
	asnPattern       = regexp.MustCompile(`^[Aa][Ss]\d+$`)

	// Show output state keywords
	statesGood = map[string]bool{
		"up": true, "connected": true, "established": true,
		"full": true, "enabled": true, "active": true,
		"forwarding": true, "ok": true, "online": true,
		"running": true, "ready": true, "complete": true,
	}

	// Compound state patterns matched as whole words
	statesGoodCompound = []string{"up/up"}

	statesBad = map[string]bool{
		"down": true, "notconnect": true, "err-disabled": true,
		"disabled": true, "failed": true, "idle": true,
		"connect": true, "opensent": true, "openconfirm": true,
		"error": true, "offline": true, "unreachable": true,
	}

	statesBadCompound = []string{"down/down", "administratively"}

	statesWarning = map[string]bool{
		"init": true, "2way": true, "exstart": true,
		"exchange": true, "loading": true, "attempt": true,
		"flapping": true, "pending": true, "waiting": true,
		"starting": true, "stopping": true,
	}

	statesNeutral = map[string]bool{
		"inactive": true, "standby": true, "backup": true,
		"suspended": true, "n/a": true, "none": true,
	}

	columnHeaders = map[string]bool{
		"interface": true, "status": true, "protocol": true,
		"address": true, "admin": true, "link": true,
		"speed": true, "type": true, "duplex": true,
		"neighbor": true, "peer": true, "state": true,
		"as": true, "inpkt": true, "outpkt": true,
		"uptime": true, "dead": true, "pri": true,
		"mtu": true, "metric": true, "local": true,
		"remote": true, "outq": true, "up/dn": true,
		"flaps": true, "prefixes": true, "paths": true,
		"vlan": true, "description": true,
	}

	statusSymbols = map[string]bool{
		"*": true, "+": true, "-": true, ">": true,
		"B": true, "O": true, "I": true, "S": true,
		"L": true, "D": true, "C": true, "R": true,
	}

	// Show output regex patterns
	timeDurationPattern  = regexp.MustCompile(`^(\d+[wdhms])+$|^\d+:\d{2}(:\d{2})?$`)
	percentagePattern    = regexp.MustCompile(`^\d+(\.\d+)?%$`)
	byteSizePattern      = regexp.MustCompile(`^\d+(\.\d+)?[KMGTP][Bb]?$`)
	routeProtocolPattern = regexp.MustCompile(`^\[(BGP|OSPF|EIGRP|RIP|ISIS|Static|Direct|Local|Connected|Aggregate)/\d+\]$`)
	tabularPattern       = regexp.MustCompile(`\w+\s{2,}\w+\s{2,}\w+`)

	// Cisco prompt pattern
	// Matches: Router>, Router#, Router(config)#, Router(config-if)#
	// Also: hostname with dots/dashes: core-rtr-01.example>, CORE-RTR-01(config-router)#
	// Group 1 = leading whitespace/control chars (like \r)
	// Group 2 = hostname
	// Group 3 = mode string e.g. (config-if) - optional
	// Group 4 = prompt char (> or #)
	// Group 5 = spaces between prompt char and command (optional)
	// Group 6 = command after prompt (optional)
	// Group 7 = line ending (optional)
	promptPattern = regexp.MustCompile(`^([\s\x00-\x1f]*)([\w.-]+)(\([\w-]+\))?([>#])([ \t]*)(.*?)(\r?\n)?$`)
)

// New creates a new Lexer for the given input.
func New(input string) *Lexer {
	return &Lexer{
		input: input,
		pos:   0,
		line:  1,
		col:   1,
	}
}

// Tokenize processes the input and returns all tokens.
func (l *Lexer) Tokenize() []Token {
	var tokens []Token

	// Check if the entire input is a prompt line
	if promptTokens := l.tryTokenizePrompt(l.input); promptTokens != nil {
		return promptTokens
	}

	for l.pos < len(l.input) {
		token := l.nextToken()
		if token.Type != TokenText || token.Value != "" {
			tokens = append(tokens, token)
		}
	}

	return tokens
}

// tryTokenizePrompt checks if input matches a Cisco prompt and returns tokens if so
func (l *Lexer) tryTokenizePrompt(input string) []Token {
	matches := promptPattern.FindStringSubmatch(input)
	if matches == nil {
		return nil
	}

	var tokens []Token
	col := 1

	// matches[1] = leading whitespace/control chars (optional)
	// matches[2] = hostname
	// matches[3] = mode string (config), (config-if), etc. (optional)
	// matches[4] = prompt char (> or #)
	// matches[5] = spaces before the command (optional)
	// matches[6] = command after prompt (optional)
	// matches[7] = line ending (optional)

	// Preserve leading whitespace/control chars
	if matches[1] != "" {
		tokens = append(tokens, Token{
			Type:   TokenText,
			Value:  matches[1],
			Line:   1,
			Column: col,
		})
		col += len(matches[1])
	}

	// Add hostname
	isConfig := matches[4] == "#"
	tokens = append(tokens, Token{
		Type:   TokenPromptHost,
		Value:  matches[2],
		Line:   1,
		Column: col,
	})
	col += len(matches[2])

	// Add mode string if present (e.g., "(config-if)")
	if matches[3] != "" {
		tokens = append(tokens, Token{
			Type:   TokenPromptMode,
			Value:  matches[3],
			Line:   1,
			Column: col,
		})
		col += len(matches[3])
	}

	// Add prompt character
	promptTokenType := TokenPromptOper
	if isConfig {
		promptTokenType = TokenPromptConf
	}
	tokens = append(tokens, Token{
		Type:   promptTokenType,
		Value:  matches[4],
		Line:   1,
		Column: col,
	})
	col++

	// Keep the spaces before the command as typed
	if matches[5] != "" {
		tokens = append(tokens, Token{
			Type:   TokenText,
			Value:  matches[5],
			Line:   1,
			Column: col,
		})
		col += len(matches[5])
	}

	// Add command after prompt if present
	if matches[6] != "" {
		cmdLexer := New(matches[6])
		cmdTokens := cmdLexer.Tokenize()
		for _, tok := range cmdTokens {
			tok.Column = col
			tokens = append(tokens, tok)
			col += len(tok.Value)
		}
	}

	// Preserve the line ending
	if matches[7] != "" {
		tokens = append(tokens, Token{
			Type:   TokenText,
			Value:  matches[7],
			Line:   1,
			Column: col,
		})
	}

	return tokens
}

// nextToken extracts the next token from the input
func (l *Lexer) nextToken() Token {
	startLine, startCol := l.line, l.col

	if l.pos >= len(l.input) {
		return Token{Type: TokenText, Value: "", Line: startLine, Column: startCol}
	}

	ch := l.input[l.pos]

	switch {
	case ch == '!' && l.col == 1:
		return l.scanComment()
	case ch == '"':
		isValue := l.expectingValue
		l.expectingValue = false
		token := l.scanString('"')
		if isValue {
			token.Type = TokenValue
		}
		return token
	case ch == '\'':
		isValue := l.expectingValue
		l.expectingValue = false
		token := l.scanString('\'')
		if isValue {
			token.Type = TokenValue
		}
		return token
	case isWhitespace(ch):
		return l.scanWhitespace()
	default:
		if l.expectingValue {
			l.expectingValue = false
			return l.scanValueToEndOfLine()
		}
		return l.scanWord()
	}
}

// scanComment scans a ! comment line (Cisco section separator)
func (l *Lexer) scanComment() Token {
	startLine, startCol := l.line, l.col
	start := l.pos

	for l.pos < len(l.input) && l.input[l.pos] != '\n' {
		l.advance()
	}

	return Token{
		Type:   TokenComment,
		Value:  l.input[start:l.pos],
		Line:   startLine,
		Column: startCol,
	}
}

// scanString scans a quoted string
func (l *Lexer) scanString(quote byte) Token {
	startLine, startCol := l.line, l.col
	start := l.pos

	l.advance() // opening quote

	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if ch == quote {
			l.advance() // closing quote
			break
		}
		if ch == '\\' && l.pos+1 < len(l.input) {
			l.advance() // escape char
		}
		l.advance()
	}

	return Token{
		Type:   TokenString,
		Value:  l.input[start:l.pos],
		Line:   startLine,
		Column: startCol,
	}
}

// scanValueToEndOfLine scans an unquoted value until end of line (for description, hostname, etc.)
func (l *Lexer) scanValueToEndOfLine() Token {
	startLine, startCol := l.line, l.col
	start := l.pos

	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if ch == '\n' {
			break
		}
		l.advance()
	}

	return Token{
		Type:   TokenValue,
		Value:  l.input[start:l.pos],
		Line:   startLine,
		Column: startCol,
	}
}

// scanWhitespace scans whitespace characters
func (l *Lexer) scanWhitespace() Token {
	startLine, startCol := l.line, l.col
	start := l.pos

	for l.pos < len(l.input) && isWhitespace(l.input[l.pos]) {
		l.advance()
	}

	return Token{
		Type:   TokenText,
		Value:  l.input[start:l.pos],
		Line:   startLine,
		Column: startCol,
	}
}

// scanWord scans a word token
func (l *Lexer) scanWord() Token {
	startLine, startCol := l.line, l.col
	start := l.pos

	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if isWhitespace(ch) || ch == '"' || ch == '\'' {
			break
		}
		l.advance()
	}

	word := l.input[start:l.pos]
	tokenType := l.classifyWord(word)

	return Token{
		Type:   tokenType,
		Value:  word,
		Line:   startLine,
		Column: startCol,
	}
}

// classifyWord determines the token type for a word
func (l *Lexer) classifyWord(word string) TokenType {
	if l.parseMode == ParseModeAuto && !l.detectedMode {
		l.parseMode, _ = Detect(l.input)
		l.detectedMode = true
	}

	lower := strings.ToLower(word)

	if l.parseMode == ParseModeShow {
		return l.classifyShowWord(word, lower)
	}

	return l.classifyConfigWord(word, lower)
}

// classifyConfigWord handles Cisco configuration syntax classification
func (l *Lexer) classifyConfigWord(word, lower string) TokenType {
	// Check for "no" prefix (negation)
	if lower == "no" {
		l.lastToken = lower
		return TokenNegation
	}

	// Check for AS number format (AS65000, as65001)
	if asnPattern.MatchString(word) {
		return TokenASN
	}

	if tokenType, ok := configWords[lower]; ok {
		if valueKeywords[lower] {
			l.expectingValue = true
		}
		l.lastToken = lower
		return tokenType
	}

	return l.classifySharedPatterns(word)
}

// classifyShowWord handles show command output classification
func (l *Lexer) classifyShowWord(word, lower string) TokenType {
	// Compound states
	for _, s := range statesGoodCompound {
		if lower == s {
			return TokenStateGood
		}
	}
	for _, s := range statesBadCompound {
		if lower == s {
			return TokenStateBad
		}
	}

	// State classification
	if statesGood[lower] {
		return TokenStateGood
	}
	if statesBad[lower] {
		return TokenStateBad
	}
	if statesWarning[lower] {
		return TokenStateWarning
	}
	if statesNeutral[lower] {
		return TokenStateNeutral
	}

	// Status symbols
	if len(word) <= 2 && statusSymbols[word] {
		return TokenStatusSymbol
	}

	// Show-specific patterns
	if timeDurationPattern.MatchString(word) {
		return TokenTimeDuration
	}
	if percentagePattern.MatchString(word) {
		return TokenPercentage
	}
	if byteSizePattern.MatchString(word) {
		return TokenByteSize
	}
	if routeProtocolPattern.MatchString(word) {
		return TokenRouteProtocol
	}

	// Column headers
	if columnHeaders[lower] {
		return TokenColumnHeader
	}

	return l.classifySharedPatterns(word)
}

// classifySharedPatterns handles patterns common to both config and show modes
func (l *Lexer) classifySharedPatterns(word string) TokenType {
	// Cisco interface names
	if interfacePattern.MatchString(word) {
		return TokenInterface
	}

	// IP patterns - more specific first
	if ipv4PrefixPattern.MatchString(word) {
		return TokenIPv4Prefix
	}
	if ipv4Pattern.MatchString(word) {
		return TokenIPv4
	}

	// MAC addresses (Cisco dotted and colon format)
	if macPatternCisco.MatchString(word) {
		return TokenMAC
	}
	if macPatternColon.MatchString(word) {
		return TokenMAC
	}

	// BGP community - only after "community" keyword to avoid false positives (e.g., "12:00")
	if l.lastToken == "community" && communityPattern.MatchString(word) {
		return TokenCommunity
	}

	// IPv6 patterns
	if ipv6PrefixPattern.MatchString(word) {
		return TokenIPv6Prefix
	}
	if ipv6Pattern.MatchString(word) {
		return TokenIPv6
	}

	// Numbers
	if isAllDigits(word) {
		return TokenNumber
	}

	return TokenIdentifier
}

// Helper methods

func (l *Lexer) advance() {
	if l.pos < len(l.input) {
		if l.input[l.pos] == '\n' {
			l.line++
			l.col = 1
		} else {
			l.col++
		}
		l.pos++
	}
}

func isWhitespace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

// isAllDigits returns true if s is non-empty and contains only ASCII digits.
func isAllDigits(s string) bool {
	if len(s) == 0 {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// configIndicators are statements that start a line of Cisco configuration.
var configIndicators = []string{
	"hostname ", "interface ", "router ", "ip address ",
	"switchport ", "access-list ", "no ", "line vty",
	"line con", "service ", "enable ", "username ",
	"ip route ", "snmp-server ", "logging ", "ntp ",
	"crypto ", "aaa ", "spanning-tree ", "vlan ",
	"banner ", "ip access-list ", "transport ",
	"exec-timeout ", "channel-group ",
}

// showIndicators are words and phrases that suggest show command output.
var showIndicators = []string{
	"line protocol", "up/up", "down/down",
	"notconnect", "err-disabled", "connected",
	"bgp summary", "ospf neighbor",
	"show", "last input", "last output",
	"5 minute", "input rate", "output rate",
	"show version", "cisco ios",
}

// Detect reports whether input looks like Cisco configuration or show
// output, and which parse mode fits it. It samples the start of input.
// When ok is false the returned mode is ParseModeConfig.
//
// Config indicators must start a line (after indentation). Show indicators
// match as whole words. One config indicator is enough, but show output
// needs two hits, because show vocabulary ("connected") is common prose.
func Detect(input string) (mode ParseMode, ok bool) {
	sample := input
	if len(sample) > parseModeDetectionSampleSize {
		sample = sample[:parseModeDetectionSampleSize]
	}

	if isPrompt(sample) {
		return ParseModeConfig, true
	}

	lower := strings.ToLower(sample)
	lines := strings.Split(lower, "\n")

	configScore := 0
	for _, ind := range configIndicators {
		for _, line := range lines {
			if strings.HasPrefix(strings.TrimLeft(line, " \t\r"), ind) {
				configScore++
				break
			}
		}
	}
	// ! section separators are a strong config indicator
	bangs := 0
	for _, line := range lines {
		if strings.TrimSpace(line) == "!" {
			bangs++
		}
	}
	if bangs >= 2 {
		configScore += 2
	}

	showScore := 0
	for _, ind := range showIndicators {
		if containsWord(lower, ind) {
			showScore++
		}
	}
	if tabularPattern.MatchString(sample) {
		showScore += 2
	}

	if showScore >= 2 && showScore > configScore {
		return ParseModeShow, true
	}
	return ParseModeConfig, configScore > 0 || showScore >= 2
}

// containsWord reports whether phrase occurs in s with no letter or digit
// directly before or after it.
func containsWord(s, phrase string) bool {
	for i := 0; ; {
		j := strings.Index(s[i:], phrase)
		if j < 0 {
			return false
		}
		start, end := i+j, i+j+len(phrase)
		if (start == 0 || !isWordByte(s[start-1])) && (end == len(s) || !isWordByte(s[end])) {
			return true
		}
		i = start + 1
	}
}

func isWordByte(b byte) bool {
	return b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' || b >= '0' && b <= '9'
}

// isPrompt checks if the input matches a Cisco CLI prompt pattern.
func isPrompt(input string) bool {
	return promptPattern.MatchString(strings.TrimSpace(input))
}

// SetParseMode explicitly sets the parsing mode
func (l *Lexer) SetParseMode(mode ParseMode) {
	l.parseMode = mode
	l.detectedMode = true
}

// GetParseMode returns the current parse mode
func (l *Lexer) GetParseMode() ParseMode {
	return l.parseMode
}
