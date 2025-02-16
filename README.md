# Gitcloner

Ease of mind when it comes to your data safety. Easily mirror your repositories to a different provider. 

By hooking Gitcloner as a global webhook into your source provider, it will start listening for events. When it detects a new repository, it will create a mirror of your repository on the destination provider. It will also keep the mirror up to date when there are push events to the original repository, or when the mirror repository is updated.

## Features

- Automatically creates mirror repositories when new repositories are created
- Syncs repositories on push events to the default branch
- One-time import of repositories using the CLI
- Supports mirroring to:
  - Gitea
  - GitHub
  - GitLab
- Prefixes mirrored repositories with original owner name
- Handles private repositories with authentication
- Docker support for easy deployment
- Automatically updates mirrors when the original repository is updated

## Usage

### Webhook Server

The default mode runs a webhook server that listens for repository events and automatically creates/updates mirrors.

```bash
./gitcloner
```
## Deployment Options

### Using Docker (Recommended)

1. Pull the image from GitHub Container Registry:
   ```bash
   docker pull ghcr.io/janyksteenbeek/gitcloner:latest
   ```

2. Create a `.env` file with your configuration:
   ```bash
   cp .env.example .env
   ```

3. Run with Docker Compose:
   ```bash
   docker-compose up -d
   ```

### Manual Deployment

1. Clone the repository:
   ```bash
   git clone https://github.com/janyksteenbeek/gitcloner.git
   cd gitcloner
   ```

2. Copy and configure environment variables:
   ```bash
   cp .env.example .env
   ```

3. Build and run:
   ```bash
   go build ./cmd/gitcloner
   ./gitcloner
   ```

Install Supervisor to run the service in the background. More info [here](https://google.com/search?q=How+to+install+supervisord).

## Configuration

### Environment Variables

- `PORT`: The port the webhook server will listen on (default: 8080)
- `DESTINATION_TYPE`: Either "gitea", "github", or "gitlab"
- `DESTINATION_URL`: The URL of your destination instance
- `DESTINATION_TOKEN`: API token with repository creation permissions
- `DESTINATION_ORG`: The organization/owner name where mirrors will be created
- `SOURCE_TOKEN`: Token for accessing private source repositories
- `ALWAYS_PUSH`: Whether to push to the destination even if the mirror already exists. By default, this is ommited.

### Webhook Configuration

#### Gitea
In Gitea organization settings, add a Gitea webhook with:
- URL: `http://your-server:8080/webhook`
- Method: POST 
- Events: Repository Created, Push
- Branch filter: *

#### GitHub
In GitHub organization settings, add webhook with:
- URL: `http://your-server:8080/webhook`
- Content type: `application/json`
- Events: Repository, Push

#### GitLab
In GitLab group settings, add webhook with:
- URL: `http://your-server:8080/webhook`
- Triggers: Project events, Push events
- SSL verification: Optional

## One-time Import

You can use Gitcloner to import one or more repositories using the CLI:

```bash
# Mirror multiple GitHub repositories
./gitcloner --import "github octocat/Hello-World,golang/go,kubernetes/kubernetes"

# Mirror a GitLab repository
./gitcloner --import "gitlab gitlab-org/gitlab"
```

The import command requires:
1. The `--import` flag
1. Platform name (`github`, `gitlab`, or `gitea`)
3. One or more repository paths in `username/repo` format, separated by commas

Note: When importing multiple repositories, if one import fails, the process will continue with the remaining repositories.

Make sure you have valid environment variables in the `.env` file


## Repository Naming

Mirrored repositories are created with the format: `originalOwner-repoName`
Example:
- Original: `janyksteenbeek/myrepo`
- Mirrored: `yourbackuporg/janyksteenbeek-myrepo`

### Private Access Tokens

For private repositories, you need to set the `SOURCE_TOKEN` environment variable. This token needs to have access to the private repositories you want to mirror.

Make sure the token has the necessary permissions for the source provider. Read more about the permissions needed in the documentation of the source provider. Most of the times, it's just a matter of read/write access to the repositories and organizations & users scopes.

## License

Licensed under the MIT License.