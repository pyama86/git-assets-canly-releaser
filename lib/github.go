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

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type GitHub struct {
	client                *github.Client
	config                *Config
	owner                 string
	repo                  string
	logger                *slog.Logger
	regPackageNamePattern *regexp.Regexp
}

func NewGitHub(config *Config) (*GitHub, error) {
	logger := slog.Default().With("package", "github")

	// HTTPクライアントを設定
	var tc *http.Client
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config.GitHubToken},
	)
	tc = oauth2.NewClient(context.Background(), ts)

	client := github.NewClient(tc)

	ownerRepo := strings.Split(config.Repo, "/")
	if len(ownerRepo) != 2 {
		return nil, fmt.Errorf("invalid repo: %s", config.Repo)
	}
	return &GitHub{
		client:                client,
		config:                config,
		logger:                logger,
		owner:                 ownerRepo[0],
		repo:                  ownerRepo[1],
		regPackageNamePattern: regexp.MustCompile(config.PackageNamePattern),
	}, nil
}

var AssetsNotFound = errors.New("no match assets")

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

	g.logger.Info("tag info", "Latest release Tag", *release.TagName)
	for _, asset := range release.Assets {
		g.logger.Info("assets info", "name", asset.Name, "download url", asset.URL)
		if g.regPackageNamePattern.MatchString(*asset.Name) {

			filePath := filepath.Join(g.config.AssetsDownloadPath, *asset.Name)

			if _, err := os.Stat(filePath); err == nil {
				return *release.TagName, filePath, nil
			} else if !os.IsNotExist(err) {
				return "", "", err
			}

			ret, _, err := g.client.Repositories.DownloadReleaseAsset(context.Background(), g.owner, g.repo, *asset.ID)
			if err != nil {
				return "", "", fmt.Errorf("repositories.DownloadReleaseAsset returned error: %v", err)
			}

			defer ret.Close()

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
	return "", "", AssetsNotFound
}
