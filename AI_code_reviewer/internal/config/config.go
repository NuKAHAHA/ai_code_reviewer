package config

import (
	"fmt"
	"os"
)

type Config struct {
	GitLabBaseURL string
	GitLabToken   string
	WebhookSecret string
	HTTPPort      string
	CodeSageURL   string
	LLAMA_URL     string
}

func Load() (Config, error) {
	cfg := Config{
		GitLabBaseURL: os.Getenv("GITLAB_BASE_URL"),
		GitLabToken:   os.Getenv("GITLAB_TOKEN"),
		WebhookSecret: os.Getenv("GITLAB_WEBHOOK_SECRET"),
		HTTPPort:      os.Getenv("PORT"),
		CodeSageURL:   os.Getenv("CODESAGE_URL"),
		LLAMA_URL:     os.Getenv("LLAMA_URL"),
	}

	if cfg.GitLabToken == "" {
		return Config{}, fmt.Errorf("GITLAB_TOKEN is required")
	}

	return cfg, nil
}
