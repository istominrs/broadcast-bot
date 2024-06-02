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

// ToAccessURLFromRepo convert model to entity.
func ToAccessURLFromRepo(accessURL model.AccessURL) entity.AccessURL {
	id := strconv.Itoa(accessURL.ID)

	return entity.AccessURL{
		ID:        id,
		AccessKey: accessURL.AccessKey,
		ApiURL:    accessURL.ApiURL,
	}
}

// ToAccessURLFromRepo convert entity to model.
func ToRepoFromAccessURL(accessURL entity.AccessURL) (model.AccessURL, error) {
	id, err := strconv.Atoi(accessURL.ID)
	if err != nil {
		return model.AccessURL{}, err
	}

	return model.AccessURL{
		ID:        id,
		AccessKey: accessURL.AccessKey,
		ApiURL:    accessURL.ApiURL,
	}, nil
}
