package mirror

import (
	"fmt"
	"net/url"
	"strings"
)

// MirrorService defines the interface for repository mirroring
type MirrorService interface {
	CreateMirror(repo Repository) error
	SyncRepository(repo Repository) error
	NeedsManualSync() bool
	CheckRepository(repo Repository) (exists bool, isMirror bool, needsUpdate bool, err error)
	UpdateRepository(repo Repository) error
}

// Repository represents a generic repository structure
type Repository struct {
	Name        string
	Description string
	Private     bool
	CloneURL    string
	Owner       string
}

// GetAuthenticatedCloneURL returns the clone URL with authentication if needed
func (r *Repository) GetAuthenticatedCloneURL(sourceToken string) (string, error) {
	if !r.Private {
		return r.CloneURL, nil
	}

	if sourceToken == "" {
		return "", ErrSourceTokenRequired
	}

	parsedURL, err := url.Parse(r.CloneURL)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidCloneURL, err)
	}

	parsedURL.User = url.UserPassword("oauth2", sourceToken)
	return parsedURL.String(), nil
}

// Config holds the configuration for a mirror service
type Config struct {
	URL         string
	Token       string
	OrgID       string // Can be empty for personal accounts
	Type        string
	SourceToken string // Token used for authenticating with source repositories
}

// NewMirrorService creates a new mirror service based on the configuration
func NewMirrorService(config Config) (MirrorService, error) {
	if config.URL == "" || config.Token == "" {
		return nil, ErrInvalidConfig
	}

	switch config.Type {
	case "gitea":
		return NewGiteaMirrorService(config)
	case "gitlab":
		return NewGitlabMirrorService(config)
	case "github":
		return NewGithubMirrorService(config)
	default:
		return nil, ErrUnsupportedProvider
	}
}

// ParseRepositoryURL parses a repository URL and returns a Repository struct
func ParseRepositoryURL(repoURL string) (Repository, error) {
	parsedURL, err := url.Parse(repoURL)
	if err != nil {
		return Repository{}, fmt.Errorf("%w: %v", ErrInvalidCloneURL, err)
	}

	// Extract owner and name from path
	path := strings.Trim(parsedURL.Path, "/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return Repository{}, fmt.Errorf("%w: invalid repository path", ErrInvalidCloneURL)
	}

	owner := parts[0]
	name := strings.TrimSuffix(parts[1], ".git")

	// Determine if repository is private based on URL scheme
	// This is a best guess, the actual privacy status will be determined when creating the mirror
	private := parsedURL.Scheme == "git+ssh" || parsedURL.Scheme == "ssh"

	return Repository{
		Name:     name,
		Owner:    owner,
		CloneURL: repoURL,
		Private:  private,
	}, nil
}
