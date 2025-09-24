package config

import "github.com/ilyakaznacheev/cleanenv"

type AppConfig struct {
	HttpHostname string `env:"HTTP_HOSTNAME" env-default:":8080"`
}

// Load environment variables to AppConfig instance
func LoadAppConfig() (*AppConfig, error) {
	cfg := &AppConfig{}
	err := cleanenv.ReadEnv(cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
