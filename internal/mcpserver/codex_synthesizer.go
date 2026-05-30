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

type CodexQuerySynthesizer struct {
	Command string
	Args    []string
	Timeout time.Duration
}

func NewCodexQuerySynthesizer(command string) CodexQuerySynthesizer {
	if command == "" {
		command = "codex"
	}
	return CodexQuerySynthesizer{Command: command, Args: []string{"exec"}, Timeout: 3 * time.Minute}
}

func (c CodexQuerySynthesizer) SynthesizeQuery(ctx context.Context, req querySynthesisRequest) (querySynthesisResult, error) {
	timeout := c.Timeout
	if timeout == 0 {
		timeout = 3 * time.Minute
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	prompt, err := buildQueryPrompt(req)
	if err != nil {
		return querySynthesisResult{}, err
	}
	args := append([]string{}, c.Args...)
	args = append(args, prompt)
	cmd := exec.CommandContext(ctx, c.Command, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return querySynthesisResult{}, fmt.Errorf("codex query synthesis failed: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	var result querySynthesisResult
	if err := decodeQueryJSONObject(out, &result); err != nil {
		text := strings.TrimSpace(string(out))
		if text == "" {
			return querySynthesisResult{}, err
		}
		return querySynthesisResult{Answer: text}, nil
	}
	return result, nil
}

func buildQueryPrompt(req querySynthesisRequest) (string, error) {
	b, err := json.MarshalIndent(req, "", "  ")
	if err != nil {
		return "", err
	}
	return `You are Workbench's knowledge query synthesizer.

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

func decodeQueryJSONObject(data []byte, v any) error {
	data = bytes.TrimSpace(data)
	start := bytes.IndexByte(data, '{')
	end := bytes.LastIndexByte(data, '}')
	if start < 0 || end < start {
		return fmt.Errorf("no JSON object in output")
	}
	return json.Unmarshal(data[start:end+1], v)
}
