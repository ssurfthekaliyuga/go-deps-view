package obsidian

import (
	"fmt"
	"obsidian-deps-view/errs"
	"os"
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
