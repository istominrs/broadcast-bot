package config

import (
	"flag"
	"os"
	"strconv"
)

type Config struct {
	DSN       string
	Token     string
	ChannelID int64
}

func MustLoad() *Config {
	var cfg Config

	dsn, token, channelID := fetchConfigData()
	if dsn == "" {
		panic("empty dsn string")
	}
	cfg.DSN = dsn

	if token == "" {
		panic("empty token string")
	}
	cfg.Token = token

	chID, err := strconv.Atoi(channelID)
	if err != nil {
		panic("invalid channel id")
	}
	cfg.ChannelID = int64(chID)

	return &cfg
}

func fetchConfigData() (string, string, string) {
	var dsn, token, channelID string

	flag.StringVar(&dsn, "dsn", "", "dsn string")
	flag.StringVar(&token, "token", "", "token string")
	flag.StringVar(&channelID, "channel-id", "", "channel id")
	flag.Parse()

	if token == "" {
		token = os.Getenv("TOKEN")
	}

	if dsn == "" {
		dsn = os.Getenv("DSN")
	}

	if channelID == "" {
		channelID = os.Getenv("CHANNEL_ID")
	}

	return dsn, token, channelID
}
