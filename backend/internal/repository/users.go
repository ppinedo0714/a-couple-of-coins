package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ppinedo/a-couple-of-coins/backend/internal/models"
)

var ErrNotFound = errors.New("not found")

type UserRepository interface {
	Create(ctx context.Context, email, passwordHash string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByOAuthProvider(ctx context.Context, provider, providerUserID string) (*models.User, error)
	Update(ctx context.Context, id uuid.UUID, email string) (*models.User, error)
	CreateOAuthConnection(ctx context.Context, userID uuid.UUID, provider, providerUserID string) error
}

type pgxUserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &pgxUserRepository{pool: pool}
}

func (r *pgxUserRepository) Create(ctx context.Context, email, passwordHash string) (*models.User, error) {
	var u models.User
	err := r.pool.QueryRow(ctx,
		`INSERT INTO users (email, password_hash)
		 VALUES (LOWER($1), $2)
		 RETURNING id, email, password_hash, created_at`,
		email, passwordHash,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *pgxUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var u models.User
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, password_hash, created_at
		 FROM users
		 WHERE LOWER(email) = LOWER($1)`,
		email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *pgxUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var u models.User
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, password_hash, created_at
		 FROM users
		 WHERE id = $1`,
		id,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *pgxUserRepository) GetByOAuthProvider(ctx context.Context, provider, providerUserID string) (*models.User, error) {
	var u models.User
	err := r.pool.QueryRow(ctx,
		`SELECT u.id, u.email, u.password_hash, u.created_at
		 FROM users u
		 JOIN oauth_connections oc ON oc.user_id = u.id
		 WHERE oc.provider = $1 AND oc.provider_user_id = $2`,
		provider, providerUserID,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *pgxUserRepository) Update(ctx context.Context, id uuid.UUID, email string) (*models.User, error) {
	var u models.User
	err := r.pool.QueryRow(ctx,
		`UPDATE users
		 SET email = LOWER($2)
		 WHERE id = $1
		 RETURNING id, email, password_hash, created_at`,
		id, email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *pgxUserRepository) CreateOAuthConnection(ctx context.Context, userID uuid.UUID, provider, providerUserID string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO oauth_connections (user_id, provider, provider_user_id)
		 VALUES ($1, $2, $3)`,
		userID, provider, providerUserID,
	)
	return err
}
