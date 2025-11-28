package models

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
