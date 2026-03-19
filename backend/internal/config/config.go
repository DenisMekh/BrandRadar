package config

import (
	"fmt"
	"strings"
	"sync"

	"github.com/spf13/viper"
)

var (
	instance *Config
	once     sync.Once
)

type Config struct {
	App         App                   `mapstructure:"app"`
	Postgres    Postgres              `mapstructure:"postgres"`
	Redis       Redis                 `mapstructure:"redis"`
	Crawler     Crawler               `mapstructure:"crawler"`
	SentimentML SentimentML           `mapstructure:"ml_sentiment"`
	Workers     WorkerIntervalMinutes `mapstructure:"worker"`
	Telegram    Telegram              `mapstructure:"telegram"`
}

type Telegram struct {
	BotToken string `mapstructure:"bot_token"`
	ChatID   string `mapstructure:"chat_id"`
}

type App struct {
	Env            string    `mapstructure:"env"`
	Port           int       `mapstructure:"port"`
	LogLevel       string    `mapstructure:"log_level"`
	RateLimit      RateLimit `mapstructure:"rate_limit"`
	TrustedProxies []string  `mapstructure:"trusted_proxies"`
}

type WorkerIntervalMinutes struct {
	Clustering int `mapstructure:"clustering"`
	Analytics  int `mapstructure:"analytics"`
	Anomalies  int `mapstructure:"anomalies"`
}
type SentimentML struct {
	URL string `mapstructure:"url"`
}

type RateLimit struct {
	RPS   float64 `mapstructure:"rps"`
	Burst int     `mapstructure:"burst"`
}

type Postgres struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
	SSLMode  string `mapstructure:"ssl_mode"`
	MaxConns int32  `mapstructure:"max_conns"`
	MinConns int32  `mapstructure:"min_conns"`
}

type Redis struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type Crawler struct {
	Timeout          int      `mapstructure:"timeout"`
	PostsLimit       int      `mapstructure:"posts_limit"`
	TelegramChannels []string `mapstructure:"telegram_channels"`
	WebURLs          []string `mapstructure:"web_urls"`
}

func Get() *Config {
	once.Do(func() {
		instance = &Config{}
	})
	return instance
}

func Init(configPath string) error {
	var initErr error
	once.Do(func() {
		viper.SetConfigFile(configPath)
		viper.SetConfigType("yaml")

		viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		viper.AutomaticEnv()

		if err := viper.ReadInConfig(); err != nil {
			initErr = fmt.Errorf("config: read: %w", err)
			return
		}

		instance = &Config{}
		if err := viper.Unmarshal(instance); err != nil {
			initErr = fmt.Errorf("config: unmarshal: %w", err)
			return
		}
	})
	return initErr
}
