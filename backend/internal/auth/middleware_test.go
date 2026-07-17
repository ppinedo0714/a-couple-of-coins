package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func TestMiddlewareNoCookie(t *testing.T) {
	r := chi.NewRouter()
	r.Use(Middleware(testSecret))
	r.Get("/protected", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestMiddlewareInvalidJWT(t *testing.T) {
	r := chi.NewRouter()
	r.Use(Middleware(testSecret))
	r.Get("/protected", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: "not.a.valid.jwt"})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestMiddlewareValidJWT(t *testing.T) {
	userID := uuid.New()
	token, err := Issue(testSecret, userID, time.Hour)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	var gotUserID uuid.UUID
	r := chi.NewRouter()
	r.Use(Middleware(testSecret))
	r.Get("/protected", func(w http.ResponseWriter, r *http.Request) {
		id, ok := UserIDFromContext(r.Context())
		if !ok {
			t.Error("UserIDFromContext() returned false")
		}
		gotUserID = id
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if gotUserID != userID {
		t.Errorf("userID = %v, want %v", gotUserID, userID)
	}
}
