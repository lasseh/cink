# cink

**Cisco INK.** Terminal syntax highlighter for Cisco IOS/IOS-XE configuration with real-time SSH session highlighting and pipe support.

![Demo](https://img.shields.io/badge/go-1.22+-blue.svg)

## Features

- Real-time syntax highlighting for SSH sessions
- Pipe configuration files for highlighted output
- Multiple color themes (Tokyo Night, Monokai, Nord, Solarized, etc.)
- Auto-detection of Cisco IOS content with force mode override
- Recognizes Cisco-specific syntax:
  - Commands (`interface`, `router`, `ip`, `show`, `configure`, etc.)
  - Sections (`interface`, `router`, `line`, `access-list`, `route-map`, etc.)
  - Protocols (`ospf`, `bgp`, `eigrp`, `tcp`, `udp`, `ssh`, etc.)
  - Interfaces (`GigabitEthernet0/0/0`, `Gi0/0/0`, `Loopback0`, `Vlan100`, `Po1`, etc.)
  - IP addresses (IPv4, IPv6, prefixes)
  - MAC addresses (Cisco dotted format `0011.2233.4455`)
  - ACL actions (`permit`, `deny`) and operators (`eq`, `gt`, `any`, `host`)
  - Negation (`no` prefix highlighted distinctly)
  - Comments (`!` section separators)
  - Show output states (`up`/`down`, `connected`/`notconnect`, `err-disabled`, etc.)
  - Cisco CLI prompts (`Router>`, `Router#`, `Router(config-if)#`)

![Theme Demo](.github/cink-demo-theme.png "Themes")

## See Also

If you're lucky enough to work with Juniper instead, check out [jink](https://github.com/lasseh/jink) — syntax highlighting for a syntax worth highlighting.

## Installation

### From Source

```bash
git clone https://github.com/lasseh/cink.git
cd cink
make build
make install  # Installs to $GOPATH/bin or ~/go/bin
```

### Go Install

```bash
go install github.com/lasseh/cink/cmd/cink@latest
```

## Usage

### SSH Sessions

Wrap your SSH command with `cink` for real-time highlighting:

```bash
cink ssh admin@192.168.1.1
cink ssh -p 2222 admin@router.example.com
```

### Pipe Configuration Files

```bash
cat router-config.conf | cink
cink < backup-config.txt
```

### Select a Theme

```bash
cink -t monokai ssh admin@router
cink -t nord < config.conf
cat config.conf | cink -t solarized
```

### Force Highlighting

Skip auto-detection and always highlight (useful when detection fails):

```bash
cat config.conf | cink -f
ssh router "show running-config" | cink --force
```

## Themes

| Theme | Description |
|-------|-------------|
| `tokyonight` | Tokyo Night - soft, modern colors (default) |
| `vibrant` | Bright, high-contrast colors |
| `solarized` | Solarized Dark color scheme |
| `monokai` | Monokai-inspired colors |
| `nord` | Nord color palette |
| `catppuccin` | Catppuccin Mocha - pastel colors |
| `dracula` | Dracula - popular dark theme |
| `gruvbox` | Gruvbox Dark - retro groove |
| `onedark` | Atom One Dark |

Preview all themes:

```bash
make demo-all
```

## Shell Aliases

Create an alias to use `cink` as a drop-in replacement for `ssh`:

### Bash (~/.bashrc)

```bash
alias ssh='cink ssh'
alias cssh='cink ssh'
```

### Zsh (~/.zshrc)

```zsh
alias ssh='cink ssh'
alias cssh='cink ssh'
```

### Fish (~/.config/fish/config.fish)

```fish
alias ssh 'cink ssh'
alias cssh 'cink ssh'
```

After adding the alias, reload your shell or run `source ~/.bashrc` (or equivalent).

Now you can use `ssh` or `cssh` with automatic highlighting:

```bash
ssh admin@router           # Uses cink automatically
cssh admin@router          # Dedicated alias for Cisco highlighting
ssh -p 2222 admin@router   # All SSH arguments work normally
```

## Examples

### Basic SSH Session

```bash
$ cink ssh admin@core-router
core-router> show running-config
```

### View Configuration with Highlighting

```bash
# From a file
cat /var/log/cisco-backup.conf | cink

# From clipboard (macOS)
pbpaste | cink

# From clipboard (Linux)
xclip -o | cink
```

### Compare Themes

```bash
# Show sample config in each theme
make demo        # Default theme (Tokyo Night)
make demo-all    # All themes side by side
make demo-show   # Show command output demo
```

## Building

```bash
make build       # Build binaries to build/
make install     # Install to Go bin directory
make test        # Run tests
make clean       # Clean build artifacts
```

## Command Line Reference

```
cink [OPTIONS] [command] [args...]

OPTIONS:
    -f, --force           Always highlight (skip auto-detection)
    -t, --theme <name>    Color theme (see Themes section)
    -n, --no-highlight    Disable highlighting (pass-through mode)
    -v, --version         Show version
    -h, --help            Show help

EXAMPLES:
    cink ssh admin@192.168.1.1
    cink -t monokai ssh admin@router
    cat config.conf | cink
    cat config.conf | cink -f
    cink < config.conf
```

## Library Usage

Use `cink` as a Go library in your own projects:

```bash
go get github.com/lasseh/cink
```

### Simple Highlighting

```go
import "github.com/lasseh/cink/highlighter"

// One-liner with default theme
colored := highlighter.Highlight(config)
fmt.Println(colored)
```

### With Custom Theme

```go
import "github.com/lasseh/cink/highlighter"

// Use a specific theme
hl := highlighter.NewWithTheme(highlighter.MonokaiTheme())
colored := hl.Highlight(config)

// Or get theme by name
theme := highlighter.ThemeByName("nord")
hl := highlighter.NewWithTheme(theme)

// List available themes
themes := highlighter.ThemeNames() // ["tokyonight", "vibrant", "solarized", ...]
```

### Tokenization (for custom rendering)

```go
import "github.com/lasseh/cink/lexer"

lex := lexer.New(config)
tokens := lex.Tokenize()

for _, tok := range tokens {
    fmt.Printf("%s: %q\n", tok.Type, tok.Value)
    // Output: Command: "interface"
    //         Interface: "GigabitEthernet0/0/0"
    //         Negation: "no"
    //         Command: "shutdown"
    //         IPv4: "192.168.1.1"
}
```

### Available Packages

| Package | Description |
|---------|-------------|
| `highlighter` | ANSI color highlighting with theme support |
| `lexer` | Tokenizer for Cisco IOS config and show output |
| `terminal` | PTY wrapper for real-time highlighting (CLI-specific) |

## How It Works

1. **Lexer**: Tokenizes Cisco IOS/IOS-XE configuration text into meaningful tokens (commands, sections, IPs, interfaces, negation, etc.)
2. **Highlighter**: Applies ANSI color codes based on token types and selected theme
3. **Terminal**: Wraps commands with a PTY for real-time output processing

The highlighter includes heuristics to detect Cisco IOS configuration and show output, avoiding highlighting of unrelated text.

## License

MIT
