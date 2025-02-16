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

	var payload types.GitHubWebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return fmt.Errorf("failed to parse GitHub webhook payload: %v", err)
	}

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
		if payload.Ref == "refs/heads/"+payload.Repository.DefaultBranch {
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
