package main

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/amarin/lexicon"
)

func TestParseProfile(t *testing.T) {
	cases := []struct {
		in   string
		want lexicon.Profile
	}{
		{"text", lexicon.Profile{Name: "text"}},
		{"text:", lexicon.Profile{Name: "text"}},
		{"name:base[Name|Surn|Patr],surname,given,patronymic", lexicon.Profile{
			Name:      "name",
			Kinds:     []lexicon.Kind{"base", "surname", "given", "patronymic"},
			Grammemes: map[lexicon.Kind][]string{"base": {"Name", "Surn", "Patr"}},
		}},
	}
	for _, c := range cases {
		if got, err := parseProfile(c.in); err != nil || !reflect.DeepEqual(got, c.want) {
			t.Errorf("parseProfile(%q) = %+v, %v; want %+v", c.in, got, err, c.want)
		}
	}

	for _, bad := range []string{"", ":base", "x:Base", "x:base[", "x:base[]"} {
		if _, err := parseProfile(bad); err == nil {
			t.Errorf("parseProfile(%q): no error", bad)
		}
	}
}

func TestParseRulesAndMode(t *testing.T) {
	if r, err := parseRules("prereform"); err != nil || r.Name != "prereform" {
		t.Errorf("parseRules(prereform) = %v, %v", r.Name, err)
	}

	if _, err := parseRules("old"); err == nil {
		t.Error("parseRules(old): no error")
	}

	if m, err := parseMode("full"); err != nil || m != lexicon.ModeFull {
		t.Errorf("parseMode(full) = %v, %v", m, err)
	}

	if _, err := parseMode("x"); err == nil {
		t.Error("parseMode(x): no error")
	}
}

func TestDefaultDictsDir(t *testing.T) {
	t.Setenv("LEXICON_DICTS", "/x")

	if got := defaultDictsDir(); got != "/x" {
		t.Errorf("LEXICON_DICTS: %q", got)
	}

	t.Setenv("LEXICON_DICTS", "")
	t.Setenv("XDG_DATA_HOME", "/d")

	if got := defaultDictsDir(); got != filepath.Join("/d", "lexicon", "dicts") {
		t.Errorf("XDG_DATA_HOME: %q", got)
	}
}
