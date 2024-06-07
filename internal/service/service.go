package service

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"telegram-bot/internal/entity"
	"telegram-bot/internal/handlers/client"
	"time"

	tgbotapi "github.com/Syfaro/telegram-bot-api"
)

type store interface {
	Server(context.Context) ([]entity.Server, error)

	ExpiredURLs(context.Context) ([]entity.AccessURL, error)
	AddURL(context.Context, entity.AccessURL) error
	DeleteURL(context.Context, string) error
}

type Service struct {
	bot       *tgbotapi.BotAPI
	channelID int64

	client *client.Client
	store  store
}

func New(token string,
	channelID int64,
	client *client.Client,
	store store,
) (*Service, error) {
	const op = "service.New"

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Service{
		bot:       bot,
		channelID: channelID,
		client:    client,
		store:     store,
	}, nil
}

// StartBroadcast starts telegram channel broadcast.
func (s *Service) StartBroadcast(ctx context.Context) error {
	const op = "service.StartBroadcast"

	servers, err := s.store.Server(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	go s.startSending(ctx, servers)
	go s.startCleanup(ctx)

	// Wait until the context is done
	<-ctx.Done()

	return nil
}

// startSending handles the periodic sending of access URLs to the telegram channel.
func (s *Service) startSending(ctx context.Context, servers []entity.Server) {
	const op = "service.startSending"

	sendTicker := time.NewTicker(24 * time.Hour)
	defer sendTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-sendTicker.C:
			randIndex := rand.Intn(len(servers))

			accessURL, err := s.client.CreateAccessURL(servers[randIndex])
			if err != nil {
				log.Printf("%s: %s", op, err)
				continue
			}

			if err := s.store.AddURL(ctx, accessURL); err != nil {
				log.Printf("%s: %s", op, err)
			}

			msg := tgbotapi.NewMessage(s.channelID, accessURL.AccessKey)
			if _, err := s.bot.Send(msg); err != nil {
				log.Printf("%s: %s", op, err)
			}
		}
	}
}

// startCleanup handles the periodic cleanup of expired access URLs.
func (s *Service) startCleanup(ctx context.Context) {
	const op = "service.startCleanup"

	cleanupTicker := time.NewTicker(1 * time.Hour)
	defer cleanupTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-cleanupTicker.C:
			expiredURLs, err := s.store.ExpiredURLs(ctx)
			if err != nil {
				log.Printf("%s: %s", op, err)
			}

			if err := s.client.RemoveAccessURLs(expiredURLs); err != nil {
				log.Printf("%s: %s", op, err)
			}

			for _, u := range expiredURLs {
				if err := s.store.DeleteURL(ctx, u.ID); err != nil {
					log.Printf("%s: %s", op, err)
				}
			}
		}
	}
}
