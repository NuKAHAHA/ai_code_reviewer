package main

import (
	"log"

	"AI_code_reviewer/internal/config"
	"AI_code_reviewer/internal/gitlab"
	"AI_code_reviewer/internal/http_server"
	"AI_code_reviewer/internal/service"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load("./.env")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	gitlabClient := gitlab.NewClient(cfg.GitLabBaseURL, cfg.GitLabToken)
	reviewSvc := service.NewService(gitlabClient, cfg.CodeSageURL, cfg.LLAMA_URL)
	server := httpserver.New(cfg, reviewSvc)

	if err := server.Start(); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
