package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/naglezhang/fingersaver/internal/tmux"
)

func NewRespondConfirmationTool(tc TmuxClient) Tool {
	return Tool{
		Name:        "respond_confirmation",
		Description: "Send an approval or rejection to a session's pending confirmation prompt. Sends 'Yes' or 'No' followed by Enter.",
		Parameters: []Param{
			{Name: "session_name", Type: "string", Description: "Session name to respond to", Required: true},
			{Name: "approve", Type: "boolean", Description: "true to approve (Yes), false to reject (No)", Required: true},
		},
		Execute: func(ctx context.Context, args map[string]any) (string, error) {
			sessionName, _ := args["session_name"].(string)
			if sessionName == "" {
				return "", fmt.Errorf("session_name is required")
			}

			approve := false
			if v, ok := args["approve"].(bool); ok {
				approve = v
			}

			text := "No"
			if approve {
				text = "Yes"
			}

			if _, err := tc.Exec(tmux.SendKeysLiteralCmd(sessionName, text)); err != nil {
				return "", fmt.Errorf("send %s to %q: %w", text, sessionName, err)
			}
			if _, err := tc.Exec(tmux.SendEnterCmd(sessionName)); err != nil {
				return "", fmt.Errorf("send enter to %q: %w", sessionName, err)
			}

			data, _ := json.Marshal(map[string]any{
				"session": sessionName,
				"sent":    text,
			})
			return string(data), nil
		},
	}
}
