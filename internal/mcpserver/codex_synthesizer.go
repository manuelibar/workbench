package mcpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type CodexAskSynthesizer struct {
	Command string
	Args    []string
	Timeout time.Duration
}

func NewCodexAskSynthesizer(command string) CodexAskSynthesizer {
	if command == "" {
		command = "codex"
	}
	return CodexAskSynthesizer{Command: command, Args: []string{"exec"}, Timeout: 3 * time.Minute}
}

func (c CodexAskSynthesizer) SynthesizeAsk(ctx context.Context, req askSynthesisRequest) (askSynthesisResult, error) {
	timeout := c.Timeout
	if timeout == 0 {
		timeout = 3 * time.Minute
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	prompt, err := buildAskPrompt(req)
	if err != nil {
		return askSynthesisResult{}, err
	}
	args := append([]string{}, c.Args...)
	args = append(args, prompt)
	cmd := exec.CommandContext(ctx, c.Command, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return askSynthesisResult{}, fmt.Errorf("codex ask synthesis failed: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	var result askSynthesisResult
	if err := decodeAskJSONObject(out, &result); err != nil {
		text := strings.TrimSpace(string(out))
		if text == "" {
			return askSynthesisResult{}, err
		}
		return askSynthesisResult{Answer: text}, nil
	}
	return result, nil
}

func buildAskPrompt(req askSynthesisRequest) (string, error) {
	b, err := json.MarshalIndent(req, "", "  ")
	if err != nil {
		return "", err
	}
	return `You are Workbench's final answer synthesizer.

Use only the retrieved KB content and knowledge below. Return one JSON object.

Return an inline answer when a concise grounded answer is enough:
{"answer":"..."}

Return flat ad hoc skill resources when the user needs a reusable learning pack:
{"resources":[
  {"uri":"skill://<dir>/SKILL.md","mime_type":"text/markdown","text":"..."},
  {"uri":"skill://<dir>/resources/overview.md","mime_type":"text/markdown","text":"..."},
  {"uri":"skill://<dir>/resources/concepts.md","mime_type":"text/markdown","text":"..."},
  {"uri":"skill://<dir>/resources/learning-paths.md","mime_type":"text/markdown","text":"..."},
  {"uri":"skill://<dir>/resources/deep-dives.md","mime_type":"text/markdown","text":"..."},
  {"uri":"skill://<dir>/resources/related-areas.md","mime_type":"text/markdown","text":"..."},
  {"uri":"skill://<dir>/resources/sources.md","mime_type":"text/markdown","text":"..."}
]}

Retrieved input:
` + string(b), nil
}

func decodeAskJSONObject(data []byte, v any) error {
	data = bytes.TrimSpace(data)
	start := bytes.IndexByte(data, '{')
	end := bytes.LastIndexByte(data, '}')
	if start < 0 || end < start {
		return fmt.Errorf("no JSON object in output")
	}
	return json.Unmarshal(data[start:end+1], v)
}
