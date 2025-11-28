package httpserver

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"AI_code_reviewer/internal/config"
	"AI_code_reviewer/internal/gitlab"
	"AI_code_reviewer/internal/service"
)

type Server struct {
	cfg        config.Config
	Svc        *service.Service
	httpServer *http.Server
}

func New(cfg config.Config, Svc *service.Service) *Server {
	return &Server{
		cfg: cfg,
		Svc: Svc,
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/gitlab/webhook", s.handleGitLabWebhook)

	s.httpServer = &http.Server{
		Addr:    ":" + s.cfg.HTTPPort,
		Handler: mux,
	}

	log.Printf("httpserver: starting on :%s ...", s.cfg.HTTPPort)
	return s.httpServer.ListenAndServe()
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}

func (s *Server) handleGitLabWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.cfg.WebhookSecret != "" {
		token := r.Header.Get("X-Gitlab-Token")
		if token == "" || token != s.cfg.WebhookSecret {
			log.Printf("httpserver: invalid webhook token: %q", token)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
	}

	eventHeader := r.Header.Get("X-Gitlab-Event")
	log.Printf("httpserver: received GitLab event: %s", eventHeader)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("httpserver: failed to read body: %v", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var mrEvent gitlab.MergeRequestEvent
	if err := json.Unmarshal(body, &mrEvent); err != nil {
		log.Printf("httpserver: failed to unmarshal MR event: %v", err)
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if mrEvent.ObjectKind != "merge_request" {
		log.Printf("httpserver: skip event with object_kind=%s", mrEvent.ObjectKind)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ignored"))
		return
	}

	payload, skipped, err := s.Svc.HandleMergeRequest(r.Context(), mrEvent)
	if err != nil {
		log.Printf("httpserver: HandleMergeRequest error: %v", err)
		http.Error(w, "failed to process MR", http.StatusInternalServerError)
		return
	}

	if skipped {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("skipped"))
		return
	}

	// Здесь пока НИЧЕГО не шлём в CodeSage.
	// Только логируем, что payload готов.
	log.Printf("httpserver: CodeSage payload prepared: project=%s MR=!%d changes=%d",
		payload.Project.PathWithNamespace,
		payload.MR.IID,
		len(payload.Changes),
	)

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
