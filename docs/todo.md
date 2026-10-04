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
| v0.3 — inflection API: `Registry.Inflect`, `NumeralGrammemes` (word forms by grammemes, number agreement; requested by genodex); scenario 22 (EN + RU), example `inflect` | 0.3.0, 2026-10-01 | [implementation/v0.3-inflect.md](implementation/v0.3-inflect.md), [plan](plans/2026-10-01-v0.3-inflect.md) |

## Next milestones

- [x] **v0.4 — patterns** (was v0.3): `rules` sequence patterns and facts
  (rule sets by document tags already ship in 0.2).
  [Plan](plans/2026-10-04-v0.4-patterns.md),
  [implementation](implementation/v0.4-patterns.md). Implemented, not yet
  released: the release commit adds its row to "Done".

  **Blocks genodex E9** (mention suggestions; owner decision 2026-10-04):
  assembling a person's name from adjacent name words is pattern work in
  lexicon, not host code. The plan predates this requirement — review it
  against the list below before execution. Observed on a 1905 metric
  record in pre-reform spelling through genodex `extract` (0.3.0): places
  and plain «given patronymic surname» sequences come out right as
  separate word spans; the cases below do not.

  - [x] **Person as one unit.** A sequence of name words (given,
    given + patronymic, given + surname, given + patronymic + surname,
    surname + given + patronymic) must reach the host as one range with
    its parts by role, each part keeping its `Refs`. Decide the carrier:
    an `Emit`ted fact (`person` with args `given`/`patronymic`/`surname`
    — the plan as written, the host derives the range from the args) or
    a composite span of a new type nested over the part spans (needs a
    group-role relabel, P6, and a `Nesting` pair per part type). The host
    needs the whole range for the mention and the parts for matching
    against its persons.
  - [x] **Unknown word inside a name.** A capitalised word with no
    dictionary hit between or right after name words is a name part
    («Анна Михаилова Кузнецова» with «Михаилова» absent from the
    patronymic dictionary): token-role relabel (P6) to a `Candidate`
    patronymic or surname. Needs a selector for capitalisation — the
    plan has none (selectors: span type, lemma, grammemes, token kind).
  - [x] **Surname that is also a place name.** After given + patronymic,
    a word that won resolution as `settlement` without a settlement
    keyword before it («Анна Петрова Головина», the gazetteer has the
    village «Головино») is a surname: relabel, the settlement reading
    stays in `Alternatives`.
  - [x] **Given-name reading in the patronymic slot.** «Алексей Степанов
    Сидоров» where «Степанов» resolved as a form of the given name
    «Степан»: the second word of a name sequence is a patronymic.
  - [x] **Several persons in one sentence.** Metric records list parents
    and godparents in one sentence, sometimes back to back with no
    separator. Every person must match on its own — revisit Q-v03-2
    (greedy `any*` binds the last name pair) and Q-v03-8 (repeated role
    captures after «восприемники:»).
  - [x] **Golden cases** for each item above, in pre-reform and modern
    spelling (synthetic fixtures; the real texts stay outside the repo).

  Relations and `age` facts ship in the same milestone as planned but do
  not block E9.

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

## Open questions (from v0.4)

Carried over from [implementation/v0.4-patterns.md](implementation/v0.4-patterns.md#open-questions).

- [ ] **Q-v03-2. Lazy repetitions.** `*?` is not supported; repetitions are
  greedy. Needed for relation lists? Scenario 23.
- [ ] **Q-v03-4. Facts in golden sets.** `nertest` scores spans only; add
  fact cases to the JSONL? Scenario 25.
- [ ] **Q-v03-8. Repeated role captures.** A repeated `role` keeps its last
  repetition; «восприемники: …» needs all of them. Scenarios 23, 25.
- [ ] **Q-v03-10. Agreed normal forms** — see the follow-up below.

## Follow-ups

- [ ] **Agreed normal forms** — see
  [Q-v03-10](implementation/v0.4-patterns.md#open-questions):
  «Калужская губерния» instead of lemma sequences for spans without
  dictionary hits. The API question is settled (0.3: `Registry.Inflect`,
  D29); left: the head-word/agreement rule for multi-word spans and how
  `ner` reaches inflection (`lexicon.Dictionaries` has only `Parse`).
  ⚠ re-extract. Also usable for the roadmap's morphology overlays
  (inflected surnames).
- [ ] **Base from OpenCorpora XML?** gomorphy 1.3.0 builds prediction for
  `CompileFromXML`, so `basefetch` could drop the pymorphy2 loader (and
  `logging`/zap). A different dictionary: prediction quality, file size
  (14.9 MB) and terms change — ⚠ reindex, owner decision. On such a file
  «край» had no `sing gent` form through gomorphy `Inflect` (checked
  2026-10-01); recheck `Registry.Inflect` on it before switching.
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
