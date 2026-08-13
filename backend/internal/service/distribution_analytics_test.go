package service

import (
	"testing"
	"time"
)

func TestNormalizeDistributionAnalyticsDays(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input int
		want  int
	}{{7, 7}, {30, 30}, {90, 90}, {0, 30}, {14, 30}, {-1, 30}}
	for _, test := range tests {
		if got := normalizeDistributionAnalyticsDays(test.input); got != test.want {
			t.Fatalf("normalizeDistributionAnalyticsDays(%d) = %d, want %d", test.input, got, test.want)
		}
	}
}

func TestNormalizeDistributionAnalyticsFilter(t *testing.T) {
	t.Parallel()
	location := time.FixedZone("Asia/Shanghai", 8*60*60)
	now := time.Now().In(location)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)
	from := today.AddDate(0, 0, -30).Add(12 * time.Hour)
	to := today.Add(21 * time.Hour)

	normalized, err := normalizeDistributionAnalyticsFilter(DistributionAnalyticsFilter{DateFrom: &from, DateTo: &to})
	if err != nil {
		t.Fatalf("normalize valid range: %v", err)
	}
	if normalized.Days != 31 || normalized.DateFrom.Hour() != 0 || normalized.DateTo.Hour() != 0 {
		t.Fatalf("unexpected normalized range: %#v", normalized)
	}

	if _, err = normalizeDistributionAnalyticsFilter(DistributionAnalyticsFilter{DateFrom: &from}); err == nil {
		t.Fatal("expected an error when only date_from is supplied")
	}
	tooLongFrom := today.AddDate(0, 0, -366)
	if _, err = normalizeDistributionAnalyticsFilter(DistributionAnalyticsFilter{DateFrom: &tooLongFrom, DateTo: &today}); err == nil {
		t.Fatal("expected an error for a range longer than 366 days")
	}
	future := today.AddDate(0, 0, 1)
	if _, err = normalizeDistributionAnalyticsFilter(DistributionAnalyticsFilter{DateFrom: &from, DateTo: &future}); err == nil {
		t.Fatal("expected an error for a future end date")
	}
}
