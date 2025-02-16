package mirror

import (
	"fmt"
	"log"
	"net/url"

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

type gitlabMirrorService struct {
	client *gitlab.Client
	config Config
}

// NewGitlabMirrorService creates a new GitLab mirror service
func NewGitlabMirrorService(config Config) (MirrorService, error) {
	if config.URL == "" || config.Token == "" || config.OrgID == "" {
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

// findProject finds a project by name in the group and returns its ID
func (s *gitlabMirrorService) findProject(name string) (*gitlab.Project, error) {
	listOpts := &gitlab.ListGroupProjectsOptions{
		Search: gitlab.Ptr(name),
	}
	projects, _, err := s.client.Groups.ListGroupProjects(s.config.OrgID, listOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to find project: %v", err)
	}

	for _, p := range projects {
		if p.Name == name {
			return p, nil
		}
	}

	return nil, nil
}

func (s *gitlabMirrorService) CheckRepository(repo Repository) (exists bool, isMirror bool, needsUpdate bool, err error) {
	// Get group ID by path
	_, _, err = s.client.Groups.GetGroup(s.config.OrgID, nil)
	if err != nil {
		return false, false, false, fmt.Errorf("failed to get group: %v", err)
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

	log.Printf("Creating mirror for %s, %s [ %s ]", repo.Name, repo.CloneURL, s.config.OrgID)

	// Get group ID by path
	group, _, err := s.client.Groups.GetGroup(s.config.OrgID, nil)
	if err != nil {
		return fmt.Errorf("%w: failed to get group: %v", ErrMirrorCreationFailed, err)
	}

	// Get authenticated clone URL if needed
	cloneURL, err := repo.GetAuthenticatedCloneURL(s.config.SourceToken)
	if err != nil {
		return err
	}

	// Create project options
	opts := &gitlab.CreateProjectOptions{
		Name:                gitlab.Ptr(repo.Name),
		Description:         gitlab.Ptr(repo.Description),
		NamespaceID:         &group.ID,
		Visibility:          visibilityLevel(repo.Private),
		ImportURL:           gitlab.Ptr(cloneURL),
		Mirror:              gitlab.Ptr(true),
		MirrorTriggerBuilds: gitlab.Ptr(true),
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

func addAuthToURL(cloneURL, username, token string) (string, error) {
	parsedURL, err := url.Parse(cloneURL)
	if err != nil {
		return "", err
	}
	parsedURL.User = url.UserPassword(username, token)
	return parsedURL.String(), nil
}
