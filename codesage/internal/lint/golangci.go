package lint

import (
	"encoding/json"
	"os/exec"
	"codesage/internal/models"
)

func RunLint() ([]models.LintIssue, error) {
	cmd := exec.Command("golangci-lint", "run", "--out-format", "json")
	out, err := cmd.Output()
	if err != nil {
		// golangci-lint returns exit code 3 when issues found, that's okay
	}

	var result struct {
		Issues []struct {
			FromLinter string `json:"from_linter"`
			Text       string `json:"text"`
			Pos        struct {
				Filename string `json:"file"`
				Line     int    `json:"line"`
			} `json:"pos"`
			Severity string `json:"severity"`
		} `json:"Issues"`
	}

	json.Unmarshal(out, &result)

	var issues []models.LintIssue

	for _, i := range result.Issues {
		issues = append(issues, models.LintIssue{
			FromLinter: i.FromLinter,
			Text:       i.Text,
			Pos:        i.Pos.Filename,
			Severity:   i.Severity,
		})
	}

	return issues, nil
}
