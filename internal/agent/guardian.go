package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/naglezhang/fingersaver/internal/agent/tools"
	"github.com/naglezhang/fingersaver/internal/llm"
	"github.com/naglezhang/fingersaver/internal/util"
)

const defaultAssessorPrompt = `You are a session guardian. You monitor a coding agent in a terminal and decide how to respond to its confirmation prompts.

CRITICAL: You must distinguish between the agent WORKING (producing output) and the agent WAITING for user input. Look at the LAST FEW LINES of output carefully.

The agent is WORKING (return "idle") when:
- It is showing progress messages (e.g. "Phase 1", "Step 2/5", bullet points, file diffs)
- It is listing files, reading code, or showing analysis
- It is printing tool results or intermediate output
- The last line is NOT a question or prompt — it is a statement, output, or progress indicator
- The output ends mid-thought (the agent is still generating)

The agent is WAITING for input (return "approve"/"reject") ONLY when the LAST LINE is clearly a confirmation prompt:
- Claude Code: ends with "?  [Y/n]" or "Allow this action?" or similar yes/no prompt
- Copilot: ends with "(Y/n)" or "Proceed?" or "Confirm?"
- Any agent: last line is a direct question asking for user approval

When in doubt, return "idle". It is much better to keep waiting than to prematurely approve or reject.

Respond with ONLY a JSON object on a single line:
- {"decision":"approve","reason":"brief reason"} — routine confirmation (tool calls, file edits, proceed prompts)
- {"decision":"reject","reason":"brief reason"} — dangerous operation (deleting prod data, force-pushing, dropping databases, sudo)
- {"decision":"idle","reason":"brief reason"} — agent is still working or showing output, no prompt visible
- {"decision":"unknown","reason":"brief reason"} — cannot determine`

// SessionAssessor implements tools.Assessor using an LLM to evaluate
// pending confirmation prompts in coding agent sessions.
type SessionAssessor struct {
	provider llm.Provider
	model    string
	prompt   string
}

func NewSessionAssessor(provider llm.Provider, model, prompt string) *SessionAssessor {
	if prompt == "" {
		prompt = defaultAssessorPrompt
	}
	return &SessionAssessor{
		provider: provider,
		model:    model,
		prompt:   prompt,
	}
}

func (sa *SessionAssessor) Assess(ctx context.Context, sessionName, output string) (*tools.Assessment, error) {
	assessCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	msgs := []llm.Message{
		{Role: llm.RoleSystem, Content: sa.prompt},
		{Role: llm.RoleUser, Content: fmt.Sprintf("Session %q output:\n%s", sessionName, output)},
	}
	opts := llm.GenerateOptions{Model: sa.model, MaxTokens: 256}
	result, err := sa.provider.Complete(assessCtx, msgs, opts)
	if err != nil {
		return nil, fmt.Errorf("assess LLM call: %w", err)
	}

	raw := strings.TrimSpace(result.Content)
	if idx := strings.Index(raw, "{"); idx >= 0 {
		raw = raw[idx:]
		if end := strings.Index(raw, "}"); end >= 0 {
			raw = raw[:end+1]
		}
	}

	var j struct {
		Decision string `json:"decision"`
		Reason   string `json:"reason"`
	}
	if err := json.Unmarshal([]byte(raw), &j); err != nil {
		log.Printf("[assessor] parse error for %q: %v (raw: %s)", sessionName, err, util.Truncate(raw, 100))
		return &tools.Assessment{Decision: "unknown", Reason: "failed to parse LLM response"}, nil
	}

	decision := normalizeDecision(j.Decision)
	log.Printf("[assessor] %s: decision=%s reason=%s", sessionName, decision, j.Reason)
	return &tools.Assessment{Decision: decision, Reason: j.Reason}, nil
}

func normalizeDecision(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "approve", "yes", "safe", "ok", "allow", "accept":
		return "approve"
	case "reject", "no", "deny", "block", "risky", "dangerous", "unsafe":
		return "reject"
	case "idle", "working", "running", "none", "n/a":
		return "idle"
	default:
		return "unknown"
	}
}
