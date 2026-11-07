package main

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
	"desktop/internal/services"

	"github.com/joho/godotenv"
)

func integrationDB(t *testing.T) *sql.DB {
	t.Helper()
	_ = godotenv.Load(".env")
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

func TestMySQLIntegrationSprint09(t *testing.T) {
	db := integrationDB(t)
	ctx := context.Background()
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	namePrefix := "TESTE_Sprint09_" + suffix
	userEmail := strings.ToLower(namePrefix) + "@exemplo.com"
	app := &App{ctx: ctx, db: db}
	userRepo := repositories.NewUserRepository(db)
	articleRepo := repositories.NewArticleRepository(db)
	sourceRepo := repositories.NewSourceRepository(db)
	app.authService = services.NewAuthService(userRepo)
	app.articleService = services.NewArticleService(db, articleRepo, sourceRepo)
	app.searchService = services.NewSearchService(repositories.NewIndexRepository(db))
	app.logService = services.NewLogService(repositories.NewLogRepository(db))

	var articlePaths []string
	var userIDs []uint64
	t.Cleanup(func() {
		cleanupExec := func(query string, args ...any) {
			if _, err := db.ExecContext(ctx, query, args...); err != nil {
				t.Errorf("limpeza dos dados TESTE_ falhou: %v", err)
			}
		}
		cleanupExec(`DELETE FROM operation_logs WHERE details LIKE ?`, namePrefix+"%")
		for _, id := range userIDs {
			cleanupExec(`DELETE FROM operation_logs WHERE user_id = ?`, id)
		}
		rows, err := db.Query(`SELECT file_path FROM articles WHERE title LIKE ?`, namePrefix+"%")
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
		for _, path := range articlePaths {
			if err := os.Remove(filepath.FromSlash(path)); err != nil && !errors.Is(err, os.ErrNotExist) {
				t.Errorf("remoção do PDF TESTE_ falhou: %v", err)
			}
		}
		cleanupExec(`DELETE FROM articles WHERE title LIKE ?`, namePrefix+"%")
		cleanupExec(`DELETE FROM users WHERE email = ?`, userEmail)
		for _, id := range userIDs {
			cleanupExec(`DELETE FROM users WHERE id = ?`, id)
		}
		cleanupExec(`DELETE au FROM authors au LEFT JOIN article_authors aa ON aa.author_id = au.id WHERE aa.author_id IS NULL AND au.name LIKE ?`, namePrefix+"%")
		for _, check := range []struct {
			query string
			args  []any
		}{
			{query: `SELECT COUNT(*) FROM articles WHERE title LIKE ?`, args: []any{namePrefix + "%"}},
			{query: `SELECT COUNT(*) FROM users WHERE email = ?`, args: []any{userEmail}},
			{query: `SELECT COUNT(*) FROM authors WHERE name LIKE ?`, args: []any{namePrefix + "%"}},
			{query: `SELECT COUNT(*) FROM operation_logs WHERE details LIKE ?`, args: []any{namePrefix + "%"}},
		} {
			var count int
			if err := db.QueryRowContext(ctx, check.query, check.args...).Scan(&count); err != nil {
				t.Errorf("verificação da limpeza TESTE_ falhou: %v", err)
			} else if count != 0 {
				t.Errorf("limpeza deixou %d registros TESTE_ (%s)", count, check.query)
			}
		}
	})

	t.Run("sessao, cadastro, login e senha", func(t *testing.T) {
		if _, err := app.authenticatedUser(); err == nil {
			t.Fatal("sessão sem login deveria ser recusada")
		}
		if got := app.ListArticles(); len(got) != 0 {
			t.Fatalf("sessão sem login retornou artigos: %d", len(got))
		}
		if _, err := app.requireCurator(); err == nil {
			t.Fatal("sessão sem login deveria falhar na autorização")
		}

		response := app.Register(namePrefix, userEmail, "SenhaTeste09")
		if !response.Success || response.User == nil || response.User.Role != models.RolePesquisador {
			t.Fatalf("cadastro inesperado: %+v", response)
		}
		userIDs = append(userIDs, response.User.ID)
		app.Logout()
		if response := app.Login(userEmail, "errada"); response.Success {
			t.Fatal("login com senha incorreta foi aceito")
		}
		if response := app.Login(userEmail, "SenhaTeste09"); !response.Success {
			t.Fatalf("login válido falhou: %+v", response)
		}
		app.Logout()
		if app.GetCurrentUser() != nil {
			t.Fatal("logout manteve usuário autenticado")
		}
		if result := app.Register(namePrefix, strings.ToUpper(userEmail), "SenhaTeste09"); result.Success {
			t.Fatal("cadastro duplicado foi aceito")
		}
		if _, err := app.authService.Register(ctx, models.RegisterInput{Name: namePrefix, Email: "valid@example.com", Password: "curta"}); !errors.Is(err, services.ErrInvalidInput) {
			t.Fatalf("senha menor que 8 não retornou validação: %v", err)
		}
	})

	t.Run("email inválido no cadastro e criação de Admin", func(t *testing.T) {
		for _, email := range []string{"abc", "a@b", "teste@"} {
			if _, err := app.authService.Register(ctx, models.RegisterInput{Name: namePrefix, Email: email, Password: "SenhaTeste09"}); !errors.Is(err, services.ErrInvalidEmail) {
				t.Errorf("cadastro com %q: esperado erro de e-mail, obtido %v", email, err)
			}
			response := app.Register(namePrefix, email, "SenhaTeste09")
			if response.Success || !strings.Contains(strings.ToLower(response.Message), "e-mail válido") {
				t.Errorf("cadastro não apresentou mensagem clara para %q: %+v", email, response)
			}
		}
		response := app.CreateInitialAdmin(namePrefix, "teste@", "SenhaTeste09")
		if response.Success || !strings.Contains(strings.ToLower(response.Message), "e-mail válido") {
			t.Fatalf("criação de Admin não apresentou validação clara de e-mail: %+v", response)
		}
		if _, err := app.adminSetupService().CreateInitialAdmin(ctx, models.RegisterInput{Name: namePrefix, Email: "valid@example.com", Password: "curta"}); !errors.Is(err, services.ErrInvalidInput) {
			t.Fatalf("senha curta na criação de Admin não foi recusada: %v", err)
		}
	})

	var admin models.User
	if err := db.QueryRowContext(ctx, `SELECT id, name, email, password_hash, role FROM users WHERE role = 'ADMIN' ORDER BY id LIMIT 1`).Scan(&admin.ID, &admin.Name, &admin.Email, &admin.PasswordHash, &admin.Role); errors.Is(err, sql.ErrNoRows) {
		created := app.CreateInitialAdmin(namePrefix+" Admin", "admin+"+suffix+"@exemplo.com", "SenhaTeste09")
		if !created.Success || created.User == nil {
			t.Fatalf("não foi possível criar Admin temporário para teste: %+v", created)
		}
		admin = *created.User
		userIDs = append(userIDs, admin.ID)
	} else if err != nil {
		t.Fatalf("não foi possível localizar Admin para validar Gestão de Usuários: %v", err)
	}
	if app.NeedsAdminSetup() {
		t.Fatal("NeedsAdminSetup informou que falta Admin quando já existe um")
	}
	if response := app.CreateInitialAdmin(namePrefix+" Admin duplicado", "outro+"+suffix+"@exemplo.com", "SenhaTeste09"); response.Success || !strings.Contains(response.Message, "Já existe um administrador") {
		t.Fatalf("a API deveria recusar um segundo Admin: %+v", response)
	}
	if health := app.Health(); health.Status != "ok" || health.Database != "connected" {
		t.Fatalf("Health não confirmou a conexão MySQL: %+v", health)
	}

	t.Run("perfis e Gestão de Usuários", func(t *testing.T) {
		app.setCurrentUser(&models.User{ID: userIDs[0], Role: models.RolePesquisador})
		if len(app.ListUsers()) != 0 {
			t.Fatal("Pesquisador acessou a lista de usuários")
		}
		if response := app.UpdateUserRole(userIDs[0], models.RoleCurador); response.Success {
			t.Fatal("Pesquisador alterou perfil de usuário")
		}
		if response := app.DeleteUser(userIDs[0]); response.Success {
			t.Fatal("Pesquisador excluiu usuário")
		}
		app.setCurrentUser(&admin)
		if listed := app.ListUsers(); len(listed) == 0 {
			t.Fatal("Admin não conseguiu listar usuários")
		}
		if response := app.UpdateUserRole(userIDs[0], models.RoleCurador); !response.Success {
			t.Fatalf("Admin não conseguiu criar perfil Curador para usuário de teste: %+v", response)
		}
		curator, err := userRepo.FindByEmail(ctx, userEmail)
		if err != nil || curator.Role != models.RoleCurador {
			t.Fatalf("perfil Curador não persistido: user=%+v err=%v", curator, err)
		}
		app.setCurrentUser(curator)
		if len(app.ListUsers()) != 0 {
			t.Fatal("Curador acessou a lista de usuários")
		}
		if response := app.UpdateUserRole(userIDs[0], models.RolePesquisador); response.Success {
			t.Fatal("Curador alterou perfil de usuário")
		}
		if response := app.DeleteUser(userIDs[0]); response.Success {
			t.Fatal("Curador excluiu usuário")
		}
		if _, err := app.requireCurator(); err != nil {
			t.Fatalf("Curador foi recusado: %v", err)
		}
		app.setCurrentUser(&models.User{ID: userIDs[0], Role: models.RolePesquisador})
		if _, err := app.requireCurator(); err == nil {
			t.Fatal("Pesquisador recebeu autorização de Curador")
		}
		app.setCurrentUser(&admin)
		if response := app.UpdateUserRole(0, models.RoleAdmin); response.Success || !strings.Contains(response.Message, "Função inválida") {
			t.Fatalf("perfil inválido não foi recusado corretamente: %+v", response)
		}
		if response := app.UpdateUserRole(0, models.RoleCurador); response.Success || !strings.Contains(response.Message, "Usuário não encontrado") {
			t.Fatalf("alteração de usuário inexistente não foi recusada: %+v", response)
		}
		if response := app.DeleteUser(0); response.Success || !strings.Contains(response.Message, "Usuário não encontrado") {
			t.Fatalf("exclusão de usuário inexistente não foi recusada: %+v", response)
		}
		if response := app.UpdateUserRole(admin.ID, models.RoleCurador); response.Success || !strings.Contains(response.Message, "Administradores não podem ser alterados") {
			t.Fatalf("alteração de Admin protegido não foi recusada: %+v", response)
		}
		if response := app.DeleteUser(admin.ID); response.Success || !strings.Contains(response.Message, "Administradores não podem ser excluídos") {
			t.Fatalf("exclusão de Admin protegido não foi recusada: %+v", response)
		}
	})

	t.Run("CRUD, busca, exclusão do PDF", func(t *testing.T) {
		var sourceID uint64
		if err := db.QueryRowContext(ctx, `SELECT id FROM article_sources WHERE name = 'local' LIMIT 1`).Scan(&sourceID); err != nil {
			t.Fatalf("fonte local ausente: %v", err)
		}
		pdf := filepath.Join(t.TempDir(), "TESTE_artigo.pdf")
		if err := os.WriteFile(pdf, []byte("%PDF-1.4 TESTE_sprint09_concordia"), 0o600); err != nil {
			t.Fatal(err)
		}
		app.setCurrentUser(&models.User{ID: userIDs[0], Role: models.RoleCurador})
		if response := app.ImportArticle(namePrefix, "", sourceID, "", "", "", "", ""); response.Success || !strings.Contains(response.Message, "Selecione um arquivo PDF") {
			t.Fatalf("importação sem PDF deveria ser recusada: %+v", response)
		}
		invalidExtension := filepath.Join(t.TempDir(), "TESTE_artigo.txt")
		if err := os.WriteFile(invalidExtension, []byte("%PDF-1.4 TESTE conteúdo"), 0o600); err != nil {
			t.Fatal(err)
		}
		invalidHeader := filepath.Join(t.TempDir(), "TESTE_invalido.pdf")
		if err := os.WriteFile(invalidHeader, []byte("TESTE não é PDF"), 0o600); err != nil {
			t.Fatal(err)
		}
		for _, invalid := range []struct {
			name     string
			path     string
			title    string
			sourceID uint64
		}{
			{name: "título ausente", path: pdf, sourceID: sourceID},
			{name: "fonte ausente", path: pdf, title: namePrefix},
			{name: "extensão inválida", path: invalidExtension, title: namePrefix, sourceID: sourceID},
			{name: "cabeçalho PDF inválido", path: invalidHeader, title: namePrefix, sourceID: sourceID},
		} {
			app.mu.Lock()
			app.selectedPDF = invalid.path
			app.mu.Unlock()
			response := app.ImportArticle(invalid.title, "", invalid.sourceID, "", "", "", "", "")
			if response.Success || !strings.Contains(response.Message, "Título, fonte e PDF válido") {
				t.Errorf("%s deveria ser recusado com mensagem de validação: %+v", invalid.name, response)
			}
		}
		app.mu.Lock()
		app.selectedPDF = pdf
		app.mu.Unlock()
		invalid := services.ImportArticleInput{Title: namePrefix + " data inválida", SourceID: sourceID, OriginalPath: pdf, UserID: userIDs[0], PublicationDate: "31/02/2026"}
		if _, err := app.articleService.ImportPDF(ctx, invalid); !errors.Is(err, services.ErrInvalidPublicationDate) {
			t.Fatalf("importação aceitou data inválida: %v", err)
		}
		imported := app.ImportArticle(namePrefix+" artigo CRUD", "TESTE_description", sourceID, namePrefix+" Autor", "pt-BR", "2020-02-29", "", "")
		if !imported.Success {
			t.Fatalf("ImportArticle da API falhou: %+v", imported)
		}
		id := imported.ID
		if sources := app.ListSources(); len(sources) == 0 {
			t.Fatal("ListSources da API não retornou fontes")
		}
		listedArticles := app.ListArticles()
		foundArticle := false
		for _, item := range listedArticles {
			if item.ID == id {
				foundArticle = true
			}
		}
		if !foundArticle {
			t.Fatal("ListArticles da API não retornou artigo importado")
		}
		if results := app.SearchArticles(namePrefix); len(results) == 0 {
			t.Fatal("SearchArticles da API não encontrou artigo importado")
		}
		detailResponse := app.GetArticle(id)
		if !detailResponse.Success || detailResponse.Article == nil {
			t.Fatalf("GetArticle da API falhou: %+v", detailResponse)
		}
		article := detailResponse.Article
		articlePaths = append(articlePaths, article.FilePath)
		if stats := app.GetCollectionStats(); stats.TotalArticles == 0 {
			t.Fatalf("GetCollectionStats da API inesperado: %+v", stats)
		}
		if results, err := app.articleService.SearchArticles(ctx, namePrefix); err != nil || len(results) == 0 {
			t.Fatalf("busca pelo artigo de teste falhou: resultados=%d erro=%v", len(results), err)
		}
		if response := app.UpdateArticle(id, namePrefix+" artigo atualizado", "", sourceID, "", "", "2020-02-30", "", ""); response.Success || !strings.Contains(response.Message, "data de publicação inválida") {
			t.Fatalf("API aceitou data inválida na edição: %+v", response)
		}
		if response := app.UpdateArticle(id, namePrefix+" artigo atualizado", "", sourceID, "", "", "2021-01-01", "", ""); !response.Success {
			t.Fatalf("UpdateArticle da API falhou: %+v", response)
		}
		if response := app.DeleteArticle(id); !response.Success {
			t.Fatalf("DeleteArticle da API falhou: %+v", response)
		}
		if _, err := os.Stat(filepath.FromSlash(article.FilePath)); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("PDF permaneceu no disco após exclusão: %v", err)
		}
		if err := app.articleService.UpdateArticle(ctx, services.UpdateArticleInput{
			ID: id, Title: namePrefix + " artigo removido", SourceID: sourceID,
		}); !errors.Is(err, repositories.ErrArticleNotFound) {
			t.Fatalf("edição de artigo inexistente deveria retornar ErrArticleNotFound: %v", err)
		}
		if err := app.articleService.DeleteArticle(ctx, id); !errors.Is(err, repositories.ErrArticleNotFound) {
			t.Fatalf("exclusão repetida deveria retornar ErrArticleNotFound: %v", err)
		}
		if response := app.UpdateArticle(id, namePrefix+" removido", "", sourceID, "", "", "", "", ""); response.Success {
			t.Fatalf("API aceitou edição de artigo inexistente: %+v", response)
		}
		if response := app.DeleteArticle(id); response.Success {
			t.Fatalf("API aceitou exclusão de artigo inexistente: %+v", response)
		}
	})

	t.Run("falha ao remover PDF é reportada", func(t *testing.T) {
		folder := t.TempDir()
		if err := os.WriteFile(filepath.Join(folder, "keep.txt"), []byte("TESTE_"), 0o600); err != nil {
			t.Fatal(err)
		}
		result, err := db.ExecContext(ctx, `INSERT INTO articles (title, file_path, index_status) VALUES (?, ?, 'PENDING')`, namePrefix+" remoção falha", filepath.ToSlash(folder))
		if err != nil {
			t.Fatal(err)
		}
		id, _ := result.LastInsertId()
		if err := app.articleService.DeleteArticle(ctx, uint64(id)); err == nil {
			t.Fatal("falha de remoção do filesystem não foi reportada")
		}
	})

	t.Run("reconexão", func(t *testing.T) {
		reconnecting := &App{ctx: ctx}
		status := reconnecting.ReconnectDatabase()
		if status.Status != "ok" || status.Database != "connected" {
			t.Fatalf("reconexão inesperada: %+v", status)
		}
		if reconnecting.db != nil {
			_ = reconnecting.db.Close()
		}
	})

	t.Run("Admin exclui usuário de teste", func(t *testing.T) {
		app.setCurrentUser(&admin)
		response := app.DeleteUser(userIDs[0])
		if !response.Success {
			t.Fatalf("Admin não conseguiu excluir usuário de teste: %+v", response)
		}
		if _, err := userRepo.FindByID(ctx, userIDs[0]); !errors.Is(err, repositories.ErrUserNotFound) {
			t.Fatalf("usuário permaneceu após exclusão: %v", err)
		}
	})
}
