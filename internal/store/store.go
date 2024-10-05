package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"telegram-bot/internal/entity"
	"telegram-bot/internal/store/converter"
	"telegram-bot/internal/store/model"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	db *pgxpool.Pool
}

func New(dsn string) (*Store, error) {
	const op = "store.New"

	db, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Printf("%s: %s", op, err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Printf("%s: database connection established", op)
	return &Store{db: db}, nil
}

func (s *Store) Close() {
	s.db.Close()
	log.Println("store.Close: database connection closed")
}

// recoverPanic logs and recovers from panic
func recoverPanic(op string) {
	if r := recover(); r != nil {
		log.Printf("%s: recovered from panic: %v", op, r)
	}
}

// Server returns server data.
func (s *Store) Server(ctx context.Context) ([]entity.Server, error) {
	const op = "store.Server"
	defer recoverPanic(op)

	log.Println("Fetching server data")
	data, err := s.receiveServerData(ctx)
	if err != nil {
		log.Printf("%s: %s", op, err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	servers := make([]entity.Server, 0, len(data))
	for _, d := range data {
		server, err := converter.ToServerFromRepo(d)
		if err != nil {
			log.Printf("%s: %s", op, err)
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		servers = append(servers, server)
	}

	log.Println("Fetched server data successfully")
	return servers, nil
}

// AddURL adds access URL to database.
func (s *Store) AddURL(ctx context.Context, url entity.AccessURL) error {
	const op = "store.AddURL"
	defer recoverPanic(op)

	log.Println("Adding new access URL")
	data, err := converter.ToRepoFromAccessURL(url)
	if err != nil {
		log.Println("Error converting URL data:", err)
		log.Printf("%s: %s", op, err)
		return fmt.Errorf("%s: %w", op, err)
	}

	createdAt := time.Now()
	expiredAt := createdAt.Add(time.Hour * 48)

	log.Println("Created 'expiredAt' and 'createdAt' variables:", createdAt, expiredAt)
	log.Println("Executing insert query")

	query := `
		INSERT INTO access_url (id, access_key, api_url, created_at, expired_at) 
		VALUES ($1, $2, $3, $4, $5)
	`

	log.Printf("Executing query: %s with values: %v, %v, %v, %v", query, data.ID, data.AccessKey, createdAt, expiredAt)

	_, err = s.db.Exec(ctx, query, data.ID, data.AccessKey, data.ApiURL, createdAt, expiredAt)
	if err != nil {
		log.Printf("%s: Error executing query: %s", op, err)
		log.Println(err)
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Println("Added new access URL successfully")
	return nil
}

// ExpiredURLs receives expired URLs data.
func (s *Store) ExpiredURLs(ctx context.Context) ([]entity.AccessURL, error) {
	const op = "store.ExpiredURLs"
	defer recoverPanic(op)

	log.Println("Fetching expired URLs")
	data, err := s.receiveAccessURLData(ctx)
	if err != nil {
		log.Printf("%s: %s", op, err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	accessURLs := make([]entity.AccessURL, 0, len(data))
	for _, d := range data {
		accessURL := converter.ToAccessURLFromRepo(d)
		accessURLs = append(accessURLs, accessURL)
	}

	log.Println("Fetched expired URLs successfully")
	return accessURLs, nil
}

// DeleteURL deletes URLs from the database.
func (s *Store) DeleteURL(ctx context.Context, ids []string) error {
	const op = "store.DeleteURL"
	defer recoverPanic(op)

	log.Printf("Deleting URLs with IDs: %v", ids)

	if len(ids) == 0 {
		log.Println("No IDs provided to delete")
		return nil
	}

	// Convert string IDs to integer IDs
	intIDs := make([]int, len(ids))
	for i, idStr := range ids {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			log.Printf("%s: invalid ID %s: %v", op, idStr, err)
			return fmt.Errorf("%s: invalid ID %s: %w", op, idStr, err)
		}
		intIDs[i] = id
	}

	// Create the query with the correct number of placeholders
	query := "DELETE FROM access_url WHERE id = ANY($1::int[])"

	_, err := s.db.Exec(ctx, query, intIDs)
	if err != nil {
		log.Printf("%s: %s", op, err)
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Println("Deleted URLs successfully")
	return nil
}

// LastURLSentTime returns the creation time of the most recently added access URL.
func (s *Store) LastURLSentTime(ctx context.Context) (time.Time, error) {
	const op = "store.LastURLSentTime"
	defer recoverPanic(op)

	log.Println("Fetching the last URL sent time")
	rows, err := s.db.Query(ctx, "SELECT created_at FROM access_url ORDER BY created_at DESC LIMIT 1")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Println("No URLs found")
			return time.Time{}, nil
		}

		log.Printf("%s: %s", op, err)
		return time.Time{}, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var lastTime time.Time
	for rows.Next() {
		if err := rows.Scan(&lastTime); err != nil {
			log.Printf("%s: %s", op, err)
			return time.Time{}, fmt.Errorf("%s: %w", op, err)
		}
	}

	log.Printf("Fetched the last URL sent time: %v", lastTime)
	return lastTime, nil
}

// receiveAccessURLData selects expired URLs data from database.
func (s *Store) receiveAccessURLData(ctx context.Context) ([]model.AccessURL, error) {
	const op = "store.receiveAccessURLData"
	defer recoverPanic(op)

	current := time.Now()
	log.Println("Fetching expired access URLs")
	rows, err := s.db.Query(ctx, "SELECT id, access_key, api_url FROM access_url WHERE $1 > expired_at", current)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Println("No expired URLs found")
			return nil, nil
		}
		log.Printf("%s: %s", op, err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var accessURLs []model.AccessURL
	for rows.Next() {
		var accessURL model.AccessURL
		if err := rows.Scan(&accessURL.ID, &accessURL.AccessKey, &accessURL.ApiURL); err != nil {
			log.Printf("%s: %s", op, err)
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		accessURLs = append(accessURLs, accessURL)
	}

	log.Println("Fetched expired access URLs successfully")
	return accessURLs, nil
}

// receiveServerData selects server data from database.
func (s *Store) receiveServerData(ctx context.Context) ([]model.Server, error) {
	const op = "store.receiveServerData"
	defer recoverPanic(op)

	log.Println("Fetching server data")
	rows, err := s.db.Query(ctx, "SELECT ip_address, port, key FROM servers WHERE is_active = $1 AND is_test = $2 AND protocol = $3", true, true, "shadowsocks")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Println("No servers found")
			return nil, nil
		}
		log.Printf("%s: %s", op, err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var servers []model.Server
	for rows.Next() {
		var server model.Server
		if err := rows.Scan(&server.IPAddr, &server.Port, &server.Key); err != nil {
			log.Printf("%s: %s", op, err)
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		servers = append(servers, server)
	}
	log.Println(servers)
	log.Println("Fetched server data successfully")
	return servers, nil
}
