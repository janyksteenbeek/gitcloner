package mirror

import (
	"fmt"
	"log"

	"code.gitea.io/sdk/gitea"
)

type giteaMirrorService struct {
	client *gitea.Client
	config Config
}

// NewGiteaMirrorService creates a new Gitea mirror service
func NewGiteaMirrorService(config Config) (MirrorService, error) {
	if config.URL == "" || config.Token == "" || config.OrgID == "" {
		return nil, ErrInvalidConfig
	}

	client, err := gitea.NewClient(config.URL, gitea.SetToken(config.Token))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gitea client: %w", err)
	}

	return &giteaMirrorService{
		client: client,
		config: config,
	}, nil
}

// getRepo safely gets a repository and handles 404 errors
func (s *giteaMirrorService) getRepo(name string) (*gitea.Repository, error) {
	repo, resp, err := s.client.GetRepo(s.config.OrgID, name)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get repository: %v", err)
	}
	return repo, nil
}

func (s *giteaMirrorService) CheckRepository(repo Repository) (exists bool, isMirror bool, needsUpdate bool, err error) {
	existingRepo, err := s.getRepo(repo.Name)
	if err != nil {
		return false, false, false, err
	}

	if existingRepo != nil {
		needsUpdate = existingRepo.Description != repo.Description
		return true, existingRepo.Mirror, needsUpdate, nil
	}

	return false, false, false, nil
}

func (s *giteaMirrorService) UpdateRepository(repo Repository) error {
	log.Printf("Updating repository %s description", repo.Name)

	updateOpts := gitea.EditRepoOption{
		Description: &repo.Description,
	}

	_, _, err := s.client.EditRepo(s.config.OrgID, repo.Name, updateOpts)
	if err != nil {
		return fmt.Errorf("failed to update repository: %v", err)
	}

	return nil
}

func (s *giteaMirrorService) CreateMirror(repo Repository) error {
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

	// Set up mirroring using the migration API
	migrationOpts := gitea.MigrateRepoOption{
		RepoName:       repo.Name,
		RepoOwner:      s.config.OrgID,
		CloneAddr:      repo.CloneURL,
		Mirror:         true,
		Private:        repo.Private,
		Description:    repo.Description,
		Service:        gitea.GitServiceGitea,
		AuthToken:      s.config.SourceToken,
		MirrorInterval: "1h0m0s",
	}

	_, _, err = s.client.MigrateRepo(migrationOpts)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrMirrorCreationFailed, err)
	}

	return nil
}

func (s *giteaMirrorService) SyncRepository(repo Repository) error {
	log.Printf("Syncing mirror for %s", repo.Name)

	// Get the repository
	giteaRepo, err := s.getRepo(repo.Name)
	if err != nil {
		return fmt.Errorf("failed to get repository: %v", err)
	}

	if giteaRepo == nil {
		return fmt.Errorf("%w: repository not found", ErrMirrorSyncFailed)
	}

	// Trigger a mirror sync
	_, err = s.client.MirrorSync(s.config.OrgID, giteaRepo.Name)
	if err != nil {
		return fmt.Errorf("failed to trigger mirror sync: %v", err)
	}

	return nil
}

// NeedsManualSync returns false as Gitea handles mirror syncing automatically
func (s *giteaMirrorService) NeedsManualSync() bool {
	return false
}
