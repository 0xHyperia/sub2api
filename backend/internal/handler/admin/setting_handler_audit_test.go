package admin

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestDiffSettingsIncludesChannelMonitorContracts(t *testing.T) {
	before := &service.SystemSettings{
		ChannelMonitorMode:           "v1",
		ChannelMonitorHideThroughput: true,
		ChannelMonitorShowQuota:      false,
	}
	after := *before
	after.ChannelMonitorMode = "v2"
	after.ChannelMonitorHideThroughput = false
	after.ChannelMonitorShowQuota = true

	changed := diffSettings(before, &after, nil, nil, UpdateSettingsRequest{})
	require.Contains(t, changed, "channel_monitor_mode")
	require.Contains(t, changed, "channel_monitor_hide_throughput")
	require.Contains(t, changed, "channel_monitor_show_quota")
}
