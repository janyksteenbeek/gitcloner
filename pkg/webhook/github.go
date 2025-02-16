package webhook

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/janyksteenbeek/gitcloner/pkg/mirror"
	"github.com/janyksteenbeek/gitcloner/pkg/webhook/types"
)

func (h *Handler) handleGitHubWebhook(r *http.Request) error {
	eventType := r.Header.Get("X-GitHub-Event")

	// Parse the form if content type is form-urlencoded
	if r.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
		if err := r.ParseForm(); err != nil {
			return fmt.Errorf("failed to parse form data: %v", err)
		}
		// GitHub sends the JSON payload in a "payload" form field
		if len(r.Form["payload"]) == 0 {
			return fmt.Errorf("no payload found in form data")
		}
		payloadStr := r.Form["payload"][0]
		var payload types.GitHubWebhookPayload
		if err := json.Unmarshal([]byte(payloadStr), &payload); err != nil {
			return fmt.Errorf("failed to parse GitHub webhook payload: %v", err)
		}
		return h.handleGitHubPayload(eventType, payload)
	}

	// Handle regular JSON payload
	var payload types.GitHubWebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return fmt.Errorf("failed to parse GitHub webhook payload: %v", err)
	}
	return h.handleGitHubPayload(eventType, payload)
}

func (h *Handler) handleGitHubPayload(eventType string, payload types.GitHubWebhookPayload) error {
	mirrorService, err := mirror.NewMirrorService(h.mirrorConfig)
	if err != nil {
		return fmt.Errorf("failed to create mirror service: %v", err)
	}

	switch eventType {
	case "repository":
		if payload.Action == "created" {
			repo := mirror.Repository{
				Name:        formatRepoName(payload.Repository.Owner.Login, payload.Repository.Name),
				Description: payload.Repository.Description,
				Private:     payload.Repository.Private,
				CloneURL:    payload.Repository.CloneURL,
				Owner:       payload.Repository.Owner.Login,
			}
			return mirrorService.CreateMirror(repo)
		}
	case "push":
		if payload.Ref == fmt.Sprintf("refs/heads/%s", payload.Repository.DefaultBranch) {
			repo := mirror.Repository{
				Name:        formatRepoName(payload.Repository.Owner.Login, payload.Repository.Name),
				Description: payload.Repository.Description,
				Private:     payload.Repository.Private,
				CloneURL:    payload.Repository.CloneURL,
				Owner:       payload.Repository.Owner.Login,
			}
			return h.handlePushEvent(mirrorService, repo)
		}
	}

	return nil
}
