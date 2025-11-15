package config

import (
	"log"
	"os"
	"time"

	"github.com/spf13/viper"
)

//type Config struct {
//	BotToken          string        `env:"MAX_BOT_TOKEN,required"`
//	RedisAddr         string        `env:"REDIS_ADDR" envDefault:"127.0.0.1:6379"`
//	RedisPassword     string        `env:"REDIS_PASSWORD"`
//	RedisDB           int           `env:"REDIS_DB" envDefault:"0"`
//	PollingTimeout    time.Duration `env:"POLLING_TIMEOUT" envDefault:"30s"`
//	LogLevel          string        `env:"LOG_LEVEL" envDefault:"info"`
//	MockScheduleLag   time.Duration `env:"MOCK_SCHEDULE_LAG" envDefault:"500ms"`
//	YandexGPTAPIKey   string        `env:"YANDEX_GPT_API_KEY"`
//	YandexGPTFolderID string        `env:"YANDEX_GPT_FOLDER_ID"`
//}

type Config struct {
	BotToken          string        `mapstructure:"MAX_BOT_TOKEN"`
	RedisAddr         string        `mapstructure:"REDIS_ADDR"`
	RedisPassword     string        `mapstructure:"REDIS_PASSWORD"`
	RedisDB           int           `mapstructure:"REDIS_DB"`
	PollingTimeout    time.Duration `mapstructure:"POLLING_TIMEOUT"`
	LogLevel          string        `mapstructure:"LOG_LEVEL"`
	MockScheduleLag   time.Duration `mapstructure:"MOCK_SCHEDULE_LAG"`
	YandexGPTAPIKey   string        `mapstructure:"YANDEX_GPT_API_KEY"`
	YandexGPTFolderID string        `mapstructure:"YANDEX_GPT_FOLDER_ID"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	v := viper.NewWithOptions(viper.ExperimentalBindStruct())
	v.AddConfigPath(".")
	v.SetConfigName(".env")
	v.SetConfigType("env")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if os.Getenv("REDIS_PORT") == "" {
			log.Fatalf("Error reading config file, %s", err)
		}
	}

	if err := v.Unmarshal(&cfg); err != nil {
		log.Fatalf("Unable to decode into struct: %v", err)
	}
	//if err := env.Parse(cfg); err != nil {
	//	return nil, err
	//}

	//cfg.BotToken = "f9LHodD0cOKtdNXYJSGw08b_JxydItPjFrhEtkukQjfSybcpG7swcMV1ytwnZHJLSEtUSQSG9PVRG3Zd-EHB"
	//cfg.RedisAddr = "redis.my-first-bot.orb.local:6379"
	//cfg.RedisPassword = ""
	//cfg.RedisDB = 0
	//cfg.LogLevel = "info"
	//cfg.YandexGPTAPIKey = ""
	//cfg.YandexGPTFolderID = ""

	return cfg, nil
}
