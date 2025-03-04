package obsidian

import (
	"errors"
	"golang.org/x/sync/errgroup"
	"iter"
	"maps"
	"obsidian-deps-view/errs"
	"os"
	"strings"
	"text/template"
)

type Node struct {
	Name         string
	Tags         []string
	Imports      []string
	FromInternal []string
}

type GraphCreator struct {
	vault    *os.Root
	template *template.Template
	delim    string
	core     map[string]struct{}
}

func NewGraphCreator(vault *os.Root, template *template.Template, delim string) *GraphCreator {
	gc := &GraphCreator{
		vault:    vault,
		template: template,
		delim:    delim,
		core:     make(map[string]struct{}),
	}

	slice := []string{
		"bufio", "bytes", "cmp", "context",
		"crypto/rand", "database/sql", "database/driver",
		"embed", "encoding", "encoding/json",
		"errors", "flag", "fmt", "io", "io/fs", "iter",
		"log", "log/slog", "log/syslog",
		"maps", "math", "math/bits", "math/big", "math/rand/v2",
		"net", "net/http", "os", "os/exec", "os/signal", "os/user",
		"path", "path/filepath", "reflect",
		"regexp", "runtime", "slices", "strconv",
		"strings", "sync", "sync/atomic", "testing",
		"time", "unicode", "unicode", "unicode/utf8",
		"unsafe",
	}

	for _, pkg := range slice {
		pkg = gc.nodeFilename(pkg)
		gc.core[pkg] = struct{}{}
	}

	return gc
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

	node := Node{Name: c.nodeFilename(name)}

	for _, pkg := range imports {
		pkg = c.fixPackageName(pkg)

		if c.isInternalPackage(pkg) {
			node.FromInternal = append(node.FromInternal, pkg)
		} else {
			node.Imports = append(node.Imports, pkg)
		}
	}

	node.Tags = append(node.Tags, c.tags(node)...)

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
	return strings.
		NewReplacer("/", c.delim, "vendor/golang.org/", "").
		Replace(name)
}

func (c *GraphCreator) nodeFilename(name string) string {
	return c.fixPackageName(name) + ".md"
}

func (c *GraphCreator) tags(node Node) []string {
	tags := make([]string, 0)

	hierarch := []string{"go", "pkg", "std"}
	switch {
	case c.isInternalPackage(node.Name):
		hierarch = append(hierarch, "internal")
	case c.isCore(node.Name):
		hierarch = append(hierarch, "core")
	default:
		hierarch = append(hierarch, "specific")
	}

	tags = append(tags, strings.Join(hierarch, "/"))

	return tags
}

func (c *GraphCreator) isInternalPackage(pkg string) bool {
	return strings.Contains(pkg, "internal") || strings.Contains(pkg, "golang.org")
}

func (c *GraphCreator) isCore(pkg string) bool {
	_, ok := c.core[pkg]
	return ok
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
