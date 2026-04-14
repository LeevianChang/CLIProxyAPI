# Deployment Guide

This guide explains how to build and deploy your own Docker images to a remote server using GitHub Actions.

## Prerequisites

1. A Docker Hub account (or other container registry like GitHub Container Registry)
2. A remote server with Docker and Docker Compose installed
3. GitHub repository with Actions enabled

## Local Development

### Quick Start with Docker Compose

For local development, use `docker-compose.local.yml` which builds locally without pulling from a registry:

```bash
# Build and start (uses local Dockerfile)
docker-compose -f docker-compose.local.yml up --build -d

# View logs
docker-compose -f docker-compose.local.yml logs -f

# Stop
docker-compose -f docker-compose.local.yml down
```

### Local Development Without Docker Compose

You can also run directly with Go:

```bash
# Build
go build -o cli-proxy-api ./cmd/server

# Run
./cli-proxy-api --config config.yaml
```

## Step 1: Configure GitHub Container Registry

GitHub Container Registry (ghcr.io) is automatically available for your repository. No additional setup needed!

The workflow uses `GITHUB_TOKEN` which is automatically provided by GitHub Actions.

## Step 2: Enable Package Permissions (Optional)

If you want the package to be public:

1. Go to your GitHub repository
2. After first push, go to Packages (on your profile or repo)
3. Find your package and click it
4. Go to Package settings
5. Change visibility to Public (if desired)

## Step 3: Update Configuration Files

### 3.1 Verify Configuration

The configuration is already set to use GitHub Container Registry:

**docker-compose.yml:**
```yaml
image: ${CLI_PROXY_IMAGE:-ghcr.io/your-github-username/cliproxyapi:latest}
```

**GitHub Actions workflow:**
```yaml
REGISTRY: ghcr.io
IMAGE_NAME: ${{ github.repository_owner }}/cliproxyapi
```

No manual changes needed - it uses your GitHub username automatically!

## Step 4: Trigger Build

### Option A: Push a Git Tag (Recommended)

```bash
git tag v1.0.0
git push origin v1.0.0
```

This will trigger the GitHub Actions workflow to build and push multi-arch images (amd64 and arm64) to GitHub Container Registry.

### Option B: Manual Workflow Dispatch

1. Go to GitHub → Actions → docker-image workflow
2. Click "Run workflow"
3. Select branch and run

## Step 5: Deploy to Remote Server

### 5.1 Prepare Remote Server

SSH into your server and create the deployment directory:

```bash
mkdir -p ~/cli-proxy-api
cd ~/cli-proxy-api
```

### 5.2 Copy Configuration Files

Copy these files to your server:

```bash
# On your local machine
scp docker-compose.yml user@your-server:~/cli-proxy-api/
scp config.example.yaml user@your-server:~/cli-proxy-api/config.yaml
```

### 5.3 Create Required Directories

```bash
# On your server
mkdir -p auths logs
```

### 5.4 Configure Environment Variables (Optional)

Create a `.env` file on your server:

```bash
cat > .env << 'EOF'
# Pull from registry on server (default behavior)
PULL_POLICY=always

# Use your published image from GitHub Container Registry
CLI_PROXY_IMAGE=ghcr.io/your-github-username/cliproxyapi:latest
VERSION=v1.0.0

# Configuration paths
CLI_PROXY_CONFIG_PATH=./config.yaml
CLI_PROXY_AUTH_PATH=./auths
CLI_PROXY_LOG_PATH=./logs
EOF
```

If the package is private, you'll need to login on your server:
```bash
echo $GITHUB_TOKEN | docker login ghcr.io -u your-github-username --password-stdin
```

### 5.5 Deploy

```bash
# Pull the latest image and start
docker-compose pull
docker-compose up -d

# Check logs
docker-compose logs -f

# Update to latest version
docker-compose pull && docker-compose up -d
```

## Step 6: Automatic Deployment (Optional)

### Option A: Using Watchtower

Watchtower automatically updates running containers when new images are available:

```bash
docker run -d \
  --name watchtower \
  -v /var/run/docker.sock:/var/run/docker.sock \
  containrrr/watchtower \
  --interval 300 \
  cli-proxy-api
```

### Option B: Using GitHub Actions with SSH

Add a deployment step to `.github/workflows/docker-image.yml`:

```yaml
  deploy:
    runs-on: ubuntu-latest
    needs: docker_manifest
    steps:
      - name: Deploy to Remote Server
        uses: appleboy/ssh-action@v1.0.0
        with:
          host: ${{ secrets.SERVER_HOST }}
          username: ${{ secrets.SERVER_USER }}
          key: ${{ secrets.SERVER_SSH_KEY }}
          script: |
            cd ~/cli-proxy-api
            docker-compose pull
            docker-compose up -d
            docker-compose logs --tail=50
```

Add these secrets to GitHub:
- `SERVER_HOST`: Your server IP or domain
- `SERVER_USER`: SSH username
- `SERVER_SSH_KEY`: Private SSH key for authentication

## Using Docker Hub Instead (Alternative)

If you prefer Docker Hub over GitHub Container Registry:

1. Register at https://hub.docker.com
2. Create a repository named `cli-proxy-api`
3. Get an access token from Account Settings → Security
4. Add GitHub secrets: `DOCKERHUB_USERNAME` and `DOCKERHUB_TOKEN`
5. Update configurations to use `username/cli-proxy-api:latest` format

## Troubleshooting

### Image Pull Failed

For private packages, login to GitHub Container Registry on your server:

```bash
# Create a GitHub Personal Access Token with read:packages scope
# Then login:
echo $GITHUB_TOKEN | docker login ghcr.io -u your-github-username --password-stdin
```

### Permission Denied

```bash
# Fix volume permissions
sudo chown -R $USER:$USER auths logs
```

### Container Won't Start

```bash
# Check logs
docker-compose logs cli-proxy-api

# Check container status
docker-compose ps
```

### Port Already in Use

Edit `docker-compose.yml` and change the port mappings:

```yaml
ports:
  - "8318:8317"  # Change 8317 to 8318 or any available port
```

## Maintenance

### Update to Latest Version

```bash
docker-compose pull
docker-compose up -d
```

### View Logs

```bash
docker-compose logs -f cli-proxy-api
```

### Restart Service

```bash
docker-compose restart
```

### Stop Service

```bash
docker-compose down
```

### Backup Configuration

```bash
tar -czf backup-$(date +%Y%m%d).tar.gz config.yaml auths/
```

## Security Recommendations

1. Use environment variables for sensitive data
2. Enable firewall and only expose necessary ports
3. Use HTTPS with reverse proxy (nginx/caddy)
4. Regularly update Docker images
5. Monitor logs for suspicious activity
6. Use strong management passwords
7. Restrict Docker Hub/GHCR access tokens to minimum required permissions
