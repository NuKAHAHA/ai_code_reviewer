package codesage

import "AI_code_reviewer/internal/gitlab"

type FileDiff struct {
	OldPath     string `json:"old_path"`
	NewPath     string `json:"new_path"`
	NewFile     bool   `json:"new_file"`
	RenamedFile bool   `json:"renamed_file"`
	DeletedFile bool   `json:"deleted_file"`
	Diff        string `json:"diff"`
}

type Payload struct {
	Project gitlab.Project      `json:"project"`
	MR      gitlab.MRAttributes `json:"mr"`
	User    gitlab.User         `json:"user"`
	Changes []FileDiff          `json:"changes"`
}
