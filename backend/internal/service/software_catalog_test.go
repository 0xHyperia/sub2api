package service

import (
	"context"
	"errors"
	"testing"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestParseGitHubRepositoryURL(t *testing.T) {
	repository, normalized, err := ParseGitHubRepositoryURL(" https://github.com/USA-Zero/ZeroAgent ")
	require.NoError(t, err)
	require.Equal(t, "USA-Zero/ZeroAgent", repository)
	require.Equal(t, "https://github.com/USA-Zero/ZeroAgent", normalized)

	invalid := []string{
		"http://github.com/owner/repo", "https://api.github.com/owner/repo",
		"https://user@github.com/owner/repo", "https://github.com:443/owner/repo",
		"https://github.com/owner/repo/releases", "https://github.com/owner/repo?q=1",
		"https://github.com/owner/repo#readme", "https://github.com/owner",
	}
	for _, raw := range invalid {
		t.Run(raw, func(t *testing.T) {
			_, _, err := ParseGitHubRepositoryURL(raw)
			require.Error(t, err)
			require.True(t, infraerrors.IsBadRequest(err))
		})
	}
}

type softwareCatalogRepositoryStub struct {
	items   []SoftwareCatalogItem
	updates int
	created *SoftwareCatalogItem
	item    *SoftwareCatalogItem
}

func (r *softwareCatalogRepositoryStub) List(context.Context, bool) ([]SoftwareCatalogItem, error) {
	return append([]SoftwareCatalogItem(nil), r.items...), nil
}
func (r *softwareCatalogRepositoryStub) GetByID(context.Context, int64) (*SoftwareCatalogItem, error) {
	if r.item != nil {
		return r.item, nil
	}
	return nil, errors.New("not implemented")
}
func (r *softwareCatalogRepositoryStub) Create(_ context.Context, item *SoftwareCatalogItem) (*SoftwareCatalogItem, error) {
	r.created = item
	return item, nil
}
func (r *softwareCatalogRepositoryStub) Update(context.Context, int64, UpdateSoftwareCatalogInput) (*SoftwareCatalogItem, error) {
	return nil, errors.New("not implemented")
}
func (r *softwareCatalogRepositoryStub) UpdateRelease(context.Context, int64, *GitHubRelease, []string, time.Time, string) error {
	r.updates++
	return nil
}
func (r *softwareCatalogRepositoryStub) Delete(context.Context, int64) error {
	return errors.New("not implemented")
}

type softwareCatalogGitHubStub struct{}

func (softwareCatalogGitHubStub) FetchRepository(context.Context, string) (*SoftwareRepositoryMetadata, error) {
	return nil, errors.New("not implemented")
}

type softwareReleaseClientStub struct {
	release *GitHubRelease
	err     error
	calls   int
}

func (c *softwareReleaseClientStub) FetchLatestRelease(context.Context, string) (*GitHubRelease, error) {
	c.calls++
	return c.release, c.err
}
func (*softwareReleaseClientStub) FetchRecentReleases(context.Context, string, int) ([]*GitHubRelease, error) {
	return nil, errors.New("not implemented")
}
func (*softwareReleaseClientStub) DownloadFile(context.Context, string, string, int64) error {
	return errors.New("not implemented")
}
func (*softwareReleaseClientStub) FetchChecksumFile(context.Context, string) ([]byte, error) {
	return nil, errors.New("not implemented")
}

func TestSoftwareCatalogListPublicUsesFreshSnapshot(t *testing.T) {
	now := time.Date(2026, 7, 29, 12, 0, 0, 0, time.UTC)
	snapshot := &GitHubRelease{TagName: "v1.0.0"}
	repository := &softwareCatalogRepositoryStub{items: []SoftwareCatalogItem{{ID: 1, Repository: "owner/repo", Release: snapshot, ReleaseFetchedAt: softwareTimePtr(now.Add(-time.Minute))}}}
	releases := &softwareReleaseClientStub{err: errors.New("must not be called")}
	service := NewSoftwareCatalogService(repository, softwareCatalogGitHubStub{}, releases)
	service.now = func() time.Time { return now }
	items, err := service.ListPublic(context.Background())
	require.NoError(t, err)
	require.Equal(t, "v1.0.0", items[0].Release.TagName)
	require.Zero(t, releases.calls)
	require.Zero(t, repository.updates)
}

func TestSoftwareCatalogListPublicKeepsSnapshotWhenRefreshFails(t *testing.T) {
	now := time.Date(2026, 7, 29, 12, 0, 0, 0, time.UTC)
	snapshot := &GitHubRelease{TagName: "v1.0.0"}
	repository := &softwareCatalogRepositoryStub{items: []SoftwareCatalogItem{{ID: 1, Repository: "owner/repo", Release: snapshot, ReleaseFetchedAt: softwareTimePtr(now.Add(-time.Hour))}}}
	releases := &softwareReleaseClientStub{err: errors.New("rate limited")}
	service := NewSoftwareCatalogService(repository, softwareCatalogGitHubStub{}, releases)
	service.now = func() time.Time { return now }
	items, err := service.ListPublic(context.Background())
	require.NoError(t, err)
	require.Equal(t, "v1.0.0", items[0].Release.TagName)
	require.Equal(t, 1, releases.calls)
	require.Equal(t, 1, repository.updates)
}

func TestSoftwareCatalogCreateManualDoesNotCallGitHub(t *testing.T) {
	repository := &softwareCatalogRepositoryStub{}
	releases := &softwareReleaseClientStub{err: errors.New("must not be called")}
	catalog := NewSoftwareCatalogService(repository, softwareCatalogGitHubStub{}, releases)
	published := time.Date(2026, 7, 29, 0, 0, 0, 0, time.UTC)
	item, err := catalog.Create(context.Background(), CreateSoftwareCatalogInput{
		SourceType: SoftwareSourceManual, SourceURL: "https://chatgpt.com/download/", Name: "Codex App",
		Version: "1.0.0", ReleaseNotes: "## Changes", PublishedAt: &published, Enabled: true,
		DownloadAssets: []SoftwareDownloadAsset{{Name: "CodexSetup.exe", Label: "Windows 安装程序", Platform: "windows", Architecture: "x64", Format: "EXE", Kind: "installer", URL: "https://downloads.example.com/CodexSetup.exe"}},
	})
	require.NoError(t, err)
	require.Equal(t, SoftwareSourceManual, item.SourceType)
	require.Equal(t, []string{"windows"}, item.SupportedPlatforms)
	require.Equal(t, "1.0.0", item.Release.TagName)
	require.Equal(t, "https://downloads.example.com/CodexSetup.exe", item.Release.Assets[0].BrowserDownloadURL)
	require.Zero(t, releases.calls)
	require.Same(t, item, repository.created)
}

func TestSoftwareCatalogCreateManualRejectsUnsafeURLs(t *testing.T) {
	catalog := NewSoftwareCatalogService(&softwareCatalogRepositoryStub{}, softwareCatalogGitHubStub{}, &softwareReleaseClientStub{})
	for _, raw := range []string{"javascript:alert(1)", "http://downloads.example.com/app.exe", "https://user@example.com/app.exe"} {
		t.Run(raw, func(t *testing.T) {
			_, err := catalog.Create(context.Background(), CreateSoftwareCatalogInput{SourceType: SoftwareSourceManual, SourceURL: "https://example.com", Name: "App", Version: "1.0", DownloadAssets: []SoftwareDownloadAsset{{Name: "app.exe", Label: "Windows", Platform: "windows", Architecture: "x64", Format: "EXE", Kind: "installer", URL: raw}}})
			require.Error(t, err)
			require.True(t, infraerrors.IsBadRequest(err))
		})
	}
}

func TestSoftwareCatalogRefreshRejectsManualSource(t *testing.T) {
	repository := &softwareCatalogRepositoryStub{item: &SoftwareCatalogItem{ID: 1, SourceType: SoftwareSourceManual}}
	catalog := NewSoftwareCatalogService(repository, softwareCatalogGitHubStub{}, &softwareReleaseClientStub{})
	_, err := catalog.Refresh(context.Background(), 1)
	require.Error(t, err)
	require.True(t, infraerrors.IsBadRequest(err))
}

func softwareTimePtr(value time.Time) *time.Time { return &value }
