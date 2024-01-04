package lib

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/pyama86/git-assets-canary-releaser/testutils"
	"github.com/tj/assert"
)

func newTestConfig() *Config {
	redisHost := os.Getenv("GACR_REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost"
	}

	return &Config{
		Redis: &RedisConfig{
			Host:      redisHost,
			Port:      6379,
			Password:  "",
			DB:        0,
			KeyPrefix: "test_prefix",
		},
		Repo:                "test_repo",
		VersionCommand:      "echo v1.0.0",
		RolloutWindow:       time.Minute,
		CanaryRolloutWindow: 2 * time.Minute,
	}
}

func TestSaveMemberState(t *testing.T) {
	redisClient := testutils.RedisClient()
	state, err := NewState(newTestConfig())
	if err != nil {
		t.Fatalf("failed to setup test: %v", err)
	}

	// テストケースを実行
	err = state.SaveMemberState()
	// エラーがないことを確認します。
	assert.NoError(t, err)

	// Redisにデータが保存されたことを確認します。
	memberData, err := redisClient.Get(context.Background(), state.me).Result()
	assert.NoError(t, err)

	var ms MemberState
	err = json.Unmarshal([]byte(memberData), &ms)
	assert.NoError(t, err)

	assert.Equal(t, "v1.0.0", ms.CurrentVersion)
}

func TestGetRolloutProgress(t *testing.T) {
	state, err := NewState(newTestConfig())
	if err != nil {
		t.Fatalf("failed to setup test: %v", err)
	}

	err = state.SaveMemberState()
	assert.NoError(t, err)

	installed, all, err := state.GetRolloutProgress("v1.0.0")
	assert.NoError(t, err)
	assert.Equal(t, 1, installed)
	assert.Equal(t, 1, all)

	installed, all, err = state.GetRolloutProgress("non-existent-tag")
	assert.NoError(t, err)
	assert.Equal(t, 0, installed)
	assert.Equal(t, 1, all)
}
