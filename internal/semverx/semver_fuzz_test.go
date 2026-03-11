package semverx

import "testing"

func FuzzParseRoundTrip(f *testing.F) {
	seeds := []string{
		"0.0.1",
		"1.2.3",
		"1.2.3-alpha.1",
		"1.2.3+build.7",
		"2.10.42-rc.1+meta.9",
		"not-semver",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, input string) {
		v, err := Parse(input)
		if err != nil {
			return
		}

		roundTrip := v.String()
		v2, err := Parse(roundTrip)
		if err != nil {
			t.Fatalf("Parse(roundTrip=%q) error: %v", roundTrip, err)
		}
		if v2.String() != roundTrip {
			t.Fatalf("round trip mismatch: %q != %q", v2.String(), roundTrip)
		}
		if !IsValid(roundTrip) {
			t.Fatalf("round trip semver should be valid: %q", roundTrip)
		}
	})
}
