package main

import (
	"strings"
	"testing"
)

func TestRewriteCoversEachInboundProtocol(t *testing.T) {
	cfg := defaultFilterConfig()
	cases := []struct {
		name   string
		format string
		body   string
		banned string
	}{
		{
			name:   "claude",
			format: "claude",
			body:   `{"system":"You are Claude Code.","messages":[{"role":"user","content":"hi"}]}`,
			banned: "Claude Code",
		},
		{
			name:   "responses",
			format: "openai-response",
			body:   `{"instructions":"You are Codex.","input":"hi"}`,
			banned: "Codex",
		},
		{
			name:   "chat-completions",
			format: "openai",
			body:   `{"messages":[{"role":"system","content":"You are Codex."},{"role":"user","content":"hi"}]}`,
			banned: "Codex",
		},
		{
			name:   "gemini",
			format: "gemini",
			body:   `{"systemInstruction":{"parts":[{"text":"You are Codex."}]},"contents":[]}`,
			banned: "Codex",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			out, changed := rewriteRequestBodyForFormat([]byte(testCase.body), testCase.format, cfg)
			if !changed {
				t.Fatalf("format %q did not rewrite", testCase.format)
			}
			if strings.Contains(string(out), testCase.banned) {
				t.Fatalf("payload still contains %q: %s", testCase.banned, out)
			}
			if !strings.Contains(string(out), "Antigravity") {
				t.Fatalf("payload has no replacement: %s", out)
			}
		})
	}
}

func TestRewriteLeavesUnknownFormatAndUserMessagesAlone(t *testing.T) {
	cfg := defaultFilterConfig()

	if _, changed := rewriteRequestBodyForFormat([]byte(`{"system":"Claude Code"}`), "unknown-protocol", cfg); changed {
		t.Fatal("unknown protocol must not be rewritten")
	}

	userOnly := `{"messages":[{"role":"user","content":"You are Codex"}]}`
	if _, changed := rewriteRequestBodyForFormat([]byte(userOnly), "openai", cfg); changed {
		t.Fatal("user messages must not be rewritten")
	}
}

func TestClassifyCoversEachInboundProtocol(t *testing.T) {
	cfg := defaultFilterConfig()
	cases := []struct {
		name   string
		format string
		body   string
	}{
		{"claude", "claude", `{"system":"You are Claude Code."}`},
		{"responses", "openai-response", `{"instructions":"You are Codex."}`},
		{"chat-completions", "openai", `{"messages":[{"role":"system","content":"You are Codex."}]}`},
		{"gemini", "gemini", `{"systemInstruction":{"parts":[{"text":"You are Codex."}]}}`},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			decision := classifyRequestForFormat([]byte(testCase.body), testCase.format, cfg)
			if !decision.Blocked {
				t.Fatalf("format %q was not classified as blocked", testCase.format)
			}
		})
	}
}
