package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/lib/pq"
)

type softwareCatalogRepository struct{ db *sql.DB }

func NewSoftwareCatalogRepository(db *sql.DB) service.SoftwareCatalogRepository {
	return &softwareCatalogRepository{db: db}
}

const softwareCatalogColumns = `id, source_type, source_url, repository, repository_url, name, description, logo_url,
	featured, enabled, sort_order, supported_platforms, release_snapshot, release_fetched_at,
	last_error, download_assets, manual_version, manual_release_name, manual_release_notes, manual_published_at,
	created_at, updated_at`

type softwareCatalogScanner interface{ Scan(...any) error }

func scanSoftwareCatalogItem(scanner softwareCatalogScanner) (*service.SoftwareCatalogItem, error) {
	var item service.SoftwareCatalogItem
	var platformsJSON []byte
	var releaseJSON []byte
	var fetchedAt sql.NullTime
	var publishedAt sql.NullTime
	var assetsJSON []byte
	var repository, repositoryURL sql.NullString
	if err := scanner.Scan(&item.ID, &item.SourceType, &item.SourceURL, &repository, &repositoryURL, &item.Name, &item.Description,
		&item.LogoURL, &item.Featured, &item.Enabled, &item.SortOrder, &platformsJSON, &releaseJSON,
		&fetchedAt, &item.LastError, &assetsJSON, &item.Version, &item.ReleaseName, &item.ReleaseNotes, &publishedAt,
		&item.CreatedAt, &item.UpdatedAt); err != nil {
		return nil, err
	}
	item.Repository = repository.String
	item.RepositoryURL = repositoryURL.String
	if len(assetsJSON) > 0 {
		if err := json.Unmarshal(assetsJSON, &item.DownloadAssets); err != nil {
			return nil, fmt.Errorf("decode software assets: %w", err)
		}
	}
	if publishedAt.Valid {
		item.PublishedAt = &publishedAt.Time
	}
	if err := json.Unmarshal(platformsJSON, &item.SupportedPlatforms); err != nil {
		return nil, fmt.Errorf("decode software platforms: %w", err)
	}
	if len(releaseJSON) > 0 {
		var release *service.GitHubRelease
		if err := json.Unmarshal(releaseJSON, &release); err != nil {
			return nil, fmt.Errorf("decode software release snapshot: %w", err)
		}
		item.Release = release
	}
	if fetchedAt.Valid {
		item.ReleaseFetchedAt = &fetchedAt.Time
	}
	if item.SourceType == service.SoftwareSourceManual {
		published := ""
		if item.PublishedAt != nil {
			published = item.PublishedAt.UTC().Format(time.RFC3339)
		}
		assets := make([]service.GitHubAsset, 0, len(item.DownloadAssets))
		for _, asset := range item.DownloadAssets {
			assets = append(assets, service.GitHubAsset{Name: asset.Name, BrowserDownloadURL: asset.URL, Size: asset.Size})
		}
		item.Release = &service.GitHubRelease{TagName: item.Version, Name: item.ReleaseName, Body: item.ReleaseNotes, PublishedAt: published, HTMLURL: item.SourceURL, Assets: assets}
	}
	return &item, nil
}

func (r *softwareCatalogRepository) List(ctx context.Context, enabledOnly bool) ([]service.SoftwareCatalogItem, error) {
	query := `SELECT ` + softwareCatalogColumns + ` FROM software_catalog_items`
	if enabledOnly {
		query += ` WHERE enabled = TRUE`
	}
	query += ` ORDER BY featured DESC, sort_order, id`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]service.SoftwareCatalogItem, 0)
	for rows.Next() {
		item, err := scanSoftwareCatalogItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	return items, rows.Err()
}

func (r *softwareCatalogRepository) GetByID(ctx context.Context, id int64) (*service.SoftwareCatalogItem, error) {
	item, err := scanSoftwareCatalogItem(r.db.QueryRowContext(ctx, `SELECT `+softwareCatalogColumns+` FROM software_catalog_items WHERE id = $1`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, infraerrors.NotFound("SOFTWARE_NOT_FOUND", "软件不存在")
	}
	return item, err
}

func (r *softwareCatalogRepository) Create(ctx context.Context, item *service.SoftwareCatalogItem) (*service.SoftwareCatalogItem, error) {
	platforms, _ := json.Marshal(item.SupportedPlatforms)
	release, _ := json.Marshal(item.Release)
	assets, _ := json.Marshal(item.DownloadAssets)
	created, err := scanSoftwareCatalogItem(r.db.QueryRowContext(ctx, `INSERT INTO software_catalog_items
		(source_type, source_url, repository, repository_url, name, description, logo_url, featured, enabled, sort_order,
		 supported_platforms, release_snapshot, release_fetched_at, last_error, download_assets, manual_version,
		 manual_release_name, manual_release_notes, manual_published_at)
		VALUES ($1,$2,NULLIF($3,''),NULLIF($4,''),$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19) RETURNING `+softwareCatalogColumns,
		item.SourceType, item.SourceURL, item.Repository, item.RepositoryURL, item.Name, item.Description, item.LogoURL, item.Featured,
		item.Enabled, item.SortOrder, platforms, release, item.ReleaseFetchedAt, item.LastError, assets, item.Version,
		item.ReleaseName, item.ReleaseNotes, item.PublishedAt))
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return nil, infraerrors.Conflict("SOFTWARE_ALREADY_EXISTS", "该软件已上架")
		}
	}
	return created, err
}

func (r *softwareCatalogRepository) Update(ctx context.Context, id int64, input service.UpdateSoftwareCatalogInput) (*service.SoftwareCatalogItem, error) {
	sets := []string{"updated_at = NOW()"}
	args := []any{}
	add := func(column string, value any) {
		args = append(args, value)
		sets = append(sets, fmt.Sprintf("%s = $%d", column, len(args)))
	}
	if input.Name != nil {
		add("name", strings.TrimSpace(*input.Name))
	}
	if input.Description != nil {
		add("description", strings.TrimSpace(*input.Description))
	}
	if input.LogoURL != nil {
		add("logo_url", strings.TrimSpace(*input.LogoURL))
	}
	if input.Featured != nil {
		add("featured", *input.Featured)
	}
	if input.Enabled != nil {
		add("enabled", *input.Enabled)
	}
	if input.SortOrder != nil {
		add("sort_order", *input.SortOrder)
	}
	if input.SourceURL != nil {
		add("source_url", strings.TrimSpace(*input.SourceURL))
	}
	if input.Version != nil {
		add("manual_version", strings.TrimSpace(*input.Version))
	}
	if input.ReleaseName != nil {
		add("manual_release_name", strings.TrimSpace(*input.ReleaseName))
	}
	if input.ReleaseNotes != nil {
		add("manual_release_notes", *input.ReleaseNotes)
	}
	if input.PublishedAt != nil {
		add("manual_published_at", *input.PublishedAt)
	}
	if input.DownloadAssets != nil {
		assets, _ := json.Marshal(*input.DownloadAssets)
		add("download_assets", assets)
		platforms := make([]string, 0)
		seen := map[string]bool{}
		for _, asset := range *input.DownloadAssets {
			if !seen[asset.Platform] {
				seen[asset.Platform] = true
				platforms = append(platforms, asset.Platform)
			}
		}
		platformsJSON, _ := json.Marshal(platforms)
		add("supported_platforms", platformsJSON)
	}
	args = append(args, id)
	item, err := scanSoftwareCatalogItem(r.db.QueryRowContext(ctx, `UPDATE software_catalog_items SET `+strings.Join(sets, ", ")+fmt.Sprintf(" WHERE id = $%d RETURNING ", len(args))+softwareCatalogColumns, args...))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, infraerrors.NotFound("SOFTWARE_NOT_FOUND", "软件不存在")
	}
	return item, err
}

func (r *softwareCatalogRepository) UpdateRelease(ctx context.Context, id int64, release *service.GitHubRelease, platforms []string, fetchedAt time.Time, lastError string) error {
	releaseJSON, _ := json.Marshal(release)
	platformsJSON, _ := json.Marshal(platforms)
	result, err := r.db.ExecContext(ctx, `UPDATE software_catalog_items SET release_snapshot=$1, supported_platforms=$2, release_fetched_at=$3, last_error=$4, updated_at=NOW() WHERE id=$5`, releaseJSON, platformsJSON, fetchedAt, lastError, id)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err == nil && count == 0 {
		return infraerrors.NotFound("SOFTWARE_NOT_FOUND", "软件不存在")
	}
	return err
}

func (r *softwareCatalogRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM software_catalog_items WHERE id=$1`, id)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err == nil && count == 0 {
		return infraerrors.NotFound("SOFTWARE_NOT_FOUND", "软件不存在")
	}
	return err
}
