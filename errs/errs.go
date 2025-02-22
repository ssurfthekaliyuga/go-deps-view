package errs

import (
	"fmt"
)

type Op string

func W(op Op, err error) error {
	return fmt.Errorf("%s: %w", op, err)
}
