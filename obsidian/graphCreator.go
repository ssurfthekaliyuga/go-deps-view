package obsidian

import (
	"errors"
	"fmt"
	"golang.org/x/sync/errgroup"
	"iter"
	"maps"
	"obsidian-deps-view/errs"
	"os"
	"strings"
	"text/template"
)

type ConflictsError struct {
	Filenames []string
	Vault     string
}

func newConflictsError(op errs.Op, filenames []string, vault string) error {
	err := &ConflictsError{
		Filenames: filenames,
		Vault:     vault,
	}

	return errs.W(op, err)
}

func (e *ConflictsError) Error() string {
	return fmt.Sprintf("files: %v already exists in %s", e.Filenames, e.Vault)
}

func (e *ConflictsError) Unwrap() error {
	return os.ErrExist
}

type Node struct {
	Name         string
	Tags         []string
	Imports      []string
	FromInternal []string
}

type GraphCreator struct {
	vault     *os.Root
	template  *template.Template
	tags      []string
	pathDelim string
}

func NewGraphCreator(vault *os.Root, template *template.Template, pathDelim string, tags []string) *GraphCreator {
	return &GraphCreator{
		vault:     vault,
		template:  template,
		pathDelim: pathDelim,
		tags:      tags,
	}
}

func (c *GraphCreator) CreateGraph(packages map[string][]string) error {
	const op errs.Op = "obsidian.GraphCreator.CreateGraph"

	conflicts, err := c.filenameConflicts(maps.Keys(packages))
	if err != nil {
		return errs.W(op, err)
	}
	if len(conflicts) != 0 {
		return newConflictsError(op, conflicts, c.vault.Name())
	}

	var eg errgroup.Group

	for name, imports := range packages {
		eg.Go(func() error {
			return c.CreateNode(name, imports)
		})
	}

	if err = eg.Wait(); err != nil {
		return errs.W(op, err)
	}

	return nil
}

func (c *GraphCreator) CreateNode(name string, imports []string) error {
	const op errs.Op = "obsidian.GraphCreator.CreateNode"

	node := Node{
		Name:         c.nodeFilename(name),
		Tags:         c.tags,
		Imports:      make([]string, 0, len(imports)),
		FromInternal: make([]string, 0, len(imports)),
	}

	if c.isInternalPackage(name) {
		node.Tags = append(node.Tags, "internal")
	}

	for _, pkg := range imports {
		pkg = c.fixPackageName(pkg)

		if c.isInternalPackage(pkg) {
			node.FromInternal = append(node.FromInternal, pkg)
		} else {
			node.Imports = append(node.Imports, pkg)
		}
	}

	file, err := c.vault.OpenFile(node.Name, os.O_WRONLY|os.O_CREATE, 0666) //todo perm
	if err != nil {
		return errs.W(op, err)
	}
	defer file.Close()

	if err = c.template.Execute(file, node); err != nil {
		return errs.W(op, err)
	}

	return nil
}

func (c *GraphCreator) fixPackageName(name string) string {
	r := strings.NewReplacer("/", c.pathDelim, "vendor/golang.org/", "")
	return r.Replace(name)
}

func (c *GraphCreator) nodeFilename(name string) string {
	return c.fixPackageName(name) + ".md"
}

func (c *GraphCreator) isInternalPackage(pkg string) bool {
	return strings.Contains(pkg, "internal") || strings.Contains(pkg, "x")
}

func (c *GraphCreator) filenameConflicts(packages iter.Seq[string]) ([]string, error) {
	const op errs.Op = "obsidian.GraphCreator.nodesExist"

	existing := make([]string, 0)

	for pkg := range packages {
		filename := c.nodeFilename(pkg)
		_, err := c.vault.Stat(filename)
		if err == nil {
			existing = append(existing, filename)
		}
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, errs.W(op, err)
		}
	}

	return existing, nil
}
