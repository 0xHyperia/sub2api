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
	billing       *BillingService
	channels      *ChannelService
}

func NewModelMonitorService(repo ModelMonitorRepository, groupRepo GroupRepository, accountRepo AccountRepository, accountTester *AccountTestService, gateway *GatewayService, openAIGateway *OpenAIGatewayService, billing *BillingService, channels *ChannelService) *ModelMonitorService {
	return &ModelMonitorService{repo: repo, groupRepo: groupRepo, accountRepo: accountRepo, accountTester: accountTester, gateway: gateway, openAIGateway: openAIGateway, billing: billing, channels: channels}
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

func (s *ModelMonitorService) ListRows(ctx context.Context, resolution ModelMonitorResolution) ([]ModelMonitorRow, error) {
	if resolution != ModelMonitorResolutionHour {
		resolution = ModelMonitorResolutionMinute
	}
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
	for _, cfg := range configs {
		key := ModelMonitorKey(cfg.Platform, cfg.Model)
		if !catalogSet[key] {
			keys = append(keys, ModelCatalogEntry{Platform: cfg.Platform, Model: cfg.Model})
		}
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
	metricScopes := make([]ModelMonitorMetricScope, 0, len(configs))
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
		row := ModelMonitorRow{ModelMonitor: cfg, CatalogAvailable: catalogSet[key], Configured: configured, Groups: append([]ModelMonitorGroupOption{}, entry.Groups...)}
		if configuredGroups, ok := groupConfigs[cfg.ID]; ok {
			row.GroupsConfigured = true
			configByID := make(map[int64]ModelMonitorGroupConfig, len(configuredGroups))
			for _, groupConfig := range configuredGroups {
				configByID[groupConfig.GroupID] = groupConfig
			}
			for i := range row.Groups {
				groupConfig, selected := configByID[row.Groups[i].GroupID]
				row.Groups[i].Selected = selected
				if selected {
					row.Groups[i].Priority = groupConfig.Priority
					row.Groups[i].Enabled = groupConfig.Enabled
					row.Groups[i].IntervalSeconds = groupConfig.IntervalSeconds
					row.Groups[i].FailureCompensationEnabled = groupConfig.FailureCompensationEnabled
					row.Groups[i].FailureCompensationPending = groupConfig.FailureCompensationPending
					row.Groups[i].LastTrafficAt = groupConfig.LastTrafficAt
					row.Groups[i].LastProbeAt = groupConfig.LastProbeAt
					row.Groups[i].LastScheduledSlotAt = groupConfig.LastScheduledSlotAt
					row.Groups[i].NextCompensationAt = groupConfig.NextCompensationAt
					row.Groups[i].ConsecutiveProbeFailures = groupConfig.ConsecutiveProbeFailures
				} else {
					row.Groups[i].Priority = len(configuredGroups) + i
					row.Groups[i].IntervalSeconds = ModelMonitorDefaultIntervalSeconds
				}
			}
			sort.SliceStable(row.Groups, func(i, j int) bool { return row.Groups[i].Priority < row.Groups[j].Priority })
		} else {
			for i := range row.Groups {
				row.Groups[i].Enabled = configured && cfg.Enabled
				row.Groups[i].IntervalSeconds = max(cfg.IntervalSeconds, ModelMonitorMinIntervalSeconds)
			}
		}
		if configured && cfg.ID > 0 {
			groupIDs := make([]int64, 0, len(row.Groups))
			for i := range row.Groups {
				groupIDs = append(groupIDs, row.Groups[i].GroupID)
			}
			metricScopes = append(metricScopes, ModelMonitorMetricScope{MonitorID: cfg.ID, GroupIDs: groupIDs})
		}
		if summary, ok := summaries[key]; ok {
			copy := summary
			row.Summary = &copy
		}
		rows = append(rows, row)
	}
	metricsByMonitor, err := s.repo.GroupMetricsBatch(ctx, metricScopes, resolution, time.Now())
	if err != nil {
		return nil, err
	}
	for rowIndex := range rows {
		metrics := metricsByMonitor[rows[rowIndex].ID]
		for groupIndex := range rows[rowIndex].Groups {
			if metric, ok := metrics[rows[rowIndex].Groups[groupIndex].GroupID]; ok {
				copy := metric
				rows[rowIndex].Groups[groupIndex].Metrics = &copy
			}
		}
		if metric, ok := averageModelMonitorMetrics(metrics, groupIDsForModel(rows[rowIndex].Groups)); ok && rows[rowIndex].Summary != nil {
			copy := metric
			rows[rowIndex].Summary.Metrics = &copy
		}
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

func (s *ModelMonitorService) ConfigureGroup(ctx context.Context, platform, model string, groupID int64, enabled bool, intervalSeconds int, failureCompensationEnabled bool, createdBy int64) (*ModelMonitor, error) {
	if intervalSeconds == 0 {
		intervalSeconds = ModelMonitorDefaultIntervalSeconds
	}
	if intervalSeconds < ModelMonitorMinIntervalSeconds || intervalSeconds > ModelMonitorMaxIntervalSeconds {
		return nil, fmt.Errorf("interval_seconds must be between %d and %d", ModelMonitorMinIntervalSeconds, ModelMonitorMaxIntervalSeconds)
	}
	platform, model = strings.TrimSpace(platform), strings.TrimSpace(model)
	catalog, err := s.DiscoverCatalog(ctx)
	if err != nil {
		return nil, err
	}
	valid := false
	for _, entry := range catalog {
		if entry.Platform != platform || entry.Model != model {
			continue
		}
		for _, group := range entry.Groups {
			if group.GroupID == groupID {
				valid = true
				break
			}
		}
	}
	if !valid {
		return nil, fmt.Errorf("group %d does not currently support model %s", groupID, model)
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
	if err := s.repo.UpsertGroupConfig(ctx, ModelMonitorGroupConfig{
		MonitorID: m.ID, GroupID: groupID, Enabled: enabled, IntervalSeconds: intervalSeconds,
		FailureCompensationEnabled: failureCompensationEnabled,
	}); err != nil {
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

func (s *ModelMonitorService) RunGroupByKey(ctx context.Context, platform, model string, groupID int64, createdBy int64) (*ModelMonitorHistory, error) {
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
	groups, err := s.DiscoverCatalog(ctx)
	if err != nil {
		return nil, err
	}
	for _, entry := range groups {
		if entry.Platform != platform || entry.Model != model {
			continue
		}
		for _, group := range entry.Groups {
			if group.GroupID == groupID {
				history, runErr := s.RunGroup(ctx, m, group)
				if runErr != nil {
					return nil, runErr
				}
				if err := s.repo.FinishGroupProbe(ctx, m.ID, groupID, nil, history.CheckedAt, history.Status); err != nil {
					return nil, err
				}
				return history, nil
			}
		}
	}
	return nil, fmt.Errorf("group %d does not currently support model %s", groupID, model)
}

func (s *ModelMonitorService) RunGroup(ctx context.Context, monitor *ModelMonitor, group ModelMonitorGroupOption) (*ModelMonitorHistory, error) {
	started := time.Now()
	excluded := make(map[int64]struct{})
	attempts := 0
	lastMessage := ""
	var finalResult *ScheduledTestResult
	costs := probeCostAccumulator{known: true}
	for attempts < modelMonitorMaxAttempts {
		account, selectErr := s.selectAccount(ctx, monitor.Platform, monitor.Model, group.GroupID, excluded)
		if selectErr != nil {
			lastMessage = selectErr.Error()
			break
		}
		attempts++
		excluded[account.ID] = struct{}{}
		result, testErr := s.accountTester.RunTestBackground(ctx, account.ID, monitor.Model)
		finalResult = result
		costs.add(result, s.probeCost(ctx, monitor.Model, group.GroupID, account, result))
		if testErr == nil && result != nil && result.Status == "success" {
			break
		}
		if testErr != nil {
			lastMessage = testErr.Error()
		} else if result != nil {
			lastMessage = result.ErrorMessage
		}
	}
	checkedAt := time.Now()
	latency := int(checkedAt.Sub(started).Milliseconds())
	status := MonitorStatusFailed
	if attempts == 0 {
		status = MonitorStatusError
	}
	if finalResult != nil && finalResult.Status == "success" {
		status = MonitorStatusOperational
		if attempts > 1 || time.Duration(latency)*time.Millisecond > modelMonitorDegradedThreshold {
			status = MonitorStatusDegraded
		}
	}
	groupID := group.GroupID
	h := &ModelMonitorHistory{MonitorID: monitor.ID, Status: status, LatencyMs: &latency, Attempts: attempts, Message: truncateMessage(sanitizeErrorMessage(lastMessage)), GroupID: &groupID, GroupName: group.Name, CheckedAt: checkedAt}
	h.ProbeCostKnown = costs.known
	h.ProbeCost = costs.value()
	if finalResult != nil {
		h.FirstTokenMs = finalResult.FirstTokenMs
		h.InputTokens = finalResult.InputTokens
		h.OutputTokens = finalResult.OutputTokens
		h.GenerationMs = finalResult.GenerationMs
	}
	if err := s.persistGroupResult(ctx, monitor, h); err != nil {
		return nil, err
	}
	return h, nil
}

type probeCostAccumulator struct {
	total    float64
	billable bool
	known    bool
}

func (a *probeCostAccumulator) add(result *ScheduledTestResult, cost *float64) {
	if result == nil || (result.InputTokens <= 0 && result.OutputTokens <= 0) {
		return
	}
	a.billable = true
	if cost == nil {
		a.known = false
		return
	}
	a.total += *cost
}

func (a probeCostAccumulator) value() *float64 {
	if !a.billable || !a.known {
		return nil
	}
	value := a.total
	return &value
}

func (s *ModelMonitorService) probeCost(ctx context.Context, model string, groupID int64, account *Account, result *ScheduledTestResult) *float64 {
	if account == nil || result == nil || (result.InputTokens <= 0 && result.OutputTokens <= 0) {
		return nil
	}
	tokens := UsageTokens{InputTokens: result.InputTokens, OutputTokens: result.OutputTokens}
	if cost := resolveAccountStatsCost(ctx, s.channels, s.billing, account.ID, groupID, model, tokens, 1, 0); cost != nil {
		return applyProbeCostMultiplier(cost, account)
	}
	if s.billing == nil {
		return nil
	}
	breakdown, err := s.billing.CalculateCost(model, tokens, account.BillingRateMultiplier())
	if err != nil || breakdown == nil {
		return nil
	}
	value := breakdown.TotalCost
	return &value
}

func applyProbeCostMultiplier(cost *float64, account *Account) *float64 {
	if cost == nil || account == nil {
		return nil
	}
	value := *cost * account.BillingRateMultiplier()
	return &value
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
	configs, ok := configured[monitor.ID]
	if !ok {
		return groups, nil
	}
	byID := make(map[int64]ModelMonitorGroupOption, len(groups))
	for _, group := range groups {
		byID[group.GroupID] = group
	}
	out := make([]ModelMonitorGroupOption, 0, len(configs))
	for _, config := range configs {
		if group, exists := byID[config.GroupID]; exists {
			group.Priority = config.Priority
			group.Enabled = config.Enabled
			group.IntervalSeconds = config.IntervalSeconds
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

func (s *ModelMonitorService) persistGroupResult(ctx context.Context, m *ModelMonitor, h *ModelMonitorHistory) error {
	if h.GroupID == nil {
		return fmt.Errorf("monitor group is required")
	}
	if err := s.repo.InsertHistory(ctx, h); err != nil {
		return err
	}
	cost := 0.0
	if h.ProbeCost != nil {
		cost = *h.ProbeCost
	}
	delta := ModelMonitorMetricDelta{
		MonitorID: m.ID, GroupID: *h.GroupID, Source: "probe", BucketStart: h.CheckedAt,
		RequestCount: 1, ProbeCost: cost, ProbeCostKnown: h.ProbeCostKnown,
	}
	if h.Status == MonitorStatusOperational || h.Status == MonitorStatusDegraded {
		delta.SuccessCount = 1
		if h.LatencyMs != nil {
			delta.LatencySumMs = int64(*h.LatencyMs)
			delta.LatencyCount = 1
		}
		if h.FirstTokenMs != nil {
			delta.TTFTSumMs = int64(*h.FirstTokenMs)
			delta.TTFTCount = 1
		}
		delta.OutputTokens = int64(h.OutputTokens)
		delta.GenerationMs = h.GenerationMs
	}
	if err := s.repo.UpsertMetricDelta(ctx, delta); err != nil {
		return err
	}
	if err := s.repo.UpdateGroupProbeAt(ctx, m.ID, *h.GroupID, h.CheckedAt); err != nil {
		return err
	}
	m.LastCheckedAt = &h.CheckedAt
	return s.repo.UpdateLastChecked(ctx, m.ID, h.CheckedAt)
}
func (s *ModelMonitorService) History(ctx context.Context, id int64, limit int) ([]ModelMonitorHistory, error) {
	return s.repo.ListHistory(ctx, id, limit)
}
func (s *ModelMonitorService) PublicSummaries(ctx context.Context, keys []ModelCatalogEntry, resolution ModelMonitorResolution) (map[string]ModelMonitorSummary, error) {
	if resolution != ModelMonitorResolutionMinute {
		resolution = ModelMonitorResolutionHour
	}
	summaries, err := s.repo.Summaries(ctx, keys, ModelMonitorTimelinePoints)
	if err != nil {
		return nil, err
	}
	monitors, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	catalog, err := s.DiscoverCatalog(ctx)
	if err != nil {
		return nil, err
	}
	catalogByKey := make(map[string]ModelCatalogEntry, len(catalog))
	for _, entry := range catalog {
		catalogByKey[ModelMonitorKey(entry.Platform, entry.Model)] = entry
	}
	scopes := make([]ModelMonitorMetricScope, 0, len(monitors))
	for _, monitor := range monitors {
		entry, available := catalogByKey[ModelMonitorKey(monitor.Platform, monitor.Model)]
		if !available {
			continue
		}
		groupIDs := make([]int64, 0, len(entry.Groups))
		for _, group := range entry.Groups {
			groupIDs = append(groupIDs, group.GroupID)
		}
		scopes = append(scopes, ModelMonitorMetricScope{MonitorID: monitor.ID, GroupIDs: groupIDs})
	}
	now := time.Now()
	metricsByMonitor, err := s.repo.GroupMetricsBatch(ctx, scopes, resolution, now)
	if err != nil {
		return nil, err
	}
	var hourlyByMonitor map[int64]map[int64]ModelMonitorGroupMetrics
	if resolution == ModelMonitorResolutionMinute {
		hourlyByMonitor, err = s.repo.GroupMetricsBatch(ctx, scopes, ModelMonitorResolutionHour, now)
		if err != nil {
			return nil, err
		}
	}
	for _, monitor := range monitors {
		key := ModelMonitorKey(monitor.Platform, monitor.Model)
		summary, ok := summaries[key]
		entry, available := catalogByKey[key]
		if !ok || !available {
			continue
		}
		metrics := metricsByMonitor[monitor.ID]
		if metric, exists := averageModelMonitorMetrics(metrics, groupIDsForModel(entry.Groups)); exists {
			copy := metric
			copy.ProbeCost = nil
			summary.Metrics = &copy
			if resolution == ModelMonitorResolutionHour {
				summary.HourlyMetrics = &copy
			}
		}
		if resolution == ModelMonitorResolutionMinute {
			hourly := hourlyByMonitor[monitor.ID]
			if metric, exists := averageModelMonitorMetrics(hourly, groupIDsForModel(entry.Groups)); exists {
				copy := metric
				copy.ProbeCost = nil
				summary.HourlyMetrics = &copy
			}
		}
		for _, group := range entry.Groups {
			if metric, exists := metrics[group.GroupID]; exists {
				metric.ProbeCost = nil
				summary.Groups = append(summary.Groups, ModelMonitorPublicGroupMetrics{GroupID: group.GroupID, Name: group.Name, Metrics: metric})
			}
		}
		RedactModelMonitorSampleCounts(&summary)
		summaries[key] = summary
	}
	return summaries, nil
}

func groupIDsForModel(groups []ModelMonitorGroupOption) []int64 {
	ids := make([]int64, 0, len(groups))
	for _, group := range groups {
		ids = append(ids, group.GroupID)
	}
	return ids
}

// averageModelMonitorMetrics builds model-level performance from independent
// group metrics. The repository's group_id=0 aggregate remains the source for
// total request counts and probe cost, while performance values are averaged
// across groups so a high-volume group cannot dominate the model view.
func averageModelMonitorMetrics(metrics map[int64]ModelMonitorGroupMetrics, groupIDs []int64) (ModelMonitorGroupMetrics, bool) {
	aggregate, hasAggregate := metrics[0]
	candidates := make([]ModelMonitorGroupMetrics, 0, len(groupIDs))
	seen := make(map[int64]struct{}, len(groupIDs))
	for _, groupID := range groupIDs {
		if groupID <= 0 {
			continue
		}
		if _, ok := seen[groupID]; ok {
			continue
		}
		seen[groupID] = struct{}{}
		metric, ok := metrics[groupID]
		if ok && modelMonitorMetricHasData(metric) {
			candidates = append(candidates, metric)
		}
	}
	if len(candidates) == 0 {
		return aggregate, hasAggregate && modelMonitorMetricHasData(aggregate)
	}

	result := aggregate
	result.TPS = averageMetricValue(candidates, func(metric ModelMonitorGroupMetrics) *float64 { return metric.TPS })
	result.TTFTMs = averageMetricValue(candidates, func(metric ModelMonitorGroupMetrics) *float64 { return metric.TTFTMs })
	result.AverageLatencyMs = averageMetricValue(candidates, func(metric ModelMonitorGroupMetrics) *float64 { return metric.AverageLatencyMs })
	result.SuccessRate = averageMetricValue(candidates, func(metric ModelMonitorGroupMetrics) *float64 { return metric.SuccessRate })
	result.Buckets = averageModelMonitorBuckets(candidates, aggregate.Buckets)
	return result, true
}

func modelMonitorMetricHasData(metric ModelMonitorGroupMetrics) bool {
	if metric.TPS != nil || metric.TTFTMs != nil || metric.AverageLatencyMs != nil || metric.SuccessRate != nil || metric.RequestCount != nil || metric.SuccessCount != nil || metric.FailureCount != nil {
		return true
	}
	for _, bucket := range metric.Buckets {
		if bucket.SuccessRate != nil || bucket.TTFTMs != nil {
			return true
		}
	}
	return false
}

func averageMetricValue(metrics []ModelMonitorGroupMetrics, value func(ModelMonitorGroupMetrics) *float64) *float64 {
	var total float64
	count := 0
	for _, metric := range metrics {
		if current := value(metric); current != nil {
			total += *current
			count++
		}
	}
	if count == 0 {
		return nil
	}
	result := total / float64(count)
	return &result
}

func averageModelMonitorBuckets(metrics []ModelMonitorGroupMetrics, aggregate []ModelMonitorMetricBucket) []ModelMonitorMetricBucket {
	times := make(map[time.Time]struct{}, len(aggregate))
	for _, bucket := range aggregate {
		times[bucket.StartedAt.UTC()] = struct{}{}
	}
	for _, metric := range metrics {
		for _, bucket := range metric.Buckets {
			times[bucket.StartedAt.UTC()] = struct{}{}
		}
	}
	if len(times) == 0 {
		return nil
	}
	startedAt := make([]time.Time, 0, len(times))
	for value := range times {
		startedAt = append(startedAt, value)
	}
	sort.Slice(startedAt, func(i, j int) bool { return startedAt[i].Before(startedAt[j]) })
	result := make([]ModelMonitorMetricBucket, 0, len(startedAt))
	for _, started := range startedAt {
		var successTotal, ttftTotal float64
		var successCount, ttftCount int
		for _, metric := range metrics {
			for _, bucket := range metric.Buckets {
				if !bucket.StartedAt.UTC().Equal(started) {
					continue
				}
				if bucket.SuccessRate != nil {
					successTotal += *bucket.SuccessRate
					successCount++
				}
				if bucket.TTFTMs != nil {
					ttftTotal += *bucket.TTFTMs
					ttftCount++
				}
			}
		}
		bucket := ModelMonitorMetricBucket{StartedAt: started}
		if successCount > 0 {
			value := successTotal / float64(successCount)
			bucket.SuccessRate = &value
		}
		if ttftCount > 0 {
			value := ttftTotal / float64(ttftCount)
			bucket.TTFTMs = &value
		}
		result = append(result, bucket)
	}
	return result
}

type ModelMonitorRunner struct {
	service     *ModelMonitorService
	settings    *SettingService
	stop        chan struct{}
	once        sync.Once
	sem         chan struct{}
	inFlight    sync.Map
	lastCleanup time.Time
	lastRollup  time.Time
}

func NewModelMonitorRunner(service *ModelMonitorService, settings *SettingService) *ModelMonitorRunner {
	return &ModelMonitorRunner{service: service, settings: settings, stop: make(chan struct{}), sem: make(chan struct{}, 5)}
}

func (r *ModelMonitorRunner) Start() {
	go func() {
		r.tick()
		for {
			now := time.Now()
			// Keep the first scan on each natural minute boundary, while polling
			// within the slot so concurrency-limited backlogs are not delayed by
			// a full minute.
			nextScan := now.Truncate(15 * time.Second).Add(15 * time.Second)
			timer := time.NewTimer(time.Until(nextScan))
			select {
			case <-timer.C:
				r.tick()
			case <-r.stop:
				if !timer.Stop() {
					<-timer.C
				}
				return
			}
		}
	}()
}

func (r *ModelMonitorRunner) Stop() { r.once.Do(func() { close(r.stop) }) }

func (r *ModelMonitorRunner) tick() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	now := time.Now()
	catalog, err := r.service.DiscoverCatalog(ctx)
	if err != nil {
		return
	}
	r.reconcileLegacyGroups(ctx, catalog)
	_ = r.service.repo.RefreshTrafficMetrics(ctx, now.Add(-3*time.Minute), now)
	rollupHour := now.UTC().Truncate(time.Hour).Add(-time.Hour)
	if !r.lastRollup.Equal(rollupHour) {
		if err := r.service.repo.RollupHourlyMetrics(ctx, rollupHour); err == nil {
			r.lastRollup = rollupHour
		}
	}
	if r.settings == nil || !r.settings.GetModelMonitorRuntime(ctx).Enabled {
		r.cleanup(ctx, now)
		return
	}
	available := make(map[string]map[int64]ModelMonitorGroupOption, len(catalog))
	for _, entry := range catalog {
		groups := make(map[int64]ModelMonitorGroupOption, len(entry.Groups))
		for _, group := range entry.Groups {
			groups[group.GroupID] = group
		}
		available[ModelMonitorKey(entry.Platform, entry.Model)] = groups
	}
	due, err := r.service.repo.ListEnabledDueGroups(ctx, now, 100)
	if err != nil {
		return
	}
	for i := range due {
		item := due[i]
		group, ok := available[ModelMonitorKey(item.Monitor.Platform, item.Monitor.Model)][item.Group.GroupID]
		if !ok {
			continue
		}
		select {
		case r.sem <- struct{}{}:
		default:
			continue
		}
		flightKey := fmt.Sprintf("%d:%d", item.Monitor.ID, item.Group.GroupID)
		if _, running := r.inFlight.LoadOrStore(flightKey, struct{}{}); running {
			<-r.sem
			continue
		}
		compensation := modelMonitorCompensationDue(item.Group, now)
		scheduledSlot := alignedModelMonitorSlot(now, item.Group.IntervalSeconds)
		if compensation {
			scheduledSlot = now.UTC().Truncate(time.Minute)
		}
		claimed, claimErr := r.service.repo.ClaimDueGroup(ctx, item.Monitor.ID, item.Group.GroupID, now, now.Add(3*time.Minute), scheduledSlot, compensation)
		if claimErr != nil || !claimed {
			r.inFlight.Delete(flightKey)
			<-r.sem
			continue
		}
		if !compensation {
			if modelMonitorSlotHasTraffic(item.Group, scheduledSlot) {
				_ = r.service.repo.CompleteGroupSlotWithoutProbe(ctx, item.Monitor.ID, item.Group.GroupID, scheduledSlot)
				r.inFlight.Delete(flightKey)
				<-r.sem
				continue
			}
		}
		go func(item ModelMonitorDueGroup, group ModelMonitorGroupOption, key string, slot time.Time) {
			defer func() { <-r.sem; r.inFlight.Delete(key) }()
			runCtx, runCancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer runCancel()
			history, runErr := r.service.RunGroup(runCtx, &item.Monitor, group)
			if runErr != nil || history == nil {
				return
			}
			_ = r.service.repo.FinishGroupProbe(runCtx, item.Monitor.ID, item.Group.GroupID, &slot, history.CheckedAt, history.Status)
		}(item, group, flightKey, scheduledSlot)
	}
	r.cleanup(ctx, now)
}

func alignedModelMonitorSlot(now time.Time, intervalSeconds int) time.Time {
	if intervalSeconds < ModelMonitorMinIntervalSeconds {
		intervalSeconds = ModelMonitorDefaultIntervalSeconds
	}
	unix := now.UTC().Unix()
	interval := int64(intervalSeconds)
	return time.Unix(unix-(unix%interval), 0).UTC()
}

func modelMonitorCompensationDue(group ModelMonitorGroupConfig, now time.Time) bool {
	return group.Enabled && group.FailureCompensationEnabled && group.FailureCompensationPending &&
		group.NextCompensationAt != nil && !group.NextCompensationAt.After(now)
}

func modelMonitorSlotHasTraffic(group ModelMonitorGroupConfig, scheduledSlot time.Time) bool {
	if group.LastTrafficAt == nil {
		return false
	}
	windowStart := scheduledSlot.Add(-time.Duration(group.IntervalSeconds) * time.Second)
	return !group.LastTrafficAt.Before(windowStart)
}

func (r *ModelMonitorRunner) reconcileLegacyGroups(ctx context.Context, catalog []ModelCatalogEntry) {
	monitors, err := r.service.repo.List(ctx)
	if err != nil {
		return
	}
	configs, err := r.service.repo.ListGroupConfigs(ctx)
	if err != nil {
		return
	}
	byKey := make(map[string]ModelMonitor, len(monitors))
	for _, monitor := range monitors {
		byKey[ModelMonitorKey(monitor.Platform, monitor.Model)] = monitor
	}
	for _, entry := range catalog {
		key := ModelMonitorKey(entry.Platform, entry.Model)
		monitor, existed := byKey[key]
		if !existed {
			created, createErr := r.service.Upsert(ctx, entry.Platform, entry.Model, false, ModelMonitorDefaultIntervalSeconds, 0, "", 0)
			if createErr != nil {
				continue
			}
			monitor = *created
			byKey[key] = monitor
		}
		existingConfigs := configs[monitor.ID]
		configuredGroupIDs := make(map[int64]struct{}, len(existingConfigs))
		for _, config := range existingConfigs {
			configuredGroupIDs[config.GroupID] = struct{}{}
		}
		for _, group := range entry.Groups {
			if _, configured := configuredGroupIDs[group.GroupID]; configured {
				continue
			}
			enabled := false
			interval := ModelMonitorDefaultIntervalSeconds
			if existed && len(existingConfigs) == 0 {
				enabled = monitor.Enabled
				interval = max(monitor.IntervalSeconds, ModelMonitorMinIntervalSeconds)
			}
			_ = r.service.repo.UpsertGroupConfig(ctx, ModelMonitorGroupConfig{MonitorID: monitor.ID, GroupID: group.GroupID, Enabled: enabled, IntervalSeconds: interval})
		}
	}
}

func (r *ModelMonitorRunner) cleanup(ctx context.Context, now time.Time) {
	if time.Since(r.lastCleanup) >= 24*time.Hour {
		_, _ = r.service.repo.DeleteHistoryBefore(ctx, now.AddDate(0, 0, -ModelMonitorHistoryRetentionDays))
		_, _ = r.service.repo.DeleteMetricBucketsBefore(ctx, ModelMonitorResolutionMinute, now.Add(-ModelMonitorMinuteRetentionHours*time.Hour))
		_, _ = r.service.repo.DeleteMetricBucketsBefore(ctx, ModelMonitorResolutionHour, now.AddDate(0, 0, -ModelMonitorHistoryRetentionDays))
		r.lastCleanup = now
	}
}
