package main

import (
	"encoding/json"
	"net/http"

	"codesage/internal/ast"
	"codesage/internal/lint"
	"codesage/internal/models"
)

func analyzeHandler(w http.ResponseWriter, r *http.Request) {
	var req models.AnalysisRequest
	json.NewDecoder(r.Body).Decode(&req)

	// линтер
	lintIssues, _ := lint.RunLint()

	// AST анализируем DIFF а не файлы
	astIssues := ast.AnalyzeFiles(nil, req.Diff)

	resp := models.AnalysisResponse{
		LintIssues: lintIssues,
		ASTIssues:  astIssues,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func main() {
	http.HandleFunc("/analyze", analyzeHandler)
	http.ListenAndServe(":8081", nil)
}
