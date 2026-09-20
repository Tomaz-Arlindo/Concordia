package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"

	"desktop/internal/config"
	"desktop/internal/models"
	"desktop/internal/repositories"
	"desktop/internal/services"

	"github.com/joho/godotenv"
)

type App struct {
	ctx             context.Context
	db              *sql.DB
	authService     *services.AuthService
	articleService  *services.ArticleService
	indexingService *services.IndexingService
	searchService   *services.SearchService

	mu          sync.RWMutex
	currentUser *models.User
	selectedPDF string
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	_ = godotenv.Load()

	db, err := config.ConnectDatabase()
	if err != nil {
		fmt.Printf("Concordia: falha ao conectar ao banco: %v\n", err)
		return
	}

	a.db = db

	userRepo := repositories.NewUserRepository(db)
	articleRepo := repositories.NewArticleRepository(db)
	sourceRepo := repositories.NewSourceRepository(db)
	indexRepo := repositories.NewIndexRepository(db)

	if err := sourceRepo.EnsureDefaults(ctx); err != nil {
		fmt.Printf("Concordia: aviso ao garantir fontes padrão: %v\n", err)
	}

	a.authService = services.NewAuthService(userRepo)
	a.articleService = services.NewArticleService(db, articleRepo, sourceRepo)
	a.indexingService = services.NewIndexingService(indexRepo)
	a.searchService = services.NewSearchService(indexRepo)

	fmt.Println("Concordia: conexão com MySQL estabelecida.")
}

func (a *App) shutdown(ctx context.Context) {
	if a.db != nil {
		_ = a.db.Close()
	}
}

type HealthResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
}

func (a *App) Health() HealthResponse {
	if a.db == nil {
		return HealthResponse{Status: "error", Database: "disconnected"}
	}

	if err := a.db.PingContext(a.ctx); err != nil {
		return HealthResponse{Status: "error", Database: "disconnected"}
	}

	return HealthResponse{Status: "ok", Database: "connected"}
}

type AuthResponse struct {
	Success bool         `json:"success"`
	Message string       `json:"message"`
	User    *models.User `json:"user,omitempty"`
}

func (a *App) Register(name, email, password string) AuthResponse {
	if a.authService == nil {
		return AuthResponse{Success: false, Message: "Banco de dados não conectado."}
	}

	user, err := a.authService.Register(a.ctx, models.RegisterInput{
		Name:     name,
		Email:    email,
		Password: password,
	})
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidInput):
			return AuthResponse{Success: false, Message: "Preencha nome e email e use uma senha com pelo menos 8 caracteres."}
		case errors.Is(err, services.ErrEmailAlreadyExists):
			return AuthResponse{Success: false, Message: "Este email já está cadastrado."}
		default:
			return AuthResponse{Success: false, Message: "Não foi possível criar o usuário."}
		}
	}

	a.setCurrentUser(user)
	return AuthResponse{Success: true, Message: "Cadastro realizado com sucesso.", User: user}
}

func (a *App) Login(email, password string) AuthResponse {
	if a.authService == nil {
		return AuthResponse{Success: false, Message: "Banco de dados não conectado."}
	}

	user, err := a.authService.Login(a.ctx, strings.TrimSpace(email), password)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			return AuthResponse{Success: false, Message: "Email ou senha inválidos."}
		}
		return AuthResponse{Success: false, Message: "Não foi possível realizar o login."}
	}

	a.setCurrentUser(user)
	return AuthResponse{Success: true, Message: "Login realizado com sucesso.", User: user}
}

func (a *App) Logout() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.currentUser = nil
	a.selectedPDF = ""
}

func (a *App) GetCurrentUser() *models.User {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.currentUser == nil {
		return nil
	}

	copyUser := *a.currentUser
	return &copyUser
}

func (a *App) setCurrentUser(user *models.User) {
	a.mu.Lock()
	defer a.mu.Unlock()
	copyUser := *user
	a.currentUser = &copyUser
}

func (a *App) authenticatedUser() (*models.User, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.currentUser == nil {
		return nil, errors.New("usuário não autenticado")
	}

	copyUser := *a.currentUser
	return &copyUser, nil
}

func (a *App) requireCurator() (*models.User, error) {
	user, err := a.authenticatedUser()
	if err != nil {
		return nil, err
	}
	if user.Role != models.RoleCurador {
		return nil, errors.New("ação permitida somente para CURADOR")
	}
	return user, nil
}
