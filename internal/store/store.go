package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"telegram-bot/internal/entity"
	"telegram-bot/internal/store/converter"
	"telegram-bot/internal/store/model"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type store struct {
	db *pgxpool.Pool
}

func New(dsn string) (*store, error) {
	const op = "store.New"

	db, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &store{db: db}, nil
}

func (s *store) Close() {
	s.db.Close()
}

// Server returns server data.
func (s *store) Server(ctx context.Context) ([]entity.Server, error) {
	const op = "store.Server"

	data, err := s.receiveServerData(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	servers := make([]entity.Server, 0, len(data))
	for _, d := range data {
		server, err := converter.ToServerFromRepo(d)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		servers = append(servers, server)
	}

	return servers, nil
}

// AddURL add access url in database.
func (s *store) AddURL(ctx context.Context, url entity.AccessURL) error {
	const op = "store.AddURL"

	data, err := converter.ToRepoFromAccessURL(url)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	createdAt := time.Now()
	expiredAt := createdAt.Add(time.Hour * 48)

	_, err = s.db.Query(ctx, "INSERT INTO access_url (id, access_key, api_url, created_at, expired_at) VALUES ($1, $2, $3, $4, $5)",
		data.ID, data.AccessKey, data.ApiURL, createdAt, expiredAt,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// ExpiredURLs recieve expired urls data.
func (s *store) ExpiredURLs(ctx context.Context) ([]entity.AccessURL, error) {
	const op = "store.ExpiredURLs"

	data, err := s.receiveAccessURLData(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	accessURLs := make([]entity.AccessURL, 0, len(data))
	for _, d := range data {
		accessURL := converter.ToAccessURLFromRepo(d)

		accessURLs = append(accessURLs, accessURL)
	}

	return accessURLs, nil
}

// DeleteURL deletes url from database.
func (s *store) DeleteURL(ctx context.Context, id string) error {
	const op = "store.DeleteURL"

	_, err := s.db.Query(ctx, "DELETE FROM access_url WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// LastURLSentTime returns the creation time of the most recently added access URL.
func (s *store) LastURLSentTime(ctx context.Context) (time.Time, error) {
	const op = "store.LastURLSentTime"

	rows, err := s.db.Query(ctx, "SELECT created_at FROM access_url ORDER BY created_at DESC LIMIT 1")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return time.Time{}, nil
		}

		return time.Time{}, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var lastTime time.Time
	for rows.Next() {
		if err := rows.Scan(&lastTime); err != nil {
			return time.Time{}, fmt.Errorf("%s: %w", op, err)
		}
	}

	return lastTime, nil
}

// receiveAccessURLData select expired urls data from database.
func (s *store) receiveAccessURLData(ctx context.Context) ([]model.AccessURL, error) {
	const op = "store.receiveAccessURLData"

	current := time.Now()

	rows, err := s.db.Query(ctx, "SELECT id, access_key, api_url FROM access_url WHERE $1 > expired_at", current)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var accessURLs []model.AccessURL
	for rows.Next() {
		var accessURL model.AccessURL
		if err := rows.Scan(&accessURL.ID, &accessURL.AccessKey, &accessURL.ApiURL); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		accessURLs = append(accessURLs, accessURL)
	}

	return accessURLs, nil
}

// receiveServerData select server data from database.
func (s *store) receiveServerData(ctx context.Context) ([]model.Server, error) {
	const op = "store.receiveServerData"

	rows, err := s.db.Query(ctx, "SELECT ip_address, port, key FROM servers WHERE is_active = $1 AND is_test = $2", true, true)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var servers []model.Server
	for rows.Next() {
		var server model.Server
		if err := rows.Scan(&server.IPAddr, &server.Port, &server.Key); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		servers = append(servers, server)
	}

	return servers, nil
}
