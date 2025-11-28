package gitlab

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ========= Типы для webhook =========

type MergeRequestEvent struct {
	ObjectKind       string       `json:"object_kind"`
	EventType        string       `json:"event_type"`
	User             User         `json:"user"`
	Project          Project      `json:"project"`
	ObjectAttributes MRAttributes `json:"object_attributes"`
}

type User struct {
	Username string `json:"username"`
	Name     string `json:"name"`
}

type Project struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	PathWithNamespace string `json:"path_with_namespace"`
}

type MRAttributes struct {
	ID           int    `json:"id"`
	IID          int    `json:"iid"`
	Title        string `json:"title"`
	State        string `json:"state"`
	Action       string `json:"action"`
	SourceBranch string `json:"source_branch"`
	TargetBranch string `json:"target_branch"`
	URL          string `json:"url"`
}

// ========= Типы для /merge_requests/:iid/changes =========

type MRChangesResponse struct {
	ID      int                `json:"id"`
	IID     int                `json:"iid"`
	Title   string             `json:"title"`
	Changes []MRFileChangeItem `json:"changes"`
}

type MRFileChangeItem struct {
	OldPath     string `json:"old_path"`
	NewPath     string `json:"new_path"`
	NewFile     bool   `json:"new_file"`
	RenamedFile bool   `json:"renamed_file"`
	DeletedFile bool   `json:"deleted_file"`
	Diff        string `json:"diff"`
}

// ========= Клиент GitLab =========

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetMergeRequestChanges получает diff по MR:
// GET /api/v4/projects/:id/merge_requests/:iid/changes
func (c *Client) GetMergeRequestChanges(projectID, mergeIID int) (*MRChangesResponse, error) {
	url := fmt.Sprintf("%s/api/v4/projects/%d/merge_requests/%d/changes", c.baseURL, projectID, mergeIID)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("PRIVATE-TOKEN", c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gitlab returned status %d: %s", resp.StatusCode, string(body))
	}

	var mrChanges MRChangesResponse
	if err := json.NewDecoder(resp.Body).Decode(&mrChanges); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &mrChanges, nil
}

func (c *Client) PostMRComment(projectID, mrIID int, body string) error {
	url := fmt.Sprintf("%s/api/v4/projects/%d/merge_requests/%d/notes",
		c.baseURL, projectID, mrIID)

	reqBody := map[string]string{"body": body}
	b, _ := json.Marshal(reqBody)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(b))
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}

	req.Header.Set("PRIVATE-TOKEN", c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("gitlab error %d: %s", resp.StatusCode, string(raw))
	}

	return nil
}
