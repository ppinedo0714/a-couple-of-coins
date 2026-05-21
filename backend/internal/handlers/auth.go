package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/auth"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/models"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/repository"
	"golang.org/x/oauth2"
)

type AuthHandler struct {
	repo        repository.UserRepository
	jwtSecret   string
	frontendURL string
	googleCfg   *oauth2.Config
	githubCfg   *oauth2.Config
}

func NewAuthHandler(
	repo repository.UserRepository,
	jwtSecret string,
	frontendURL string,
	googleCfg *oauth2.Config,
	githubCfg *oauth2.Config,
) *AuthHandler {
	return &AuthHandler{
		repo:        repo,
		jwtSecret:   jwtSecret,
		frontendURL: frontendURL,
		googleCfg:   googleCfg,
		githubCfg:   githubCfg,
	}
}

func setAuthCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   int(auth.TokenTTL.Seconds()),
	})
}

func clearAuthCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   0,
	})
}

func isDuplicateEmail(err error) bool {
	var pgErr *pgconn.PgError
	// 23505 = unique_violation
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return true
	}
	return false
}

func userResponse(u *models.User) map[string]interface{} {
	return map[string]interface{}{
		"id":         u.ID,
		"email":      u.Email,
		"created_at": u.CreatedAt,
	}
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" || !strings.Contains(req.Email, "@") {
		writeError(w, http.StatusBadRequest, "valid email is required")
		return
	}
	if len(req.Password) < 8 {
		writeError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}

	hash, err := auth.Hash(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	user, err := h.repo.Create(r.Context(), req.Email, hash)
	if err != nil {
		if isDuplicateEmail(err) {
			writeError(w, http.StatusConflict, "email already registered")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	token, err := auth.Issue(h.jwtSecret, user.ID, auth.TokenTTL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	setAuthCookie(w, token)
	writeJSON(w, http.StatusCreated, map[string]interface{}{"user": userResponse(user)})
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	user, err := h.repo.GetByEmail(r.Context(), req.Email)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if user.PasswordHash == nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	if err := auth.Compare(*user.PasswordHash, req.Password); err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, err := auth.Issue(h.jwtSecret, user.ID, auth.TokenTTL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	setAuthCookie(w, token)
	writeJSON(w, http.StatusOK, map[string]interface{}{"user": userResponse(user)})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	clearAuthCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (h *AuthHandler) GoogleOAuth(w http.ResponseWriter, r *http.Request) {
	state, err := generateState()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   int((10 * time.Minute).Seconds()),
	})

	url := h.googleCfg.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusFound)
}

func (h *AuthHandler) GoogleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	h.handleOAuthCallback(w, r, "google", h.googleCfg, func(ctx context.Context, cfg *oauth2.Config, token *oauth2.Token) (string, string, error) {
		return auth.FetchGoogleProfile(ctx, cfg, token)
	})
}

func (h *AuthHandler) GitHubOAuth(w http.ResponseWriter, r *http.Request) {
	state, err := generateState()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   int((10 * time.Minute).Seconds()),
	})

	url := h.githubCfg.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusFound)
}

func (h *AuthHandler) GitHubOAuthCallback(w http.ResponseWriter, r *http.Request) {
	h.handleOAuthCallback(w, r, "github", h.githubCfg, func(ctx context.Context, cfg *oauth2.Config, token *oauth2.Token) (string, string, error) {
		return auth.FetchGitHubProfile(ctx, cfg, token)
	})
}

type profileFetchFn func(ctx context.Context, cfg *oauth2.Config, token *oauth2.Token) (email, providerUserID string, err error)

func (h *AuthHandler) handleOAuthCallback(w http.ResponseWriter, r *http.Request, provider string, cfg *oauth2.Config, fetchProfile profileFetchFn) {
	// Validate CSRF state
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || r.URL.Query().Get("state") != stateCookie.Value {
		writeError(w, http.StatusBadRequest, "invalid oauth state")
		return
	}

	// Clear state cookie
	http.SetCookie(w, &http.Cookie{
		Name:   "oauth_state",
		Value:  "",
		MaxAge: -1,
		Path:   "/",
	})

	code := r.URL.Query().Get("code")
	if code == "" {
		writeError(w, http.StatusBadRequest, "missing oauth code")
		return
	}

	token, err := cfg.Exchange(r.Context(), code)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "oauth token exchange failed")
		return
	}

	email, providerUserID, err := fetchProfile(r.Context(), cfg, token)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch oauth profile")
		return
	}

	// Find existing user by OAuth provider
	user, err := h.repo.GetByOAuthProvider(r.Context(), provider, providerUserID)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if errors.Is(err, repository.ErrNotFound) {
		// Check if a user with this email already exists (link accounts)
		existingUser, emailErr := h.repo.GetByEmail(r.Context(), email)
		if emailErr != nil && !errors.Is(emailErr, repository.ErrNotFound) {
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		if existingUser != nil {
			// Link OAuth to existing account
			if connErr := h.repo.CreateOAuthConnection(r.Context(), existingUser.ID, provider, providerUserID); connErr != nil {
				writeError(w, http.StatusInternalServerError, "internal server error")
				return
			}
			user = existingUser
		} else {
			// Create new user (no password for OAuth-only users, pass empty string)
			user, err = h.repo.Create(r.Context(), email, "")
			if err != nil {
				writeError(w, http.StatusInternalServerError, "internal server error")
				return
			}
			if connErr := h.repo.CreateOAuthConnection(r.Context(), user.ID, provider, providerUserID); connErr != nil {
				writeError(w, http.StatusInternalServerError, "internal server error")
				return
			}
		}
	}

	jwtToken, err := auth.Issue(h.jwtSecret, user.ID, auth.TokenTTL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	setAuthCookie(w, jwtToken)
	http.Redirect(w, r, h.frontendURL+"/login?oauth=success", http.StatusFound)
}
