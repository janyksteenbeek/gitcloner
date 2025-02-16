package webhook

import (
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/janyksteenbeek/gitcloner/pkg/mirror"
)

type Handler struct {
	mirrorConfig mirror.Config
	repoCache    *sync.Map
}

func NewHandler(config mirror.Config) *Handler {
	return &Handler{
		mirrorConfig: config,
		repoCache:    &sync.Map{},
	}
}

func (h *Handler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Detect webhook type based on headers
	var err error
	switch {
	case r.Header.Get("X-Gitea-Event") != "":
		err = h.handleGiteaWebhook(r)
	case r.Header.Get("X-GitHub-Event") != "":
		err = h.handleGitHubWebhook(r)
	case r.Header.Get("X-Gitlab-Event") != "":
		err = h.handleGitLabWebhook(r)
	default:
		http.Error(w, "Unknown webhook source", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to handle webhook: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) handlePushEvent(mirrorService mirror.MirrorService, repo mirror.Repository) error {
	exists, isMirror, needsUpdate, err := mirrorService.CheckRepository(repo)
	if err != nil {
		return fmt.Errorf("failed to check repository: %v", err)
	}

	if !exists {
		if err := mirrorService.CreateMirror(repo); err != nil {
			return fmt.Errorf("failed to create repository: %v", err)
		}
		return nil
	}

	if !isMirror {
		return fmt.Errorf("repository exists but is not a mirror")
	}

	if needsUpdate {
		if err := mirrorService.UpdateRepository(repo); err != nil {
			return fmt.Errorf("failed to update repository: %v", err)
		}
	}

	// Skip sync if provider handles it automatically and ALWAYS_PUSH is not set
	if !mirrorService.NeedsManualSync() && os.Getenv("ALWAYS_PUSH") == "" {
		return nil
	}

	// Sync the repository
	return mirrorService.SyncRepository(repo)
}
