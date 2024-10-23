package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"telegram-bot/internal/domain/models"
	"telegram-bot/internal/repository"
	"time"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(dsn string) (*Repository, error) {
	const op = "repository.New"

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Repository{pool: pool}, nil
}

func (r *Repository) Close() {
	r.pool.Close()
}

func (r *Repository) Server(ctx context.Context) (models.Server, error) {
	const op = "repository.Server"

	row := r.pool.QueryRow(ctx, "SELECT * FROM servers WHERE is_active = $1 ORDER BY RANDOM() LIMIT 1", true)

	var server models.Server
	err := row.Scan(&server.UUID, &server.IpAddress, &server.Port, &server.Key, &server.IsActive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Server{}, fmt.Errorf("%s: %w", op, repository.ErrServerNotFound)
		}

		return models.Server{}, fmt.Errorf("%s: %w", op, err)
	}

	return server, nil
}

func (r *Repository) SaveAccessKey(ctx context.Context, accessKey models.AccessKey) error {
	const op = "repository.SaveAccessKey"

	createdAt := time.Now()
	expiredAt := createdAt.Add(time.Hour * 48)

	query := "INSERT INTO access_keys (uuid, key, api_url, created_at, expired_at) VALUES ($1, $2, $3, $4, $5)"

	_ = r.pool.QueryRow(ctx, query, uuid.New(), accessKey.Key, accessKey.ApiURL, createdAt, expiredAt)

	return nil
}

func (r *Repository) DeleteAccessKeys(ctx context.Context, uuids []uuid.UUID) error {
	const op = "repository.DeleteAccessKeys"

	if len(uuids) == 0 {
		return nil
	}

	_ = r.pool.QueryRow(ctx, "DELETE FROM access_keys WHERE uuid = ANY($1::uuid[])", uuids)

	return nil
}

func (r *Repository) AccessKeyTime(ctx context.Context) (time.Time, error) {
	const op = "repository.AccessKeyTime"

	row := r.pool.QueryRow(ctx, "SELECT created_at FROM access_keys ORDER BY created_at DESC LIMIT 1")

	var lastTime time.Time
	if err := row.Scan(&lastTime); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return time.Time{}, fmt.Errorf("%s: %w", op, repository.AccessKeysNotFound)
		}

		return time.Time{}, fmt.Errorf("%s: %w", op, err)
	}

	return lastTime, nil
}

func (r *Repository) ExpiredAccessKeys(ctx context.Context) ([]models.AccessKey, error) {
	const op = "repository.ExpiredAccessKeys"

	query := "SELECT * FROM access_keys WHERE $1 > expired_at"

	current := time.Now()
	rows, err := r.pool.Query(ctx, query, current)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, repository.AccessKeysNotFound)
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()
	
	var accessKeys []models.AccessKey
	for rows.Next() {
		var accessKey models.AccessKey
		err := rows.Scan(&accessKey.UUID, &accessKey.Key, &accessKey.ApiURL, &accessKey.CreatedAt, &accessKey.ExpiredAt)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		accessKeys = append(accessKeys, accessKey)
	}

	return accessKeys, nil
}
