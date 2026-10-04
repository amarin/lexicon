// Command persons assembles person names with rule patterns: name words
// next to each other become one person span with its parts nested inside,
// an unknown capitalised word between name words becomes a candidate
// patronymic, a surname that is also a village keeps the village as an
// alternative, a defining word («ответчик») makes a person of one surname,
// and a pattern emits a child_of fact whose argument is the whole person.
// The rule set is data: a host owns its own file and type names. The
// morphology is a tiny TSV registered in code, so the example needs no
// download. That base knows only a few words, so spans a pattern created
// over other words carry the flag predicted (with the real base dictionary it
// appears only for words the dictionary has to guess).
//
// Run: go run ./examples/persons
package main

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strings"

	"github.com/amarin/lexicon"
	"github.com/amarin/lexicon/gazetteer"
	"github.com/amarin/lexicon/ner"
	"github.com/amarin/lexicon/rules"
	"github.com/amarin/lexicon/textnorm"
)

// base is a stand-in for the OpenCorpora base: lemma<TAB>wordform<TAB>tag.
const base = `иван	иван	NOUN,anim,masc,Name sing,nomn
иван	ивана	NOUN,anim,masc,Name sing,gent
дочь	дочь	NOUN,anim,femn sing,nomn
сын	сын	NOUN,anim,masc sing,nomn
`

// names is a gazetteer source in TSV: type<TAB>ref<TAB>canonical<TAB>alias.
const names = `given_name	g1	Иван	Иван
given_name	g2	Анна	Анна
given_name	g3	Мария	Мария
patronymic	p1	Петров	Петров
patronymic	p1	Петров	Петрова
surname	s1	Сидоров	Сидоров
surname	s2	Кузнецов	Кузнецова
division	d1	Головино	Головина
`

// ruleFile: fix-up patterns first, assembly after them, facts last.
const ruleFile = `
sets:
  - name: persons
    patterns:
      - name: place-as-surname
        elements:
          - {type: given_name}
          - {type: patronymic}
          - {type: division, shape: {case: title}, role: s}
        actions:
          - relabel: {role: s, type: surname}
      - name: unknown-patronymic
        elements:
          - {type: given_name}
          - {token: word, shape: {case: title}, not: {type: given_name|patronymic|surname}, role: p}
          - {type: surname}
        actions:
          - label: {role: p, type: patronymic}
      - name: person
        elements:
          - group:
              - {type: given_name}
              - {type: patronymic|surname}
              - {type: surname, repeat: "?"}
            role: who
        actions:
          - label: {role: who, type: person}
      - name: person-by-defining-word
        elements:
          - {lemma: ответчик|истец}
          - group:
              - {type: surname}
            role: who
        actions:
          - label: {role: who, type: person}
  - name: kinship
    patterns:
      - name: child-of
        elements:
          - {type: person|given_name, role: child}
          - {token: punct, repeat: "?"}
          - lemma: сын|дочь
          - {type: person|given_name, role: parent}
        actions:
          - emit: {kind: child_of, args: {child: child, parent: parent}}
`

func main() {
	ctx := context.Background()

	reg, err := lexicon.Open(ctx, lexicon.Options{
		Builtin: []lexicon.BuiltinDict{{Name: "base.demo", Format: lexicon.FormatTSV, Data: []byte(base)}},
	})
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = reg.Close() }()

	an := lexicon.NewAnalyzer(reg, textnorm.PreReform, lexicon.AnalyzerOptions{})
	text := lexicon.Profile{Name: "text"}

	gz, err := gazetteer.New(ctx, gazetteer.Config{
		Analyzer:       an,
		DefaultProfile: text,
		Sources:        []gazetteer.Source{gazetteer.NewTSVSourceData("names", []byte(names))},
	})
	if err != nil {
		log.Fatal(err)
	}

	rf, err := rules.LoadNamed("persons.yaml", strings.NewReader(ruleFile))
	if err != nil {
		log.Fatal(err)
	}
	book, err := rules.Compile(rf)
	if err != nil {
		log.Fatal(err)
	}

	p, err := ner.New(ner.Config{
		Analyzer: an, Gazetteer: gz, Rules: book,
		Profiles: map[string]lexicon.Profile{"text": text}, DefaultProfile: "text",
		// The parts of a person are output inside it only for these pairs.
		Nesting: map[string][]string{"person": {"given_name", "patronymic", "surname"}},
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, doc := range []string{
		"Иван Петров Сидоров",
		"Анна Михаилова Кузнецова",
		"Анна Петрова Головина",
		"ответчик Сидоров",
		"Мария, дочь Ивана Петрова",
	} {
		res, err := p.Extract(ctx, ner.Doc{Text: doc})
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(doc)
		for _, s := range res.Spans {
			notes := s.Flags.Names()
			if len(s.Refs) > 0 {
				notes = append(notes, "ref="+strings.Join(s.Refs, "|"))
			}
			for _, a := range s.Alternatives {
				notes = append(notes, "alt="+a.Type)
			}
			fmt.Printf("  %-10s %-24s %s\n", s.Type, s.Surface, strings.Join(notes, " "))
		}
		for _, f := range res.Facts {
			roles := make([]string, 0, len(f.Args))
			for r := range f.Args {
				roles = append(roles, r)
			}
			slices.Sort(roles)
			for i, r := range roles {
				roles[i] = r + "=«" + res.Spans[f.Args[r]].Surface + "»"
			}
			fmt.Printf("  fact %s %s\n", f.Kind, strings.Join(roles, " "))
		}
	}

	// Output:
	// Иван Петров Сидоров
	//   person     Иван Петров Сидоров
	//   given_name Иван                     nested ref=g1
	//   patronymic Петров                   nested ref=p1
	//   surname    Сидоров                  nested ref=s1
	// Анна Михаилова Кузнецова
	//   person     Анна Михаилова Кузнецова predicted
	//   given_name Анна                     nested ref=g2
	//   patronymic Михаилова                predicted candidate nested
	//   surname    Кузнецова                nested ref=s2
	// Анна Петрова Головина
	//   person     Анна Петрова Головина    predicted
	//   given_name Анна                     nested ref=g2
	//   patronymic Петрова                  nested ref=p1
	//   surname    Головина                 candidate nested alt=division
	// ответчик Сидоров
	//   person     Сидоров
	//   surname    Сидоров                  nested ref=s1
	// Мария, дочь Ивана Петрова
	//   given_name Мария                    ref=g3
	//   person     Ивана Петрова            predicted
	//   given_name Ивана                    nested ref=g1
	//   patronymic Петрова                  nested ref=p1
	//   fact child_of child=«Мария» parent=«Ивана Петрова»
}
