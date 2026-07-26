package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
)

type ModelMonitorService struct {
	repo          ModelMonitorRepository
	groupRepo     GroupRepository
	accountRepo   AccountRepository
	accountTester *AccountTestService
	gateway       *GatewayService
	openAIGateway *OpenAIGatewayService
}

func NewModelMonitorService(repo ModelMonitorRepository, groupRepo GroupRepository, accountRepo AccountRepository, accountTester *AccountTestService, gateway *GatewayService, openAIGateway *OpenAIGatewayService) *ModelMonitorService {
	return &ModelMonitorService{repo: repo, groupRepo: groupRepo, accountRepo: accountRepo, accountTester: accountTester, gateway: gateway, openAIGateway: openAIGateway}
}

func (s *ModelMonitorService) DiscoverCatalog(ctx context.Context) ([]ModelCatalogEntry, error) {
	groups, err := s.groupRepo.ListActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("list active groups: %w", err)
	}
	seen := make(map[string]ModelCatalogEntry)
	for i := range groups {
		group := groups[i]
		platform := strings.TrimSpace(group.Platform)
		// Composite aliases span concrete providers and cannot be probed as a
		// standalone upstream platform. Their concrete child groups remain in
		// the catalog and retain independent monitoring histories.
		if platform == "" || platform == PlatformComposite {
			continue
		}
		accounts, listErr := s.accountRepo.ListSchedulableByGroupIDAndPlatform(ctx, group.ID, platform)
		if listErr != nil {
			return nil, fmt.Errorf("list schedulable accounts for group %d: %w", group.ID, listErr)
		}
		if len(accounts) == 0 {
			continue
		}
		for _, model := range ModelCatalogModels(group, accounts) {
			key := ModelMonitorKey(platform, model)
			entry := seen[key]
			entry.Platform = platform
			entry.Model = model
			entry.Groups = append(entry.Groups, ModelMonitorGroupOption{
				GroupID: group.ID, Name: group.Name, RateMultiplier: group.RateMultiplier, Selected: true,
			})
			seen[key] = entry
		}
	}
	out := make([]ModelCatalogEntry, 0, len(seen))
	for _, entry := range seen {
		sort.SliceStable(entry.Groups, func(i, j int) bool {
			if entry.Groups[i].RateMultiplier != entry.Groups[j].RateMultiplier {
				return entry.Groups[i].RateMultiplier < entry.Groups[j].RateMultiplier
			}
			return entry.Groups[i].Name < entry.Groups[j].Name
		})
		for i := range entry.Groups {
			entry.Groups[i].Priority = i
		}
		out = append(out, entry)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Platform != out[j].Platform {
			return out[i].Platform < out[j].Platform
		}
		return out[i].Model < out[j].Model
	})
	return out, nil
}

// ModelCatalogModels returns the concrete models explicitly allowed by
// schedulable accounts in a group. A configured group model list is an
// additional upper bound, not an independent source of model availability.
func ModelCatalogModels(group Group, accounts []Account) []string {
	explicitMappings := make([]map[string]string, 0, len(accounts))
	for i := range accounts {
		mapping := stringMappingFromRaw(accounts[i].Credentials["model_mapping"])
		if len(mapping) > 0 {
			explicitMappings = append(explicitMappings, mapping)
		}
	}

	models := make([]string, 0)
	if group.ModelsListConfig.Enabled && len(group.ModelsListConfig.Models) > 0 {
		for _, model := range group.ModelsListConfig.Models {
			for _, mapping := range explicitMappings {
				if mappingSupportsRequestedModel(mapping, strings.TrimSpace(model)) {
					models = append(models, model)
					break
				}
			}
		}
	} else {
		for _, mapping := range explicitMappings {
			for model := range mapping {
				models = append(models, model)
			}
		}
	}
	seen := make(map[string]struct{}, len(models))
	out := make([]string, 0, len(models))
	for _, model := range models {
		model = strings.TrimSpace(model)
		if model == "" || strings.Contains(model, "*") {
			continue
		}
		if _, ok := seen[model]; ok {
			continue
		}
		seen[model] = struct{}{}
		out = append(out, model)
	}
	sort.Strings(out)
	return out
}

func (s *ModelMonitorService) ListRows(ctx context.Context) ([]ModelMonitorRow, error) {
	catalog, err := s.DiscoverCatalog(ctx)
	if err != nil {
		return nil, err
	}
	configs, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	groupConfigs, err := s.repo.ListGroupConfigs(ctx)
	if err != nil {
		return nil, err
	}
	keys := append([]ModelCatalogEntry(nil), catalog...)
	catalogSet := make(map[string]bool, len(catalog))
	for _, entry := range catalog {
		catalogSet[ModelMonitorKey(entry.Platform, entry.Model)] = true
	}
	configByKey := make(map[string]ModelMonitor, len(configs))
	for _, cfg := range configs {
		key := ModelMonitorKey(cfg.Platform, cfg.Model)
		configByKey[key] = cfg
	}
	summaries, err := s.repo.Summaries(ctx, keys, ModelMonitorTimelinePoints)
	if err != nil {
		return nil, err
	}
	rows := make([]ModelMonitorRow, 0, len(keys))
	seen := make(map[string]struct{}, len(keys))
	for _, entry := range keys {
		key := ModelMonitorKey(entry.Platform, entry.Model)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		cfg, configured := configByKey[key]
		if !configured {
			cfg = ModelMonitor{Platform: entry.Platform, Model: entry.Model, IntervalSeconds: ModelMonitorDefaultIntervalSeconds}
		}
		row := ModelMonitorRow{ModelMonitor: cfg, CatalogAvailable: catalogSet[key], Configured: configured, Groups: append([]ModelMonitorGroupOption(nil), entry.Groups...)}
		if configuredIDs, ok := groupConfigs[cfg.ID]; ok {
			row.GroupsConfigured = true
			priorityByID := make(map[int64]int, len(configuredIDs))
			for priority, groupID := range configuredIDs {
				priorityByID[groupID] = priority
			}
			for i := range row.Groups {
				priority, selected := priorityByID[row.Groups[i].GroupID]
				row.Groups[i].Selected = selected
				if selected {
					row.Groups[i].Priority = priority
				} else {
					row.Groups[i].Priority = len(configuredIDs) + i
				}
			}
			sort.SliceStable(row.Groups, func(i, j int) bool { return row.Groups[i].Priority < row.Groups[j].Priority })
		}
		if summary, ok := summaries[key]; ok {
			copy := summary
			row.Summary = &copy
		}
		rows = append(rows, row)
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Platform != rows[j].Platform {
			return rows[i].Platform < rows[j].Platform
		}
		return rows[i].Model < rows[j].Model
	})
	return rows, nil
}

func (s *ModelMonitorService) ConfigureGroups(ctx context.Context, platform, model string, groupIDs []int64, createdBy int64) (*ModelMonitor, error) {
	if len(groupIDs) == 0 {
		return nil, fmt.Errorf("at least one monitoring group is required")
	}
	catalog, err := s.DiscoverCatalog(ctx)
	if err != nil {
		return nil, err
	}
	valid := make(map[int64]struct{})
	for _, entry := range catalog {
		if entry.Platform == platform && entry.Model == model {
			for _, group := range entry.Groups {
				valid[group.GroupID] = struct{}{}
			}
		}
	}
	seen := make(map[int64]struct{}, len(groupIDs))
	for _, groupID := range groupIDs {
		if _, ok := valid[groupID]; !ok {
			return nil, fmt.Errorf("group %d does not currently support model %s", groupID, model)
		}
		if _, duplicate := seen[groupID]; duplicate {
			return nil, fmt.Errorf("duplicate monitoring group %d", groupID)
		}
		seen[groupID] = struct{}{}
	}
	m, err := s.repo.GetByKey(ctx, platform, model)
	if err != nil {
		return nil, err
	}
	if m == nil {
		m, err = s.Upsert(ctx, platform, model, false, ModelMonitorDefaultIntervalSeconds, 0, "", createdBy)
		if err != nil {
			return nil, err
		}
	}
	if err := s.repo.ReplaceGroups(ctx, m.ID, groupIDs); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *ModelMonitorService) Upsert(ctx context.Context, platform, model string, enabled bool, intervalSeconds, displayOrder int, label string, createdBy int64) (*ModelMonitor, error) {
	platform = strings.TrimSpace(platform)
	model = strings.TrimSpace(model)
	label = strings.TrimSpace(label)
	if !validModelMonitorPlatform(platform) || model == "" || len(model) > 200 {
		return nil, fmt.Errorf("invalid platform or model")
	}
	if intervalSeconds == 0 {
		intervalSeconds = ModelMonitorDefaultIntervalSeconds
	}
	if intervalSeconds < ModelMonitorMinIntervalSeconds || intervalSeconds > ModelMonitorMaxIntervalSeconds {
		return nil, fmt.Errorf("interval_seconds must be between %d and %d", ModelMonitorMinIntervalSeconds, ModelMonitorMaxIntervalSeconds)
	}
	if displayOrder < 0 || displayOrder > 9999 {
		return nil, fmt.Errorf("display_order must be between 0 and 9999")
	}
	if len([]rune(label)) > 12 {
		return nil, fmt.Errorf("label must not exceed 12 characters")
	}
	m := &ModelMonitor{Platform: platform, Model: model, Enabled: enabled, IntervalSeconds: intervalSeconds, DisplayOrder: displayOrder, Label: label, CreatedBy: createdBy}
	if err := s.repo.Upsert(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

func validModelMonitorPlatform(platform string) bool {
	switch platform {
	case PlatformAnthropic, PlatformOpenAI, PlatformGemini, PlatformAntigravity, PlatformGrok:
		return true
	}
	return false
}

func (s *ModelMonitorService) RunByKey(ctx context.Context, platform, model string, createdBy int64) (*ModelMonitorHistory, error) {
	m, err := s.repo.GetByKey(ctx, platform, model)
	if err != nil {
		return nil, err
	}
	if m == nil {
		m, err = s.Upsert(ctx, platform, model, false, ModelMonitorDefaultIntervalSeconds, 0, "", createdBy)
		if err != nil {
			return nil, err
		}
	}
	return s.Run(ctx, m)
}

func (s *ModelMonitorService) Run(ctx context.Context, monitor *ModelMonitor) (*ModelMonitorHistory, error) {
	started := time.Now()
	excluded := make(map[int64]struct{})
	attempts := 0
	lastMessage := ""
	finalStatus := MonitorStatusFailed
	groups, err := s.monitorGroups(ctx, monitor)
	if err != nil {
		return nil, err
	}
	var lastGroup *ModelMonitorGroupOption
	for i := range groups {
		group := groups[i]
		for attempts < modelMonitorMaxAttempts {
			account, selectErr := s.selectAccount(ctx, monitor.Platform, monitor.Model, group.GroupID, excluded)
			if selectErr != nil {
				lastMessage = fmt.Sprintf("%s: %s", group.Name, selectErr.Error())
				break
			}
			lastGroup = &group
			attempts++
			excluded[account.ID] = struct{}{}
			result, testErr := s.accountTester.RunTestBackground(ctx, account.ID, monitor.Model)
			if testErr == nil && result != nil && result.Status == "success" {
				latency := int(time.Since(started).Milliseconds())
				finalStatus = MonitorStatusOperational
				if attempts > 1 || time.Duration(latency)*time.Millisecond > modelMonitorDegradedThreshold {
					finalStatus = MonitorStatusDegraded
				}
				groupID := group.GroupID
				h := &ModelMonitorHistory{MonitorID: monitor.ID, Status: finalStatus, LatencyMs: &latency, Attempts: attempts, GroupID: &groupID, GroupName: group.Name, CheckedAt: time.Now()}
				if err := s.persistResult(ctx, monitor, h); err != nil {
					return nil, err
				}
				return h, nil
			}
			if testErr != nil {
				lastMessage = testErr.Error()
			} else if result != nil {
				lastMessage = result.ErrorMessage
			}
		}
		if attempts >= modelMonitorMaxAttempts {
			break
		}
	}
	if attempts == 0 {
		finalStatus = MonitorStatusError
	}
	latency := int(time.Since(started).Milliseconds())
	h := &ModelMonitorHistory{MonitorID: monitor.ID, Status: finalStatus, LatencyMs: &latency, Attempts: attempts, Message: truncateMessage(sanitizeErrorMessage(lastMessage)), CheckedAt: time.Now()}
	if lastGroup != nil {
		groupID := lastGroup.GroupID
		h.GroupID, h.GroupName = &groupID, lastGroup.Name
	}
	if err := s.persistResult(ctx, monitor, h); err != nil {
		return nil, err
	}
	return h, nil
}

func (s *ModelMonitorService) monitorGroups(ctx context.Context, monitor *ModelMonitor) ([]ModelMonitorGroupOption, error) {
	catalog, err := s.DiscoverCatalog(ctx)
	if err != nil {
		return nil, err
	}
	var groups []ModelMonitorGroupOption
	for _, entry := range catalog {
		if entry.Platform == monitor.Platform && entry.Model == monitor.Model {
			groups = append(groups, entry.Groups...)
			break
		}
	}
	configured, err := s.repo.ListGroupConfigs(ctx)
	if err != nil {
		return nil, err
	}
	ids, ok := configured[monitor.ID]
	if !ok {
		return groups, nil
	}
	byID := make(map[int64]ModelMonitorGroupOption, len(groups))
	for _, group := range groups {
		byID[group.GroupID] = group
	}
	out := make([]ModelMonitorGroupOption, 0, len(ids))
	for priority, id := range ids {
		if group, exists := byID[id]; exists {
			group.Priority = priority
			out = append(out, group)
		}
	}
	return out, nil
}

func (s *ModelMonitorService) selectAccount(ctx context.Context, platform, model string, groupID int64, excluded map[int64]struct{}) (*Account, error) {
	switch platform {
	case PlatformOpenAI, PlatformGrok:
		return s.openAIGateway.selectAccountForModelWithExclusions(ctx, &groupID, platform, "", model, excluded, false, 0, "", false)
	case PlatformGemini, PlatformAntigravity:
		ctx = context.WithValue(ctx, ctxkey.ForcePlatform, platform)
		return s.gateway.SelectAccountForModelWithExclusions(ctx, &groupID, "", model, excluded)
	default:
		return s.gateway.SelectAccountForModelWithExclusions(ctx, &groupID, "", model, excluded)
	}
}

func (s *ModelMonitorService) persistResult(ctx context.Context, m *ModelMonitor, h *ModelMonitorHistory) error {
	if err := s.repo.InsertHistory(ctx, h); err != nil {
		return err
	}
	m.LastCheckedAt = &h.CheckedAt
	return s.repo.UpdateLastChecked(ctx, m.ID, h.CheckedAt)
}
func (s *ModelMonitorService) History(ctx context.Context, id int64, limit int) ([]ModelMonitorHistory, error) {
	return s.repo.ListHistory(ctx, id, limit)
}
func (s *ModelMonitorService) PublicSummaries(ctx context.Context, keys []ModelCatalogEntry) (map[string]ModelMonitorSummary, error) {
	return s.repo.Summaries(ctx, keys, ModelMonitorTimelinePoints)
}

type ModelMonitorRunner struct {
	service     *ModelMonitorService
	settings    *SettingService
	stop        chan struct{}
	once        sync.Once
	sem         chan struct{}
	inFlight    sync.Map
	lastCleanup time.Time
}

func NewModelMonitorRunner(service *ModelMonitorService, settings *SettingService) *ModelMonitorRunner {
	return &ModelMonitorRunner{service: service, settings: settings, stop: make(chan struct{}), sem: make(chan struct{}, 5)}
}

func (r *ModelMonitorRunner) Start() {
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		r.tick()
		for {
			select {
			case <-ticker.C:
				r.tick()
			case <-r.stop:
				return
			}
		}
	}()
}

func (r *ModelMonitorRunner) Stop() { r.once.Do(func() { close(r.stop) }) }

func (r *ModelMonitorRunner) tick() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	if r.settings == nil || !r.settings.GetModelMonitorRuntime(ctx).Enabled {
		return
	}
	monitors, err := r.service.repo.ListEnabledDue(ctx, time.Now())
	if err != nil {
		return
	}
	catalog, err := r.service.DiscoverCatalog(ctx)
	if err != nil {
		return
	}
	available := make(map[string]bool, len(catalog))
	for _, entry := range catalog {
		available[ModelMonitorKey(entry.Platform, entry.Model)] = true
	}
	for i := range monitors {
		monitor := monitors[i]
		if !available[ModelMonitorKey(monitor.Platform, monitor.Model)] {
			continue
		}
		if _, running := r.inFlight.LoadOrStore(monitor.ID, struct{}{}); running {
			continue
		}
		select {
		case r.sem <- struct{}{}:
			go func(m ModelMonitor) {
				defer func() { <-r.sem; r.inFlight.Delete(m.ID) }()
				runCtx, runCancel := context.WithTimeout(context.Background(), 2*time.Minute)
				defer runCancel()
				_, _ = r.service.Run(runCtx, &m)
			}(monitor)
		default:
			r.inFlight.Delete(monitor.ID)
		}
	}
	if time.Since(r.lastCleanup) >= 24*time.Hour {
		_, _ = r.service.repo.DeleteHistoryBefore(ctx, time.Now().AddDate(0, 0, -ModelMonitorHistoryRetentionDays))
		r.lastCleanup = time.Now()
	}
}
