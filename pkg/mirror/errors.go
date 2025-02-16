package mirror

import "errors"

var (
	// ErrUnsupportedProvider is returned when the mirror provider type is not supported
	ErrUnsupportedProvider = errors.New("unsupported mirror provider")
	// ErrInvalidConfig is returned when the configuration is invalid
	ErrInvalidConfig = errors.New("invalid configuration")
	// ErrMirrorCreationFailed is returned when mirror creation fails
	ErrMirrorCreationFailed = errors.New("failed to create mirror")
	// ErrInvalidCloneURL is returned when the clone URL is invalid
	ErrInvalidCloneURL = errors.New("invalid clone URL")
	// ErrMirrorSyncFailed is returned when mirror sync fails
	ErrMirrorSyncFailed = errors.New("failed to sync mirror")
	// ErrSourceTokenRequired is returned when a source token is required but not provided
	ErrSourceTokenRequired = errors.New("SOURCE_TOKEN required for private repositories")
	// ErrRepositoryExists is returned when a repository already exists but is not a mirror
	ErrRepositoryExists = errors.New("repository already exists")
)
