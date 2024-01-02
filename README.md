# README for git-assets-canary-releaser
## Overview
The git-assets-canary-releaser is a tool designed to automate the deployment of the latest release assets from GitHub repositories. It facilitates a Canary release strategy where new software versions are rolled out incrementally to a subset of users before making it available to everyone. This helps in detecting and addressing any potential issues early in the release process.

```mermaid
graph TD
    A[Download Assets from GitHub] --> B[Canary Release]
    B --> C[Perform Health Check]
    C -->|Health Check OK| D[Full Release]
    C -->|Health Check Fail| E[Rollback]
    D --> F[Deployment Complete]
    E --> G[Deployment Rollback]
```

## Features
- Automatically downloads and deploys latest release assets from GitHub.
- Implements Canary Release strategy with health checks and rollback functionality.
- Configurable deployment, health check, and rollback commands.
- Supports locking mechanisms to control the rollout process.
- Customizable logging level, asset download paths, and release timings.
- Utilizes Redis for managing release states and locks.

## Prerequisites
To use this command-line tool, you will need:
- Access to a GitHub repository with release assets.
- A GitHub token with permissions to access the repository.
- A deployment environment with Redis installed and configured.
- Go programming language environment to build the application.

## Configuration
Configuration can be done via command-line flags or a TOML configuration file. You can specify the configuration file using the --config flag when running the command-line tool. The following configurations need to be specified or customized:
- GitHub repository information (name, API endpoint, token).
- Commands for deploying, rolling back, and performing health checks.
- Redis connection information (host, port, password, and DB).
- Logging and Slack webhook settings.
- Release asset download path and pattern.
- Canary rollout window and health check intervals.

## Usage
After building the tool, you can run it with the appropriate flags or configuration file. If you're using a configuration file, it might look like this:

```sh
./git-assets-canary-releaser --config path/to/your/config.toml
```

Here's an example of how to use command-line flags (note: not all options are covered):
```sh
./git-assets-canary-releaser \
  --repo "your-github-repo/name" \
  --github-token "your_github_token" \
  --deploy-command "/path/to/your/deploy/script" \
  --rollback-command "/path/to/your/rollback/script" \
  --healthcheck-command "/path/to/your/health/check/script"
  --version-command "/path/to/your/version/script"
```

## Command-Line Arguments

- `--config`: Specifies the path to the configuration file. Default is `$HOME/gacr.conf`.
- `--repo`: Sets the GitHub repository name.
- `--github-token`: Specifies the GitHub token for authentication.(env:GITHUB_TOKEN)
- `--github-api`: Sets the GitHub API endpoint. Default is `https://api.github.com`.(env:GITHUB_API_URL)

- `--deploy-command`: Defines the command for deployment.
- `--rollback-command`: Specifies the command for rollback operations.
- `--healthcheck-command`: Sets the command for health checks.
- `--version-command`: Defines the command to check the current version.
- `--slack-webhook-url`: Sets the Slack webhook URL for notifications.
- `--slack-channel`: Specifies the Slack channel for notifications.
- `--redis-host`: Defines the Redis host. Default is `127.0.0.1`.
- `--redis-port`: Sets the Redis port. Default is `6379`.
- `--redis-password`: Specifies the Redis password.
- `--redis-db`: Sets the Redis database number. Default is `1`.
- `--redis-key-prefix`: Defines the Redis key prefix. Default is the repository name.
- `--package-name-pattern`: Sets the package name pattern.
- `--log-level`: Specifies the log level. Default is `info`.
- `--save-assets-path`: Defines the path to save downloaded assets. Default is `/usr/local/src`.
- `--canary-rollout-window`: Sets the time window for the canary release rollout. Default is `5 minutes`.
- `--rollout-window`: Specifies the time window for the release rollout. Default is `1 minute`.
- `--health-check-interval`: Sets the interval for health checks. Default is `1 minute`.
- `--repository-polling-interval`: Defines the interval for repository polling. Default is `5 minutes`.
- `--once`: Enables one-shot mode. The application exits after one execution cycle.
- `--healthcheck-retries`: Sets the number of retries for health checks. Default is `3`.
- `--healthcheck-timeout`: Specifies the timeout for health checks. Default is `30 seconds`.

## Configuration File (TOML Format)

```toml
# GitHub token for authentication
github_token = "your_github_token"

# GitHub repository name
repo = "user/repository"

# Path to save downloaded assets
save_assets_path = "/path/to/save/assets"

# GitHub API endpoint
github_api = "https://api.github.com"

# Command for deployment
deploy_command = "deploy_script.sh"

# Command for rollback operations
rollback_command = "rollback_script.sh"

# Command for health checks
healthcheck_command = "health_check_script.sh"

# Command to check the current version
version_command = "version_check_script.sh"

# Interval for health checks
healthcheck_interval = "1m"

# Time window for canary release rollout
canary_rollout_window = "5m"

# Time window for release rollout
rollout_window = "1m"

# Interval for repository polling
repository_polling_interval = "5m"

# Package name pattern
package_name_pattern = "pattern"

# Slack webhook URL for notifications
slack_webhook_url = "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX"

# Slack channel for notifications
slack_channel = "#channel"

# Redis configuration
[redis]
  host = "127.0.0.1"
  port = 6379
  password = "password"
  db = 1
  key_prefix = "prefix"

# Log level
log_level = "info"

# Retry count of health check
healthcheck_retries = 3

# Timeout of health check
healthcheck_timeout = "30s"
```

## Available Environment Variables

- `GACR_CONFIG`: Path to the configuration file. Overrides `--config` argument. Default is `$HOME/gacr.conf`.
- `GACR_REPO`: Sets the GitHub repository name. Overrides `--repo` argument.
- `GACR_GITHUB_TOKEN`: Specifies the GitHub token for authentication. Overrides `--github-token` argument.
- `GACR_GITHUB_API`: Sets the GitHub API endpoint. Overrides `--github-api` argument. Default is `https://api.github.com`.
- `GACR_DEPLOY_COMMAND`: Defines the command for deployment. Overrides `--deploy-command` argument.
- `GACR_ROLLBACK_COMMAND`: Specifies the command for rollback operations. Overrides `--rollback-command` argument.
- `GACR_HEALTHCHECK_COMMAND`: Sets the command for health checks. Overrides `--healthcheck-command` argument.
- `GACR_VERSION_COMMAND`: Defines the command to check the current version. Overrides `--version-command` argument.
- `GACR_SLACK_WEBHOOK_URL`: Sets the Slack webhook URL for notifications. Overrides `--slack-webhook-url` argument.
- `GACR_SLACK_CHANNEL`: Specifies the Slack channel for notifications. Overrides `--slack-channel` argument.
- `GACR_REDIS_HOST`: Defines the Redis host. Overrides `--redis-host` argument. Default is `127.0.0.1`.
- `GACR_REDIS_PORT`: Sets the Redis port. Overrides `--redis-port` argument. Default is `6379`.
- `GACR_REDIS_PASSWORD`: Specifies the Redis password. Overrides `--redis-password` argument.
- `GACR_REDIS_DB`: Sets the Redis database number. Overrides `--redis-db` argument. Default is `1`.
- `GACR_REDIS_KEY_PREFIX`: Defines the Redis key prefix. Overrides `--redis-key-prefix` argument. Default is the repository name.
- `GACR_PACKAGE_NAME_PATTERN`: Sets the package name pattern. Overrides `--package-name-pattern` argument.
- `GACR_LOG_LEVEL`: Specifies the log level. Overrides `--log-level` argument. Default is `info`.
- `GACR_SAVE_ASSETS_PATH`: Defines the path to save downloaded assets. Overrides `--save-assets-path` argument. Default is `/usr/local/src`.
- `GACR_CANARY_ROLLOUT_WINDOW`: Sets the time window for the canary release rollout. Overrides `--canary-rollout-window` argument. Default is `5 minutes`.
- `GACR_ROLLOUT_WINDOW`: Specifies the time window for the release rollout. Overrides `--rollout-window` argument. Default is `1 minute`.
- `GACR_HEALTH_CHECK_INTERVAL`: Sets the interval for health checks. Overrides `--health-check-interval` argument. Default is `1 minute`.
- `GACR_REPOSITORY_POLLING_INTERVAL`: Defines the interval for repository polling. Overrides `--repository-polling-interval` argument. Default is `5 minutes`.
- `GACR_ONCE`: Enables one-shot mode. Overrides `--once` argument. The application exits after one execution cycle.
- `GACR_HEALTHCHECK_RETRIES`: Sets the number of retries for health checks. Overrides `--healthcheck-retries` argument. Default is `3`.
- `GACR_HEALTHCHECK_TIMEOUT`: Specifies the timeout for health checks. Overrides `--healthcheck-timeout` argument. Default is `30 seconds`.

## example
The example of using docker-compose can be checked with the following command:

```bash
$ make run_example
```
