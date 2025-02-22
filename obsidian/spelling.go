package obsidian

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"obsidian-deps-view/errs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func SpellingsLocation() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	switch runtime.GOOS {
	case "windows":
		fmt.Println("123")
		path := []string{home, "AppData", "Roaming", "obsidian", "Custom Dictionary.txt"}
		return filepath.Join(path...), nil
	default:
		return "", nil
	}
}

func AddSpellings(dst io.ReadWriter, words []string) error {
	const op errs.Op = "obsidian.AddSpellings"

	set := make(map[string]struct{}, len(words))
	for _, word := range words {
		set[word] = struct{}{}
	}

	scanner := bufio.NewScanner(dst)
	for scanner.Scan() {
		set[scanner.Text()] = struct{}{}
	}

	buf := bytes.NewBuffer(nil)
	for word := range set {
		line := strings.TrimSpace(word) + "\n"

		if strings.Contains(line, "checksum") {
			continue
		}

		if _, err := buf.WriteString(line); err != nil {
			return errs.W(op, err)
		}
	}

	if _, err := buf.WriteTo(dst); err != nil {
		return errs.W(op, err)
	}

	return nil
}
