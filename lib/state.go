package lib

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	redis "github.com/redis/go-redis/v9"
)

type State struct {
	client *redis.Client
	config *Config
}

const stateFile = "/var/lib/gacr/state.json"

func NewState(config *Config) (*State, error) {
	rc := redis.NewClient(&redis.Options{
		Addr:     config.Redis.Host,
		Password: config.Redis.Password,
		DB:       config.Redis.DB,
	})

	if err := rc.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to create redis client: %s", err)
	}
	return &State{
		client: rc,
		config: config,
	}, nil
}

func (s *State) UnlockCanaryRelease() error {
	return s.client.Del(context.Background(), s.config.Repo+s.config.CanaryReleaseVersionKey).Err()
}

func (s *State) TryCanaryReleaseLock(tag string) (bool, error) {
	return s.getLock(s.config.CanaryReleaseVersionKey, tag, s.config.CanaryRolloutWindow*2)
}

func (s *State) TryRolloutLock(tag string) (bool, error) {
	return s.getLock(s.config.RolloutKey, tag, s.config.RolloutWindow)
}

func (s *State) getLock(key string, tag string, window time.Duration) (bool, error) {
	ok, err := s.client.SetNX(context.Background(), s.config.Repo+key, tag, 0).Result()
	if err != nil {
		return false, err
	}
	if ok {
		err := s.client.Expire(context.Background(), s.config.Repo+key, window).Err()
		if err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

func (s *State) saveRelease(key, tag string) error {
	return s.client.SetEx(context.Background(), s.config.Repo+key, tag, 0).Err()
}

func (s *State) saveReleases(key string, tags ...string) error {
	return s.client.SAdd(context.Background(), s.config.Repo+key, tags).Err()
}

func (s *State) SaveStableReleaseTag(tag string) error {
	return s.saveRelease(s.config.StableReleaseVersionKey, tag)
}

func (s *State) SaveAvoidReleaseTag(tag string) error {
	return s.saveReleases(s.config.AvoidReleaseVersionKey, tag)
}

func (s *State) getRelease(key string) (string, error) {
	return s.client.Get(context.Background(), s.config.Repo+key).Result()
}

func (s *State) getReleases(key string) ([]string, error) {
	return s.client.SMembers(context.Background(), s.config.Repo+key).Result()
}

func (s *State) GetStableReleaseTag() (string, error) {
	return s.getRelease(s.config.StableReleaseVersionKey)
}

func (s *State) GetAvoidReleaseTag() ([]string, error) {
	return s.getReleases(s.config.AvoidReleaseVersionKey)
}

type localState struct {
	LastInstalledTag string
}

func (s *State) saveFile(path string, c interface{}) error {
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

func (s *State) SaveLocalState(state *localState) error {
	return s.saveFile(stateFile, state)
}

func Exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func (s *State) readFile(path string, st interface{}) error {
	if !Exists(path) {
		return nil
	}

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewDecoder(f).Decode(st)
}

func (s *State) GetLocalState() (*localState, error) {
	ls := &localState{}
	if err := s.readFile(stateFile, ls); err != nil {
		return nil, err
	}
	return ls, nil
}
