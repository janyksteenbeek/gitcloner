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
	Ref        string `json:"ref,omitempty"`
	Before     string `json:"before,omitempty"`
	After      string `json:"after,omitempty"`
	Created    bool   `json:"created,omitempty"`
	Deleted    bool   `json:"deleted,omitempty"`
	Forced     bool   `json:"forced,omitempty"`
	Repository struct {
		ID            int64  `json:"id"`
		Name          string `json:"name"`
		FullName      string `json:"full_name"`
		Description   string `json:"description"`
		Private       bool   `json:"private"`
		Fork          bool   `json:"fork"`
		DefaultBranch string `json:"default_branch"`
		CloneURL      string `json:"clone_url"`
		SSHURL        string `json:"ssh_url"`
		GitURL        string `json:"git_url"`
		HTMLURL       string `json:"html_url"`
		Visibility    string `json:"visibility"`
		Owner         struct {
			Name      string `json:"name,omitempty"`
			Email     string `json:"email,omitempty"`
			Login     string `json:"login"`
			ID        int64  `json:"id"`
			AvatarURL string `json:"avatar_url"`
			Type      string `json:"type"`
		} `json:"owner"`
		Organization struct {
			Login     string `json:"login"`
			ID        int64  `json:"id"`
			AvatarURL string `json:"avatar_url"`
		} `json:"organization,omitempty"`
	} `json:"repository"`
	Organization struct {
		Login       string `json:"login"`
		ID          int64  `json:"id"`
		NodeID      string `json:"node_id"`
		URL         string `json:"url"`
		AvatarURL   string `json:"avatar_url"`
		Description string `json:"description"`
	} `json:"organization,omitempty"`
	Pusher struct {
		Name  string `json:"name,omitempty"`
		Email string `json:"email,omitempty"`
	} `json:"pusher,omitempty"`
	Sender struct {
		Login     string `json:"login"`
		ID        int64  `json:"id"`
		AvatarURL string `json:"avatar_url"`
		Type      string `json:"type"`
	} `json:"sender"`
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
