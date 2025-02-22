package analyzer

import (
	"bufio"
	"context"
	"fmt"
	"obsidian-deps-view/errs"
	"os/exec"
	"regexp"
	"strings"
)

type ImportsParser struct {
	inPattern string
	outRegexp *regexp.Regexp
}

func NewImportsParser() *ImportsParser {
	return &ImportsParser{
		inPattern: `{{.ImportPath}}: {{.Imports}}`,
		outRegexp: regexp.MustCompile(`([\w./]+): \[([^\]]*)\]`),
	}
}

func (p *ImportsParser) ParseImports(ctx context.Context, pkg string) (map[string][]string, error) {
	const op = "analyzer.ImportsParser.ParseImports"

	packages := make(map[string][]string)

	cmd := p.createCommand(ctx, pkg)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, errs.W(op, err)
	}

	if err = cmd.Start(); err != nil {
		return nil, errs.W(op, err)
	}

	scanner := bufio.NewScanner(stdout)

	for scanner.Scan() {
		line := scanner.Text()
		name, imports := p.parseImports(line)
		packages[name] = imports
	}

	if err = cmd.Wait(); err != nil {
		return nil, fmt.Errorf("wait exec: %w", err)
	}

	return packages, nil
}

func (p *ImportsParser) parseImports(line string) (string, []string) {
	line = strings.TrimSpace(line)
	match := p.outRegexp.FindStringSubmatch(line)

	if len(match) < 3 {
		return "", nil
	}

	name := match[1]
	imports := strings.Fields(match[2])

	return name, imports
}

func (p *ImportsParser) createCommand(ctx context.Context, pkg string) *exec.Cmd {
	args := []string{"list", "-f", p.inPattern, pkg}

	if pkg != "std" {
		args = append(args, "./...")
	}

	return exec.CommandContext(ctx, "go", args...)
}
