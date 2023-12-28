package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

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
				m.On("DownloadReleaseAsset", "stable").Return("stable", "/path/to/stable", nil)
			},
			tag:     "stable",
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
					Host: "localhost",
					Port: 6379,
					DB:   0,
				},
				DeployCommand: "../scripts/deploy",
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
