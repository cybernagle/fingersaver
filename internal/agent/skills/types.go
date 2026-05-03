package skills

import (
	"fmt"
	"strings"
)

// Skill represents a reusable prompt template loaded from a SKILL.md file.
type Skill struct {
	Name         string
	Description  string
	Prompt       string
	AllowedTools []string
	Args         []string
}

// ExpandPrompt replaces template variables in the skill prompt with the
// provided argument values. Supported templates: {{.Args}} (all args joined),
// {{.Arg0}}, {{.Arg1}}, etc.
func (s *Skill) ExpandPrompt(args []string) string {
	p := s.Prompt
	p = strings.ReplaceAll(p, "{{.Args}}", strings.Join(args, " "))
	for i, a := range args {
		p = strings.ReplaceAll(p, fmt.Sprintf("{{.Arg%d}}", i), a)
	}
	return p
}
