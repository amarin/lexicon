package gazetteer

import (
	"context"
	"errors"
	"testing"
)

func TestSliceSource(t *testing.T) {
	src := NewSliceSource("names", "v7", []Entry{
		{Alias: "Иван", Type: "given_name", Ref: "given_name:1"},
		{Alias: "Иоанн", Type: "given_name", Ref: "given_name:1"},
	})
	if src.Name() != "names" {
		t.Fatalf("Name() = %q", src.Name())
	}
	if v, err := src.Version(context.Background()); err != nil || v != "v7" {
		t.Fatalf("Version() = %q, %v", v, err)
	}
	var got []string
	err := src.Entries(context.Background(), func(e Entry) error {
		got = append(got, e.Alias)
		return nil
	})
	if err != nil || len(got) != 2 || got[0] != "Иван" || got[1] != "Иоанн" {
		t.Fatalf("Entries: %v %v", got, err)
	}
	stop := errors.New("stop")
	if err := src.Entries(context.Background(), func(Entry) error { return stop }); !errors.Is(err, stop) {
		t.Fatalf("yield error not propagated: %v", err)
	}
}
