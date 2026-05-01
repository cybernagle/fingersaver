package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
)

// ViewerModel renders tmux session output in the right pane.
type ViewerModel struct {
	sessions map[string]string // session name -> output buffer
	order    []string          // authoritative session list from SessionListMsg
	active   string            // currently displayed session
	width    int
	height   int
	focused  bool
}

func NewViewerModel() ViewerModel {
	return ViewerModel{
		sessions: make(map[string]string),
	}
}

func (v ViewerModel) Init() tea.Cmd {
	return nil
}

func (v ViewerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height

	case SessionListMsg:
		v.order = msg.Sessions
		v.pruneSessions(msg.Sessions)
		// If active session was removed, switch to the first remaining one.
		if v.active == "" || !v.sessionExists(v.active) {
			if len(msg.Sessions) > 0 {
				v.active = msg.Sessions[0]
			} else {
				v.active = ""
			}
		}

	case tea.KeyPressMsg:
		if !v.focused {
			return v, nil
		}
		v.handleKey(msg.String())
	}

	return v, nil
}

func (v ViewerModel) View() tea.View {
	var b strings.Builder

	// Session tabs.
	if len(v.order) > 0 {
		b.WriteString(v.renderTabs())
		b.WriteString("\n")
	}

	content := v.sessions[v.active]
	lines := strings.Split(content, "\n")
	visibleHeight := v.height - 5
	if visibleHeight < 1 {
		visibleHeight = 1
	}

	start := len(lines) - visibleHeight
	if start < 0 {
		start = 0
	}
	visible := lines[start:]

	// Render entire block once.
	b.WriteString(viewerContentStyle.Render(strings.Join(visible, "\n")))

	for i := len(visible); i < visibleHeight; i++ {
		b.WriteString("\n")
	}

	return tea.NewView(b.String())
}

func (v *ViewerModel) renderTabs() string {
	var parts []string
	for _, s := range v.order {
		if s == v.active {
			parts = append(parts, viewerTitleStyle.Render("["+s+"]"))
		} else {
			parts = append(parts, statusStyle.Render(" "+s+" "))
		}
	}
	return strings.Join(parts, " ")
}

func (v *ViewerModel) handleKey(key string) {
	switch key {
	case "[":
		v.switchSession(-1)
	case "]":
		v.switchSession(1)
	}
}

func (v *ViewerModel) switchSession(dir int) {
	if len(v.order) == 0 {
		return
	}
	idx := 0
	for i, s := range v.order {
		if s == v.active {
			idx = i
			break
		}
	}
	idx += dir
	if idx < 0 {
		idx = len(v.order) - 1
	} else if idx >= len(v.order) {
		idx = 0
	}
	v.active = v.order[idx]
}

func (v *ViewerModel) sessionExists(name string) bool {
	for _, s := range v.order {
		if s == name {
			return true
		}
	}
	return false
}

func (v *ViewerModel) pruneSessions(active []string) {
	activeSet := make(map[string]bool, len(active))
	for _, s := range active {
		activeSet[s] = true
	}
	for name := range v.sessions {
		if !activeSet[name] {
			delete(v.sessions, name)
		}
	}
}

func (v *ViewerModel) SetFocused(f bool)         { v.focused = f }
func (v *ViewerModel) SetSize(w, h int)          { v.width = w; v.height = h }
func (v *ViewerModel) ActiveSession() string     { return v.active }
func (v *ViewerModel) SetActiveSession(s string) { v.active = s }

func (v *ViewerModel) AppendOutput(session, content string) {
	// capture-pane returns the full screen each time; replace, don't append.
	v.sessions[session] = content
	if v.active == "" {
		v.active = session
	}
}
