package lib

import "time"

type RedisConfig struct {
	Host     string `mapstracture:"host" validate:"required"`
	Port     int    `mapstracture:"port" validate:"required"`
	Password string `mapstracture:"password"`
	DB       int    `mapstracture:"db" validate:"required"`
}
type Config struct {
	GitHubToken              string        `mapstracture:"github_token" validate:"required"`
	Repo                     string        `mapstracture:"repo" validate:"required"`
	AssetsDownloadPath       string        `mapstracture:"assets_download_path" validate:"required"`
	GitHubAPIEndpoint        string        `mapstracture:"github_api"`
	DeployCommand            string        `mapstracture:"deploy_command"  validate:"required"`
	RollbackCommand          string        `mapstracture:"rollback_command" validate:"required"`
	HealthCheckCommand       string        `mapstracture:"healthcheck_command" validate:"required"`
	HealthCheckInterval      time.Duration `mapstracture:"healthcheck_interval" validate:"required"`
	CanaryRolloutWindow      time.Duration `mapstracture:"canary_rollout_window" validate:"required"`
	RolloutWindow            time.Duration `mapstracture:"rollout_window" validate:"required"`
	RepositryPollingInterval time.Duration `mapstracture:"repository_polling_interval" validate:"required"`
	PackageNamePattern       string        `mapstracture:"package_name_pattern" validate:"required"`
	CanaryReleaseVersionKey  string        `mapstracture:"canary_release_version_key" validate:"required"`
	StableReleaseVersionKey  string        `mapstracture:"stable_release_version_key" validate:"required"`
	AvoidReleaseVersionKey   string        `mapstracture:"avoid_release_version_key" validate:"required"`
	RolloutKey               string        `mapstracture:"rollout_key" validate:"required"`
	SlackWebhookURL          string        `mapstracture:"slack_webhook_url"`
	Redis                    *RedisConfig  `mapstracture:"redis" validate:"required"`
	LogLevel                 string        `mapstracture:"log_level"`
}
