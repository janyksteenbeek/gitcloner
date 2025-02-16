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
	if config.URL == "" || config.Token == "" {
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

// getCurrentUser gets the current authenticated user
func (s *githubMirrorService) getCurrentUser() (string, error) {
	user, _, err := s.client.Users.Get(s.ctx, "")
	if err != nil {
		return "", fmt.Errorf("failed to get current user: %v", err)
	}
	return *user.Login, nil
}

// getOwner returns the owner (organization or user) for repository operations
func (s *githubMirrorService) getOwner() (string, error) {
	if s.config.OrgID != "" {
		return s.config.OrgID, nil
	}
	return s.getCurrentUser()
}

// getRepo safely gets a repository and handles 404 errors
func (s *githubMirrorService) getRepo(name string) (*github.Repository, error) {
	owner, err := s.getOwner()
	if err != nil {
		return nil, err
	}

	repo, resp, err := s.client.Repositories.Get(s.ctx, owner, name)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get repository: %v", err)
	}
	return repo, nil
}

// getDescription safely gets the description from a GitHub repository
func getDescription(repo *github.Repository) string {
	if repo != nil && repo.Description != nil {
		return *repo.Description
	}
	return ""
}

func (s *githubMirrorService) CheckRepository(repo Repository) (exists bool, isMirror bool, needsUpdate bool, err error) {
	existingRepo, err := s.getRepo(repo.Name)
	if err != nil {
		return false, false, false, err
	}

	if existingRepo != nil {
		needsUpdate = getDescription(existingRepo) != repo.Description
		return true, existingRepo.MirrorURL != nil, needsUpdate, nil
	}

	return false, false, false, nil
}

func (s *githubMirrorService) UpdateRepository(repo Repository) error {
	owner, err := s.getOwner()
	if err != nil {
		return err
	}

	log.Printf("Updating repository %s description", repo.Name)

	updateRepo := &github.Repository{
		Description: &repo.Description,
	}

	_, _, err = s.client.Repositories.Edit(s.ctx, owner, repo.Name, updateRepo)
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

	owner, err := s.getOwner()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrMirrorCreationFailed, err)
	}

	log.Printf("Creating mirror for %s, %s [ %s ]", repo.Name, repo.CloneURL, owner)

	// Create a new repository
	newRepo := &github.Repository{
		Name:        &repo.Name,
		Description: &repo.Description,
		Private:     &repo.Private,
	}

	if s.config.OrgID != "" {
		_, _, err = s.client.Repositories.Create(s.ctx, owner, newRepo)
	} else {
		_, _, err = s.client.Repositories.Create(s.ctx, "", newRepo)
	}
	if err != nil {
		return fmt.Errorf("%w: %v", ErrMirrorCreationFailed, err)
	}

	// Get authenticated clone URL if needed
	cloneURL, err := repo.GetAuthenticatedCloneURL(s.config.SourceToken)
	if err != nil {
		// Clean up the created repository
		_, _ = s.client.Repositories.Delete(s.ctx, owner, repo.Name)
		return err
	}

	// Set up mirroring
	mirrorConfig := &github.Repository{
		MirrorURL: &cloneURL,
	}

	_, _, err = s.client.Repositories.Edit(s.ctx, owner, repo.Name, mirrorConfig)
	if err != nil {
		// Clean up the created repository
		_, _ = s.client.Repositories.Delete(s.ctx, owner, repo.Name)
		return fmt.Errorf("%w: %v", ErrMirrorCreationFailed, err)
	}

	return nil
}

func (s *githubMirrorService) SyncRepository(repo Repository) error {
	owner, err := s.getOwner()
	if err != nil {
		return fmt.Errorf("failed to get owner: %v", err)
	}

	log.Printf("Syncing mirror for %s", repo.Name)

	eventType := "sync_mirror"
	_, _, err = s.client.Repositories.Dispatch(s.ctx, owner, repo.Name, github.DispatchRequestOptions{
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
