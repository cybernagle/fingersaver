package tools

import (
	"fmt"

	"github.com/naglezhang/fingersaver/internal/tmux"
)

func sendText(tc TmuxClient, sessionName, text string) error {
	payload := fmt.Sprintf("\033[200~%s\033[201~\r", text)
	if _, err := tc.Exec(tmux.SendKeysLiteralCmd(sessionName, payload)); err != nil {
		return fmt.Errorf("send to %q: %w", sessionName, err)
	}
	return nil
}
