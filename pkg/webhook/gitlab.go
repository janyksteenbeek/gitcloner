package webhook

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/janyksteenbeek/gitcloner/pkg/mirror"
	"github.com/janyksteenbeek/gitcloner/pkg/webhook/types"
)

func (h *Handler) handleGitLabWebhook(r *http.Request) error {
	var payload types.GitLabWebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return fmt.Errorf("failed to parse GitLab webhook payload: %v", err)
	}

	mirrorService, err := mirror.NewMirrorService(h.mirrorConfig)
	if err != nil {
		return fmt.Errorf("failed to create mirror service: %v", err)
	}

	switch {
	case payload.ObjectKind == "project" && payload.EventType == "project_create":
		repo := mirror.Repository{
			Name:        formatRepoName(getOwnerFromPath(payload.Project.PathWithNamespace), payload.Project.Name),
			Description: payload.Project.Description,
			Private:     payload.Project.VisibilityLevel < 20,
			CloneURL:    payload.Project.GitHTTPURL,
			Owner:       getOwnerFromPath(payload.Project.PathWithNamespace),
		}
		return mirrorService.CreateMirror(repo)
	case payload.ObjectKind == "push":
		if payload.Ref == "refs/heads/"+payload.Project.DefaultBranch {
			repo := mirror.Repository{
				Name:        formatRepoName(getOwnerFromPath(payload.Project.PathWithNamespace), payload.Project.Name),
				Description: payload.Project.Description,
				Private:     payload.Project.VisibilityLevel < 20,
				CloneURL:    payload.Project.GitHTTPURL,
				Owner:       getOwnerFromPath(payload.Project.PathWithNamespace),
			}
			return h.handlePushEvent(mirrorService, repo)
		}
	}

	return nil
}
