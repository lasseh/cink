package lexer

// TokenType represents the type of a lexical token
type TokenType int

const (
	TokenText       TokenType = iota
	TokenCommand              // interface, router, ip, show, configure, etc.
	TokenSection              // interface, router, line, access-list, etc. (section headers)
	TokenProtocol             // ospf, bgp, eigrp, tcp, udp, etc.
	TokenAction               // permit, deny, log, match, set
	TokenInterface            // GigabitEthernet0/0/0, Gi0/0/0, Loopback0, etc.
	TokenIPv4                 // 192.168.1.1
	TokenIPv4Prefix           // 192.168.1.0/24
	TokenIPv6                 // 2001:db8::1
	TokenIPv6Prefix           // 2001:db8::/32
	TokenMAC                  // 0011.2233.4455 (Cisco dotted format)
	TokenNumber               // 100, 1000
	TokenString               // "quoted string"
	TokenComment              // ! comment/section separator
	TokenIdentifier           // generic identifier
	TokenKeyword              // other important keywords
	TokenOperator             // operators like eq, gt, lt, neq, range, ge, le
	TokenASN                  // AS numbers
	TokenCommunity            // BGP communities
	TokenValue                // Values after keywords (description, hostname, etc.)
	TokenNegation             // "no" prefix for negation

	// Show output semantic tokens
	TokenStateGood    // up, connected, established, full, enabled
	TokenStateBad     // down, notconnect, err-disabled, disabled
	TokenStateWarning // init, 2way, exstart, exchange, loading
	TokenStateNeutral // inactive, standby, backup, suspended

	// Show output structural tokens
	TokenColumnHeader  // Table column headers
	TokenStatusSymbol  // *, +, -, > (route markers)
	TokenTimeDuration  // 1d 2:30:45, 1w2d, 0:05:10
	TokenPercentage    // 50%, 99.9%
	TokenByteSize      // 1.5G, 500M, 10K
	TokenRouteProtocol // [BGP/170], [OSPF/10], [Static/5]

	// Prompt tokens (simplified for Cisco: no user@host format)
	TokenPromptHost // hostname portion of prompt
	TokenPromptMode // (config), (config-if), etc.
	TokenPromptOper // > (user EXEC mode prompt char)
	TokenPromptConf // # (privileged EXEC / config mode prompt char)
)

// Token represents a single lexical token
type Token struct {
	Type   TokenType
	Value  string
	Line   int
	Column int
}

// String returns a string representation of the token type
func (t TokenType) String() string {
	switch t {
	case TokenText:
		return "Text"
	case TokenCommand:
		return "Command"
	case TokenSection:
		return "Section"
	case TokenProtocol:
		return "Protocol"
	case TokenAction:
		return "Action"
	case TokenInterface:
		return "Interface"
	case TokenIPv4:
		return "IPv4"
	case TokenIPv4Prefix:
		return "IPv4Prefix"
	case TokenIPv6:
		return "IPv6"
	case TokenIPv6Prefix:
		return "IPv6Prefix"
	case TokenMAC:
		return "MAC"
	case TokenNumber:
		return "Number"
	case TokenString:
		return "String"
	case TokenComment:
		return "Comment"
	case TokenIdentifier:
		return "Identifier"
	case TokenKeyword:
		return "Keyword"
	case TokenOperator:
		return "Operator"
	case TokenASN:
		return "ASN"
	case TokenCommunity:
		return "Community"
	case TokenValue:
		return "Value"
	case TokenNegation:
		return "Negation"
	case TokenStateGood:
		return "StateGood"
	case TokenStateBad:
		return "StateBad"
	case TokenStateWarning:
		return "StateWarning"
	case TokenStateNeutral:
		return "StateNeutral"
	case TokenColumnHeader:
		return "ColumnHeader"
	case TokenStatusSymbol:
		return "StatusSymbol"
	case TokenTimeDuration:
		return "TimeDuration"
	case TokenPercentage:
		return "Percentage"
	case TokenByteSize:
		return "ByteSize"
	case TokenRouteProtocol:
		return "RouteProtocol"
	case TokenPromptHost:
		return "PromptHost"
	case TokenPromptMode:
		return "PromptMode"
	case TokenPromptOper:
		return "PromptOper"
	case TokenPromptConf:
		return "PromptConf"
	default:
		return "Unknown"
	}
}
