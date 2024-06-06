package config

import (
	"os"
	"strconv"
)

type Config struct {
	DSN            string
	Token          string
	ChannelID      int64
	MigrationsPath string
}

// MustLoad return instance of config struct.
func MustLoad() *Config {
	var cfg Config

	dsn, token, channelID, migrationsPath := fetchEnv()
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
		panic("invalid channel id " + err.Error())
	}
	cfg.ChannelID = int64(chID)

	if migrationsPath == "" {
		panic("empty mirgations path string")
	}

	return &cfg
}

// fetchEnv receive variables from env.
func fetchEnv() (string, string, string, string) {
	dsn := os.Getenv("DSN")
	token := os.Getenv("TOKEN")
	channelID := os.Getenv("CHANNEL_ID")
	migrationsPath := os.Getenv("MIGRATIONS_PATH")

	return dsn, token, channelID, migrationsPath
}
