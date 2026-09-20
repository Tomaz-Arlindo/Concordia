package services

import "testing"

func TestQueryTermsRemovesStopWordsAndDuplicates(t *testing.T) {
	terms := QueryTerms("Supino reto e supino com carga")
	if len(terms) != 3 {
		t.Fatalf("esperava 3 termos, recebeu %d: %#v", len(terms), terms)
	}
	if terms[0] != "supino" || terms[1] != "reto" || terms[2] != "carga" {
		t.Fatalf("termos inesperados: %#v", terms)
	}
}

func TestBuildIndexedDocumentTracksFrequencyAndPositions(t *testing.T) {
	doc := BuildIndexedDocument(10, "go paralelo go busca")
	if doc.DocumentLength != 4 {
		t.Fatalf("document_length esperado 4, recebeu %d", doc.DocumentLength)
	}
	found := false
	for _, term := range doc.Terms {
		if term.Term == "go" {
			found = true
			if term.Frequency != 2 {
				t.Fatalf("frequência de go esperada 2, recebeu %d", term.Frequency)
			}
			if len(term.Positions) != 2 || term.Positions[0] != 0 || term.Positions[1] != 2 {
				t.Fatalf("posições inesperadas: %#v", term.Positions)
			}
		}
	}
	if !found {
		t.Fatal("termo go não encontrado")
	}
}


func TestNormalizeTermCollapsesAccentAndCaseVariants(t *testing.T) {
	cases := map[string]string{
		"Martínez": "martinez",
		"MARTINEZ": "martinez",
		"Gülçehre":  "gulcehre",
		"Krämer":    "kramer",
		"François":  "francois",
		"Stéphane":  "stephane",
		"É":         "e",
		"２D":        "2d",
	}

	for input, expected := range cases {
		if got := NormalizeTerm(input); got != expected {
			t.Fatalf("NormalizeTerm(%q): esperado %q, recebeu %q", input, expected, got)
		}
	}
}

func TestBuildIndexedDocumentMergesCollationEquivalentTokens(t *testing.T) {
	doc := BuildIndexedDocument(
		99,
		"Martinez Martínez MARTINEZ Gülçehre Gulcehre François Francois",
	)

	expected := map[string]uint64{
		"martinez": 3,
		"gulcehre": 2,
		"francois": 2,
	}

	if len(doc.Terms) != len(expected) {
		t.Fatalf("esperava %d termos normalizados, recebeu %d: %#v", len(expected), len(doc.Terms), doc.Terms)
	}

	for _, term := range doc.Terms {
		want, ok := expected[term.Term]
		if !ok {
			t.Fatalf("termo inesperado: %q", term.Term)
		}
		if term.Frequency != want {
			t.Fatalf("frequência de %q: esperava %d, recebeu %d", term.Term, want, term.Frequency)
		}
	}
}
