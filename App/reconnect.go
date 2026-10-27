package main

import (
	"fmt"
	"sync"

	"desktop/internal/config"
	"desktop/internal/repositories"
	"desktop/internal/services"

	"github.com/joho/godotenv"
)

var reconnectMu sync.Mutex

// tenta reconectar ao banco
func (a *App) ReconnectDatabase() HealthResponse {
	reconnectMu.Lock()
	defer reconnectMu.Unlock()

	// se já estiver conectado ele dá o retorno
	if a.db != nil {
		return a.Health()
	}

	// overload para aceitar um env mesmo aberto
	_ = godotenv.Overload()

	db, err := config.ConnectDatabase()
	if err != nil {
		fmt.Printf("Concordia: reconexão falhou: %v\n", err)
		return HealthResponse{Status: "error", Database: "disconnected"}
	}

	userRepo := repositories.NewUserRepository(db)
	articleRepo := repositories.NewArticleRepository(db)
	sourceRepo := repositories.NewSourceRepository(db)
	indexRepo := repositories.NewIndexRepository(db)
	logRepo := repositories.NewLogRepository(db)

	if err := sourceRepo.EnsureDefaults(a.ctx); err != nil {
		fmt.Printf("Concordia: aviso ao garantir fontes padrão: %v\n", err)
	}

	a.authService = services.NewAuthService(userRepo)
	a.articleService = services.NewArticleService(db, articleRepo, sourceRepo)
	a.indexingService = services.NewIndexingService(indexRepo)
	a.searchService = services.NewSearchService(indexRepo)
	a.logService = services.NewLogService(logRepo)
	a.db = db

	fmt.Println("Concordia: conexão com MySQL estabelecida.")
	return a.Health()
}