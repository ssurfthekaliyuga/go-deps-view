package obsidian

import (
	"bufio"
	"bytes"
	"maps"
	"slices"
	"strings"
	"testing"
)

func TestAddSpellings(t *testing.T) {
	tests := []struct {
		initLines []string
		newLines  []string
	}{
		{
			initLines: []string{"sasha\n", "leonov"},
			newLines:  []string{"golang"},
		},
	}

	for _, test := range tests {
		spellings := bytes.NewBuffer(nil)

		for _, line := range test.initLines {
			spellings.WriteString(line)
		}

		err := AddSpellings(spellings, test.newLines)
		if err != nil {
			t.Error(err)
		}

		afterAddLines := make(map[string]struct{})

		scanner := bufio.NewScanner(spellings)
		for scanner.Scan() {
			afterAddLines[scanner.Text()] = struct{}{}
		}

		expected := make(map[string]struct{})

		for _, word := range slices.Concat(test.initLines, test.newLines) {
			expected[strings.TrimSpace(word)] = struct{}{}
		}

		if !maps.Equal(afterAddLines, expected) {
			t.Errorf("words was not added, expected: %v got: %v", expected, afterAddLines)
		}
	}
}
