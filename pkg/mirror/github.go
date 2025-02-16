package mirror

import (
	"context"
	"fmt"
	"log"

	"github.com/google/go-github/v60/github"
	"golang.org/x/oauth2"
)

type githubMirrorService struct {
	client *github.Client
	config Config
	ctx    context.Context
}

// NewGithubMirrorService creates a new GitHub mirror service
func NewGithubMirrorService(config Config) (MirrorService, error) {
	if config.URL == "" || config.Token == "" || config.OrgID == "" {
		return nil, ErrInvalidConfig
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config.Token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	return &githubMirrorService{
		client: client,
		config: config,
		ctx:    ctx,
	}, nil
}

// getDescription safely gets the description from a GitHub repository
func getDescription(repo *github.Repository) string {
	if repo != nil && repo.Description != nil {
		return *repo.Description
	}
	return ""
}

func (s *githubMirrorService) CheckRepository(repo Repository) (exists bool, isMirror bool, needsUpdate bool, err error) {
	existingRepo, resp, err := s.client.Repositories.Get(s.ctx, s.config.OrgID, repo.Name)
	if err == nil && existingRepo != nil {
		needsUpdate = getDescription(existingRepo) != repo.Description
		return true, existingRepo.MirrorURL != nil, needsUpdate, nil
	}

	// Check if it's a 404 (not found) error
	if resp != nil && resp.StatusCode == 404 {
		return false, false, false, nil
	}

	// Unexpected error
	return false, false, false, fmt.Errorf("failed to check repository: %v", err)
}

func (s *githubMirrorService) UpdateRepository(repo Repository) error {
	log.Printf("Updating repository %s description", repo.Name)

	updateRepo := &github.Repository{
		Description: &repo.Description,
	}

	_, _, err := s.client.Repositories.Edit(s.ctx, s.config.OrgID, repo.Name, updateRepo)
	if err != nil {
		return fmt.Errorf("failed to update repository: %v", err)
	}

	return nil
}

func (s *githubMirrorService) CreateMirror(repo Repository) error {
	exists, isMirror, needsUpdate, err := s.CheckRepository(repo)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrMirrorCreationFailed, err)
	}

	if exists {
		if !isMirror {
			return ErrRepositoryExists
		}

		// Repository exists and is a mirror, update it if needed
		if needsUpdate {
			if err := s.UpdateRepository(repo); err != nil {
				return fmt.Errorf("failed to update repository: %v", err)
			}
		}

		if err := s.SyncRepository(repo); err != nil {
			return fmt.Errorf("failed to sync repository: %v", err)
		}
		return nil
	}

	log.Printf("Creating mirror for %s, %s [ %s ]", repo.Name, repo.CloneURL, s.config.OrgID)

	// Create a new repository in GitHub
	newRepo := &github.Repository{
		Name:        &repo.Name,
		Description: &repo.Description,
		Private:     &repo.Private,
	}

	_, _, err = s.client.Repositories.Create(s.ctx, s.config.OrgID, newRepo)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrMirrorCreationFailed, err)
	}

	// Get authenticated clone URL if needed
	cloneURL, err := repo.GetAuthenticatedCloneURL(s.config.SourceToken)
	if err != nil {
		// Clean up the created repository
		_, _ = s.client.Repositories.Delete(s.ctx, s.config.OrgID, repo.Name)
		return err
	}

	// Set up mirroring
	mirrorConfig := &github.Repository{
		MirrorURL: &cloneURL,
	}

	_, _, err = s.client.Repositories.Edit(s.ctx, s.config.OrgID, repo.Name, mirrorConfig)
	if err != nil {
		// Clean up the created repository
		_, _ = s.client.Repositories.Delete(s.ctx, s.config.OrgID, repo.Name)
		return fmt.Errorf("%w: %v", ErrMirrorCreationFailed, err)
	}

	return nil
}

func (s *githubMirrorService) SyncRepository(repo Repository) error {
	log.Printf("Syncing mirror for %s", repo.Name)

	eventType := "sync_mirror"
	_, _, err := s.client.Repositories.Dispatch(s.ctx, s.config.OrgID, repo.Name, github.DispatchRequestOptions{
		EventType: eventType,
	})
	if err != nil {
		return fmt.Errorf("%w: %v", ErrMirrorSyncFailed, err)
	}

	return nil
}

// NeedsManualSync returns true as GitHub requires manual sync triggers
func (s *githubMirrorService) NeedsManualSync() bool {
	return true
}
