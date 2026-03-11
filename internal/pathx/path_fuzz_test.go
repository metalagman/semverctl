package pathx

import (
	"fmt"
	"reflect"
	"testing"
)

func FuzzParseJoinRoundTrip(f *testing.F) {
	seeds := []string{"version", ".version", "app.version", ".app.version", "a.b.c"}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, input string) {
		parts, err := Parse(input)
		if err != nil {
			return
		}

		reparsed, err := Parse(Join(parts))
		if err != nil {
			t.Fatalf("Parse(Join(parts)) error: %v", err)
		}
		if !reflect.DeepEqual(parts, reparsed) {
			t.Fatalf("round trip mismatch: %v != %v", parts, reparsed)
		}

		doc := make(map[string]any)
		current := doc
		for i := 0; i < len(parts)-1; i++ {
			next := make(map[string]any)
			current[parts[i]] = next
			current = next
		}
		if len(parts) > 0 {
			current[parts[len(parts)-1]] = "old"
			if err := Set(doc, parts, "new"); err != nil {
				t.Fatalf("Set() error: %v", err)
			}
			got, err := Get(doc, parts)
			if err != nil {
				t.Fatalf("Get() error: %v", err)
			}
			if got != "new" {
				t.Fatalf("Get() = %v, want new", got)
			}
		}

		_ = fmt.Sprintf("%v", doc)
	})
}
