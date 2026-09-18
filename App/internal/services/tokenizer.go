package services

import (
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"desktop/internal/models"

	"golang.org/x/text/unicode/norm"
)

var tokenPattern = regexp.MustCompile(`[\p{L}\p{N}]+`)

var stopWords = map[string]struct{}{
	"a": {}, "ao": {}, "aos": {}, "as": {}, "com": {}, "como": {}, "da": {}, "das": {},
	"de": {}, "do": {}, "dos": {}, "e": {}, "em": {}, "entre": {}, "foi": {},
	"na": {}, "nas": {}, "no": {}, "nos": {}, "o": {}, "os": {}, "ou": {}, "para": {},
	"por": {}, "que": {}, "se": {}, "sem": {}, "ser": {}, "sua": {}, "suas": {}, "seu": {},
	"seus": {}, "um": {}, "uma": {}, "uns": {}, "umas": {}, "the": {}, "and": {}, "of": {},
	"to": {}, "in": {}, "for": {}, "on": {}, "with": {}, "is": {}, "are": {}, "by": {},
	"an": {}, "this": {}, "that": {}, "from": {},
}

// NormalizeTerm transforma tokens Unicode em uma forma estável para o índice.
//
// O banco usa utf8mb4_unicode_ci, uma collation que é case-insensitive e,
// para muitos caracteres latinos, accent-insensitive. Sem normalização,
// "Martinez" e "Martínez", por exemplo, podem ser chaves diferentes no map
// do Go, mas apontar para o mesmo terms.id no MySQL. Isso gerava duas
// tentativas de inserir a mesma PK (term_id, article_id).
//
// NFKD também transforma algumas formas Unicode compatíveis (por exemplo,
// caracteres de largura diferente) antes da remoção de marcas diacríticas.
func NormalizeTerm(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return ""
	}

	decomposed := norm.NFKD.String(value)

	var builder strings.Builder
	builder.Grow(len(decomposed))

	for _, r := range decomposed {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		builder.WriteRune(r)
	}

	return strings.TrimSpace(builder.String())
}

func BuildIndexedDocument(articleID uint64, text string) models.IndexedDocument {
	raw := tokenPattern.FindAllString(text, -1)
	positions := make(map[string][]int)
	indexedPosition := 0

	for _, rawToken := range raw {
		token := NormalizeTerm(rawToken)
		if token == "" || utf8.RuneCountInString(token) < 2 || utf8.RuneCountInString(token) > 120 {
			continue
		}
		if _, blocked := stopWords[token]; blocked {
			continue
		}

		positions[token] = append(positions[token], indexedPosition)
		indexedPosition++
	}

	terms := make([]models.TermStat, 0, len(positions))
	for term, pos := range positions {
		terms = append(terms, models.TermStat{
			Term:      term,
			Frequency: uint64(len(pos)),
			Positions: pos,
		})
	}

	sort.Slice(terms, func(i, j int) bool {
		return terms[i].Term < terms[j].Term
	})

	return models.IndexedDocument{
		ArticleID:      articleID,
		DocumentLength: uint64(indexedPosition),
		Terms:          terms,
	}
}

func QueryTerms(query string) []string {
	raw := tokenPattern.FindAllString(query, -1)
	seen := make(map[string]struct{})
	out := make([]string, 0)

	for _, rawToken := range raw {
		token := NormalizeTerm(rawToken)
		if token == "" || utf8.RuneCountInString(token) < 2 {
			continue
		}
		if _, blocked := stopWords[token]; blocked {
			continue
		}
		if _, exists := seen[token]; exists {
			continue
		}

		seen[token] = struct{}{}
		out = append(out, token)
	}

	return out
}
