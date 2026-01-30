package lexer

import (
	"testing"
)

func TestTokenizeCommands(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"interface", TokenCommand},
		{"router", TokenCommand},
		{"ip", TokenCommand},
		{"ipv6", TokenCommand},
		{"show", TokenCommand},
		{"configure", TokenCommand},
		{"hostname", TokenCommand},
		{"username", TokenCommand},
		{"enable", TokenCommand},
		{"service", TokenCommand},
		{"line", TokenCommand},
		{"logging", TokenCommand},
		{"shutdown", TokenCommand},
		{"write", TokenCommand},
		{"copy", TokenCommand},
		{"reload", TokenCommand},
		{"ping", TokenCommand},
		{"traceroute", TokenCommand},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			tokens := l.Tokenize()
			if len(tokens) != 1 {
				t.Fatalf("expected 1 token, got %d", len(tokens))
			}
			if tokens[0].Type != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, tokens[0].Type)
			}
		})
	}
}

func TestTokenizeNegation(t *testing.T) {
	l := New("no")
	tokens := l.Tokenize()
	if len(tokens) != 1 {
		t.Fatalf("expected 1 token, got %d", len(tokens))
	}
	if tokens[0].Type != TokenNegation {
		t.Errorf("expected TokenNegation, got %v", tokens[0].Type)
	}
}

func TestTokenizeNegationInContext(t *testing.T) {
	input := "no shutdown"
	l := New(input)
	tokens := l.Tokenize()

	// Should have: no, space, shutdown
	if len(tokens) < 3 {
		t.Fatalf("expected at least 3 tokens, got %d", len(tokens))
	}
	if tokens[0].Type != TokenNegation {
		t.Errorf("expected TokenNegation for 'no', got %v", tokens[0].Type)
	}
	if tokens[0].Value != "no" {
		t.Errorf("expected value 'no', got %q", tokens[0].Value)
	}
	// "shutdown" should be TokenCommand
	if tokens[2].Type != TokenCommand {
		t.Errorf("expected TokenCommand for 'shutdown', got %v", tokens[2].Type)
	}
}

func TestTokenizeProtocols(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"ospf", TokenProtocol},
		{"bgp", TokenProtocol},
		{"eigrp", TokenProtocol},
		{"rip", TokenProtocol},
		{"isis", TokenProtocol},
		{"mpls", TokenProtocol},
		{"hsrp", TokenProtocol},
		{"vrrp", TokenProtocol},
		{"tcp", TokenProtocol},
		{"udp", TokenProtocol},
		{"icmp", TokenProtocol},
		{"ssh", TokenProtocol},
		{"dhcp", TokenProtocol},
		{"bfd", TokenProtocol},
		{"cdp", TokenProtocol},
		{"lldp", TokenProtocol},
		{"lacp", TokenProtocol},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			tokens := l.Tokenize()
			if len(tokens) != 1 {
				t.Fatalf("expected 1 token, got %d", len(tokens))
			}
			if tokens[0].Type != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, tokens[0].Type)
			}
		})
	}
}

func TestTokenizeActions(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"permit", TokenAction},
		{"deny", TokenAction},
		{"log", TokenAction},
		{"log-input", TokenAction},
		{"match", TokenAction},
		{"set", TokenAction},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			tokens := l.Tokenize()
			if len(tokens) != 1 {
				t.Fatalf("expected 1 token, got %d", len(tokens))
			}
			if tokens[0].Type != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, tokens[0].Type)
			}
		})
	}
}

func TestTokenizeOperators(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"eq", TokenOperator},
		{"gt", TokenOperator},
		{"lt", TokenOperator},
		{"neq", TokenOperator},
		{"range", TokenOperator},
		{"ge", TokenOperator},
		{"le", TokenOperator},
		{"any", TokenOperator},
		{"host", TokenOperator},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			tokens := l.Tokenize()
			if len(tokens) != 1 {
				t.Fatalf("expected 1 token, got %d", len(tokens))
			}
			if tokens[0].Type != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, tokens[0].Type)
			}
		})
	}
}

func TestTokenizeCiscoInterfaces(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		// Full names
		{"GigabitEthernet0/0/0", TokenInterface},
		{"GigabitEthernet0/0/1", TokenInterface},
		{"FastEthernet0/0", TokenInterface},
		{"TenGigabitEthernet1/0/0", TokenInterface},
		{"Loopback0", TokenInterface},
		{"Loopback99", TokenInterface},
		{"Vlan100", TokenInterface},
		{"Vlan1", TokenInterface},
		{"Port-channel1", TokenInterface},
		{"Port-channel10", TokenInterface},
		{"Tunnel0", TokenInterface},
		{"Tunnel1", TokenInterface},
		{"Serial0/0/0", TokenInterface},
		{"Null0", TokenInterface},
		{"BDI1", TokenInterface},
		{"nve1", TokenInterface},
		// Abbreviated names
		{"Gi0/0/0", TokenInterface},
		{"Gi0/0/1", TokenInterface},
		{"Fa0/0", TokenInterface},
		{"Te1/0/0", TokenInterface},
		{"Lo0", TokenInterface},
		{"Vl100", TokenInterface},
		{"Po1", TokenInterface},
		{"Tu0", TokenInterface},
		{"Se0/0/0", TokenInterface},
		// With subinterfaces
		{"GigabitEthernet0/0/0.100", TokenInterface},
		{"Gi0/0/0.10", TokenInterface},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			tokens := l.Tokenize()
			if len(tokens) != 1 {
				t.Fatalf("expected 1 token for %q, got %d", tt.input, len(tokens))
			}
			if tokens[0].Type != tt.expected {
				t.Errorf("expected %v for %q, got %v", tt.expected, tt.input, tokens[0].Type)
			}
		})
	}
}

func TestTokenizeIPv4(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"192.168.1.1", TokenIPv4},
		{"10.0.0.1", TokenIPv4},
		{"172.16.0.1", TokenIPv4},
		{"255.255.255.255", TokenIPv4},
		{"0.0.0.0", TokenIPv4},
		{"203.0.113.1", TokenIPv4},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			tokens := l.Tokenize()
			if len(tokens) != 1 {
				t.Fatalf("expected 1 token, got %d", len(tokens))
			}
			if tokens[0].Type != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, tokens[0].Type)
			}
		})
	}
}

func TestTokenizeIPv4Prefix(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"192.168.1.0/24", TokenIPv4Prefix},
		{"10.0.0.0/8", TokenIPv4Prefix},
		{"0.0.0.0/0", TokenIPv4Prefix},
		{"192.168.1.1/32", TokenIPv4Prefix},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			tokens := l.Tokenize()
			if len(tokens) != 1 {
				t.Fatalf("expected 1 token, got %d", len(tokens))
			}
			if tokens[0].Type != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, tokens[0].Type)
			}
		})
	}
}

func TestTokenizeIPv6(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"2001:db8::1", TokenIPv6},
		{"::1", TokenIPv6},
		{"fe80::1", TokenIPv6},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			tokens := l.Tokenize()
			if len(tokens) != 1 {
				t.Fatalf("expected 1 token, got %d", len(tokens))
			}
			if tokens[0].Type != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, tokens[0].Type)
			}
		})
	}
}

func TestTokenizeIPv6Prefix(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"2001:db8::/32", TokenIPv6Prefix},
		{"::/0", TokenIPv6Prefix},
		{"fe80::/10", TokenIPv6Prefix},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			tokens := l.Tokenize()
			if len(tokens) != 1 {
				t.Fatalf("expected 1 token, got %d", len(tokens))
			}
			if tokens[0].Type != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, tokens[0].Type)
			}
		})
	}
}

func TestTokenizeCiscoMAC(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		// Cisco dotted format
		{"0011.2233.4455", TokenMAC},
		{"aabb.ccdd.eeff", TokenMAC},
		{"AABB.CCDD.EEFF", TokenMAC},
		// Colon format
		{"00:11:22:33:44:55", TokenMAC},
		{"aa:bb:cc:dd:ee:ff", TokenMAC},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			tokens := l.Tokenize()
			if len(tokens) != 1 {
				t.Fatalf("expected 1 token for %q, got %d", tt.input, len(tokens))
			}
			if tokens[0].Type != tt.expected {
				t.Errorf("expected %v for %q, got %v", tt.expected, tt.input, tokens[0].Type)
			}
		})
	}
}

func TestTokenizeComment(t *testing.T) {
	input := "!"
	l := New(input)
	tokens := l.Tokenize()
	if len(tokens) != 1 {
		t.Fatalf("expected 1 token, got %d", len(tokens))
	}
	if tokens[0].Type != TokenComment {
		t.Errorf("expected TokenComment, got %v", tokens[0].Type)
	}
}

func TestTokenizeCommentWithText(t *testing.T) {
	input := "! Section header"
	l := New(input)
	tokens := l.Tokenize()
	if len(tokens) != 1 {
		t.Fatalf("expected 1 token, got %d", len(tokens))
	}
	if tokens[0].Type != TokenComment {
		t.Errorf("expected TokenComment, got %v", tokens[0].Type)
	}
	if tokens[0].Value != input {
		t.Errorf("expected value %q, got %q", input, tokens[0].Value)
	}
}

func TestTokenizeStrings(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"double quoted", `"Uplink to ISP"`},
		{"with spaces", `"Main Data Center, Rack 42"`},
		{"empty", `""`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			tokens := l.Tokenize()
			if len(tokens) != 1 {
				t.Fatalf("expected 1 token, got %d", len(tokens))
			}
			if tokens[0].Type != TokenString {
				t.Errorf("expected TokenString, got %v", tokens[0].Type)
			}
			if tokens[0].Value != tt.input {
				t.Errorf("expected value %q, got %q", tt.input, tokens[0].Value)
			}
		})
	}
}

func TestTokenizeNumbers(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"100"},
		{"1000"},
		{"65535"},
		{"0"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			tokens := l.Tokenize()
			if len(tokens) != 1 {
				t.Fatalf("expected 1 token, got %d", len(tokens))
			}
			if tokens[0].Type != TokenNumber {
				t.Errorf("expected TokenNumber for %q, got %v", tt.input, tokens[0].Type)
			}
		})
	}
}

func TestTokenizeCommunity(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		community string
	}{
		{"basic", "community 65000:100", "65000:100"},
		{"single digit", "community 65001:1", "65001:1"},
		{"small numbers", "community 100:200", "100:200"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			tokens := l.Tokenize()
			// Find the community token
			found := false
			for _, tok := range tokens {
				if tok.Type == TokenCommunity {
					found = true
					if tok.Value != tt.community {
						t.Errorf("expected community value %q, got %q", tt.community, tok.Value)
					}
				}
			}
			if !found {
				t.Errorf("expected TokenCommunity in %q, token types: %v", tt.input, tokenTypes(tokens))
			}
		})
	}
}

func TestCommunityFalsePositive(t *testing.T) {
	// Without "community" keyword context, digit:digit should not match as community
	tests := []string{"12:00", "3:45", "100:200"}
	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			l := New(input)
			tokens := l.Tokenize()
			for _, tok := range tokens {
				if tok.Type == TokenCommunity {
					t.Errorf("%q should not be TokenCommunity without community keyword context", input)
				}
			}
		})
	}
}

// tokenTypes is a test helper that returns token type names for debugging
func tokenTypes(tokens []Token) []string {
	types := make([]string, len(tokens))
	for i, t := range tokens {
		types[i] = t.Type.String()
	}
	return types
}

func TestTokenizeASN(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"AS65000", TokenASN},
		{"AS1", TokenASN},
		{"as65001", TokenASN},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			tokens := l.Tokenize()
			if len(tokens) != 1 {
				t.Fatalf("expected 1 token, got %d", len(tokens))
			}
			if tokens[0].Type != tt.expected {
				t.Errorf("expected %v for %q, got %v", tt.expected, tt.input, tokens[0].Type)
			}
		})
	}
}

func TestTokenizeFullLine(t *testing.T) {
	input := "ip address 192.168.1.1 255.255.255.0"

	l := New(input)
	tokens := l.Tokenize()

	expected := []struct {
		tokenType TokenType
		value     string
	}{
		{TokenCommand, "ip"},
		{TokenText, " "},
		{TokenKeyword, "address"},
		{TokenText, " "},
		{TokenIPv4, "192.168.1.1"},
		{TokenText, " "},
		{TokenIPv4, "255.255.255.0"},
	}

	if len(tokens) != len(expected) {
		t.Fatalf("expected %d tokens, got %d", len(expected), len(tokens))
	}

	for i, exp := range expected {
		if tokens[i].Type != exp.tokenType {
			t.Errorf("token %d: expected type %v, got %v (value: %q)", i, exp.tokenType, tokens[i].Type, tokens[i].Value)
		}
		if tokens[i].Value != exp.value {
			t.Errorf("token %d: expected value %q, got %q", i, exp.value, tokens[i].Value)
		}
	}
}

func TestTokenizeAccessList(t *testing.T) {
	input := "permit tcp 10.0.0.0 0.0.255.255 any eq 22"

	l := New(input)
	tokens := l.Tokenize()

	// Verify key token types
	foundAction := false
	foundProtocol := false
	foundOperator := false
	foundIP := false

	for _, tok := range tokens {
		switch tok.Type {
		case TokenAction:
			if tok.Value == "permit" {
				foundAction = true
			}
		case TokenProtocol:
			if tok.Value == "tcp" {
				foundProtocol = true
			}
		case TokenOperator:
			if tok.Value == "eq" || tok.Value == "any" {
				foundOperator = true
			}
		case TokenIPv4:
			foundIP = true
		}
	}

	if !foundAction {
		t.Error("expected to find action 'permit'")
	}
	if !foundProtocol {
		t.Error("expected to find protocol 'tcp'")
	}
	if !foundOperator {
		t.Error("expected to find operator")
	}
	if !foundIP {
		t.Error("expected to find IP address")
	}
}

func TestTokenizeKeywords(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"description", TokenKeyword},
		{"address", TokenKeyword},
		{"switchport", TokenKeyword},
		{"speed", TokenKeyword},
		{"duplex", TokenKeyword},
		{"mtu", TokenKeyword},
		{"bandwidth", TokenKeyword},
		{"network", TokenKeyword},
		{"neighbor", TokenKeyword},
		{"redistribute", TokenKeyword},
		{"area", TokenKeyword},
		{"remote-as", TokenKeyword},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			tokens := l.Tokenize()
			if len(tokens) != 1 {
				t.Fatalf("expected 1 token, got %d", len(tokens))
			}
			if tokens[0].Type != tt.expected {
				t.Errorf("expected %v for %q, got %v", tt.expected, tt.input, tokens[0].Type)
			}
		})
	}
}

func TestTokenPosition(t *testing.T) {
	input := "interface\nhostname"
	l := New(input)
	tokens := l.Tokenize()

	if tokens[0].Line != 1 {
		t.Errorf("expected line 1 for 'interface', got %d", tokens[0].Line)
	}

	for _, tok := range tokens {
		if tok.Value == "hostname" {
			if tok.Line != 2 {
				t.Errorf("expected line 2 for 'hostname', got %d", tok.Line)
			}
			break
		}
	}
}

func TestEmptyInput(t *testing.T) {
	l := New("")
	tokens := l.Tokenize()
	if len(tokens) != 0 {
		t.Errorf("expected 0 tokens for empty input, got %d", len(tokens))
	}
}

func TestWhitespaceOnly(t *testing.T) {
	l := New("   \t\n  ")
	tokens := l.Tokenize()
	for _, tok := range tokens {
		if tok.Type != TokenText {
			t.Errorf("expected TokenText for whitespace, got %v", tok.Type)
		}
	}
}

// ==================== Show Output Tests ====================

func TestTokenizeStatesGood(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"up", TokenStateGood},
		{"connected", TokenStateGood},
		{"established", TokenStateGood},
		{"full", TokenStateGood},
		{"enabled", TokenStateGood},
		{"active", TokenStateGood},
		{"forwarding", TokenStateGood},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			l.SetParseMode(ParseModeShow)
			tokens := l.Tokenize()
			if len(tokens) != 1 {
				t.Fatalf("expected 1 token, got %d", len(tokens))
			}
			if tokens[0].Type != tt.expected {
				t.Errorf("expected %v for %q, got %v", tt.expected, tt.input, tokens[0].Type)
			}
		})
	}
}

func TestTokenizeStatesBad(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"down", TokenStateBad},
		{"notconnect", TokenStateBad},
		{"err-disabled", TokenStateBad},
		{"disabled", TokenStateBad},
		{"failed", TokenStateBad},
		{"idle", TokenStateBad},
		{"connect", TokenStateBad},
		{"opensent", TokenStateBad},
		{"openconfirm", TokenStateBad},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			l.SetParseMode(ParseModeShow)
			tokens := l.Tokenize()
			if len(tokens) != 1 {
				t.Fatalf("expected 1 token, got %d", len(tokens))
			}
			if tokens[0].Type != tt.expected {
				t.Errorf("expected %v for %q, got %v", tt.expected, tt.input, tokens[0].Type)
			}
		})
	}
}

func TestTokenizeStatesWarning(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"init", TokenStateWarning},
		{"2way", TokenStateWarning},
		{"exstart", TokenStateWarning},
		{"exchange", TokenStateWarning},
		{"loading", TokenStateWarning},
		{"attempt", TokenStateWarning},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			l.SetParseMode(ParseModeShow)
			tokens := l.Tokenize()
			if len(tokens) != 1 {
				t.Fatalf("expected 1 token, got %d", len(tokens))
			}
			if tokens[0].Type != tt.expected {
				t.Errorf("expected %v for %q, got %v", tt.expected, tt.input, tokens[0].Type)
			}
		})
	}
}

func TestTokenizeStatesNeutral(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"inactive", TokenStateNeutral},
		{"standby", TokenStateNeutral},
		{"backup", TokenStateNeutral},
		{"suspended", TokenStateNeutral},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			l.SetParseMode(ParseModeShow)
			tokens := l.Tokenize()
			if len(tokens) != 1 {
				t.Fatalf("expected 1 token, got %d", len(tokens))
			}
			if tokens[0].Type != tt.expected {
				t.Errorf("expected %v for %q, got %v", tt.expected, tt.input, tokens[0].Type)
			}
		})
	}
}

func TestTokenizeTimeDurations(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"1w", TokenTimeDuration},
		{"2d", TokenTimeDuration},
		{"3h", TokenTimeDuration},
		{"1w2d", TokenTimeDuration},
		{"1w2d3h", TokenTimeDuration},
		{"0:45", TokenTimeDuration},
		{"0:45:30", TokenTimeDuration},
		{"12:00:00", TokenTimeDuration},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			l.SetParseMode(ParseModeShow)
			tokens := l.Tokenize()
			if len(tokens) != 1 {
				t.Fatalf("expected 1 token, got %d", len(tokens))
			}
			if tokens[0].Type != tt.expected {
				t.Errorf("expected %v for %q, got %v", tt.expected, tt.input, tokens[0].Type)
			}
		})
	}
}

func TestTokenizeCompoundStates(t *testing.T) {
	l := New("up/up")
	l.SetParseMode(ParseModeShow)
	tokens := l.Tokenize()
	if len(tokens) != 1 {
		t.Fatalf("expected 1 token, got %d", len(tokens))
	}
	if tokens[0].Type != TokenStateGood {
		t.Errorf("expected TokenStateGood for 'up/up', got %v", tokens[0].Type)
	}
}

func TestParseModeDetection(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected ParseMode
	}{
		{
			name:     "cisco config",
			input:    "hostname router\ninterface GigabitEthernet0/0/0\n ip address 10.0.0.1 255.255.255.0\n no shutdown",
			expected: ParseModeConfig,
		},
		{
			name:     "cisco config with bangs",
			input:    "!\nhostname router\n!\ninterface GigabitEthernet0/0/0\n ip address 10.0.0.1 255.255.255.0\n!",
			expected: ParseModeConfig,
		},
		{
			name:     "show interface",
			input:    "GigabitEthernet0/0/0 is up, line protocol is up\n  Internet address is 203.0.113.1/24\n  5 minute input rate 1000 bits/sec",
			expected: ParseModeShow,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			mode := l.detectParseMode()
			if mode != tt.expected {
				t.Errorf("expected mode %v, got %v", tt.expected, mode)
			}
		})
	}
}

func TestSetParseMode(t *testing.T) {
	l := New("up")

	if l.GetParseMode() != ParseModeAuto {
		t.Errorf("expected default mode ParseModeAuto, got %v", l.GetParseMode())
	}

	l.SetParseMode(ParseModeShow)
	if l.GetParseMode() != ParseModeShow {
		t.Errorf("expected ParseModeShow, got %v", l.GetParseMode())
	}

	tokens := l.Tokenize()
	if len(tokens) != 1 || tokens[0].Type != TokenStateGood {
		t.Errorf("expected TokenStateGood for 'up' in show mode")
	}
}

func TestShowModePreservesSharedPatterns(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"192.168.1.1", TokenIPv4},
		{"10.0.0.0/24", TokenIPv4Prefix},
		{"GigabitEthernet0/0/0", TokenInterface},
		{"65001", TokenNumber},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			l.SetParseMode(ParseModeShow)
			tokens := l.Tokenize()
			if len(tokens) != 1 {
				t.Fatalf("expected 1 token, got %d", len(tokens))
			}
			if tokens[0].Type != tt.expected {
				t.Errorf("expected %v for %q in show mode, got %v", tt.expected, tt.input, tokens[0].Type)
			}
		})
	}
}

func TestCiscoPromptDetection(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"Router>", true},
		{"Router#", true},
		{"Router(config)#", true},
		{"Router(config-if)#", true},
		{"core-rtr-01>", true},
		{"core-rtr-01#", true},
		{"core-rtr-01(config-router)#", true},
		{"Hello world", false},
		{"SELECT * FROM users", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := IsPrompt(tt.input)
			if result != tt.expected {
				t.Errorf("IsPrompt(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestTokenizePrompt(t *testing.T) {
	input := "Router>"
	l := New(input)
	tokens := l.Tokenize()

	if len(tokens) < 2 {
		t.Fatalf("expected at least 2 tokens for prompt, got %d", len(tokens))
	}

	// First token should be hostname
	if tokens[0].Type != TokenPromptHost {
		t.Errorf("expected TokenPromptHost, got %v", tokens[0].Type)
	}
	if tokens[0].Value != "Router" {
		t.Errorf("expected 'Router', got %q", tokens[0].Value)
	}

	// Second token should be prompt char
	if tokens[1].Type != TokenPromptOper {
		t.Errorf("expected TokenPromptOper, got %v", tokens[1].Type)
	}
}

func TestTokenizePromptWithMode(t *testing.T) {
	input := "Router(config-if)#"
	l := New(input)
	tokens := l.Tokenize()

	if len(tokens) < 3 {
		t.Fatalf("expected at least 3 tokens for prompt with mode, got %d", len(tokens))
	}

	if tokens[0].Type != TokenPromptHost {
		t.Errorf("expected TokenPromptHost, got %v", tokens[0].Type)
	}
	if tokens[1].Type != TokenPromptMode {
		t.Errorf("expected TokenPromptMode, got %v (value: %q)", tokens[1].Type, tokens[1].Value)
	}
	if tokens[2].Type != TokenPromptConf {
		t.Errorf("expected TokenPromptConf, got %v", tokens[2].Type)
	}
}

func TestTokenizePromptWithCommand(t *testing.T) {
	input := "Router# show ip interface brief"
	l := New(input)
	tokens := l.Tokenize()

	// Should have hostname, #, space, and then command tokens
	if tokens[0].Type != TokenPromptHost {
		t.Errorf("expected TokenPromptHost, got %v", tokens[0].Type)
	}

	// Find the prompt char
	foundConf := false
	for _, tok := range tokens {
		if tok.Type == TokenPromptConf {
			foundConf = true
		}
	}
	if !foundConf {
		t.Error("expected to find TokenPromptConf")
	}
}
