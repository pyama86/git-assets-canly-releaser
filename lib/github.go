package lib

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/go-github/v55/github"
	"github.com/k1LoW/go-github-client/v55/factory"
)

type GitHub struct {
	client                *github.Client
	config                *Config
	owner                 string
	repo                  string
	regPackageNamePattern *regexp.Regexp
}

type GitHuber interface {
	DownloadReleaseAsset(tag string) (string, string, error)
}

func NewGitHub(config *Config) (*GitHub, error) {
	token := config.GitHubToken
	if os.Getenv("GITHUB_TOKEN") == "" {
		os.Setenv("GITHUB_TOKEN", token)
	}

	client, _ := factory.NewGithubClient()
	ownerRepo := strings.Split(config.Repo, "/")
	if len(ownerRepo) != 2 {
		return nil, fmt.Errorf("invalid repo: %s", config.Repo)
	}
	return &GitHub{
		client:                client,
		config:                config,
		owner:                 ownerRepo[0],
		repo:                  ownerRepo[1],
		regPackageNamePattern: regexp.MustCompile(config.PackageNamePattern),
	}, nil
}

var ErrAssetsNotFound = errors.New("no match assets")

const LatestTag = "latest"

func (g *GitHub) DownloadReleaseAsset(tag string) (string, string, error) {
	var release *github.RepositoryRelease
	if tag == LatestTag {
		r, _, err := g.client.Repositories.GetLatestRelease(context.Background(), g.owner, g.repo)
		if err != nil {
			return "", "", fmt.Errorf("repositories.GetRelease returned tag:%s error: %v", tag, err)
		}
		release = r
	} else {
		r, _, err := g.client.Repositories.GetReleaseByTag(context.Background(), g.owner, g.repo, tag)
		if err != nil {
			return "", "", fmt.Errorf("repositories.GetRelease returned tag:%s error: %v", tag, err)
		}
		release = r
	}

	slog.Debug("tag info", "latest release Tag", *release.TagName)
	for _, asset := range release.Assets {
		slog.Debug("assets info", "name", *asset.Name, "download url", *asset.URL)
		if g.regPackageNamePattern.MatchString(*asset.Name) {
			filePath := filepath.Join(g.config.SaveAssetsPath, *asset.Name)

			if _, err := os.Stat(filePath); err == nil {
				return *release.TagName, filePath, nil
			} else if !os.IsNotExist(err) {
				return "", "", err
			}

			ret, loc, err := g.client.Repositories.DownloadReleaseAsset(context.Background(), g.owner, g.repo, *asset.ID, nil)
			if err != nil {
				return "", "", fmt.Errorf("repositories.DownloadReleaseAsset returned error: %v", err)
			}

			if loc != "" {
				req, err := http.NewRequestWithContext(context.Background(), "GET", loc, nil)
				if err != nil {
					return "", "", err
				}
				res, err := g.client.Client().Do(req)
				if err != nil {
					return "", "", err
				}
				ret = res.Body
				if ret != nil {
					defer ret.Close()
				}
			}

			out, err := os.Create(filePath)
			if err != nil {
				return "", "", err
			}
			defer out.Close()
			_, err = io.Copy(out, ret)
			if err != nil {
				return "", "", err
			}
			return *release.TagName, filePath, nil
		}
	}
	return "", "", ErrAssetsNotFound
}
