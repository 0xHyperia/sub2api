package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/httpclient"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type softwareCatalogGitHubClient struct {
	client  *http.Client
	token   string
	initErr error
}

func NewSoftwareCatalogGitHubClient(cfg *config.Config) service.SoftwareCatalogGitHubClient {
	client, err := httpclient.GetClient(httpclient.Options{Timeout: 30 * time.Second, ProxyURL: cfg.Update.ProxyURL})
	if err != nil {
		if strings.TrimSpace(cfg.Update.ProxyURL) != "" && !cfg.Security.ProxyFallback.AllowDirectOnError {
			return &softwareCatalogGitHubClient{initErr: fmt.Errorf("proxy client init failed: %w", err)}
		}
		slog.Warn("software catalog proxy init failed; falling back to direct", "error", err)
		client = &http.Client{Timeout: 30 * time.Second}
	}
	client = cloneHTTPClient(client)
	client.CheckRedirect = githubAPICheckRedirect(client.CheckRedirect)
	return &softwareCatalogGitHubClient{client: client, token: os.Getenv("UPDATE_GITHUB_TOKEN")}
}

func (c *softwareCatalogGitHubClient) FetchRepository(ctx context.Context, repository string) (*service.SoftwareRepositoryMetadata, error) {
	if c.initErr != nil {
		return nil, c.initErr
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/repos/"+repository, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "Sub2API-Software-Center")
	if c.token != "" && isGitHubAPIURL(req.URL) {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case http.StatusNotFound:
			return nil, infraerrors.NotFound("GITHUB_REPOSITORY_NOT_FOUND", "GitHub 项目不存在或不可公开访问")
		case http.StatusForbidden, http.StatusTooManyRequests:
			return nil, infraerrors.TooManyRequests("GITHUB_RATE_LIMITED", "GitHub 请求额度已用尽，请稍后重试")
		default:
			return nil, infraerrors.ServiceUnavailable("GITHUB_UNAVAILABLE", fmt.Sprintf("GitHub API returned %d", resp.StatusCode))
		}
	}
	var payload struct {
		FullName    string `json:"full_name"`
		HTMLURL     string `json:"html_url"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Owner       struct {
			AvatarURL string `json:"avatar_url"`
		} `json:"owner"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	return &service.SoftwareRepositoryMetadata{Repository: payload.FullName, HTMLURL: payload.HTMLURL, Name: payload.Name, Description: payload.Description, AvatarURL: payload.Owner.AvatarURL}, nil
}
