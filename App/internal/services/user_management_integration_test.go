package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"desktop/internal/config"
	"desktop/internal/models"
	"desktop/internal/repositories"

	"github.com/joho/godotenv"
)

func servicesIntegrationDB(t *testing.T) *sql.DB {
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

func TestUserManagementServiceMySQLErrorsAndRoles(t *testing.T) {
	db := servicesIntegrationDB(t)
	ctx := context.Background()
	prefix := fmt.Sprintf("TESTE_UserMgmt_%d", time.Now().UnixNano())
	repo := repositories.NewUserRepository(db)
	var userID, adminID uint64
	t.Cleanup(func() {
		deleteUser := func(id uint64) {
			if _, err := db.ExecContext(ctx, `DELETE FROM users WHERE id = ?`, id); err != nil {
				t.Errorf("limpeza do usuário TESTE_ falhou: %v", err)
			}
		}
		if userID != 0 {
			deleteUser(userID)
		}
		if adminID != 0 {
			deleteUser(adminID)
		}
		for _, email := range []string{strings.ToLower(prefix) + "@example.test", strings.ToLower(prefix) + "_admin@example.test"} {
			var count int
			if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE email = ?`, email).Scan(&count); err != nil {
				t.Errorf("verificação da limpeza de usuário TESTE_ falhou: %v", err)
			} else if count != 0 {
				t.Errorf("limpeza deixou usuário TESTE_ %q", email)
			}
		}
	})

	auth := NewAuthService(repo)
	user, err := auth.Register(ctx, models.RegisterInput{
		Name:     prefix + " usuário",
		Email:    strings.ToLower(prefix) + "@example.test",
		Password: "SenhaTeste09",
	})
	if err != nil {
		t.Fatalf("AuthService.Register falhou: %v", err)
	}
	userID = user.ID
	if _, err := auth.Register(ctx, models.RegisterInput{Name: prefix, Email: strings.ToLower(prefix) + "@example.test", Password: "SenhaTeste09"}); !errors.Is(err, ErrEmailAlreadyExists) {
		t.Fatalf("cadastro duplicado deveria retornar ErrEmailAlreadyExists: %v", err)
	}
	if _, err := auth.Login(ctx, user.Email, "SenhaIncorreta"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("senha incorreta deveria retornar ErrInvalidCredentials: %v", err)
	}
	if loggedIn, err := auth.Login(ctx, user.Email, "SenhaTeste09"); err != nil || loggedIn.ID != userID {
		t.Fatalf("login correto falhou: user=%+v erro=%v", loggedIn, err)
	}

	admin := &models.User{
		Name:         prefix + " administrador",
		Email:        strings.ToLower(prefix) + "_admin@example.test",
		PasswordHash: "TESTE_hash",
		Role:         models.RoleAdmin,
	}
	if err := repo.Create(ctx, admin); err != nil {
		t.Fatalf("não foi possível criar Admin TESTE_: %v", err)
	}
	adminID = admin.ID

	service := NewUserManagementService(repo)
	listed, err := service.List(ctx)
	if err != nil || len(listed) == 0 {
		t.Fatalf("UserManagementService.List falhou: usuários=%d erro=%v", len(listed), err)
	}
	setup := NewAdminSetupService(repo)
	needsSetup, err := setup.NeedsSetup(ctx)
	if err != nil || needsSetup {
		t.Fatalf("NeedsSetup deveria detectar Admin existente: needs=%v erro=%v", needsSetup, err)
	}
	if _, err := setup.CreateInitialAdmin(ctx, models.RegisterInput{
		Name: prefix + " admin duplicado", Email: strings.ToLower(prefix) + "_other@example.test", Password: "SenhaTeste09",
	}); !errors.Is(err, ErrAdminAlreadyExists) {
		t.Fatalf("criação de Admin adicional deveria retornar ErrAdminAlreadyExists: %v", err)
	}
	if _, err := service.ChangeRole(ctx, userID, models.RoleAdmin); !errors.Is(err, ErrInvalidRole) {
		t.Fatalf("alteração para Admin deveria retornar ErrInvalidRole: %v", err)
	}
	if _, err := service.ChangeRole(ctx, 0, models.RoleCurador); !errors.Is(err, repositories.ErrUserNotFound) {
		t.Fatalf("alteração de usuário inexistente deveria retornar ErrUserNotFound: %v", err)
	}
	if _, err := service.Delete(ctx, 0); !errors.Is(err, repositories.ErrUserNotFound) {
		t.Fatalf("exclusão de usuário inexistente deveria retornar ErrUserNotFound: %v", err)
	}
	if _, err := service.ChangeRole(ctx, adminID, models.RoleCurador); !errors.Is(err, ErrProtectedUser) {
		t.Fatalf("Admin não deveria poder ser alterado: %v", err)
	}
	if _, err := service.Delete(ctx, adminID); !errors.Is(err, ErrProtectedUser) {
		t.Fatalf("Admin não deveria poder ser excluído: %v", err)
	}
	unchanged, err := service.ChangeRole(ctx, userID, models.RolePesquisador)
	if err != nil || unchanged.Role != models.RolePesquisador {
		t.Fatalf("manter perfil atual deveria ser idempotente: user=%+v erro=%v", unchanged, err)
	}
	updated, err := service.ChangeRole(ctx, userID, models.RoleCurador)
	if err != nil || updated.Role != models.RoleCurador {
		t.Fatalf("alteração para Curador falhou: user=%+v erro=%v", updated, err)
	}
	deleted, err := service.Delete(ctx, userID)
	if err != nil || deleted.ID != userID {
		t.Fatalf("exclusão de usuário TESTE_ falhou: user=%+v erro=%v", deleted, err)
	}
	if _, err := service.Delete(ctx, userID); !errors.Is(err, repositories.ErrUserNotFound) {
		t.Fatalf("segunda exclusão deveria retornar ErrUserNotFound: %v", err)
	}
}

func TestArticleServiceMySQLWorkflowAndMissingArticles(t *testing.T) {
	db := servicesIntegrationDB(t)
	ctx := context.Background()
	prefix := fmt.Sprintf("TESTE_ArticleService_%d", time.Now().UnixNano())
	email := strings.ToLower(prefix) + "@example.test"
	userRepo := repositories.NewUserRepository(db)
	user := &models.User{
		Name:         prefix + " usuário",
		Email:        email,
		PasswordHash: "TESTE_hash",
		Role:         models.RoleCurador,
	}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("não foi possível criar usuário TESTE_: %v", err)
	}

	var articlePath string
	var indexTerm string
	t.Cleanup(func() {
		rows, err := db.QueryContext(ctx, `SELECT file_path FROM articles WHERE title LIKE ?`, prefix+"%")
		if err == nil {
			for rows.Next() {
				var path string
				if scanErr := rows.Scan(&path); scanErr != nil {
					t.Errorf("leitura do caminho do artigo TESTE_ falhou: %v", scanErr)
				} else if path != "" {
					if removeErr := os.Remove(filepath.FromSlash(path)); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
						t.Errorf("remoção do PDF TESTE_ falhou: %v", removeErr)
					}
				}
			}
			if rowsErr := rows.Err(); rowsErr != nil {
				t.Errorf("leitura dos artigos TESTE_ falhou: %v", rowsErr)
			}
			if closeErr := rows.Close(); closeErr != nil {
				t.Errorf("fechamento dos resultados de artigos TESTE_ falhou: %v", closeErr)
			}
		} else {
			t.Errorf("consulta dos artigos TESTE_ para limpeza falhou: %v", err)
		}
		if articlePath != "" {
			if err := os.Remove(filepath.FromSlash(articlePath)); err != nil && !errors.Is(err, os.ErrNotExist) {
				t.Errorf("remoção do PDF TESTE_ falhou: %v", err)
			}
		}
		cleanupExec := func(query string, args ...any) {
			if _, err := db.ExecContext(ctx, query, args...); err != nil {
				t.Errorf("limpeza dos dados TESTE_ falhou: %v", err)
			}
		}
		cleanupExec(`DELETE FROM articles WHERE title IN (?, ?)`, prefix+" artigo", prefix+" artigo atualizado")
		cleanupExec(`DELETE FROM authors WHERE name IN (?, ?) AND id NOT IN (SELECT author_id FROM article_authors)`, prefix+" Autor", prefix+" Autor 2")
		cleanupExec(`DELETE FROM terms WHERE term LIKE ?`, strings.ToLower(prefix)+"%")
		if indexTerm != "" {
			cleanupExec(`DELETE FROM terms WHERE term = ?`, indexTerm)
		}
		cleanupExec(`DELETE FROM users WHERE id = ?`, user.ID)
		for _, check := range []struct {
			query string
			args  []any
		}{
			{query: `SELECT COUNT(*) FROM articles WHERE title IN (?, ?)`, args: []any{prefix + " artigo", prefix + " artigo atualizado"}},
			{query: `SELECT COUNT(*) FROM authors WHERE name IN (?, ?)`, args: []any{prefix + " Autor", prefix + " Autor 2"}},
			{query: `SELECT COUNT(*) FROM terms WHERE term LIKE ?`, args: []any{strings.ToLower(prefix) + "%"}},
			{query: `SELECT COUNT(*) FROM users WHERE email = ?`, args: []any{email}},
		} {
			var count int
			if err := db.QueryRowContext(ctx, check.query, check.args...).Scan(&count); err != nil {
				t.Errorf("verificação da limpeza TESTE_ falhou: %v", err)
			} else if count != 0 {
				t.Errorf("limpeza deixou %d registros TESTE_ (%s)", count, check.query)
			}
		}
	})

	var sourceID uint64
	if err := db.QueryRowContext(ctx, `SELECT id FROM article_sources WHERE name = 'local' LIMIT 1`).Scan(&sourceID); err != nil {
		t.Fatalf("fonte local ausente: %v", err)
	}
	service := NewArticleService(db, repositories.NewArticleRepository(db), repositories.NewSourceRepository(db))
	missingID := uint64(0)
	if err := db.QueryRowContext(ctx, `SELECT COALESCE(MAX(id), 0) + 1000000 FROM articles`).Scan(&missingID); err != nil {
		t.Fatalf("não foi possível gerar id inexistente: %v", err)
	}
	if err := service.UpdateArticle(ctx, UpdateArticleInput{
		ID: missingID, Title: prefix + " inexistente", SourceID: sourceID,
	}); !errors.Is(err, repositories.ErrArticleNotFound) {
		t.Fatalf("edição de artigo inexistente deveria retornar ErrArticleNotFound: %v", err)
	}
	if err := service.DeleteArticle(ctx, missingID); !errors.Is(err, repositories.ErrArticleNotFound) {
		t.Fatalf("exclusão de artigo inexistente deveria retornar ErrArticleNotFound: %v", err)
	}

	pdfPath := filepath.Join(t.TempDir(), "TESTE_artigo.pdf")
	if err := os.WriteFile(pdfPath, []byte("%PDF-1.4 TESTE conteúdo de integração"), 0o600); err != nil {
		t.Fatal(err)
	}
	articleID, err := service.ImportPDF(ctx, ImportArticleInput{
		Title:           prefix + " artigo",
		SourceID:        sourceID,
		AuthorsCSV:      prefix + " Autor, " + prefix + " Autor",
		LanguageCode:    "pt-BR",
		PublicationDate: "2020-02-29",
		OriginalPath:    pdfPath,
		UserID:          user.ID,
	})
	if err != nil {
		t.Fatalf("ImportPDF válido falhou: %v", err)
	}
	article, err := service.GetArticle(ctx, articleID)
	if err != nil {
		t.Fatalf("GetArticle falhou: %v", err)
	}
	articlePath = article.FilePath
	if article.Authors != prefix+" Autor" {
		t.Fatalf("autores não foram normalizados: %q", article.Authors)
	}
	if _, err := service.ImportPDF(ctx, ImportArticleInput{
		Title: prefix + " duplicado", SourceID: sourceID, OriginalPath: pdfPath, UserID: user.ID,
	}); !errors.Is(err, ErrDuplicateFile) {
		t.Fatalf("arquivo duplicado deveria retornar ErrDuplicateFile: %v", err)
	}
	if sources, err := service.ListSources(ctx); err != nil || len(sources) == 0 {
		t.Fatalf("ListSources falhou: fontes=%d erro=%v", len(sources), err)
	}
	if articles, err := service.ListArticles(ctx); err != nil || len(articles) == 0 {
		t.Fatalf("ListArticles falhou: artigos=%d erro=%v", len(articles), err)
	}
	if articles, err := service.SearchArticles(ctx, prefix); err != nil || len(articles) == 0 {
		t.Fatalf("SearchArticles falhou: artigos=%d erro=%v", len(articles), err)
	}
	if stats, err := service.GetStats(ctx); err != nil || stats.TotalArticles == 0 {
		t.Fatalf("GetStats falhou: stats=%+v erro=%v", stats, err)
	}

	if err := service.UpdateArticle(ctx, UpdateArticleInput{
		ID: articleID, Title: prefix + " artigo atualizado", SourceID: sourceID,
		AuthorsCSV: prefix + " Autor, " + prefix + " Autor 2", LanguageCode: "en",
		PublicationDate: "2021-01-01",
	}); err != nil {
		t.Fatalf("UpdateArticle válido falhou: %v", err)
	}
	updated, err := service.GetArticle(ctx, articleID)
	if err != nil || updated.Title != prefix+" artigo atualizado" || updated.Authors != prefix+" Autor, "+prefix+" Autor 2" {
		t.Fatalf("artigo atualizado inesperado: artigo=%+v erro=%v", updated, err)
	}

	index := repositories.NewIndexRepository(db)
	indexTerm = fmt.Sprintf("zzq%d", time.Now().UnixNano())
	if err := index.PersistDocument(ctx, models.IndexedDocument{
		ArticleID:      articleID,
		DocumentLength: 2,
		Terms:          []models.TermStat{{Term: indexTerm, Frequency: 2, Positions: []int{0, 1}}},
	}); err != nil {
		t.Fatalf("PersistDocument do termo de teste falhou: %v", err)
	}
	if _, err := db.ExecContext(ctx, `UPDATE terms SET document_frequency = 1 WHERE term = ?`, indexTerm); err != nil {
		t.Fatalf("não foi possível preparar frequência do termo TESTE_: %v", err)
	}
	search := NewSearchService(index)
	results, err := search.BM25(ctx, indexTerm, 10)
	if err != nil || len(results) != 1 || results[0].ID != articleID || results[0].MatchedTerms != indexTerm {
		t.Fatalf("BM25 não retornou o artigo indexado: resultados=%+v erro=%v", results, err)
	}
	filtered, err := search.BM25Filtered(ctx, indexTerm, models.SearchFilters{
		Source: "local", Author: prefix + " Autor", Language: "en",
		DateFrom: "2020-01-01", DateTo: "2022-01-01", Limit: 10,
	})
	if err != nil || len(filtered) != 1 || filtered[0].ID != articleID {
		t.Fatalf("BM25Filtered recusou filtros compatíveis: resultados=%+v erro=%v", filtered, err)
	}
	filtered, err = search.BM25Filtered(ctx, indexTerm, models.SearchFilters{Source: "TESTE_fonte_inexistente", Limit: 10})
	if err != nil || len(filtered) != 0 {
		t.Fatalf("BM25Filtered aceitou fonte incompatível: resultados=%+v erro=%v", filtered, err)
	}
	fuzzy, err := search.BM25Filtered(ctx, indexTerm+"x", models.SearchFilters{Fuzzy: true, Limit: 10})
	if err != nil || len(fuzzy) != 1 || !strings.Contains(fuzzy[0].MatchedTerms, "→") {
		t.Fatalf("busca aproximada não encontrou termo de teste: resultados=%+v erro=%v", fuzzy, err)
	}

	if err := service.DeleteArticle(ctx, articleID); err != nil {
		t.Fatalf("DeleteArticle falhou: %v", err)
	}
	if _, err := os.Stat(filepath.FromSlash(articlePath)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("PDF não foi removido após a exclusão: %v", err)
	}
	if err := service.DeleteArticle(ctx, articleID); !errors.Is(err, repositories.ErrArticleNotFound) {
		t.Fatalf("exclusão repetida deveria retornar ErrArticleNotFound: %v", err)
	}
}
