package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/auth"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/config"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/db"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/handlers"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/repository"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/services"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/services/predictor"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	pool, err := db.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	userRepo := repository.NewUserRepository(pool)
	accountRepo := repository.NewAccountRepository(pool)
	categoryRepo := repository.NewCategoryRepository(pool)
	transactionRepo := repository.NewTransactionRepository(pool)

	accountSvc := services.NewAccountService(accountRepo)
	categorySvc := services.NewCategoryService(categoryRepo)

	predictorClient := predictor.NewHTTPPredictorClient(cfg.PredictionsServiceURL)
	transactionSvc := services.NewTransactionService(pool, transactionRepo, accountRepo, categoryRepo, predictorClient)

	googleCfg := auth.GoogleConfig(
		cfg.GoogleClientID,
		cfg.GoogleClientSecret,
		cfg.FrontendURL+"/api/v1/auth/oauth/google/callback",
	)
	githubCfg := auth.GitHubConfig(
		cfg.GitHubClientID,
		cfg.GitHubClientSecret,
		cfg.FrontendURL+"/api/v1/auth/oauth/github/callback",
	)

	authHandler := handlers.NewAuthHandler(userRepo, cfg.JWTSecret, cfg.FrontendURL, googleCfg, githubCfg)
	usersHandler := handlers.NewUsersHandler(userRepo, cfg.JWTSecret)
	accountsHandler := handlers.NewAccountsHandler(accountSvc)
	categoriesHandler := handlers.NewCategoriesHandler(categorySvc)
	transactionsHandler := handlers.NewTransactionsHandler(transactionSvc)

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(corsMiddleware(cfg.FrontendURL))

	r.Get("/health", handlers.Health)

	r.Route("/api/v1", func(r chi.Router) {
		// Auth routes (no auth required)
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
			r.Post("/logout", authHandler.Logout)
			r.Get("/oauth/google", authHandler.GoogleOAuth)
			r.Get("/oauth/google/callback", authHandler.GoogleOAuthCallback)
			r.Get("/oauth/github", authHandler.GitHubOAuth)
			r.Get("/oauth/github/callback", authHandler.GitHubOAuthCallback)
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(auth.Middleware(cfg.JWTSecret))

			r.Route("/users", func(r chi.Router) {
				r.Get("/me", usersHandler.GetMe)
				r.Put("/me", usersHandler.UpdateMe)
			})

			r.Route("/accounts", func(r chi.Router) {
				// /history must be registered before /{id} so chi doesn't treat "history" as an ID
				r.Get("/history", accountsHandler.History)
				r.Get("/", accountsHandler.List)
				r.Post("/", accountsHandler.Create)
				r.Get("/{id}", accountsHandler.Get)
				r.Put("/{id}", accountsHandler.Update)
				r.Delete("/{id}", accountsHandler.Delete)
			})

			r.Route("/categories", func(r chi.Router) {
				r.Get("/", categoriesHandler.List)
				r.Post("/", categoriesHandler.Create)
				r.Put("/{id}", categoriesHandler.Update)
				r.Delete("/{id}", categoriesHandler.Delete)
			})

			r.Route("/transactions", func(r chi.Router) {
				// classify must be registered before /{id} so chi doesn't treat "classify" as an ID
				r.Post("/classify", transactionsHandler.Classify)
				r.Get("/", transactionsHandler.List)
				r.Post("/", transactionsHandler.Create)
				r.Get("/{id}", transactionsHandler.Get)
				r.Put("/{id}", transactionsHandler.Update)
				r.Delete("/{id}", transactionsHandler.Delete)
			})

			r.Mount("/import", chi.NewRouter())
		})
	})

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("starting server on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func corsMiddleware(frontendURL string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", frontendURL)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
