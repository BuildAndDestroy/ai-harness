package harness

import (
	_ "embed"
	"strings"
	"text/template"
)

//go:embed templates/prompt.md.tmpl
var promptTemplateText string

var promptTemplate = template.Must(template.New("prompt").Parse(promptTemplateText))

// promptData is the only data the template can see. It deliberately has no
// ReaperPassword field: text/template's reflection-based Execute can't leak a
// field that isn't present on the value it's given, so this makes it
// structurally impossible for BuildPrompt's output to contain the password,
// rather than relying on the template text alone never referencing it.
type promptData struct {
	ReaperURL      string
	ReaperUsername string
	Client         string
	Engagement     string
	Objectives     []string
}

// BuildPrompt renders the initial session prompt for cfg. It never includes
// cfg.ReaperPassword — the rendered prompt only ever tells the agent to read the
// password from its own process environment.
func BuildPrompt(cfg Config) (string, error) {
	data := promptData{
		ReaperURL:      cfg.ReaperURL,
		ReaperUsername: cfg.ReaperUsername,
		Client:         cfg.Client,
		Engagement:     cfg.Engagement,
		Objectives:     cfg.Objectives,
	}

	var b strings.Builder
	if err := promptTemplate.Execute(&b, data); err != nil {
		return "", err
	}
	return b.String(), nil
}
