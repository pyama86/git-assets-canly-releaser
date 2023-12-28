# README for git-assets-canly-releaser
## Overview
The git-assets-canly-releaser is a tool designed to automate the deployment of the latest release assets from GitHub repositories. It facilitates a Canary release strategy where new software versions are rolled out incrementally to a subset of users before making it available to everyone. This helps in detecting and addressing any potential issues early in the release process.

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
./git-assets-canly-releaser --config path/to/your/config.toml
```

Here's an example of how to use command-line flags (note: not all options are covered):
```sh
./git-assets-canly-releaser \
  --repo "your-github-repo/name" \
  --github-token "your_github_token" \
  --deploy-command "/path/to/your/deploy/script" \
  --rollback-command "/path/to/your/rollback/script" \
  --healthcheck-command "/path/to/your/health/check/script"
```
