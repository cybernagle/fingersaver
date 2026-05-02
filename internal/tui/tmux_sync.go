package tui

import (
	"log"

	tea "charm.land/bubbletea/v2"
	"github.com/naglezhang/fingersaver/internal/tmux"
)

type resizeTmuxMsg struct{}

// resizeTmuxCmd returns an async command that resizes the active tmux session
// to match the viewer content area dimensions. Only applies when FingerSaver
// owns the tmux server (dedicated mode) to avoid disrupting shared sessions.
func (a AppModel) resizeTmuxCmd() tea.Cmd {
	if a.tmuxClient == nil {
		return nil
	}
	active := a.viewer.ActiveSession()
	if active == "" {
		return nil
	}
	contentW := a.viewerContentWidth()
	contentH := a.viewerContentHeight()
	if contentW < 1 || contentH < 1 {
		return nil
	}
	tc := a.tmuxClient
	return func() tea.Msg {
		cmd := tmux.ResizeWindowCmd(active, contentW, contentH)
		if _, err := tc.Exec(cmd); err != nil {
			log.Printf("[app] resize-window %s: %v", active, err)
		}
		return resizeTmuxMsg{}
	}
}

func (a *AppModel) viewerContentWidth() int {
	if a.layout == LayoutPhone {
		return a.width - 2
	}
	viewerW := a.width - a.width*2/5 - 2
	return viewerW - 2
}

func (a *AppModel) viewerContentHeight() int {
	if a.layout == LayoutPhone {
		return a.height*3/5 - 4
	}
	return a.height - 4
}
