package ast

import (
	"go/parser"
	"go/token"
	"strings"

	"codesage/internal/models"
)

func AnalyzeFiles(_ []string, diff string) []models.ASTIssue {
	var issues []models.ASTIssue

	code := extractCode(diff)
	if code == "" {
		return issues
	}

	fs := token.NewFileSet()
	_, err := parser.ParseFile(fs, "patched.go", code, parser.AllErrors)
	if err != nil {
		for _, line := range strings.Split(err.Error(), "\n") {
			issues = append(issues, models.ASTIssue{
				Type:    "parse_error",
				Message: line,
				Line:    0,
			})
		}
	}

	return issues
}

func extractCode(diff string) string {
	var out strings.Builder
	for _, line := range strings.Split(diff, "\n") {
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			out.WriteString(line[1:])
			out.WriteString("\n")
		}
	}
	return out.String()
}
