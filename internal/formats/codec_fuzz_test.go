package formats

import (
	"encoding/json"
	"fmt"
	"testing"
)

func FuzzJSONCodecSetVersionRoundTrip(f *testing.F) {
	seeds := []string{
		`{"version":"1.0.0"}`,
		`{"name":"demo","version":"2.3.4","nested":{"x":1}}`,
		`{"version":"0.0.1","arr":[1,2,3],"flag":true}`,
	}
	for _, doc := range seeds {
		f.Add(doc, "9.9.9")
	}

	codec := &JSONCodec{}

	f.Fuzz(func(t *testing.T, doc string, version string) {
		target := fmt.Sprintf("%d.%d.%d", len(version)%10, (len(version)+3)%10, (len(version)+7)%10)

		old, err := codec.GetVersion([]byte(doc), []string{"version"})
		if err != nil {
			return
		}
		_ = old

		updated, err := codec.SetVersion([]byte(doc), []string{"version"}, target)
		if err != nil {
			t.Fatalf("SetVersion() error: %v", err)
		}

		got, err := codec.GetVersion(updated, []string{"version"})
		if err != nil {
			t.Fatalf("GetVersion(updated) error: %v", err)
		}
		if got != target {
			t.Fatalf("GetVersion(updated) = %q, want %q", got, target)
		}

		var parsed any
		origValid := json.Unmarshal([]byte(doc), &parsed) == nil
		if origValid {
			if err := json.Unmarshal(updated, &parsed); err != nil {
				t.Fatalf("updated JSON should stay valid when source is valid: %v", err)
			}
		}
	})
}

func FuzzYAMLCodecSetVersionRoundTrip(f *testing.F) {
	seeds := []string{
		"version: 1.0.0\n",
		"name: demo\nversion: \"2.3.4\"\n",
		"version: '0.0.1'\ncomment: keep\n",
	}
	for _, doc := range seeds {
		f.Add(doc, "3.3.3")
	}

	codec := &YAMLCodec{}

	f.Fuzz(func(t *testing.T, doc string, version string) {
		target := fmt.Sprintf("%d.%d.%d", len(version)%10, (len(version)+3)%10, (len(version)+7)%10)

		_, err := codec.GetVersion([]byte(doc), []string{"version"})
		if err != nil {
			return
		}

		updated, err := codec.SetVersion([]byte(doc), []string{"version"}, target)
		if err != nil {
			return
		}

		got, err := codec.GetVersion(updated, []string{"version"})
		if err != nil {
			t.Fatalf("GetVersion(updated) error: %v", err)
		}
		if got != target {
			t.Fatalf("GetVersion(updated) = %q, want %q", got, target)
		}
	})
}
