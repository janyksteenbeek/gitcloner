package types

import "code.gitea.io/sdk/gitea"

// GiteaWebhookPayload represents a Gitea webhook payload
type GiteaWebhookPayload struct {
	Action     string           `json:"action"`
	Repository gitea.Repository `json:"repository"`
	Ref        string           `json:"ref"`
	After      string           `json:"after"`
}

// GitHubWebhookPayload represents a GitHub webhook payload
type GitHubWebhookPayload struct {
	Action     string `json:"action"`
	Repository struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Private     bool   `json:"private"`
		CloneURL    string `json:"clone_url"`
		Owner       struct {
			Login string `json:"login"`
		} `json:"owner"`
		DefaultBranch string `json:"default_branch"`
	} `json:"repository"`
	Ref string `json:"ref"`
}

// GitLabWebhookPayload represents a GitLab webhook payload
type GitLabWebhookPayload struct {
	EventType  string `json:"event_type"`
	ObjectKind string `json:"object_kind"`
	Project    struct {
		Name              string `json:"name"`
		Description       string `json:"description"`
		Visibility        string `json:"visibility"`
		DefaultBranch     string `json:"default_branch"`
		GitHTTPURL        string `json:"git_http_url"`
		PathWithNamespace string `json:"path_with_namespace"`
	} `json:"project"`
	Ref string `json:"ref"`
}
