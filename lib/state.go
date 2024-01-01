package lib

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"time"

	redis "github.com/redis/go-redis/v9"
)

type State struct {
	client              *redis.Client
	canaryReleaseTagKey string
	stableReleaseTagKey string
	avoidReleaseTagKey  string
	rolloutKey          string
	config              *Config

	// localstate
	LastInstalledTag string `json:"last_installed_tag"`
	stateFilePath    string
}

func NewState(config *Config) (*State, error) {
	rc := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Redis.Host, config.Redis.Port),
		Password: config.Redis.Password,
		DB:       config.Redis.DB,
	})

	if err := rc.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to create redis client: %s", err)
	}
	return &State{
		client:              rc,
		config:              config,
		canaryReleaseTagKey: fmt.Sprintf("%s_canary_release_tag", config.Repo),
		stableReleaseTagKey: fmt.Sprintf("%s_stable_release_tag", config.Repo),
		avoidReleaseTagKey:  fmt.Sprintf("%s_avoid_release_tag", config.Repo),
		rolloutKey:          fmt.Sprintf("%s_rollout", config.Repo),
		stateFilePath:       config.StateFilePath,
	}, nil
}

func (s *State) UnlockCanaryRelease() error {
	return s.client.Del(context.Background(), s.canaryReleaseTagKey).Err()
}

func (s *State) TryCanaryReleaseLock(tag string) (bool, error) {
	return s.getLock(s.canaryReleaseTagKey, tag, s.config.CanaryRolloutWindow*2)
}

func (s *State) TryRolloutLock(tag string) (bool, error) {
	return s.getLock(s.rolloutKey, tag, s.config.RolloutWindow)
}

func (s *State) getLock(key string, tag string, window time.Duration) (bool, error) {
	ok, err := s.client.SetNX(context.Background(), key, tag, 0).Result()
	if err != nil {
		return false, err
	}
	if ok {
		err := s.client.Expire(context.Background(), key, window).Err()
		if err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}
func (s *State) CurrentStableTag() (string, error) {
	return s.getRelease(s.stableReleaseTagKey)
}

var ErrAvoidReleaseTag = errors.New("avoid release tag")

func (s *State) IsAvoidReleaseTag(tag string) error {
	return nil

}

func (s *State) saveRelease(key, tag string) error {
	return s.client.Set(context.Background(), key, tag, 0).Err()
}

func (s *State) saveReleases(key string, tags ...string) error {
	return s.client.SAdd(context.Background(), key, tags).Err()
}

func (s *State) SaveStableReleaseTag(tag string) error {
	return s.saveRelease(s.stableReleaseTagKey, tag)
}

func (s *State) SaveAvoidReleaseTag(tag string) error {
	return s.saveReleases(s.avoidReleaseTagKey, tag)
}

func (s *State) getRelease(key string) (string, error) {
	v, err := s.client.Get(context.Background(), key).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return v, nil
}

func (s *State) getReleases(key string) ([]string, error) {
	return s.client.SMembers(context.Background(), key).Result()
}

func (s *State) SaveLastInstalledTag(tag string) error {
	s.LastInstalledTag = tag
	return s.saveLocalState()
}

var ErrAlreadyInstalled = errors.New("already installed")

func (s *State) CanInstallTag(tag string) error {
	if tag == "" {
		return errors.New("tag is empty")
	}

	lastInstalledTag, err := s.getLastInstalledTag()
	if err != nil {
		return err
	}
	if lastInstalledTag == "" {
		return nil
	}

	slog.Debug("tags", "lastInstalledTag", lastInstalledTag, "tag", tag)
	if lastInstalledTag == tag {
		return ErrAlreadyInstalled
	}

	tags, err := s.getReleases(s.avoidReleaseTagKey)
	if err != nil {
		return err
	}
	for _, t := range tags {
		if t == tag {
			return ErrAvoidReleaseTag
		}
	}

	return nil
}

func (s *State) getLastInstalledTag() (string, error) {
	if s.config.CurrentVersionCommand != "" {
		out, err := exec.Command(s.config.CurrentVersionCommand).Output()
		if err != nil {
			return "", err
		}
		return strings.TrimRight(strings.TrimSpace(string(out)), "\n"), nil
	}
	if err := s.readLocalState(); err != nil {
		return "", err
	}
	return s.LastInstalledTag, nil
}

func (s *State) saveLocalState() error {
	dir := s.stateFilePath[:len(s.stateFilePath)-len(s.stateFilePath[strings.LastIndex(s.stateFilePath, "/"):])]
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	f, err := os.OpenFile(s.stateFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}

	defer f.Close()
	return json.NewEncoder(f).Encode(s)

}

func (s *State) readLocalState() error {
	if _, err := os.Stat(s.stateFilePath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}

	f, err := os.Open(s.stateFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	err = json.NewDecoder(f).Decode(s)
	if err != nil {
		return err
	}
	return nil
}

func (s *State) RollbackTag() (string, error) {
	stableRelease, err := s.CurrentStableTag()
	if err != nil {
		return "", err
	}

	rollbackTag := stableRelease
	if rollbackTag == "" {
		rollbackTag = s.LastInstalledTag
	}
	if rollbackTag == "" {
		return "", fmt.Errorf("can't decided rollback tag")
	}
	return rollbackTag, nil
}
