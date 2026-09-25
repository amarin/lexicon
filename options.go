package lexicon

// Options configure Open.
type Options struct {
	// Dir holds <kind>.<name>.dat|tsv files (with optional .meta sidecars).
	// "" or a missing directory means no files; it is never created.
	Dir string
	// Base is a built-in base dictionary in gomorphy binary format (nil = none),
	// registered as BaseBuiltinName. Any base.* file in Dir replaces it.
	Base []byte
	// BaseManifest is the provenance of Base (owner decision 2026-09-25):
	// hosts embed the .meta sidecar that basefetch writes next to the .dat and
	// parse it with ParseManifest. Its non-empty fields win; empty ones are
	// filled from gomorphy BuildInfo. Zero = BuildInfo only. Unused when a
	// base.* file replaces Base (that file has its own sidecar).
	BaseManifest Manifest
	// Builtin are host dictionaries registered after the base, in order.
	// A built-in whose kind is not accepted by Kinds makes Open fail.
	Builtin []BuiltinDict
	// Kinds are the host's dictionary kinds besides KindBase and KindAbbrev
	// (owner decision 7, D17). A file of any other kind is listed with an
	// "unknown kind" error and not parsed. nil/empty accepts every valid kind
	// (the lexicon CLI).
	Kinds []Kind
	// State persists enabled/disabled; nil = an in-memory MemState (all enabled).
	State StateStore
}
