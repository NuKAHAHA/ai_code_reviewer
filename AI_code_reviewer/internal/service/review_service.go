package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"AI_code_reviewer/internal/codesage"
	"AI_code_reviewer/internal/gitlab"
)

type Service struct {
	gitlabClient *gitlab.Client
	codeSageURL  string
	llamaURL     string
}

func NewService(gl *gitlab.Client, codeSageURL, llamaURL string) *Service {
	return &Service{
		gitlabClient: gl,
		codeSageURL:  codeSageURL,
		llamaURL:     llamaURL,
	}
}

func (s *Service) HandleMergeRequest(ctx context.Context, event gitlab.MergeRequestEvent) (*codesage.Payload, bool, error) {
	action := event.ObjectAttributes.Action
	if !isInterestingAction(action) {
		log.Printf("review: skip MR !%d action=%s project=%s",
			event.ObjectAttributes.IID,
			action,
			event.Project.PathWithNamespace,
		)
		return nil, true, nil
	}

	log.Printf("review: process MR !%d action=%s project=%s",
		event.ObjectAttributes.IID,
		action,
		event.Project.PathWithNamespace,
	)

	mrChanges, err := s.gitlabClient.GetMergeRequestChanges(event.Project.ID, event.ObjectAttributes.IID)
	if err != nil {
		return nil, false, fmt.Errorf("get MR changes: %w", err)
	}

	log.Printf("review: MR !%d changes received: %d files",
		event.ObjectAttributes.IID,
		len(mrChanges.Changes),
	)

	changes := make([]codesage.FileDiff, 0, len(mrChanges.Changes))
	for _, ch := range mrChanges.Changes {
		changes = append(changes, codesage.FileDiff{
			OldPath:     ch.OldPath,
			NewPath:     ch.NewPath,
			NewFile:     ch.NewFile,
			RenamedFile: ch.RenamedFile,
			DeletedFile: ch.DeletedFile,
			Diff:        ch.Diff,
		})
	}

	payload := &codesage.Payload{
		Project: event.Project,
		MR:      event.ObjectAttributes,
		User:    event.User,
		Changes: changes,
	}

	// ---- Сбор DIFF ----
	var fullDiff strings.Builder
	for _, ch := range changes {
		fullDiff.WriteString(ch.Diff)
		fullDiff.WriteString("\n")
	}

	// ---- Вызов CodeSage ----
	var codeResp *codesage.AnalysisResponse
	if s.codeSageURL != "" {
		codeResp, err = codesage.Call(s.codeSageURL, fullDiff.String())
		if err != nil {
			log.Printf("review: codesage call error: %v", err)
		} else {
			log.Printf("review: codesage lint_issues=%d ast_issues=%d",
				len(codeResp.LintIssues),
				len(codeResp.ASTIssues),
			)
		}
	}

	// ---- Формируем prompt для LLAMA ----
	prompt := fmt.Sprintf(`
You are a strict senior Go code reviewer.
Your task is to generate a deterministic Merge Request review. 
Follow THIS EXACT TEMPLATE and NEVER change section titles:

### 1. Summary
(2–3 sentences)

### 2. Lint Issues
(List ALL lint issues exactly as provided by CodeSage:
%v)

### 3. AST Parse Errors
(List ALL AST issues exactly as provided by CodeSage:
%v)

### 4. Problems Found
(List real problems from the diff. VERY IMPORTANT: reference exact line numbers and code snippets.)

### 5. Recommendations
(Give clear steps to fix the problems.)

### 6. Corrected Code Example
(Show the fully corrected code.)

MR: !%d  
Title: %s  
Project: %s  

### DIFF:
%s
`,
		codeResp.LintIssues,
		codeResp.ASTIssues,
		event.ObjectAttributes.IID,
		event.ObjectAttributes.Title,
		event.Project.PathWithNamespace,
		fullDiff.String(),
	)

	// ---- Вызов LLAMA через Ollama ----
	reviewText := ""

	if s.llamaURL != "" {
		body := map[string]interface{}{
			"model":  "llama3.1",
			"prompt": prompt,
			"stream": false,
		}

		b, _ := json.Marshal(body)

		resp, err := http.Post(s.llamaURL, "application/json", bytes.NewBuffer(b))
		if err != nil {
			log.Printf("LLAMA error: %v", err)
		} else {
			var llamaResp struct {
				Response string `json:"response"`
			}
			json.NewDecoder(resp.Body).Decode(&llamaResp)
			resp.Body.Close()

			reviewText = llamaResp.Response
			log.Printf("LLAMA review generated")
		}
	}

	// ---- Отправляем комментарий в GitLab ----
	if reviewText != "" {
		err := s.gitlabClient.PostMRComment(
			event.Project.ID,
			event.ObjectAttributes.IID,
			reviewText,
		)
		if err != nil {
			log.Printf("gitlab comment error: %v", err)
		} else {
			log.Printf("comment posted to MR !%d", event.ObjectAttributes.IID)
		}
	}

	return payload, false, nil
}

func isInterestingAction(action string) bool {
	switch action {
	case "open", "reopen", "update", "push":
		return true
	default:
		return false
	}
}
