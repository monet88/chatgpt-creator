package util

import (
	"strconv"
	"strings"
	"testing"
)

func TestRandomBirthdate_Age30PlusRange(t *testing.T) {
	for i := 0; i < 100; i++ {
		value := RandomBirthdate()
		parts := strings.Split(value, "-")
		if len(parts) != 3 {
			t.Fatalf("RandomBirthdate() = %q, want YYYY-MM-DD", value)
		}
		year, err := strconv.Atoi(parts[0])
		if err != nil {
			t.Fatalf("year parse error for %q: %v", value, err)
		}
		if year < 1985 || year > 1996 {
			t.Fatalf("year = %d, want 1985..1996", year)
		}
	}
}
