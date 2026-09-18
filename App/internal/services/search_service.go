package services

import (
	"context"
	"math"
	"sort"
	"strings"

	"desktop/internal/models"
	"desktop/internal/repositories"
)

type SearchService struct {
	index *repositories.IndexRepository
}

func NewSearchService(index *repositories.IndexRepository) *SearchService {
	return &SearchService{index: index}
}

type scoreAccumulator struct {
	Score   float64
	Matches []string
}

type weightedQueryTerm struct {
	Original string
	Term     string
	Weight   float64
}

func (s *SearchService) BM25(ctx context.Context, query string, limit int) ([]models.SearchResult, error) {
	return s.BM25Filtered(ctx, query, models.SearchFilters{Limit: limit})
}

func (s *SearchService) BM25Filtered(ctx context.Context, query string, filters models.SearchFilters) ([]models.SearchResult, error) {
	queryTerms := QueryTerms(query)
	if len(queryTerms) == 0 {
		return []models.SearchResult{}, nil
	}
	if filters.Limit <= 0 || filters.Limit > 100 {
		filters.Limit = 50
	}

	documents, avgLength, err := s.index.CorpusStats(ctx)
	if err != nil || documents == 0 || avgLength <= 0 {
		return []models.SearchResult{}, err
	}

	weightedTerms := make([]weightedQueryTerm, 0, len(queryTerms))
	for _, term := range queryTerms {
		_, _, found, err := s.index.TermInfo(ctx, term)
		if err != nil {
			return nil, err
		}
		if found {
			weightedTerms = append(weightedTerms, weightedQueryTerm{Original: term, Term: term, Weight: 1.0})
			continue
		}
		if !filters.Fuzzy {
			continue
		}

		candidate, ok, err := s.closestIndexedTerm(ctx, term)
		if err != nil {
			return nil, err
		}
		if ok {
			weightedTerms = append(weightedTerms, weightedQueryTerm{Original: term, Term: candidate, Weight: 0.72})
		}
	}

	if len(weightedTerms) == 0 {
		return []models.SearchResult{}, nil
	}

	const k1 = 1.2
	const b = 0.75
	scores := make(map[uint64]*scoreAccumulator)

	for _, queryTerm := range weightedTerms {
		termID, df, found, err := s.index.TermInfo(ctx, queryTerm.Term)
		if err != nil {
			return nil, err
		}
		if !found || df == 0 {
			continue
		}

		idf := math.Log(1 + (float64(documents)-float64(df)+0.5)/(float64(df)+0.5))
		postings, err := s.index.Postings(ctx, termID)
		if err != nil {
			return nil, err
		}

		for _, posting := range postings {
			tf := float64(posting.TermFrequency)
			dl := float64(posting.DocumentLength)
			denominator := tf + k1*(1-b+b*(dl/avgLength))
			if denominator == 0 {
				continue
			}
			score := idf * ((tf * (k1 + 1)) / denominator) * queryTerm.Weight

			acc := scores[posting.ArticleID]
			if acc == nil {
				acc = &scoreAccumulator{}
				scores[posting.ArticleID] = acc
			}
			acc.Score += score
			if queryTerm.Original == queryTerm.Term {
				acc.Matches = appendUnique(acc.Matches, queryTerm.Term)
			} else {
				acc.Matches = appendUnique(acc.Matches, queryTerm.Original+"→"+queryTerm.Term)
			}
		}
	}

	results := make([]models.SearchResult, 0, len(scores))
	for articleID, acc := range scores {
		item, err := s.index.SearchResultByArticleID(ctx, articleID)
		if err != nil {
			return nil, err
		}
		if !matchesSearchFilters(item, filters) {
			continue
		}
		item.Score = acc.Score
		item.MatchedTerms = strings.Join(acc.Matches, ", ")
		results = append(results, item)
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Score == results[j].Score {
			return results[i].ID < results[j].ID
		}
		return results[i].Score > results[j].Score
	})

	if len(results) > filters.Limit {
		results = results[:filters.Limit]
	}
	return results, nil
}

func (s *SearchService) closestIndexedTerm(ctx context.Context, term string) (string, bool, error) {
	runes := []rune(term)
	if len(runes) == 0 {
		return "", false, nil
	}

	maxDistance := 2
	if len(runes) <= 4 {
		maxDistance = 1
	}
	minLen := len(runes) - maxDistance
	if minLen < 1 {
		minLen = 1
	}
	maxLen := len(runes) + maxDistance

	candidates, err := s.index.CandidateTerms(ctx, string(runes[0]), minLen, maxLen, 200)
	if err != nil {
		return "", false, err
	}

	best := ""
	bestDistance := maxDistance + 1
	bestDF := uint64(0)
	for _, candidate := range candidates {
		distance := levenshtein(term, candidate.Term)
		if distance > maxDistance {
			continue
		}
		if distance < bestDistance || (distance == bestDistance && candidate.DocumentFrequency > bestDF) {
			best = candidate.Term
			bestDistance = distance
			bestDF = candidate.DocumentFrequency
		}
	}

	return best, best != "", nil
}

func matchesSearchFilters(item models.SearchResult, filters models.SearchFilters) bool {
	if value := strings.TrimSpace(filters.Source); value != "" && !strings.EqualFold(item.SourceName, value) {
		return false
	}
	if value := strings.ToLower(strings.TrimSpace(filters.Author)); value != "" && !strings.Contains(strings.ToLower(item.Authors), value) {
		return false
	}
	if value := strings.TrimSpace(filters.Language); value != "" && !strings.EqualFold(item.LanguageCode, value) {
		return false
	}
	if filters.DateFrom != "" && item.PublicationDate != "" && item.PublicationDate < filters.DateFrom {
		return false
	}
	if filters.DateTo != "" && item.PublicationDate != "" && item.PublicationDate > filters.DateTo {
		return false
	}
	if filters.DateFrom != "" && item.PublicationDate == "" {
		return false
	}
	if filters.DateTo != "" && item.PublicationDate == "" {
		return false
	}
	return true
}

func appendUnique(items []string, value string) []string {
	for _, item := range items {
		if item == value {
			return items
		}
	}
	return append(items, value)
}

func levenshtein(a, b string) int {
	ar := []rune(a)
	br := []rune(b)
	if len(ar) == 0 {
		return len(br)
	}
	if len(br) == 0 {
		return len(ar)
	}

	previous := make([]int, len(br)+1)
	current := make([]int, len(br)+1)
	for j := range previous {
		previous[j] = j
	}

	for i, ra := range ar {
		current[0] = i + 1
		for j, rb := range br {
			cost := 0
			if ra != rb {
				cost = 1
			}
			deletion := previous[j+1] + 1
			insertion := current[j] + 1
			substitution := previous[j] + cost
			current[j+1] = min3(deletion, insertion, substitution)
		}
		previous, current = current, previous
	}
	return previous[len(br)]
}

func min3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}
