# Antigravity Coding Filter

A CLIProxyAPI dynamic plugin that normalizes the system prompt of coding-agent
clients before the request reaches the Antigravity upstream, so the traffic looks
like Antigravity traffic instead of a third-party coding tool.

Verified end to end on CLIProxyAPI v8 with Codex (OpenAI Responses) and
DeepSeek Harness (OpenAI Chat Completions).

## What it does

The plugin hooks `request.intercept_before`, reads the system prompt out of the
inbound request, and rewrites the client's self-identification (`Codex`,
`Claude Code`, `Cursor`, ...) to `Antigravity`. Nothing else in the payload is
touched: user turns, tool calls, and tool output are left exactly as sent.

Matching is case-insensitive and restricted to the system prompt, so a user
message that happens to mention one of these names is never modified.

## Supported protocols

The rewrite target is resolved from the inbound `SourceFormat` reported by the
host, so every protocol the host accepts is covered:

| SourceFormat | Where the system text lives |
|---|---|
| `claude`, `antigravity` | `system` (string or `[{type:"text"}]`) |
| `openai` | `messages[].content` where `role` is `system` or `developer` |
| `openai-response`, `codex` | `instructions` and `input[].content` where `role` is `system` or `developer` |
| `gemini` | `systemInstruction.parts[].text` |

Unknown formats pass through untouched - the plugin never guesses a field name.

## Modes

**`rewrite`** (recommended) replaces matched names with `Antigravity` and
forwards the request. Use this when the proxy is driven *by* those clients.

**`block`** is meant to reject matched requests with HTTP `403`
(`blocked_by_antigravity_coding_filter`).

> **Known limitation:** `block` does not currently take effect. It routes to the
> plugin's own executor through `model.route`, but the host only accepts a
> self-route when the plugin resolves an executor provider identifier
> (`executor.identifier` or a registered model provider). This plugin declares
> neither, so the host logs
> `pluginhost: model router returned unavailable executor plugin` and falls back
> to normal routing - matched requests are forwarded, not rejected. Rewrite mode
> is unaffected. The fix uses the v8 `Terminate` field on the interceptor
> response.

## Built-in keywords

Enabled by default:

- Claude Code, OpenAI Codex / Codex CLI, OpenCode
- GitHub Copilot / Copilot CLI, Gemini Code Assist / Gemini CLI
- Cursor, Windsurf / Codeium, Cline, Roo Code, Kilo Code, Aider, Continue.dev
- Amazon Q Developer / CodeWhisperer, JetBrains AI Assistant / Junie, Kiro
- Qoder / Qoder CLI, Qwen Code, Trae, Tabnine, Sourcegraph Cody, Augment Code
- Replit Agent / Ghostwriter, Devin, OpenHands, SWE-agent, Goose
- Zed AI, Void Editor, PearAI, Refact.ai, Tabby, GitLab Duo, Visual Studio IntelliCode
- CodeBuddy, Blackbox AI, Pieces for Developers, Qodo / CodiumAI, Rovo Dev CLI, Factory Droid
- OpenClaw (including Clawdbot and Moltbot), Hermes Agent, WorkBuddy

Anything missing (for example `DeepSeek Harness`) can be added through
`custom_mappings`.

## Configuration

```yaml
plugins:
  enabled: true
  dir: "plugins"
  configs:
    antigravity-coding-filter:
      enabled: true
      priority: 1
      mode: rewrite
      use_default_keywords: true
      custom_mappings:
        DeepSeek Harness: Antigravity
```

In `rewrite` mode a mapping's key is replaced by its value. In `block` mode the
key is an additional blocked keyword and the value is ignored. `custom_mappings`
also accepts a comma- or newline-delimited `from: to` string. Blank entries and
duplicate source names are ignored.

CLIProxyAPI searches `plugins/<GOOS>/<GOARCH>-<variant>`, then
`plugins/<GOOS>/<GOARCH>`, then `plugins`. The plugin ID comes from the library
filename, so it must be `antigravity-coding-filter`.

## Build

CGO is required.

```powershell
$env:CGO_ENABLED = 1

# Windows
go build -buildmode=c-shared -o plugins/windows/amd64/antigravity-coding-filter.dll .
Remove-Item plugins/windows/amd64/antigravity-coding-filter.h

# Linux
go build -buildmode=c-shared -o plugins/linux/amd64/antigravity-coding-filter.so .

# macOS
go build -buildmode=c-shared -o plugins/darwin/arm64/antigravity-coding-filter.dylib .
```

## Verify at runtime

Start CLIProxyAPI and confirm the plugin was picked up:

```text
pluginhost: plugin loaded plugin_id=antigravity-coding-filter
pluginhost: plugin registered plugin_id=antigravity-coding-filter version=0.3.0
```

For an end-to-end check, enable host request logging and compare the two bodies
in `auths/logs/v1-*.log`: the inbound body still carries the client's own
identity, while the upstream body carries `Antigravity`.

## Tests

```bash
go test ./...
```

`format_test.go` covers one rewrite assertion and one classification assertion
per inbound protocol, plus pass-through for unknown formats and user turns.

## License

MIT
