# Gitcloner

Ease of mind when it comes to your data safety. Easily mirror your repositories to a different provider. 

By hooking Gitcloner as a global webhook into your source provider, it will start listening for events. When it detects a new repository, it will create a mirror of your repository on the destination provider. It will also keep the mirror up to date when there are push events to the original repository, or when the mirror repository is updated.

## Features

- Automatically creates mirror repositories when new repositories are created
- Syncs repositories on push events to the default branch
- Supports mirroring to:
  - Gitea
  - GitHub
  - GitLab
- Prefixes mirrored repositories with original owner name
- Handles private repositories with authentication
- Docker support for easy deployment
- Automatically updates mirrors when the original repository is updated

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

### Private Access Tokens

For private repositories, you need to set the `SOURCE_TOKEN` environment variable. This token needs to have access to the private repositories you want to mirror.

Make sure the token has the necessary permissions for the source provider. Read more about the permissions needed in the documentation of the source provider. Most of the times, it's just a matter of read/write access to the repositories and organizations & users scopes.

### Webhook Configuration

#### Gitea
1. In your source Gitea instance, go to your organization's settings
2. Navigate to Webhooks > Add Webhook > Gitea
3. Set the following:
   - Target URL: `http://your-server:8080/webhook`
   - HTTP Method: POST
   - Trigger On: Repository Events and Push Events
   - Branch filter: * (or your default branch)
   - Select events:
     - Repository Created
     - Push

#### GitHub
1. In your source GitHub organization, go to Settings > Webhooks
2. Click "Add webhook"
3. Set the following:
   - Payload URL: `http://your-server:8080/webhook`
   - Content type: `application/json`
   - Secret: _(optional, not yet supported)_
   - Events:
     - Repository
     - Push

#### GitLab
1. In your source GitLab group, go to Settings > Webhooks
2. Set the following:
   - URL: `http://your-server:8080/webhook`
   - Trigger:
     - Project events
     - Push events
   - SSL verification: (according to your setup)

## Repository Naming

Mirrored repositories are created with the format: `originalOwner-repoName`
Example:
- Original: `janyksteenbeek/myrepo`
- Mirrored: `yourbackuporg/janyksteenbeek-myrepo`

## Private Repositories

For private repositories, ensure you:
1. Set the `SOURCE_TOKEN` environment variable
2. Use a token with sufficient permissions in both source and destination

## License

Licensed under the MIT License. See the [LICENSE file](LICENSE) for details.