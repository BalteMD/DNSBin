package util

import (
	"time"

	"github.com/spf13/viper"
)

type NotifyConfig struct {
	Type     string `mapstructure:"type"`
	BotToken string `mapstructure:"bot_token"`
	ChatID   string `mapstructure:"chat_id"`
}

type Config struct {
	Environment       string         `mapstructure:"environment"`
	DBSource          string         `mapstructure:"db_source"`
	MigrationURL      string         `mapstructure:"migration_url"`
	HTTPServerAddress string         `mapstructure:"http_server_address"`
	DNSDomain         string         `mapstructure:"dns_domain"`
	HTTPLog           []string       `mapstructure:"http_log"`
	Interval          time.Duration  `mapstructure:"interval"`
	Notify            []NotifyConfig `mapstructure:"notify"`
	TXTValue          string         `mapstructure:"txt_value"`
	Endpoint          string         `mapstructure:"endpoint"`
	Insecure          bool           `mapstructure:"insecure"`
	ProxyURL          string         `mapstructure:"proxy_url"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
