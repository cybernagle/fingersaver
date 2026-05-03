package tools

import (
	"context"

	"github.com/naglezhang/fingersaver/internal/tmux"
)

// Tool represents a callable tool that the orchestrator can invoke.
type Tool struct {
	Name        string
	Description string
	Parameters  []Param
	Execute     func(ctx context.Context, args map[string]any) (string, error)
}

// Param describes a single tool parameter.
type Param struct {
	Name        string
	Type        string // "string", "number", "boolean"
	Description string
	Required    bool
}

// TmuxClient is the subset of tmux functionality that tools need.
type TmuxClient interface {
	Exec(cmd string) (string, error)
	State() *tmux.StateMirror
}

// Assessor evaluates session output for pending confirmation prompts
// and decides whether to approve or reject. Defined here to avoid
// circular imports (implemented in agent package).
type Assessor interface {
	Assess(ctx context.Context, sessionName, output string) (*Assessment, error)
}

// Assessment is the result of evaluating a session's pending confirmation.
type Assessment struct {
	Decision string `json:"decision"` // "approve", "reject", "idle", "unknown"
	Reason   string `json:"reason"`
}
