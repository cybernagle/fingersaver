package tools

import (
	"context"
	"fmt"

	"github.com/naglezhang/fingersaver/internal/util"
)

func NewSendToSessionTool(tc TmuxClient) Tool {
	return Tool{
		Name:        "send_to_session",
		Description: "Send a command or message to a tmux session. Follow with wait_until_idle to handle any confirmation prompts.",
		Parameters: []Param{
			{Name: "name", Type: "string", Description: "Session name", Required: true},
			{Name: "message", Type: "string", Description: "Text to send to the session", Required: true},
		},
		Execute: func(ctx context.Context, args map[string]any) (string, error) {
			name, _ := args["name"].(string)
			message, _ := args["message"].(string)
			if name == "" || message == "" {
				return "", fmt.Errorf("name and message are required")
			}

			if err := sendText(tc, name, message); err != nil {
				return "", err
			}

			return fmt.Sprintf("Sent to %q: %s", name, util.Truncate(message, 50)), nil
		},
	}
}
