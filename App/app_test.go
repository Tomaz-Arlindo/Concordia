package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"desktop/internal/models"
	"desktop/internal/services"
)

func TestValidEmail(t *testing.T) {
	for _, email := range []string{"abc", "a@b", "teste@", "@dominio.com", "a b@dominio.com"} {
		if services.ValidEmail(email) {
			t.Errorf("email inválido aceito: %q", email)
		}
	}
	for _, email := range []string{"teste@exemplo.com", "nome.sobrenome+tag@sub.dominio.org"} {
		if !services.ValidEmail(email) {
			t.Errorf("email válido recusado: %q", email)
		}
	}
}

func TestExportCopyFileCopiesAndRejectsSamePath(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "origem.pdf")
	destination := filepath.Join(dir, "destino.pdf")
	content := []byte("%PDF-1.4 TESTE_")
	if err := os.WriteFile(source, content, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := exportCopyFile(source, destination); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(destination)
	if err != nil || string(got) != string(content) {
		t.Fatalf("conteúdo exportado divergente: %q, erro=%v", got, err)
	}
	if err := exportCopyFile(source, source); err == nil || !strings.Contains(err.Error(), "mesmo arquivo") {
		t.Fatalf("esperava rejeição da mesma origem/destino, recebido: %v", err)
	}
}

func TestCuratorOperationsEnforceRolePermissions(t *testing.T) {
	type operation struct {
		name string
		call func(*App) string
	}
	operations := []operation{
		{name: "ImportPDFBatch", call: func(a *App) string { return a.ImportPDFBatch(1, "pt-BR").Message }},
		{name: "ImportArticle", call: func(a *App) string { return a.ImportArticle("TESTE artigo", "", 1, "", "", "", "", "").Message }},
		{name: "UpdateArticle", call: func(a *App) string { return a.UpdateArticle(1, "TESTE artigo", "", 1, "", "", "", "", "").Message }},
		{name: "DeleteArticle", call: func(a *App) string { return a.DeleteArticle(1).Message }},
		{name: "StartIndexing", call: func(a *App) string { return a.StartIndexing(1, false).Message }},
		{name: "RunIndexingBenchmark", call: func(a *App) string { return a.RunIndexingBenchmark().Message }},
		{name: "ExportLogsCSV", call: func(a *App) string { return a.ExportLogsCSV("", "", "", 10).Message }},
		{name: "SelectPDF", call: func(a *App) string { return a.SelectPDF().Message }},
	}
	roles := []struct {
		name       string
		role       string
		authorized bool
	}{
		{name: "anônimo"},
		{name: "Pesquisador", role: models.RolePesquisador},
		{name: "Curador", role: models.RoleCurador, authorized: true},
		{name: "Admin", role: models.RoleAdmin, authorized: true},
	}

	for _, op := range operations {
		for _, role := range roles {
			t.Run(op.name+"/"+role.name, func(t *testing.T) {
				app := &App{}
				if role.role != "" {
					app.setCurrentUser(&models.User{Role: role.role})
				}
				if op.name == "SelectPDF" && role.authorized {
					if _, err := app.requireCurator(); err != nil {
						t.Fatalf("%s deveria passar a autorização de Curador: %v", role.name, err)
					}
					return
				}

				message := op.call(app)
				if role.authorized {
					if !strings.Contains(strings.ToLower(message), "indisponível") {
						t.Fatalf("operação não avançou após autorizar %s: %q", role.name, message)
					}
					return
				}
				if !strings.Contains(strings.ToLower(message), "somente") {
					t.Fatalf("operação deveria negar acesso a %s: %q", role.name, message)
				}
			})
		}
	}
}

func TestUserManagementOperationsRequireAdmin(t *testing.T) {
	roles := []struct {
		name       string
		role       string
		authorized bool
	}{
		{name: "anônimo"},
		{name: "Pesquisador", role: models.RolePesquisador},
		{name: "Curador", role: models.RoleCurador},
		{name: "Admin", role: models.RoleAdmin, authorized: true},
	}

	for _, role := range roles {
		for _, op := range []struct {
			name string
			call func(*App) string
		}{
			{name: "UpdateUserRole", call: func(a *App) string { return a.UpdateUserRole(1, models.RoleCurador).Message }},
			{name: "DeleteUser", call: func(a *App) string { return a.DeleteUser(1).Message }},
		} {
			t.Run(op.name+"/"+role.name, func(t *testing.T) {
				app := &App{}
				if role.role != "" {
					app.setCurrentUser(&models.User{Role: role.role})
				}
				message := strings.ToLower(op.call(app))
				if role.authorized {
					if !strings.Contains(message, "banco de dados não conectado") {
						t.Fatalf("Admin deveria avançar até a verificação do banco: %q", message)
					}
					return
				}
				if !strings.Contains(message, "somente o administrador") {
					t.Fatalf("operação deveria negar acesso a %s: %q", role.name, message)
				}
			})
		}
	}
}

func TestUnauthenticatedReadOperationsDoNotExposeData(t *testing.T) {
	app := &App{}
	if got := app.ListSources(); len(got) != 0 {
		t.Fatalf("ListSources retornou dados sem autenticação: %d", len(got))
	}
	if got := app.ListArticles(); len(got) != 0 {
		t.Fatalf("ListArticles retornou dados sem autenticação: %d", len(got))
	}
	if got := app.SearchArticles("TESTE_"); len(got) != 0 {
		t.Fatalf("SearchArticles retornou dados sem autenticação: %d", len(got))
	}
	if response := app.GetArticle(1); response.Success || !strings.Contains(response.Message, "não autenticado") {
		t.Fatalf("GetArticle sem sessão deveria negar acesso: %+v", response)
	}
	if got := app.GetCollectionStats(); got.TotalArticles != 0 {
		t.Fatalf("GetCollectionStats retornou dados sem autenticação: %+v", got)
	}
	if got := app.GetIndexingOverview(); got.Documents != 0 {
		t.Fatalf("GetIndexingOverview retornou dados sem autenticação: %+v", got)
	}
	if got := app.SearchBM25("TESTE_", 10); len(got) != 0 {
		t.Fatalf("SearchBM25 retornou dados sem autenticação: %d", len(got))
	}
	if got := app.SearchBM25Filtered("TESTE_", "", "", "", "", "", true, 10); len(got) != 0 {
		t.Fatalf("SearchBM25Filtered retornou dados sem autenticação: %d", len(got))
	}
	if got := app.ListOperationLogs("", "", "", 10); len(got) != 0 {
		t.Fatalf("ListOperationLogs retornou dados sem autenticação: %d", len(got))
	}
	if response := app.GetLogSummary(); response.Success || !strings.Contains(response.Message, "não autenticado") {
		t.Fatalf("GetLogSummary sem sessão deveria negar acesso: %+v", response)
	}
	if response := app.ExportArticlePDF(1); response.Success || !strings.Contains(response.Message, "não autenticado") {
		t.Fatalf("ExportArticlePDF sem sessão deveria negar acesso: %+v", response)
	}
}
