package mirror

import (
	"fmt"
	"log"
	"strings"
)

// HandleImport processes repository import requests
func HandleImport(config Config, input string) error {
	parts := strings.Fields(input)
	if len(parts) != 2 {
		return fmt.Errorf("invalid import format. Use: --import 'platform username/repo[,username2/repo2,...]'")
	}

	platform := parts[0]
	repoPaths := strings.Split(parts[1], ",")

	for _, repoPath := range repoPaths {
		repoPath = strings.TrimSpace(repoPath)
		if repoPath == "" {
			continue
		}

		if err := handleSingleImport(config, platform, repoPath); err != nil {
			log.Printf("Warning: Failed to import repository %s: %v", repoPath, err)
			continue
		}
	}

	return nil
}

// handleSingleImport processes a single repository import
func handleSingleImport(config Config, platform, repoPath string) error {
	// Split username/repo
	repoParts := strings.Split(repoPath, "/")
	if len(repoParts) != 2 {
		return fmt.Errorf("invalid repository format. Use: username/repo")
	}

	owner := repoParts[0]
	name := repoParts[1]

	// Create base URL based on platform
	var cloneURL string
	switch platform {
	case "github":
		cloneURL = fmt.Sprintf("https://github.com/%s/%s", owner, name)
	case "gitlab":
		cloneURL = fmt.Sprintf("https://gitlab.com/%s/%s", owner, name)
	case "gitea":
		// For Gitea, we need the instance URL from config
		cloneURL = fmt.Sprintf("%s/%s/%s", strings.TrimSuffix(config.URL, "/"), owner, name)
	default:
		return fmt.Errorf("unsupported platform: %s. Use github, gitlab, or gitea", platform)
	}

	mirrorService, err := NewMirrorService(config)
	if err != nil {
		return err
	}

	repo := Repository{
		Name:     formatRepoName(owner, name),
		CloneURL: cloneURL,
		Owner:    owner,
		Private:  true,
	}

	log.Printf("Importing repository: %s from %s", repoPath, platform)
	if err := mirrorService.CreateMirror(repo); err != nil {
		return fmt.Errorf("failed to import repository %s: %v", repoPath, err)
	}
	log.Printf("Successfully imported repository: %s", repoPath)

	return nil
}

func formatRepoName(owner, name string) string {
	return fmt.Sprintf("%s-%s", owner, name)
}
