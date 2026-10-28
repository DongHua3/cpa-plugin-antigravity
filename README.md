# Antigravity Coding Filter

CLIProxyAPI dynamic plugin that rewrites or blocks non-Antigravity coding client
signals inside the request's system prompt.

Verified on CLIProxyAPI v8 with Codex (OpenAI Responses) and DeepSeek Harness
(OpenAI Chat Completions). Built against the v7.2.93 plugin SDK.

## Supported protocols

The rewrite target is resolved from the inbound `SourceFormat` reported by the
host, so every protocol the host accepts is covered:

| SourceFormat | Where the system text lives |
|---|---|
| `claude`, `antigravity` | `system` (string or `[{type:"text"}]`) |
| `openai` | `messages[].content` where `role` is `system` or `developer` |
| `openai-response`, `codex` | `instructions` and `input[].content` where `role` is `system` or `developer` |
| `gemini` | `systemInstruction.parts[].text` |

Unknown formats are passed through untouched - the plugin never guesses a field
name. Matching is case-insensitive and only ever inspects the system text, so
user turns and tool output are never modified.

## Filter modes

- `rewrite` (recommended): replace matched names with `Antigravity` and forward
  the request. This is the mode to use when you drive this proxy *with* those
  clients - blocking them would reject your own traffic.
- `block`: reject the request with HTTP `403` and a
  `blocked_by_antigravity_coding_filter` error.

**Known limitation:** `block` mode does not currently take effect. It routes to
the plugin's own executor via `model.route`, but the host requires an executor
provider identifier (`executor.identifier` or a registered model provider) before
it accepts a self-route, and this plugin declares neither. The host logs
`pluginhost: model router returned unavailable executor plugin` and falls back to
normal routing, so matched requests are forwarded instead of rejected. Rewrite
mode is unaffected because it runs through `request.intercept_before`. The fix
(using the v8 `Terminate` field on the interceptor response) is tracked for v0.4.

## Built-in keywords

The preset is enabled by default and covers major AI coding editors, assistants,
and terminal agents:

- Claude Code, OpenAI Codex / Codex CLI, OpenCode
- GitHub Copilot / Copilot CLI, Gemini Code Assist / Gemini CLI
- Cursor, Windsurf / Codeium, Cline, Roo Code, Kilo Code, Aider, Continue.dev
- Amazon Q Developer / CodeWhisperer, JetBrains AI Assistant / Junie, Kiro
- Qoder / Qoder CLI, Qwen Code, Trae, Tabnine, Sourcegraph Cody, Augment Code
- Replit Agent / Ghostwriter, Devin, OpenHands, SWE-agent, Goose
- Zed AI, Void Editor, PearAI, Refact.ai, Tabby, GitLab Duo, Visual Studio IntelliCode
- CodeBuddy, Blackbox AI, Pieces for Developers, Qodo / CodiumAI, Rovo Dev CLI, Factory Droid
- OpenClaw (including Clawdbot and Moltbot), Hermes Agent, WorkBuddy

Add anything missing (for example `DeepSeek Harness`) through `custom_mappings`.

## Mapping configuration

```yaml
plugins:
  configs:
    antigravity-coding-filter:
      enabled: true
      priority: 1
      mode: rewrite
      use_default_keywords: true
      custom_mappings:
        Cursor: Antigravity
        Windsurf: Antigravity
        DeepSeek Harness: Antigravity
```

In `block` mode each `custom_mappings` key is an additional blocked keyword and
its value is ignored. In `rewrite` mode the key is replaced with its value. The
field also accepts a comma- or newline-delimited `from: to` string. Blank entries
and duplicate source names are ignored.

## Build

CLIProxyAPI dynamic plugins require CGO. Confirm `CGO_ENABLED=1` and a C
toolchain (`gcc`/`clang`) on the build host before building.

Windows amd64:

```powershell
go build -buildmode=c-shared -o plugins/windows/amd64/antigravity-coding-filter.dll .
Remove-Item plugins/windows/amd64/antigravity-coding-filter.h
```

The plugin ID is derived from the dynamic library filename, so this build path
registers the plugin as `antigravity-coding-filter`.

## CLIProxyAPI config

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
      custom_mappings: {}
```

CLIProxyAPI searches `plugins/<GOOS>/<GOARCH>-<variant>`, then
`plugins/<GOOS>/<GOARCH>`, then `plugins`.

## Runtime verification

After starting CLIProxyAPI, check the host log for:

```text
pluginhost: plugin loaded plugin_id=antigravity-coding-filter
pluginhost: plugin registered plugin_id=antigravity-coding-filter version=0.3.0
```

Enable host request logging to confirm the rewrite end to end: the inbound body
still carries the client's own identity while the upstream body carries
`Antigravity`.

## Tests

```powershell
go test ./...
```

`format_test.go` covers one rewrite assertion and one classification assertion
per inbound protocol, plus pass-through for unknown formats and user turns.
