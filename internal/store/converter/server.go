package converter

import (
	"strconv"
	"telegram-bot/internal/entity"
	"telegram-bot/internal/store/model"
)

// ToServerFromRepo convert model to entity.
func ToServerFromRepo(server model.Server) (entity.Server, error) {
	port, err := strconv.Atoi(server.Port)
	if err != nil {
		return entity.Server{}, nil
	}

	return entity.Server{
		IPAddr: server.IPAddr,
		Port:   port,
		Key:    server.Key,
	}, nil
}
