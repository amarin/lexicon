# Roadmap and open questions

What is still ahead for lexicon: the next milestones, open questions left by
earlier versions, and follow-ups. Done work lives in
[implementation/](implementation/) and [CHANGELOG.md](../CHANGELOG.md) and
is not repeated here; the design is in the
[spec](specs/2026-09-24-lexicon-design.md).

Mark completed items with `[x]` and move their write-up to
`implementation/`.

## Done

| Milestone | Released | Details |
|---|---|---|
| v0.1 — text analysis: `textnorm`, `Registry`, `Analyzer`, `basefetch`, CLI `analyze`/`dicts` | 0.1.0, 2026-09-25 | [implementation/v0.1-analysis.md](implementation/v0.1-analysis.md) |
| User documentation: scenarios, installation, CLI, library (EN + RU), runnable examples, godoc examples | 0.2.0, 2026-09-26 | [en/index.md](en/index.md), [ru/index.md](ru/index.md), [examples/](../examples/README.md) |
| v0.2 — dictionary NER: `gazetteer`, `rules` (hints, triggers, rule sets by document tags), `ner`, `nertest`, CLI `extract`/`golden`; user docs (scenarios 15–20, EN + RU) and examples `ner`, `gazetteer`, `golden` | 0.2.0, 2026-09-26 | [implementation/v0.2-ner.md](implementation/v0.2-ner.md) |

## Next milestones

- [ ] **v0.3 — patterns**: `rules` sequence patterns and facts (rule sets
  by document tags already ship in 0.2).
  [Plan](plans/2026-09-24-v0.3-patterns.md).

Each milestone also updates the user documentation: a scenario per new
feature in `docs/en/scenarios.md` and `docs/ru/scenarios.md` (with its
"Available since"), `library.md`/`cli.md` in both languages, a runnable
example, and the "Planned" section.

## Open questions (from v0.1)

Carried over from [implementation/v0.1-analysis.md](implementation/v0.1-analysis.md#open-questions-still-open);
each is marked "Behaviour in 0.1.0" in the scenarios where users see it.

- [ ] **Q3. Abbreviation lookup and profile kinds.** A profile without
  `abbrev` in its `Kinds` still expands «с.» (genodex P1 parity, D16).
  Decide whether a host needs strict per-profile filtering. Changing it
  changes index terms: ⚠ reindex (analyzer version), owner approval
  needed. Scenario 9.
- [ ] **Q5. gomorphy string aliasing.** `Registry.Parse` defensively
  `strings.Clone`s `Reading.Normal`/`Tag` (D22). gomorphy's docs now state
  that returned strings are independent copies that stay valid after
  `Close`; verify and drop the clones (a micro-optimisation, no behaviour
  change).
- [ ] **Q6. `Reload` cost.** `Reload` reopens every dictionary even when
  its file is unchanged (D23). Reuse dictionaries whose content hash did
  not change, for hosts with many large TSVs reloaded often. Scenario 12.
- [ ] **Q8. Defaults to confirm.** `DefaultCacheSize = 50000`, CLI
  `--ortho` default `modern`, `Flag.String()` names — implemented as
  proposed, never formally confirmed.

## Open questions (from v0.2)

Carried over from [implementation/v0.2-ner.md](implementation/v0.2-ner.md#open-questions).

- [ ] **Q-CLI-1. CLI profiles.** `extract`/`golden` use one hard-coded
  `text` profile for documents and aliases; a golden case with another
  `profile` fails. Add `--profiles FILE`? Scenario 20.
- [ ] **Q-v02-8. Rule hot-reload.** `Pipeline` has no `SetRules`: changing
  rules means building a new pipeline. Scenario 16.
- [ ] **Q-v02-9. Interner growth.** The process-wide alias interner never
  shrinks; acceptable for long-running hosts with heavy alias churn?
  Scenario 15.

Settled as defaults during v0.2 and open to revisit (see
[implementation/v0.2-ner.md](implementation/v0.2-ner.md#open-questions)):

- [ ] **Q-v02-5. `Blocked` scope.** A blocked alias vetoes its type on the
  exact range, trigger candidates over that range included. Scenario 15.
- [ ] **Q-v02-7. Nesting defaults.** `Config.Nesting` is empty by default;
  hosts opt in per outer/inner type pair. Scenario 17.
- [ ] **Q-v02-10. `absorb`.** An extension beyond the spec, shipped under
  that name; since the 2026-09-26 review it also applies to a trigger that
  boosts a gazetteer span. Owner to confirm the name. Scenario 16.

## Follow-ups

- [ ] **`hasGrammeme` → gomorphy `HasGrammeme`** (D27) once a gomorphy
  release ships it (not in 1.2.1). No API change; profile filtering must
  give the same terms, otherwise ⚠ reindex.
- [ ] **gomorphy 1.2.1** (documentation release, 2026-09-26): consider
  raising the requirement from v1.2.0 at the next lexicon release.
- [ ] **Span end over an abbreviation dot** (v0.3): a span ending in a
  dotted abbreviation excludes the dot («Калужской губ»); extend it over
  the dot. Changes span ends: update the golden sets. Scenario 15.
- [ ] **`ner` integration test against the real base** (`-tags
  integration`), like v0.1's `Analyzer` one.
- [ ] **Negative triggers and trigger candidates**: a negative trigger
  never lowers a trigger-created candidate; decide whether it should.
  Scenario 16.
- [ ] **Hint order**: a hint measures its window from the span's current,
  possibly absorbed, edge, so results can depend on hint order. Scenario 16.
- [ ] **`extract --explain` table alignment**: evidence rows break the
  tabwriter columns of the following span row (cosmetic).
- [ ] **Base dictionary for examples in CI.** `examples/base` needs a
  downloaded dictionary and is not run by `go test ./examples/`; the
  integration tests (`-tags integration`, `LEXICON_BASE_DAT`) cover the
  real base.

## Not lexicon's work

Embedding the base dictionary and scribe abbreviations, a SQL
`StateStore`, dictionary order configuration — genodex (host) work.

## Roadmap beyond v0.x

From the spec: fuzzy alias matching (gomorphy `FuzzyTop` over an
alias-word dictionary), morphology overlays generated from host data
(inflected surnames), a `nerman` provider adapter (in nerman), a compact
line DSL for patterns if YAML proves verbose.
