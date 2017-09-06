package naming

import (
	"strings"
	"testing"
)

func TestGetRandomName(t *testing.T) {
	name := GetRandomName()

	parts := strings.Split(name, "-")

	if !strings.Contains(name, "-") {
		t.Fatalf("Generated name does not contain a dash - %v", name)
	}

	if !strings.ContainsAny(name, "0123456789") {
		t.Fatalf("Generated name doesn't contain trailing numbers - %v", name)
	}

	if !inSlice(parts[0], left) {
		t.Fatalf("Adjective in name not in slice of adjectives - %v", name)
	}

	if !inSlice(parts[1], right) {
		t.Fatalf("Animal in name not in slice of animals - %v", name)
	}
}

// inSlice helper func for determining if string is in slice of strings
func inSlice(needle string, haystack []string) bool {
	for _, x := range haystack {
		if x == needle {
			return true
		}
	}

	return false
}
