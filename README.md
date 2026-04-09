# postiz-cli

Postiz CLI — Social media scheduling from the command line.

## Installation

### Download Binary

Download the latest release from [GitHub Releases](https://github.com/Robben-Media/postiz-cli/releases).

### Build from Source

```bash
git clone https://github.com/Robben-Media/postiz-cli.git
cd postiz-cli
go build ./cmd/postiz
```

## Configuration

postiz-cli requires a Postiz API key. Credentials are stored securely in your system keyring.

**Store credentials:**

```bash
postiz-cli auth set-key
```

The CLI will prompt for your API key interactively. You can also pipe it:

```bash
echo "your-api-key" | postiz-cli auth set-key
```

**Environment variable override:**

```bash
export POSTIZ_API_KEY="your-api-key"
```

**Check status:**

```bash
postiz-cli auth status
```

**Remove credentials:**

```bash
postiz-cli auth remove
```

## Commands

### auth

Manage API credentials.

| Command | Description |
|---------|-------------|
| `auth set-key` | Store API key in keyring |
| `auth status` | Show authentication status |
| `auth remove` | Remove stored credentials |

### integrations

| Command | Description |
|---------|-------------|
| `integrations list` | List all connected integrations |
| `integrations check <id>` | Check connection status for an integration |
| `integrations find-slot <id>` | Find next available posting slot for an integration |

### posts

| Command | Description |
|---------|-------------|
| `posts list` | List posts |
| `posts create` | Create a new post |
| `posts delete <id>` | Delete a post |

**Flags (list):** `--from` (YYYY-MM-DD), `--to` (YYYY-MM-DD)

**Flags (create):** `--type` (now, schedule, or draft; default now), `--date` (ISO 8601, for type=schedule), `--content` (post text), `--integration` (repeatable), `--json-input` (JSON file path or `-` for stdin)

### uploads

| Command | Description |
|---------|-------------|
| `uploads file <path>` | Upload a file from disk |
| `uploads url <url>` | Upload from a URL |

## Global Flags

| Flag | Description |
|------|-------------|
| `--json` | Output JSON to stdout (best for scripting) |
| `--plain` | Output stable, parseable text to stdout (TSV; no colors) |
| `--verbose` | Enable verbose logging |
| `--force` | Skip confirmations for destructive commands |
| `--no-input` | Never prompt; fail instead (useful for CI) |
| `--color` | Color output: auto, always, or never (default auto) |

## License

MIT
