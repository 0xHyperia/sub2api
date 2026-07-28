package service

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const softwareReleaseCacheTTL = 15 * time.Minute

var githubRepositoryPartPattern = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)

type SoftwareCatalogItem struct {
	ID                 int64                   `json:"id"`
	SourceType         string                  `json:"source_type"`
	SourceURL          string                  `json:"source_url"`
	Repository         string                  `json:"repository,omitempty"`
	RepositoryURL      string                  `json:"repository_url,omitempty"`
	Name               string                  `json:"name"`
	Description        string                  `json:"description"`
	LogoURL            string                  `json:"logo_url"`
	Featured           bool                    `json:"featured"`
	Enabled            bool                    `json:"enabled"`
	SortOrder          int                     `json:"sort_order"`
	SupportedPlatforms []string                `json:"supported_platforms"`
	DownloadAssets     []SoftwareDownloadAsset `json:"asset_variants,omitempty"`
	Version            string                  `json:"version,omitempty"`
	ReleaseName        string                  `json:"release_name,omitempty"`
	ReleaseNotes       string                  `json:"release_notes,omitempty"`
	PublishedAt        *time.Time              `json:"published_at,omitempty"`
	Release            *GitHubRelease          `json:"release,omitempty"`
	ReleaseFetchedAt   *time.Time              `json:"release_fetched_at,omitempty"`
	LastError          string                  `json:"last_error,omitempty"`
	CreatedAt          time.Time               `json:"created_at"`
	UpdatedAt          time.Time               `json:"updated_at"`
}

const (
	SoftwareSourceGitHub = "github"
	SoftwareSourceManual = "manual"
)

type SoftwareDownloadAsset struct {
	Name           string `json:"name"`
	Label          string `json:"label"`
	Platform       string `json:"platform"`
	Architecture   string `json:"architecture"`
	Format         string `json:"format"`
	Kind           string `json:"kind"`
	URL            string `json:"url"`
	AcceleratedURL string `json:"accelerated_url,omitempty"`
	Size           int64  `json:"size,omitempty"`
}

type SoftwareRepositoryMetadata struct {
	Repository  string `json:"repository"`
	HTMLURL     string `json:"html_url"`
	Name        string `json:"name"`
	Description string `json:"description"`
	AvatarURL   string `json:"avatar_url"`
}

type SoftwareCatalogPreview struct {
	Repository         string         `json:"repository"`
	RepositoryURL      string         `json:"repository_url"`
	Name               string         `json:"name"`
	Description        string         `json:"description"`
	LogoURL            string         `json:"logo_url"`
	SupportedPlatforms []string       `json:"supported_platforms"`
	Release            *GitHubRelease `json:"release"`
}

type CreateSoftwareCatalogInput struct {
	SourceType     string
	RepositoryURL  string
	SourceURL      string
	Name           string
	Description    string
	LogoURL        string
	Featured       bool
	Enabled        bool
	SortOrder      int
	Version        string
	ReleaseName    string
	ReleaseNotes   string
	PublishedAt    *time.Time
	DownloadAssets []SoftwareDownloadAsset
}

type UpdateSoftwareCatalogInput struct {
	Name           *string
	Description    *string
	LogoURL        *string
	Featured       *bool
	Enabled        *bool
	SortOrder      *int
	SourceURL      *string
	Version        *string
	ReleaseName    *string
	ReleaseNotes   *string
	PublishedAt    **time.Time
	DownloadAssets *[]SoftwareDownloadAsset
}

type SoftwareCatalogRepository interface {
	List(ctx context.Context, enabledOnly bool) ([]SoftwareCatalogItem, error)
	GetByID(ctx context.Context, id int64) (*SoftwareCatalogItem, error)
	Create(ctx context.Context, item *SoftwareCatalogItem) (*SoftwareCatalogItem, error)
	Update(ctx context.Context, id int64, input UpdateSoftwareCatalogInput) (*SoftwareCatalogItem, error)
	UpdateRelease(ctx context.Context, id int64, release *GitHubRelease, platforms []string, fetchedAt time.Time, lastError string) error
	Delete(ctx context.Context, id int64) error
}

type SoftwareCatalogGitHubClient interface {
	FetchRepository(ctx context.Context, repository string) (*SoftwareRepositoryMetadata, error)
}

type SoftwareCatalogService struct {
	repository   SoftwareCatalogRepository
	github       SoftwareCatalogGitHubClient
	releases     GitHubReleaseClient
	refreshMutex sync.Mutex
	now          func() time.Time
}

func NewSoftwareCatalogService(repository SoftwareCatalogRepository, github SoftwareCatalogGitHubClient, releases GitHubReleaseClient) *SoftwareCatalogService {
	return &SoftwareCatalogService{repository: repository, github: github, releases: releases, now: time.Now}
}

func ParseGitHubRepositoryURL(raw string) (string, string, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Scheme != "https" || !strings.EqualFold(parsed.Host, "github.com") || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", "", infraerrors.BadRequest("INVALID_GITHUB_REPOSITORY", "请输入标准 GitHub 仓库地址，例如 https://github.com/owner/repo")
	}
	parts := strings.Split(strings.Trim(strings.TrimSuffix(parsed.Path, ".git"), "/"), "/")
	if len(parts) != 2 || !githubRepositoryPartPattern.MatchString(parts[0]) || !githubRepositoryPartPattern.MatchString(parts[1]) {
		return "", "", infraerrors.BadRequest("INVALID_GITHUB_REPOSITORY", "GitHub 仓库地址只能包含 owner/repo")
	}
	repository := parts[0] + "/" + parts[1]
	return repository, "https://github.com/" + repository, nil
}

func (s *SoftwareCatalogService) Preview(ctx context.Context, repositoryURL string) (*SoftwareCatalogPreview, error) {
	repository, normalizedURL, err := ParseGitHubRepositoryURL(repositoryURL)
	if err != nil {
		return nil, err
	}
	metadata, err := s.github.FetchRepository(ctx, repository)
	if err != nil {
		return nil, fmt.Errorf("读取 GitHub 项目信息失败: %w", err)
	}
	release, err := s.releases.FetchLatestRelease(ctx, repository)
	if err != nil {
		return nil, softwareGitHubError("读取 GitHub 最新 Release 失败", err, true)
	}
	return &SoftwareCatalogPreview{
		Repository: repository, RepositoryURL: normalizedURL, Name: metadata.Name,
		Description: metadata.Description, LogoURL: metadata.AvatarURL,
		SupportedPlatforms: inferSoftwarePlatforms(release.Assets), Release: release,
	}, nil
}

func (s *SoftwareCatalogService) Create(ctx context.Context, input CreateSoftwareCatalogInput) (*SoftwareCatalogItem, error) {
	if input.SourceType == SoftwareSourceManual {
		return s.createManual(ctx, input)
	}
	preview, err := s.Preview(ctx, input.RepositoryURL)
	if err != nil {
		return nil, err
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		name = preview.Name
	}
	description := strings.TrimSpace(input.Description)
	if description == "" {
		description = preview.Description
	}
	logoURL := strings.TrimSpace(input.LogoURL)
	if logoURL == "" {
		logoURL = preview.LogoURL
	}
	now := s.now()
	return s.repository.Create(ctx, &SoftwareCatalogItem{
		SourceType: SoftwareSourceGitHub, SourceURL: preview.RepositoryURL,
		Repository: preview.Repository, RepositoryURL: preview.RepositoryURL, Name: name,
		Description: description, LogoURL: logoURL, Featured: input.Featured, Enabled: input.Enabled,
		SortOrder: input.SortOrder, SupportedPlatforms: preview.SupportedPlatforms,
		Release: preview.Release, ReleaseFetchedAt: &now,
	})
}

func (s *SoftwareCatalogService) createManual(ctx context.Context, input CreateSoftwareCatalogInput) (*SoftwareCatalogItem, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, infraerrors.BadRequest("SOFTWARE_NAME_REQUIRED", "请输入软件名称")
	}
	if strings.TrimSpace(input.Version) == "" {
		return nil, infraerrors.BadRequest("SOFTWARE_VERSION_REQUIRED", "请输入当前版本")
	}
	sourceURL, err := validateSoftwareURL(input.SourceURL, "官网或来源地址")
	if err != nil {
		return nil, err
	}
	assets, platforms, err := validateManualAssets(input.DownloadAssets)
	if err != nil {
		return nil, err
	}
	release := manualSoftwareRelease(input, sourceURL, assets)
	return s.repository.Create(ctx, &SoftwareCatalogItem{
		SourceType: SoftwareSourceManual, SourceURL: sourceURL, Name: name,
		Description: strings.TrimSpace(input.Description), LogoURL: strings.TrimSpace(input.LogoURL),
		Featured: input.Featured, Enabled: input.Enabled, SortOrder: input.SortOrder,
		SupportedPlatforms: platforms, DownloadAssets: assets, Release: release,
		Version: strings.TrimSpace(input.Version), ReleaseName: strings.TrimSpace(input.ReleaseName),
		ReleaseNotes: input.ReleaseNotes, PublishedAt: input.PublishedAt,
	})
}

func manualSoftwareRelease(input CreateSoftwareCatalogInput, sourceURL string, assets []SoftwareDownloadAsset) *GitHubRelease {
	published := ""
	if input.PublishedAt != nil {
		published = input.PublishedAt.UTC().Format(time.RFC3339)
	}
	releaseAssets := make([]GitHubAsset, 0, len(assets))
	for _, asset := range assets {
		releaseAssets = append(releaseAssets, GitHubAsset{Name: asset.Name, BrowserDownloadURL: asset.URL, Size: asset.Size})
	}
	return &GitHubRelease{TagName: strings.TrimSpace(input.Version), Name: strings.TrimSpace(input.ReleaseName), Body: input.ReleaseNotes, PublishedAt: published, HTMLURL: sourceURL, Assets: releaseAssets}
}

func validateSoftwareURL(raw, label string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil {
		return "", infraerrors.BadRequest("INVALID_SOFTWARE_URL", label+"必须是有效的 HTTPS 地址")
	}
	return parsed.String(), nil
}

func validateManualAssets(input []SoftwareDownloadAsset) ([]SoftwareDownloadAsset, []string, error) {
	if len(input) == 0 {
		return nil, nil, infraerrors.BadRequest("SOFTWARE_ASSET_REQUIRED", "请至少添加一个下载项")
	}
	validPlatforms := map[string]bool{"windows": true, "macos": true, "linux": true, "android": true}
	validArch := map[string]bool{"x64": true, "arm64": true, "universal": true}
	validKinds := map[string]bool{"installer": true, "portable": true, "archive": true}
	platformSet := map[string]bool{}
	assets := make([]SoftwareDownloadAsset, 0, len(input))
	for _, value := range input {
		value.Name = strings.TrimSpace(value.Name)
		value.Label = strings.TrimSpace(value.Label)
		value.Platform = strings.ToLower(strings.TrimSpace(value.Platform))
		value.Architecture = strings.ToLower(strings.TrimSpace(value.Architecture))
		value.Format = strings.TrimSpace(value.Format)
		value.Kind = strings.ToLower(strings.TrimSpace(value.Kind))
		if value.Name == "" || value.Label == "" || !validPlatforms[value.Platform] || !validArch[value.Architecture] || !validKinds[value.Kind] {
			return nil, nil, infraerrors.BadRequest("INVALID_SOFTWARE_ASSET", "下载项名称、平台、架构、格式和类型不完整")
		}
		var err error
		if value.URL, err = validateSoftwareURL(value.URL, "下载地址"); err != nil {
			return nil, nil, err
		}
		if strings.TrimSpace(value.AcceleratedURL) != "" {
			if value.AcceleratedURL, err = validateSoftwareURL(value.AcceleratedURL, "加速地址"); err != nil {
				return nil, nil, err
			}
		}
		platformSet[value.Platform] = true
		assets = append(assets, value)
	}
	platforms := make([]string, 0, len(platformSet))
	for _, platform := range []string{"windows", "macos", "linux", "android"} {
		if platformSet[platform] {
			platforms = append(platforms, platform)
		}
	}
	return assets, platforms, nil
}

func (s *SoftwareCatalogService) ListAdmin(ctx context.Context) ([]SoftwareCatalogItem, error) {
	return s.repository.List(ctx, false)
}

func (s *SoftwareCatalogService) ListPublic(ctx context.Context) ([]SoftwareCatalogItem, error) {
	items, err := s.repository.List(ctx, true)
	if err != nil {
		return nil, err
	}
	s.refreshMutex.Lock()
	defer s.refreshMutex.Unlock()
	now := s.now()
	var wg sync.WaitGroup
	var updateErr error
	var updateErrMutex sync.Mutex
	for i := range items {
		if items[i].SourceType == SoftwareSourceManual {
			continue
		}
		if items[i].ReleaseFetchedAt != nil && now.Sub(*items[i].ReleaseFetchedAt) < softwareReleaseCacheTTL {
			items[i].LastError = ""
			continue
		}
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			release, fetchErr := s.releases.FetchLatestRelease(ctx, items[index].Repository)
			if fetchErr != nil {
				_ = s.repository.UpdateRelease(ctx, items[index].ID, items[index].Release, items[index].SupportedPlatforms, now, fetchErr.Error())
				items[index].LastError = ""
				return
			}
			platforms := inferSoftwarePlatforms(release.Assets)
			if err := s.repository.UpdateRelease(ctx, items[index].ID, release, platforms, now, ""); err != nil {
				updateErrMutex.Lock()
				if updateErr == nil {
					updateErr = err
				}
				updateErrMutex.Unlock()
				return
			}
			items[index].Release = release
			items[index].SupportedPlatforms = platforms
			items[index].ReleaseFetchedAt = &now
			items[index].LastError = ""
		}(i)
	}
	wg.Wait()
	if updateErr != nil {
		return nil, updateErr
	}
	return items, nil
}

func (s *SoftwareCatalogService) Update(ctx context.Context, id int64, input UpdateSoftwareCatalogInput) (*SoftwareCatalogItem, error) {
	item, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if item.SourceType == SoftwareSourceManual {
		if input.Version != nil && strings.TrimSpace(*input.Version) == "" {
			return nil, infraerrors.BadRequest("SOFTWARE_VERSION_REQUIRED", "请输入当前版本")
		}
		if input.SourceURL != nil {
			normalized, err := validateSoftwareURL(*input.SourceURL, "官网或来源地址")
			if err != nil {
				return nil, err
			}
			input.SourceURL = &normalized
		}
		if input.DownloadAssets != nil {
			assets, _, err := validateManualAssets(*input.DownloadAssets)
			if err != nil {
				return nil, err
			}
			input.DownloadAssets = &assets
		}
	}
	return s.repository.Update(ctx, id, input)
}

func (s *SoftwareCatalogService) Delete(ctx context.Context, id int64) error {
	return s.repository.Delete(ctx, id)
}

func (s *SoftwareCatalogService) Refresh(ctx context.Context, id int64) (*SoftwareCatalogItem, error) {
	item, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if item.SourceType != SoftwareSourceGitHub {
		return nil, infraerrors.BadRequest("SOFTWARE_REFRESH_UNSUPPORTED", "手动配置的软件无需同步 GitHub Release")
	}
	release, err := s.releases.FetchLatestRelease(ctx, item.Repository)
	if err != nil {
		return nil, softwareGitHubError("刷新 GitHub Release 失败", err, false)
	}
	now := s.now()
	platforms := inferSoftwarePlatforms(release.Assets)
	if err := s.repository.UpdateRelease(ctx, id, release, platforms, now, ""); err != nil {
		return nil, err
	}
	return s.repository.GetByID(ctx, id)
}

func softwareGitHubError(operation string, err error, missingIsBadRequest bool) error {
	message := operation + ": " + err.Error()
	lower := strings.ToLower(err.Error())
	if strings.Contains(lower, "returned 403") || strings.Contains(lower, "returned 429") || strings.Contains(lower, "rate limit") {
		return infraerrors.TooManyRequests("GITHUB_RATE_LIMITED", "GitHub 请求额度已用尽，请稍后重试")
	}
	if missingIsBadRequest && strings.Contains(lower, "returned 404") {
		return infraerrors.BadRequest("GITHUB_RELEASE_NOT_FOUND", "该项目没有可用的正式 Release")
	}
	return infraerrors.ServiceUnavailable("GITHUB_UNAVAILABLE", message)
}

func inferSoftwarePlatforms(assets []GitHubAsset) []string {
	found := map[string]bool{}
	for _, asset := range assets {
		name := strings.ToLower(asset.Name)
		switch {
		case strings.Contains(name, "android") || strings.HasSuffix(name, ".apk"):
			found["android"] = true
		case strings.Contains(name, "macos") || strings.Contains(name, "darwin") || strings.HasSuffix(name, ".dmg") || strings.HasSuffix(name, ".pkg"):
			found["macos"] = true
		case strings.Contains(name, "windows") || strings.Contains(name, "win32") || strings.HasSuffix(name, ".exe") || strings.HasSuffix(name, ".msi"):
			found["windows"] = true
		case strings.Contains(name, "linux") || strings.HasSuffix(name, ".appimage") || strings.HasSuffix(name, ".deb") || strings.HasSuffix(name, ".rpm"):
			found["linux"] = true
		}
	}
	order := []string{"windows", "macos", "linux", "android"}
	platforms := make([]string, 0, len(found))
	for _, platform := range order {
		if found[platform] {
			platforms = append(platforms, platform)
		}
	}
	return platforms
}
