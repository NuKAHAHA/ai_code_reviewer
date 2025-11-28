package codesage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type AnalysisRequest struct {
	Diff string `json:"diff"`
}

type LintIssue struct {
	FromLinter string `json:"from_linter"`
	Text       string `json:"text"`
	Pos        string `json:"pos"`
	Severity   string `json:"severity"`
}

type ASTIssue struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Line    int    `json:"line"`
}

type AnalysisResponse struct {
	LintIssues []LintIssue `json:"lint_issues"`
	ASTIssues  []ASTIssue  `json:"ast_issues"`
}

func Call(url string, diff string) (*AnalysisResponse, error) {
	reqBody := AnalysisRequest{Diff: diff}

	b, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(b))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 60 * time.Second}

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("codesage returned status %d", resp.StatusCode)
	}

	var out AnalysisResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &out, nil
}
