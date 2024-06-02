package converter

import (
	"telegram-bot/internal/entity"
	"telegram-bot/internal/handlers/model"
)

// ToEntityFromClient convert model to entity.
func ToEntityFromClient(resp model.Response, apiURL string) entity.AccessURL {
	return entity.AccessURL{
		ID:        resp.ID,
		AccessKey: resp.AccessURL,
		ApiURL:    apiURL,
	}
}
