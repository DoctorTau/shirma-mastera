package users

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"monster-screen/backend/internal/apiutil"
	"monster-screen/backend/internal/auth"
)

type repo struct{ pool *pgxpool.Pool }

func RegisterRoutes(r chi.Router, pool *pgxpool.Pool, jwtSecret string) {
	h := &handler{repo: &repo{pool: pool}, jwtSecret: jwtSecret}
	r.Post("/api/register", h.register)
	r.Post("/api/login", h.login)
}

func (rp *repo) create(ctx context.Context, email, passwordHash string) (uuid.UUID, error) {
	var id uuid.UUID
	err := rp.pool.QueryRow(ctx,
		`INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id`,
		email, passwordHash).Scan(&id)
	return id, err
}

func (rp *repo) getByEmail(ctx context.Context, email string) (uuid.UUID, string, error) {
	var id uuid.UUID
	var hash string
	err := rp.pool.QueryRow(ctx,
		`SELECT id, password_hash FROM users WHERE email = $1`, email).Scan(&id, &hash)
	return id, hash, err
}

type handler struct {
	repo      *repo
	jwtSecret string
}

type credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func (h *handler) register(w http.ResponseWriter, r *http.Request) {
	var in credentials
	if err := apiutil.DecodeJSON(r, &in); err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	email := normalizeEmail(in.Email)
	if email == "" || !strings.Contains(email, "@") {
		apiutil.WriteError(w, http.StatusBadRequest, "invalid email")
		return
	}
	if len(in.Password) < 8 {
		apiutil.WriteError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}

	hash, err := auth.HashPassword(in.Password)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "could not process password")
		return
	}

	id, err := h.repo.create(r.Context(), email, hash)
	if err != nil {
		if isUniqueViolation(err) {
			apiutil.WriteError(w, http.StatusConflict, "email already registered")
			return
		}
		apiutil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	token, err := auth.GenerateToken(id, h.jwtSecret)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "could not issue token")
		return
	}
	apiutil.WriteJSON(w, http.StatusCreated, map[string]string{"token": token})
}

func (h *handler) login(w http.ResponseWriter, r *http.Request) {
	var in credentials
	if err := apiutil.DecodeJSON(r, &in); err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}

	id, hash, err := h.repo.getByEmail(r.Context(), normalizeEmail(in.Email))
	if err != nil || !auth.CheckPassword(hash, in.Password) {
		apiutil.WriteError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	token, err := auth.GenerateToken(id, h.jwtSecret)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "could not issue token")
		return
	}
	apiutil.WriteJSON(w, http.StatusOK, map[string]string{"token": token})
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
