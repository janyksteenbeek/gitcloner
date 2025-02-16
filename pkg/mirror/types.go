package mirror

import (
	"fmt"
	"net/url"
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
	OrgID       string
	Type        string
	SourceToken string // Token used for authenticating with source repositories
}

// NewMirrorService creates a new mirror service based on the configuration
func NewMirrorService(config Config) (MirrorService, error) {
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
