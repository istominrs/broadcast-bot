package telegram

import (
	"context"
	"fmt"
	tgbotapi "github.com/Syfaro/telegram-bot-api"
	"log/slog"
	boterrors "telegram-bot/internal/bot"
	"telegram-bot/internal/errors"
	"telegram-bot/pkg/logger/sl"
	"time"
)

const (
	ParseMode = "HTML"
)

type Service interface {
	AccessKey(ctx context.Context) (string, error)
	DeleteExpiredAccessKeys(ctx context.Context) error
	IsKeySent(ctx context.Context) (bool, error)
}

type Bot struct {
	token   string
	chatID  int64
	log     *slog.Logger
	service Service

	bot *tgbotapi.BotAPI
}

type Options func(*Bot)

func WithLogger(log *slog.Logger) Options {
	return func(bot *Bot) {
		bot.log = log
	}
}

func WithToken(token string) Options {
	return func(bot *Bot) {
		bot.token = token
	}
}

func WithService(svc Service) Options {
	return func(bot *Bot) {
		bot.service = svc
	}
}

func WithChatID(chatID int64) Options {
	return func(bot *Bot) {
		bot.chatID = chatID
	}
}

func New(opts ...Options) (*Bot, error) {
	const op = "bot.New"

	bot := new(Bot)
	for _, opt := range opts {
		opt(bot)
	}

	if bot.token == "" {
		return nil, fmt.Errorf("%s: %w", op, boterrors.ErrNoTokenProvided)
	}

	if bot.chatID == 0 {
		return nil, fmt.Errorf("%s: %w", op, boterrors.ErrNoChatIDProvided)
	}

	if bot.log == nil {
		return nil, fmt.Errorf("%s: %w", op, errors.ErrNoLoggerProvided)
	}

	if bot.service == nil {
		return nil, fmt.Errorf("%s: %w", op, errors.ErrNoServiceProvided)
	}

	var err error
	bot.bot, err = tgbotapi.NewBotAPI(bot.token)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to initialize bot API: %w", op, err)
	}

	return bot, nil
}

func (b *Bot) Start(ctx context.Context) error {
	const op = "bot.Start"

	log := b.log.With(slog.String("op", op))

	log.Info("attempting to start telegram bot")

	go b.startLoop(ctx)

	log.Info("telegram bot started")

	<-ctx.Done()

	return nil
}

func (b *Bot) startLoop(ctx context.Context) {
	isKeySent, err := b.service.IsKeySent(ctx)
	if err != nil {
		b.log.Error("failed to check if key sent", sl.Err(err))
	}

	if !isKeySent {
		b.log.Info("key not sent, attempting to send access key")
		if err := b.sendMessage(ctx); err != nil {
			b.log.Error("failed to send access key", sl.Err(err))
		}

		b.log.Info("key sent")
	}

	checkTicker := time.NewTicker(5 * time.Hour)
	sendTicker := time.NewTicker(24 * time.Hour)
	defer checkTicker.Stop()
	defer sendTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			b.log.Info("stopping telegram bot")
			return
		case <-checkTicker.C:
			if err := b.service.DeleteExpiredAccessKeys(ctx); err != nil {
				b.log.Error("failed to check access keys", sl.Err(err))
			}
		case <-sendTicker.C:
			if err := b.sendMessage(ctx); err != nil {
				b.log.Error("failed to send message", sl.Err(err))
			}
		}
	}
}

func (b *Bot) sendMessage(ctx context.Context) error {
	const op = "bot.sendMessage"

	log := b.log.With(slog.String("op", op))

	log.Info("attempting to get access message from service")
	message, err := b.service.AccessKey(ctx)
	if err != nil {
		log.Error("failed to get access message from service", sl.Err(err))

		return fmt.Errorf("%s: %w", op, err)
	}

	msg := tgbotapi.NewMessage(b.chatID, message)
	msg.ParseMode = ParseMode

	log.Info("attempting to send message")
	_, err = b.bot.Send(msg)
	if err != nil {
		log.Error("failed to send message", sl.Err(err))

		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("message sent successfully")
	return nil
}
