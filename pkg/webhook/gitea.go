package webhook

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/janyksteenbeek/gitcloner/pkg/mirror"
	"github.com/janyksteenbeek/gitcloner/pkg/webhook/types"
)

func (h *Handler) handleGiteaWebhook(r *http.Request) error {
	eventType := r.Header.Get("X-Gitea-Event")

	var payload types.GiteaWebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return fmt.Errorf("failed to parse Gitea webhook payload: %v", err)
	}

	mirrorService, err := mirror.NewMirrorService(h.mirrorConfig)
	if err != nil {
		return fmt.Errorf("failed to create mirror service: %v", err)
	}

	switch eventType {
	case "repository":
		return h.handleGiteaRepositoryEvent(mirrorService, payload)
	case "push":
		return h.handleGiteaPushEvent(mirrorService, payload)
	default:
		return nil
	}
}

func (h *Handler) handleGiteaRepositoryEvent(mirrorService mirror.MirrorService, payload types.GiteaWebhookPayload) error {
	if payload.Action != "created" {
		return nil
	}

	repo := mirror.Repository{
		Name:        formatRepoName(payload.Repository.Owner.UserName, payload.Repository.Name),
		Description: payload.Repository.Description,
		Private:     payload.Repository.Private,
		CloneURL:    payload.Repository.CloneURL,
		Owner:       payload.Repository.Owner.UserName,
	}

	return mirrorService.CreateMirror(repo)
}

func (h *Handler) handleGiteaPushEvent(mirrorService mirror.MirrorService, payload types.GiteaWebhookPayload) error {
	if payload.Ref != "refs/heads/"+payload.Repository.DefaultBranch {
		return nil
	}

	repo := mirror.Repository{
		Name:        formatRepoName(payload.Repository.Owner.UserName, payload.Repository.Name),
		Description: payload.Repository.Description,
		Private:     payload.Repository.Private,
		CloneURL:    payload.Repository.CloneURL,
		Owner:       payload.Repository.Owner.UserName,
	}

	return h.handlePushEvent(mirrorService, repo)
}
