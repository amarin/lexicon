package lexicon

import "testing"

func TestFlagString(t *testing.T) {
	cases := map[Flag]string{
		0:                          "",
		FlagAmbiguous:              "ambiguous",
		FlagAmbiguous | FlagAbbrev: "ambiguous,abbrev",
		FlagPredicted:              "predicted",
		FlagUnknown:                "unknown",
		FlagStop | FlagReform:      "stop,reform",
	}
	for f, want := range cases {
		if got := f.String(); got != want {
			t.Errorf("Flag(%d).String() = %q, want %q", f, got, want)
		}
	}
}
