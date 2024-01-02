package cmd

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/pyama86/git-assets-canary-releaser/lib"
	"github.com/pyama86/git-assets-canary-releaser/testutils"
	redis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/mock"
	"github.com/tj/assert"
	"go.uber.org/mock/gomock"
)

// MockGitHuber is a mock type for the GitHuber interface
type MockGitHuber struct {
	mock.Mock
}

// DownloadReleaseAsset mocks the DownloadReleaseAsset method
func (m *MockGitHuber) DownloadReleaseAsset(tag string) (string, string, error) {
	args := m.Called(tag)
	return args.String(0), args.String(1), args.Error(2)
}

func TestDeploy(t *testing.T) {
	tests := []struct {
		name      string
		cmd       string
		tag       string
		mockSetup func(*MockGitHuber)
		wantTag   string
		wantFile  string
		wantErr   bool
	}{
		{
			name: "Successful deployment",
			cmd:  "../testdata/dummy.sh",
			tag:  "latest",
			mockSetup: func(m *MockGitHuber) {
				m.On("DownloadReleaseAsset", "latest").Return("latest", "assetfile", nil)
			},
			wantTag:  "latest",
			wantFile: "assetfile",
			wantErr:  false,
		},
		{
			name: "Failed to download asset",
			cmd:  "echo",
			tag:  "v1.0.0",
			mockSetup: func(m *MockGitHuber) {
				m.On("DownloadReleaseAsset", "v1.0.0").Return("", "", errors.New("download error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGitHub := new(MockGitHuber)
			tt.mockSetup(mockGitHub)

			tag, file, err := deploy(tt.cmd, tt.tag, mockGitHub)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantTag, tag)
				assert.Equal(t, tt.wantFile, file)
			}

			mockGitHub.AssertExpectations(t)
		})
	}
}

func TestHandleRollout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	redisClient := testutils.RedisClient()
	testCases := []struct {
		name          string
		mockSetup     func(*MockGitHuber)
		expectedError bool
		wantError     error
		before        func(redisClient *redis.Client)
	}{
		{
			name: "Successful Rollout",
			mockSetup: func(m *MockGitHuber) {
				m.On("DownloadReleaseAsset", "latest").Return("latest", "assetfile", nil)
			},
			expectedError: false,
			before: func(redisClient *redis.Client) {
				redisClient.Set(context.Background(), "foo/bar_stable_release_tag", "latest", 0)
				redisClient.Del(context.Background(), "foo/bar_avoid_release_tag")
				os.Setenv("TEST_VERSION", "stable")
			},
		},
		{
			name: "Already installed",
			mockSetup: func(m *MockGitHuber) {
				m.On("DownloadReleaseAsset", "latest").Return("latest", "assetfile", nil)
			},
			expectedError: true,
			before: func(redisClient *redis.Client) {
				redisClient.Set(context.Background(), "foo/bar_stable_release_tag", "already_installed", 0)

				os.Setenv("TEST_VERSION", "already_installed")
			},
			wantError: lib.ErrAlreadyInstalled,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			redisHost := os.Getenv("GACR_REDIS_HOST")
			if redisHost == "" {
				redisHost = "localhost"
			}
			config := &lib.Config{
				Repo: "foo/bar",
				Redis: &lib.RedisConfig{
					Host: redisHost,
					Port: 6379,
				},
				DeployCommand:  "../testdata/dummy.sh",
				VersionCommand: "../testdata/echo_version.sh",
			}

			state, err := lib.NewState(config)
			assert.NoError(t, err)
			mockGitHub := new(MockGitHuber)
			tc.mockSetup(mockGitHub)
			tc.before(redisClient)

			err = handleRollout(config, mockGitHub, state)
			if tc.expectedError {
				assert.Error(t, err)
				if tc.wantError != nil {
					assert.Equal(t, tc.wantError, err)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandleCanaryRollout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	redisClient := testutils.RedisClient()
	testCases := []struct {
		name               string
		mockSetup          func(*MockGitHuber)
		expectedError      bool
		wantError          error
		healthCheckCommand string
		rollbackCommand    string
		before             func(redisClient *redis.Client)
	}{
		{
			name: "Successful Rollout",
			mockSetup: func(m *MockGitHuber) {
				m.On("DownloadReleaseAsset", "latest").Return("latest", "assetfile", nil)
			},
			expectedError: false,
			before: func(redisClient *redis.Client) {
				redisClient.Set(context.Background(), "foo/bar_stable_release_tag", "stable", 0)
				redisClient.Del(context.Background(), "foo/bar_avoid_release_tag")
				os.Setenv("TEST_VERSION", "notinstalled")
			},
		},
		{
			name: "Already installed",
			mockSetup: func(m *MockGitHuber) {
				m.On("DownloadReleaseAsset", "latest").Return("latest", "assetfile", nil)
			},
			expectedError: true,
			before: func(redisClient *redis.Client) {
				redisClient.Set(context.Background(), "foo/bar_stable_release_tag", "nomatch", 0)
				redisClient.Del(context.Background(), "foo/bar_avoid_release_tag")
				os.Setenv("TEST_VERSION", "latest")
			},
			wantError: lib.ErrAlreadyInstalled,
		},
		{
			name: "Rollback",
			mockSetup: func(m *MockGitHuber) {
				m.On("DownloadReleaseAsset", "latest").Return("latest", "assetfile", nil)
				m.On("DownloadReleaseAsset", "rollback").Return("stable", "assetfile", nil)
			},
			expectedError: true,
			before: func(redisClient *redis.Client) {
				redisClient.Del(context.Background(), "foo/bar_avoid_release_tag")
				redisClient.Set(context.Background(), "foo/bar_stable_release_tag", "stable", 0)
				os.Setenv("TEST_VERSION", "rollback")
			},
			healthCheckCommand: "../testdata/always_fail.sh",
			rollbackCommand:    "../testdata/always_succes.sh",
			wantError:          ErrRollback,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			healthCheckCommand := "../testdata/dummy.sh"
			if tc.healthCheckCommand != "" {
				healthCheckCommand = tc.healthCheckCommand
			}
			rollbackCommand := "../testdata/dummy.sh"
			if tc.rollbackCommand != "" {
				rollbackCommand = tc.rollbackCommand
			}

			redisHost := os.Getenv("GACR_REDIS_HOST")
			if redisHost == "" {
				redisHost = "localhost"
			}
			config := &lib.Config{
				Repo: "foo/bar",
				Redis: &lib.RedisConfig{
					Host: redisHost,
					Port: 6379,
				},
				DeployCommand:       "../testdata/dummy.sh",
				VersionCommand:      "../testdata/echo_version.sh",
				HealthCheckCommand:  healthCheckCommand,
				RollbackCommand:     rollbackCommand,
				HealthCheckInterval: time.Nanosecond,
				HealthCheckTimeout:  time.Second,
				HealthCheckRetries:  1,
				CanaryRolloutWindow: time.Nanosecond,
			}

			state, err := lib.NewState(config)
			assert.NoError(t, err)
			mockGitHub := new(MockGitHuber)
			tc.mockSetup(mockGitHub)
			tc.before(redisClient)

			err = handleCanaryRelease(config, mockGitHub, state)
			if tc.expectedError {
				assert.Error(t, err)
				if tc.wantError != nil {
					assert.Equal(t, tc.wantError, err)
				}
			} else {
				assert.NoError(t, err)

				stableTag, err := redisClient.Get(context.Background(), "foo/bar_stable_release_tag").Result()
				assert.NoError(t, err)
				assert.Equal(t, "latest", stableTag)

				_, err = redisClient.Get(context.Background(), "foo/bar_canary_release_tag").Result()
				assert.Error(t, err)
			}

		})
	}
}
