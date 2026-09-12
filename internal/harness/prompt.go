package harness

import (
	_ "embed"
	"strings"
	"text/template"
)

//go:embed templates/prompt.md.tmpl
var promptTemplateText string

var promptTemplate = template.Must(template.New("prompt").Parse(promptTemplateText))

// BuildPrompt renders the initial session prompt for cfg. It never includes
// cfg.ReaperPassword — the rendered prompt only ever tells the agent to read the
// password from its own process environment.
func BuildPrompt(cfg Config) (string, error) {
	var b strings.Builder
	if err := promptTemplate.Execute(&b, cfg); err != nil {
		return "", err
	}
	return b.String(), nil
}
