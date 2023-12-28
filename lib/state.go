package lib

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
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
		canaryReleaseTagKey: fmt.Sprintf("%s_canary_release_version", config.Repo),
		stableReleaseTagKey: fmt.Sprintf("%s_stable_release_version", config.Repo),
		avoidReleaseTagKey:  fmt.Sprintf("%s_avoid_release_version", config.Repo),
		rolloutKey:          fmt.Sprintf("%s_rollout", config.Repo),
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

func (s *State) IsAvoidReleaseTag(tag string) (bool, error) {
	tags, err := s.getReleases(s.avoidReleaseTagKey)
	if err != nil {
		return false, err
	}
	for _, t := range tags {
		if t == tag {
			return true, nil
		}
	}
	return false, nil

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
	return s.client.Get(context.Background(), key).Result()
}

func (s *State) getReleases(key string) ([]string, error) {
	return s.client.SMembers(context.Background(), key).Result()
}

func (s *State) getStableReleaseTag() (string, error) {
	return s.getRelease(s.stableReleaseTagKey)
}

func (s *State) getAvoidReleaseTag() ([]string, error) {
	return s.getReleases(s.avoidReleaseTagKey)
}

const StateFilePath = "/var/lib/gacr/state.json"

type LocalState struct {
	LastInstalledTag string `json:"last_installed_tag"`
	stateFilePath    string
}

func NewLocalState(f string) (*LocalState, error) {
	return &LocalState{
		stateFilePath: f,
	}, nil
}

func (s *LocalState) SaveLastInstalledTag(tag string) error {
	s.LastInstalledTag = tag
	return s.saveLocalState(s)
}

func (s *LocalState) CanInstallTag(tag string) (bool, error) {
	lastInstalledTag, err := s.getLastInstalledTag()
	if err != nil {
		return false, err
	}
	if lastInstalledTag == "" {
		return true, nil
	}
	if lastInstalledTag == tag {
		return false, nil
	}
	return true, nil
}

func (s *LocalState) getLastInstalledTag() (string, error) {
	if err := s.readLocalState(); err != nil {
		return "", err
	}
	return s.LastInstalledTag, nil
}

func (s *LocalState) saveFile(path string, c interface{}) error {
	dir := path[:len(path)-len(path[strings.LastIndex(path, "/"):])]
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}

	defer f.Close()
	return json.NewEncoder(f).Encode(c)
}

func (s *LocalState) saveLocalState(state *LocalState) error {
	return s.saveFile(s.stateFilePath, state)
}

func (s *LocalState) readLocalState() error {
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
