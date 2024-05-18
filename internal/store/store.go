package store

import (
	"context"
	"fmt"
	"telegram-bot/internal/entity"
	"telegram-bot/internal/store/converter"
	"telegram-bot/internal/store/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type store struct {
	db *pgxpool.Pool
}

func New(dsn string) (*store, error) {
	const op = "store.server.New"

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
	const op = "store.server.Server"

	data, err := s.recieveServerData(ctx)
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

// recieveServerData select server data from database.
func (s *store) recieveServerData(ctx context.Context) ([]model.Server, error) {
	const op = "store.server.recieveServerData"

	rows, err := s.db.Query(ctx, "SELECT ip_address, port, key FROM servers WHERE is_active = $1 AND is_test = $2", true, true)
	if err != nil {
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
