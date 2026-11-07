package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"desktop/internal/config"
	"desktop/internal/models"

	"github.com/joho/godotenv"
)

func repositoriesIntegrationDB(t *testing.T) *sql.DB {
	t.Helper()
	_ = godotenv.Load("../../.env")
	db, err := config.ConnectDatabase()
	if err != nil {
		t.Skipf("MySQL de integração indisponível; testes ignorados: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("fechamento da conexão MySQL falhou: %v", err)
		}
	})
	return db
}

func TestMySQLRepositoriesIntegration(t *testing.T) {
	db := repositoriesIntegrationDB(t)
	ctx := context.Background()
	prefix := fmt.Sprintf("TESTE_Repo_%d", time.Now().UnixNano())
	termPrefix := strings.ToLower(prefix) + "_termo"
	userEmail := strings.ToLower(prefix) + "@example.test"
	adminEmail := strings.ToLower(prefix) + "_admin@example.test"
	logOperation := prefix

	var userID, adminID, articleID uint64
	t.Cleanup(func() {
		cleanupExec := func(query string, args ...any) {
			if _, err := db.ExecContext(ctx, query, args...); err != nil {
				t.Errorf("limpeza do registro TESTE_ falhou: %v", err)
			}
		}
		cleanupExec(`DELETE FROM operation_logs WHERE operation = ?`, logOperation)
		if articleID != 0 {
			cleanupExec(`DELETE FROM articles WHERE id = ?`, articleID)
		}
		cleanupExec(`DELETE FROM articles WHERE title = ?`, prefix+" artigo")
		cleanupExec(`DELETE FROM authors WHERE name = ? AND id NOT IN (SELECT author_id FROM article_authors)`, prefix+" autor")
		cleanupExec(`DELETE FROM terms WHERE term LIKE ?`, termPrefix+"%")
		if userID != 0 {
			cleanupExec(`DELETE FROM users WHERE id = ?`, userID)
		}
		if adminID != 0 {
			cleanupExec(`DELETE FROM users WHERE id = ?`, adminID)
		}
		for _, check := range []struct {
			query string
			args  []any
		}{
			{query: `SELECT COUNT(*) FROM operation_logs WHERE operation = ?`, args: []any{logOperation}},
			{query: `SELECT COUNT(*) FROM articles WHERE title = ?`, args: []any{prefix + " artigo"}},
			{query: `SELECT COUNT(*) FROM authors WHERE name = ?`, args: []any{prefix + " autor"}},
			{query: `SELECT COUNT(*) FROM terms WHERE term LIKE ?`, args: []any{termPrefix + "%"}},
			{query: `SELECT COUNT(*) FROM users WHERE email IN (?, ?)`, args: []any{userEmail, adminEmail}},
		} {
			var count int
			if err := db.QueryRowContext(ctx, check.query, check.args...).Scan(&count); err != nil {
				t.Errorf("verificação da limpeza TESTE_ falhou: %v", err)
			} else if count != 0 {
				t.Errorf("limpeza deixou %d registros TESTE_ (%s)", count, check.query)
			}
		}
	})

	sourceRepo := NewSourceRepository(db)
	if err := sourceRepo.EnsureDefaults(ctx); err != nil {
		t.Fatalf("EnsureDefaults falhou: %v", err)
	}
	sources, err := sourceRepo.List(ctx)
	if err != nil || len(sources) == 0 {
		t.Fatalf("List de fontes inesperado: quantidade=%d erro=%v", len(sources), err)
	}
	var sourceID uint64
	for _, source := range sources {
		if source.Name == "local" {
			sourceID = source.ID
			break
		}
	}
	if sourceID == 0 {
		t.Fatal("fonte padrão local não encontrada")
	}

	users := NewUserRepository(db)
	user := &models.User{
		Name:         prefix + " usuário",
		Email:        userEmail,
		PasswordHash: "TESTE_hash",
		Role:         models.RolePesquisador,
	}
	if err := users.Create(ctx, user); err != nil {
		t.Fatalf("Create de usuário falhou: %v", err)
	}
	userID = user.ID
	if user.CreatedAt.IsZero() {
		t.Fatal("Create não carregou created_at")
	}
	foundUser, err := users.FindByEmail(ctx, userEmail)
	if err != nil || foundUser.ID != userID || foundUser.Email != userEmail {
		t.Fatalf("FindByEmail inesperado: user=%+v erro=%v", foundUser, err)
	}
	if _, err := users.FindByEmail(ctx, userEmail+"_ausente"); !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("FindByEmail ausente deveria retornar ErrUserNotFound: %v", err)
	}

	admin := &models.User{
		Name:         prefix + " administrador",
		Email:        adminEmail,
		PasswordHash: "TESTE_hash",
		Role:         models.RoleAdmin,
	}
	if err := users.Create(ctx, admin); err != nil {
		t.Fatalf("Create de Admin de teste falhou: %v", err)
	}
	adminID = admin.ID

	userManagement := NewUserRepository(db)
	listed, err := userManagement.ListAll(ctx)
	if err != nil {
		t.Fatalf("ListAll falhou: %v", err)
	}
	foundInList := false
	for _, item := range listed {
		if item.ID == userID {
			foundInList = true
			if item.PasswordHash != "" {
				t.Fatal("ListAll retornou hash de senha")
			}
		}
	}
	if !foundInList {
		t.Fatal("ListAll não retornou o usuário TESTE_")
	}
	foundByID, err := userManagement.FindByID(ctx, userID)
	if err != nil || foundByID.Email != userEmail {
		t.Fatalf("FindByID inesperado: user=%+v erro=%v", foundByID, err)
	}
	if _, err := userManagement.FindByID(ctx, 0); !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("FindByID ausente deveria retornar ErrUserNotFound: %v", err)
	}
	if changed, err := userManagement.UpdateRole(ctx, userID, models.RoleCurador); err != nil || !changed {
		t.Fatalf("UpdateRole deveria alterar usuário: changed=%v erro=%v", changed, err)
	}
	if changed, err := userManagement.UpdateRole(ctx, userID, models.RoleCurador); err != nil || changed {
		t.Fatalf("UpdateRole sem mudança deveria retornar false: changed=%v erro=%v", changed, err)
	}
	if changed, err := userManagement.UpdateRole(ctx, adminID, models.RoleCurador); err != nil || changed {
		t.Fatalf("UpdateRole não deveria alterar Admin: changed=%v erro=%v", changed, err)
	}
	if deleted, err := userManagement.Delete(ctx, adminID); err != nil || deleted {
		t.Fatalf("Delete não deveria remover Admin: deleted=%v erro=%v", deleted, err)
	}

	checksum := fmt.Sprintf("%064x", time.Now().UnixNano())
	filePath := "storage/articles/" + strings.ToLower(prefix) + ".pdf"
	result, err := db.ExecContext(
		ctx,
		`INSERT INTO articles
			(title, description, filename, file_path, content_type, size, user_id, source_id,
			 external_id, language_code, publication_date, checksum_sha256, index_status)
		 VALUES (?, ?, ?, ?, 'application/pdf', 123, ?, ?, ?, 'pt-BR', '2020-02-29', ?, 'PENDING')`,
		prefix+" artigo", prefix+" descrição", prefix+".pdf", filePath,
		userID, sourceID, prefix, checksum,
	)
	if err != nil {
		t.Fatalf("não foi possível criar artigo de teste: %v", err)
	}
	rawArticleID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("LastInsertId do artigo falhou: %v", err)
	}
	articleID = uint64(rawArticleID)

	authorName := prefix + " autor"
	authorResult, err := db.ExecContext(ctx, `INSERT INTO authors (name) VALUES (?)`, authorName)
	if err != nil {
		t.Fatalf("não foi possível criar autor de teste: %v", err)
	}
	authorID, err := authorResult.LastInsertId()
	if err != nil {
		t.Fatalf("LastInsertId do autor falhou: %v", err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO article_authors (article_id, author_id, author_order) VALUES (?, ?, 1)`, articleID, authorID); err != nil {
		t.Fatalf("não foi possível vincular autor de teste: %v", err)
	}

	articles := NewArticleRepository(db)
	allArticles, err := articles.List(ctx)
	if err != nil || !containsArticle(allArticles, articleID) {
		t.Fatalf("List não retornou artigo TESTE_: encontrado=%v erro=%v", containsArticle(allArticles, articleID), err)
	}
	matches, err := articles.Search(ctx, prefix+" autor")
	if err != nil || !containsArticle(matches, articleID) {
		t.Fatalf("Search por autor não encontrou artigo: encontrado=%v erro=%v", containsArticle(matches, articleID), err)
	}
	emptySearch, err := articles.Search(ctx, " ")
	if err != nil || !containsArticle(emptySearch, articleID) {
		t.Fatalf("Search vazio deveria delegar a List: encontrado=%v erro=%v", containsArticle(emptySearch, articleID), err)
	}
	detail, err := articles.GetByID(ctx, articleID)
	if err != nil || detail.Authors != authorName || detail.SourceName != "local" {
		t.Fatalf("GetByID inesperado: artigo=%+v erro=%v", detail, err)
	}
	if _, err := articles.GetByID(ctx, 0); !errors.Is(err, ErrArticleNotFound) {
		t.Fatalf("GetByID ausente deveria retornar ErrArticleNotFound: %v", err)
	}
	if exists, err := articles.ChecksumExists(ctx, checksum); err != nil || !exists {
		t.Fatalf("ChecksumExists deveria encontrar checksum: exists=%v erro=%v", exists, err)
	}
	if exists, err := articles.ChecksumExists(ctx, strings.Repeat("0", 64)); err != nil || exists {
		t.Fatalf("ChecksumExists encontrou checksum inexistente: exists=%v erro=%v", exists, err)
	}
	stats, err := articles.GetStats(ctx)
	if err != nil || stats.TotalArticles == 0 || stats.TotalAuthors == 0 || stats.TotalSources == 0 {
		t.Fatalf("GetStats inesperado: stats=%+v erro=%v", stats, err)
	}

	index := NewIndexRepository(db)
	targets, err := index.ListTargets(ctx, false)
	if err != nil || !containsTarget(targets, articleID) {
		t.Fatalf("ListTargets não incluiu artigo pendente: targets=%+v erro=%v", targets, err)
	}
	if err := index.MarkStatus(ctx, articleID, "ERROR"); err != nil {
		t.Fatalf("MarkStatus falhou: %v", err)
	}
	targets, err = index.ListTargets(ctx, false)
	if err != nil || !containsTarget(targets, articleID) {
		t.Fatalf("ListTargets não incluiu artigo com erro: targets=%+v erro=%v", targets, err)
	}
	term := termPrefix + "_unico"
	if err := index.PersistDocument(ctx, models.IndexedDocument{
		ArticleID:      articleID,
		DocumentLength: 2,
		Terms:          []models.TermStat{{Term: term, Frequency: 2, Positions: []int{0, 1}}},
	}); err != nil {
		t.Fatalf("PersistDocument falhou: %v", err)
	}
	targets, err = index.ListTargets(ctx, false)
	if err != nil || containsTarget(targets, articleID) {
		t.Fatalf("artigo indexado ainda aparece como pendente/erro: targets=%+v erro=%v", targets, err)
	}
	allTargets, err := index.ListTargets(ctx, true)
	if err != nil || !containsTarget(allTargets, articleID) {
		t.Fatalf("ListTargets com reindexAll omitiu artigo: targets=%+v erro=%v", allTargets, err)
	}
	termID, _, found, err := index.TermInfo(ctx, term)
	if err != nil || !found {
		t.Fatalf("TermInfo não encontrou termo de teste: id=%d found=%v erro=%v", termID, found, err)
	}
	if _, _, found, err := index.TermInfo(ctx, term+"_inexistente"); err != nil || found {
		t.Fatalf("TermInfo encontrou termo inexistente: found=%v erro=%v", found, err)
	}
	postings, err := index.Postings(ctx, termID)
	if err != nil || len(postings) != 1 || postings[0].ArticleID != articleID || postings[0].TermFrequency != 2 {
		t.Fatalf("Postings inesperado: %+v erro=%v", postings, err)
	}
	resultByID, err := index.SearchResultByArticleID(ctx, articleID)
	if err != nil || resultByID.Title != prefix+" artigo" || resultByID.Authors != authorName {
		t.Fatalf("SearchResultByArticleID inesperado: %+v erro=%v", resultByID, err)
	}
	candidates, err := index.CandidateTerms(ctx, termPrefix, 1, 255, 10)
	if err != nil || !containsTerm(candidates, term) {
		t.Fatalf("CandidateTerms não encontrou termo TESTE_: candidatos=%+v erro=%v", candidates, err)
	}
	if _, _, err := index.CorpusStats(ctx); err != nil {
		t.Fatalf("CorpusStats falhou: %v", err)
	}
	overview, err := index.Overview(ctx)
	if err != nil || overview.Indexed == 0 || overview.Documents == 0 {
		t.Fatalf("Overview inesperado: overview=%+v erro=%v", overview, err)
	}

	logs := NewLogRepository(db)
	longDetails := strings.Repeat("TESTE_", 100)
	if err := logs.Create(ctx, models.LogEntry{
		UserID:      userID,
		Operation:   logOperation,
		Entity:      "TESTE_",
		EntityID:    articleID,
		Details:     longDetails,
		ResultCount: 2,
	}); err != nil {
		t.Fatalf("Create de log falhou: %v", err)
	}
	logRows, err := logs.List(ctx, models.LogFilters{
		Operation: logOperation,
		UserID:    userID,
		DateFrom:  "2000-01-01",
		DateTo:    "2999-12-31",
		Limit:     501,
	})
	if err != nil || len(logRows) != 1 {
		t.Fatalf("List de logs inesperado: quantidade=%d erro=%v", len(logRows), err)
	}
	if logRows[0].UserName != user.Name || logRows[0].EntityID != articleID || len(logRows[0].Details) != 500 {
		t.Fatalf("dados de log inesperados: %+v", logRows[0])
	}
	if err := logs.Create(ctx, models.LogEntry{Operation: logOperation, Entity: "TESTE_", Details: "TESTE sem usuário"}); err != nil {
		t.Fatalf("Create de log sem usuário falhou: %v", err)
	}
	anonymousLogs, err := logs.List(ctx, models.LogFilters{Operation: logOperation})
	if err != nil || len(anonymousLogs) != 2 {
		t.Fatalf("List de logs sem usuário inesperado: quantidade=%d erro=%v", len(anonymousLogs), err)
	}
	foundRemovedUser := false
	for _, entry := range anonymousLogs {
		if entry.UserName == "usuário removido" {
			foundRemovedUser = true
		}
	}
	if !foundRemovedUser {
		t.Fatal("List deveria identificar log sem usuário")
	}
	summary, err := logs.Summary(ctx)
	if err != nil || summary.Total == 0 {
		t.Fatalf("Summary de logs inesperado: summary=%+v erro=%v", summary, err)
	}

	if deleted, err := userManagement.Delete(ctx, userID); err != nil || !deleted {
		t.Fatalf("Delete de usuário deveria removê-lo: deleted=%v erro=%v", deleted, err)
	}
	if deleted, err := userManagement.Delete(ctx, userID); err != nil || deleted {
		t.Fatalf("Delete repetido deveria retornar false: deleted=%v erro=%v", deleted, err)
	}
	if _, err := users.FindByEmail(ctx, userEmail); !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("usuário excluído ainda foi encontrado: %v", err)
	}

	filePathResult, err := articles.Delete(ctx, articleID)
	if err != nil || filePathResult != filePath {
		t.Fatalf("Delete de artigo inesperado: path=%q erro=%v", filePathResult, err)
	}
	if _, err := articles.Delete(ctx, articleID); !errors.Is(err, ErrArticleNotFound) {
		t.Fatalf("Delete repetido deveria retornar ErrArticleNotFound: %v", err)
	}
}

func containsArticle(items []models.ArticleSummary, id uint64) bool {
	for _, item := range items {
		if item.ID == id {
			return true
		}
	}
	return false
}

func containsTarget(items []models.IndexTarget, id uint64) bool {
	for _, item := range items {
		if item.ID == id {
			return true
		}
	}
	return false
}

func containsTerm(items []TermCandidate, term string) bool {
	for _, item := range items {
		if item.Term == term {
			return true
		}
	}
	return false
}
