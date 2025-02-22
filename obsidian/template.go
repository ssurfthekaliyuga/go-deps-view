package obsidian

import (
	_ "embed"
	"fmt"
	"text/template"
)

//go:embed template.gohtml
var file string

func LoadTemplate() (*template.Template, error) {
	const name = "obsidian note template"

	tmpl, err := template.
		New(name).
		Parse(file)

	if err != nil {
		return nil, fmt.Errorf("cannot load obsidian note template: %w", err)
	}

	return tmpl, nil
}
