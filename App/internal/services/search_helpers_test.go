package services

import (
	"testing"

	"desktop/internal/models"
)

func TestLevenshtein(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"supino", "supino", 0},
		{"supno", "supino", 1},
		{"busca", "buca", 1},
		{"go", "golang", 4},
	}
	for _, tc := range cases {
		if got := levenshtein(tc.a, tc.b); got != tc.want {
			t.Fatalf("levenshtein(%q,%q)=%d; esperado %d", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestMatchesSearchFilters(t *testing.T) {
	item := models.SearchResult{
		SourceName:      "local",
		Authors:         "João Silva, Maria Santos",
		LanguageCode:    "pt-BR",
		PublicationDate: "2026-09-18",
	}
	filters := models.SearchFilters{
		Source:   "local",
		Author:   "maria",
		Language: "pt-BR",
		DateFrom: "2026-01-01",
		DateTo:   "2026-12-31",
	}
	if !matchesSearchFilters(item, filters) {
		t.Fatal("esperava que o item atendesse aos filtros")
	}
}
