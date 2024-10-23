package service

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"log/slog"
	"telegram-bot/internal/domain/models"
	"telegram-bot/internal/errors"
	"telegram-bot/internal/handlers/client"
	"telegram-bot/pkg/logger/sl"
	"time"
)

type Repository interface {
	Server(ctx context.Context) (models.Server, error)

	SaveAccessKey(ctx context.Context, accessKey models.AccessKey) error
	DeleteAccessKeys(ctx context.Context, uuids []uuid.UUID) error
	AccessKeyTime(ctx context.Context) (time.Time, error)
	ExpiredAccessKeys(ctx context.Context) ([]models.AccessKey, error)
}

type Service struct {
	log        *slog.Logger
	client     *client.Client
	repository Repository
}

type Options func(*Service)

func WithLogger(log *slog.Logger) Options {
	return func(service *Service) {
		service.log = log
	}
}

func WithClient(client *client.Client) Options {
	return func(service *Service) {
		service.client = client
	}
}

func WithRepository(repository Repository) Options {
	return func(service *Service) {
		service.repository = repository
	}
}

func New(opts ...Options) (*Service, error) {
	const op = "service.New"

	svc := new(Service)

	for _, opt := range opts {
		opt(svc)
	}

	if svc.client == nil {
		return nil, fmt.Errorf("%s: %w", op, errors.ErrNoClientProvided)
	}

	if svc.repository == nil {
		return nil, fmt.Errorf("%s: %w", op, errors.ErrNoRepositoryProvided)
	}

	if svc.log == nil {
		return nil, fmt.Errorf("%s: %w", op, errors.ErrNoLoggerProvided)
	}

	return svc, nil
}

func (s *Service) AccessKey(ctx context.Context) (string, error) {
	const op = "service.AccessKey"

	log := s.log.With(slog.String("op", op))

	log.Info("attempting to get random server")

	server, err := s.repository.Server(ctx)
	if err != nil {
		log.Error("failed to get random server", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("got random server")

	log.Info("attempting to get access key")
	accessKey, err := s.client.CreateAccessKey(server)
	if err != nil {
		log.Error("failed to get access key", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("got access key")

	if err := s.repository.SaveAccessKey(ctx, accessKey); err != nil {
		log.Error("failed to save access key", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("successfully saved access key")

	msg := fmt.Sprintf(
		"🔑 Новый ключ на <b>48 часов</b>\n"+
			"🌍 Локация: <b>Европа</b>\n"+
			"💡 Инструкция - start.okbots.ru\n\n"+
			"<code>%s</code>\n"+
			"\n🚀 <b>Купить премиум VPN со скоростью до 10 гб/с:</b>\n"+
			"@okvpn_xbot",
		accessKey.Key,
	)

	return msg, nil
}

func (s *Service) DeleteExpiredAccessKeys(ctx context.Context) error {
	const op = "service.DeleteExpiredAccessKeys"

	log := s.log.With(slog.String("op", op))

	log.Info("attempting to receive expired access keys")
	expiredKeys, err := s.repository.ExpiredAccessKeys(ctx)
	if err != nil {
		log.Error("failed to get expired access keys", sl.Err(err))

		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("got expired access keys")

	if len(expiredKeys) == 0 {
		log.Info("no expired access keys")
		return nil
	}

	log.Info("attempting to delete expired access keys")

	if err := s.client.DeleteAccessKeys(expiredKeys); err != nil {
		log.Error("failed to delete expired access keys", sl.Err(err))

		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("successfully deleted expired access keys")

	return nil
}
