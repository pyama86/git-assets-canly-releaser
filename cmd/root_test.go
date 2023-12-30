package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/pyama86/git-assets-canaly-releaser/lib"
	redis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/mock"
)

func TestDeploy(t *testing.T) {
	tests := []struct {
		name    string
		cmd     string
		tag     string
		file    string
		wantErr bool
	}{
		{
			name:    "Success",
			cmd:     "../testdata/dummy.sh",
			tag:     "latest",
			file:    "assetfile",
			wantErr: false,
		},
		{
			name:    "Failure",
			cmd:     "../testdata/dummy.sh",
			tag:     "v1.0.0",
			file:    "testfile",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := deploy(tt.cmd, tt.tag, tt.file); (err != nil) != tt.wantErr {
				t.Errorf("deploy() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type GitHubMock struct {
	mock.Mock
}

func (m *GitHubMock) DownloadReleaseAsset(tag string) (string, string, error) {
	args := m.Called(tag)
	return args.String(0), args.String(1), args.Error(2)
}

func TestHandleRollout(t *testing.T) {
	tests := []struct {
		name         string
		githubMockFn func(*GitHubMock)
		wantErr      bool
		tag          string
	}{
		{
			name: "Success",
			githubMockFn: func(m *GitHubMock) {
				m.On("DownloadReleaseAsset", "latest").Return("latest", "assetfile", nil)
			},
			tag:     "latest",
			wantErr: false,
		},
		{
			name: "Failure",
			githubMockFn: func(m *GitHubMock) {
				m.On("DownloadReleaseAsset", "stable").Return("", "", errors.New("failed to download"))
			},
			tag:     "stable",
			wantErr: true,
		},
		{
			name:    "Avoid",
			tag:     "avoid",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			githubMock := new(GitHubMock)
			if tt.githubMockFn != nil {
				tt.githubMockFn(githubMock)
			}
			f, err := os.CreateTemp("", tt.name)
			if err != nil {
				t.Fatal(err)
			}
			os.Remove(f.Name())

			config := &lib.Config{
				Redis: &lib.RedisConfig{
					Host: os.Getenv("REDIS_HOST"),
					Port: 6379,
					DB:   0,
				},
				DeployCommand: "../testdata/dummy.sh",
				StateFilePath: f.Name(),
			}

			rc := redis.NewClient(&redis.Options{
				Addr:     fmt.Sprintf("%s:%d", config.Redis.Host, config.Redis.Port),
				Password: config.Redis.Password,
				DB:       config.Redis.DB,
			})
			rc.Set(context.Background(), "_stable_release_tag", tt.tag, 0)
			rc.SAdd(context.Background(), "_avoid_release_tag", "avoid", 0)

			if err := handleRollout(config, githubMock); (err != nil) != tt.wantErr {
				t.Errorf("handleRollout() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type CommandExecutorMock struct {
	mock.Mock
}

func (m *CommandExecutorMock) ExecuteCommand(command string, tag string, file string) ([]byte, error) {
	args := m.Called(command, tag, file)
	return args.Get(0).([]byte), args.Error(1)
}

func TestRunHealthCheck(t *testing.T) {
	tests := []struct {
		name    string
		tag     string
		file    string
		wantErr bool
	}{
		{
			name:    "Success",
			tag:     "latest",
			file:    "assetfile",
			wantErr: false,
		},
		{
			name:    "Failure",
			tag:     "v1.0.0",
			file:    "testfile",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &lib.Config{
				HealthCheckRetry:    1,
				HealthCheckCommand:  "../testdata/dummy.sh",
				HealthCheckInterval: 1 * time.Nanosecond,
				CanaryRolloutWindow: 3 * time.Nanosecond,
			}

			if err := runHealthCheck(config, tt.tag, tt.file); (err != nil) != tt.wantErr {
				t.Errorf("runHealthCheck() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type StateMock struct {
	mock.Mock
}

func (m *StateMock) TryCanaryReleaseLock(tag string) (bool, error) {
	args := m.Called(tag)
	return args.Bool(0), args.Error(1)
}

type LocalStateMock struct {
	mock.Mock
}

func (m *LocalStateMock) CanInstallTag(tag string) (bool, error) {
	args := m.Called(tag)
	return args.Bool(0), args.Error(1)
}

func TestHandleCanaryRelease(t *testing.T) {
	tests := []struct {
		name         string
		tag          string
		file         string
		githubMockFn func(*GitHubMock)
		wantErr      bool
	}{
		{
			name: "Success",
			githubMockFn: func(m *GitHubMock) {
				m.On("DownloadReleaseAsset", lib.LatestTag).Return("latest", "assetfile", nil)
			},
			tag:     "latest",
			file:    "assetfile",
			wantErr: false,
		},
		{
			name: "Failure",
			githubMockFn: func(m *GitHubMock) {
				m.On("DownloadReleaseAsset", lib.LatestTag).Return("", "", errors.New("failed to download"))
			},
			tag:     "v1.0.0",
			file:    "testfile",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			githubMock := new(GitHubMock)
			tt.githubMockFn(githubMock)
			config := &lib.Config{
				DeployCommand: "../testdata/dummy.sh",
				Redis: &lib.RedisConfig{
					Host: os.Getenv("REDIS_HOST"),
					Port: 6379,
					DB:   0,
				},
				HealthCheckRetry:    1,
				HealthCheckCommand:  "../testdata/dummy.sh",
				HealthCheckInterval: 1 * time.Nanosecond,
				CanaryRolloutWindow: 3 * time.Nanosecond,
				StateFilePath:       "../testdata/state.json",
			}
			if err := handleCanaryRelease(config, githubMock); (err != nil) != tt.wantErr {
				t.Errorf("handleCanaryRelease() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
