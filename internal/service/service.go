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
	DeleteURL(context.Context, []string) error
	LastURLSentTime(context.Context) (time.Time, error)
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
		log.Printf("%s: %s", op, err)
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
		log.Printf("%s: %s", op, err)
		return fmt.Errorf("%s: %w", op, err)
	}

	for _, server := range servers {
		log.Println(server.IPAddr)
	}

	go s.safeStartSending(ctx, servers)
	go s.safeStartCleanup(ctx)

	// Wait until the context is done
	time.Sleep(time.Hour * 365 * 3 * 24)

	return nil
}

func (s *Service) safeStartSending(ctx context.Context, servers []entity.Server) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered in safeStartSending: %v", r)
		}
	}()
	s.startSending(ctx, servers)
}

func (s *Service) safeStartCleanup(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered in safeStartCleanup: %v", r)
		}
	}()
	s.startCleanup(ctx)
}

// startSending handles the periodic sending of access URLs to the telegram channel.
func (s *Service) startSending(ctx context.Context, servers []entity.Server) {
	const op = "service.startSending"

	lastURLSentTime, err := s.store.LastURLSentTime(ctx)
	if err != nil {
		log.Printf("%s: %s", op, err)
		return
	}

	if time.Since(lastURLSentTime) > 24*time.Hour {
		s.sendAccessURL(ctx, servers)
		log.Printf("First run, sent a new access URL")
	}

	sendTicker := time.NewTicker(24 * time.Hour)
	defer sendTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Context done, stopping startSending")
			return
		case <-sendTicker.C:
			if len(servers) == 0 {
				log.Println("[INFO] Servers are empty")
				continue
			}

			s.sendAccessURL(ctx, servers)
			log.Println("Sent a new access URL. Next key will be sent in 24 hours.")
		}
	}
}

// sendAccessURL sends a new access URL to the telegram channel.
func (s *Service) sendAccessURL(ctx context.Context, servers []entity.Server) {
	const op = "service.sendAccessURL"

	randIndex := rand.Intn(len(servers))

	accessURL, err := s.client.CreateAccessURL(servers[randIndex])
	if err != nil {
		log.Printf("%s: %s", op, err)
		return
	}

	if err := s.store.AddURL(ctx, accessURL); err != nil {
		log.Printf("%s: %s", op, err)
	}

	accessMessage := fmt.Sprintf(
		"🔑 Новый ключ на <b>48 часов</b>\n"+
			"🌍 Локация: <b>Европа</b>\n"+
			"💡 Инструкция - start.okbots.ru\n\n"+
			"<code>%s</code>\n"+
			"\n🚀 <b>Купить премиум VPN со скоростью до 10 гб/с:</b>\n"+
			"@okvpn_xbot",
		accessURL.AccessKey,
	)

	log.Println("Send telegram message")
	msg := tgbotapi.NewMessage(s.channelID, accessMessage)
	msg.ParseMode = "HTML"
	if _, err := s.bot.Send(msg); err != nil {
		log.Printf("%s: %s", op, err)
	}
}

// startCleanup handles the periodic cleanup of expired access URLs.
func (s *Service) startCleanup(ctx context.Context) {
	const op = "service.startCleanup"

	cleanupTicker := time.NewTicker(3 * time.Hour)
	defer cleanupTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Context done, stopping startCleanup")
			return
		case <-cleanupTicker.C:
			expiredURLs, err := s.store.ExpiredURLs(ctx)
			if err != nil {
				log.Printf("%s: %s", op, err)
			}

			log.Printf("Found %d expired URLs", len(expiredURLs))

			if err := s.client.RemoveAccessURLs(expiredURLs); err != nil {
				log.Printf("%s: %s", op, err)
			}

			if len(expiredURLs) > 0 {
				log.Printf("Successfully removed %d expired URLs from the client", len(expiredURLs))
			} else {
				log.Println("There are 0 expired urls, nothing to delete")
			}

			expiredIDs := make([]string, 0, len(expiredURLs))
			for _, u := range expiredURLs {
				expiredIDs = append(expiredIDs, u.ID)
			}

			if err := s.store.DeleteURL(ctx, expiredIDs); err != nil {
				log.Printf("%s: %s", op, err)
			} else {
				log.Printf("Successfully deleted URL with IDs %s from the store", expiredIDs)
			}
		}
	}
}
