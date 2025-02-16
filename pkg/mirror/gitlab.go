package mirror

import (
	"fmt"
	"log"

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

type gitlabMirrorService struct {
	client *gitlab.Client
	config Config
}

// NewGitlabMirrorService creates a new GitLab mirror service
func NewGitlabMirrorService(config Config) (MirrorService, error) {
	if config.URL == "" || config.Token == "" {
		return nil, ErrInvalidConfig
	}

	client, err := gitlab.NewClient(config.Token, gitlab.WithBaseURL(config.URL))
	if err != nil {
		return nil, fmt.Errorf("failed to create GitLab client: %w", err)
	}

	return &gitlabMirrorService{
		client: client,
		config: config,
	}, nil
}

// getCurrentUser gets the current authenticated user
func (s *gitlabMirrorService) getCurrentUser() (string, error) {
	user, _, err := s.client.Users.CurrentUser()
	if err != nil {
		return "", fmt.Errorf("failed to get current user: %v", err)
	}
	return user.Username, nil
}

// getOwner returns the owner (organization or user) for repository operations
func (s *gitlabMirrorService) getOwner() (string, error) {
	if s.config.OrgID != "" {
		return s.config.OrgID, nil
	}
	return s.getCurrentUser()
}

// verifyGroup verifies that the group exists if OrgID is provided
func (s *gitlabMirrorService) verifyGroup() error {
	if s.config.OrgID == "" {
		return nil
	}

	_, _, err := s.client.Groups.GetGroup(s.config.OrgID, nil)
	if err != nil {
		return fmt.Errorf("failed to get group: %v", err)
	}
	return nil
}

// findProject finds a project by name in the group/user namespace and returns its ID
func (s *gitlabMirrorService) findProject(name string) (*gitlab.Project, error) {
	if s.config.OrgID != "" {
		// Search in group projects
		listOpts := &gitlab.ListGroupProjectsOptions{
			Search: gitlab.Ptr(name),
		}
		projects, _, err := s.client.Groups.ListGroupProjects(s.config.OrgID, listOpts)
		if err != nil {
			return nil, fmt.Errorf("failed to find project in group: %v", err)
		}

		for _, p := range projects {
			if p.Name == name {
				return p, nil
			}
		}
	} else {
		// Search in user projects
		listOpts := &gitlab.ListProjectsOptions{
			Search: gitlab.Ptr(name),
		}
		projects, _, err := s.client.Projects.ListProjects(listOpts)
		if err != nil {
			return nil, fmt.Errorf("failed to find project: %v", err)
		}

		for _, p := range projects {
			if p.Name == name {
				return p, nil
			}
		}
	}

	return nil, nil
}

func (s *gitlabMirrorService) CheckRepository(repo Repository) (exists bool, isMirror bool, needsUpdate bool, err error) {
	if err := s.verifyGroup(); err != nil {
		return false, false, false, err
	}

	project, err := s.findProject(repo.Name)
	if err != nil {
		return false, false, false, err
	}

	if project != nil {
		needsUpdate = project.Description != repo.Description
		return true, project.Mirror, needsUpdate, nil
	}

	return false, false, false, nil
}

func (s *gitlabMirrorService) UpdateRepository(repo Repository) error {
	project, err := s.findProject(repo.Name)
	if err != nil {
		return err
	}

	if project == nil {
		return fmt.Errorf("project not found")
	}

	log.Printf("Updating repository %s description", repo.Name)

	updateOpts := &gitlab.EditProjectOptions{
		Description: gitlab.Ptr(repo.Description),
	}

	_, _, err = s.client.Projects.EditProject(project.ID, updateOpts)
	if err != nil {
		return fmt.Errorf("failed to update repository: %v", err)
	}

	return nil
}

func (s *gitlabMirrorService) CreateMirror(repo Repository) error {
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

	// Get authenticated clone URL if needed
	cloneURL, err := repo.GetAuthenticatedCloneURL(s.config.SourceToken)
	if err != nil {
		return err
	}

	// Create project options
	opts := &gitlab.CreateProjectOptions{
		Name:                gitlab.Ptr(repo.Name),
		Description:         gitlab.Ptr(repo.Description),
		ImportURL:           gitlab.Ptr(cloneURL),
		Mirror:              gitlab.Ptr(true),
		MirrorTriggerBuilds: gitlab.Ptr(true),
		Visibility:          visibilityLevel(repo.Private),
	}

	// Set namespace if using organization
	if s.config.OrgID != "" {
		group, _, err := s.client.Groups.GetGroup(s.config.OrgID, nil)
		if err != nil {
			return fmt.Errorf("%w: failed to get group: %v", ErrMirrorCreationFailed, err)
		}
		opts.NamespaceID = &group.ID
	}

	// Create the project
	_, _, err = s.client.Projects.CreateProject(opts)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrMirrorCreationFailed, err)
	}

	return nil
}

func (s *gitlabMirrorService) SyncRepository(repo Repository) error {
	// Get authenticated clone URL if needed
	cloneURL, err := repo.GetAuthenticatedCloneURL(s.config.SourceToken)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrMirrorSyncFailed, err)
	}

	project, err := s.findProject(repo.Name)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrMirrorSyncFailed, err)
	}

	if project == nil {
		return fmt.Errorf("%w: project not found", ErrMirrorSyncFailed)
	}

	log.Printf("Syncing mirror for %s", repo.Name)

	// Update project to force a mirror pull
	updateOpts := &gitlab.EditProjectOptions{
		ImportURL: gitlab.Ptr(cloneURL),
	}
	_, _, err = s.client.Projects.EditProject(project.ID, updateOpts)
	if err != nil {
		return fmt.Errorf("%w: failed to trigger mirror sync: %v", ErrMirrorSyncFailed, err)
	}

	return nil
}

// NeedsManualSync returns true as GitLab requires manual sync triggers
func (s *gitlabMirrorService) NeedsManualSync() bool {
	return true
}

// Helper functions

func visibilityLevel(private bool) *gitlab.VisibilityValue {
	if private {
		return gitlab.Ptr(gitlab.PrivateVisibility)
	}
	return gitlab.Ptr(gitlab.PublicVisibility)
}
