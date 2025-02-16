package webhook

import "fmt"

// formatRepoName formats the repository name with owner prefix
func formatRepoName(owner, name string) string {
	return fmt.Sprintf("%s-%s", owner, name)
}

// getOwnerFromPath extracts the owner from a path with namespace (e.g., "owner/repo" -> "owner")
func getOwnerFromPath(path string) string {
	for i := 0; i < len(path); i++ {
		if path[i] == '/' {
			return path[:i]
		}
	}
	return path
}
