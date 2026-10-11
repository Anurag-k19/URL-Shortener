package storage

import (
	"testing"
	"time"
)

func TestPartitionFor(t *testing.T) {
	ist := time.FixedZone("IST", 5*60*60+30*60)

	tests := []struct {
		name      string
		in        time.Time
		wantName  string
		wantStart string
		wantEnd   string
	}{
		{"middle of month", time.Date(2026, 10, 10, 15, 0, 0, 0, time.UTC), "short_urls_2026_10", "2026-10-01", "2026-11-01"},
		{"first moment of month", time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC), "short_urls_2026_11", "2026-11-01", "2026-12-01"},
		{"december rolls into next year", time.Date(2026, 12, 31, 23, 59, 0, 0, time.UTC), "short_urls_2026_12", "2026-12-01", "2027-01-01"},
		{"jan 31 does not skip february", time.Date(2027, 1, 31, 12, 0, 0, 0, time.UTC), "short_urls_2027_01", "2027-01-01", "2027-02-01"},
		{"leap year february", time.Date(2028, 2, 29, 12, 0, 0, 0, time.UTC), "short_urls_2028_02", "2028-02-01", "2028-03-01"},
		{"non-UTC time uses its UTC month", time.Date(2026, 11, 1, 3, 0, 0, 0, ist), "short_urls_2026_10", "2026-10-01", "2026-11-01"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, start, end := partitionFor(tt.in)
			gotStart, gotEnd := start.Format("2006-01-02"), end.Format("2006-01-02")
			if name != tt.wantName || gotStart != tt.wantStart || gotEnd != tt.wantEnd {
				t.Errorf("partitionFor(%v) = %s [%s, %s), want %s [%s, %s)",
					tt.in, name, gotStart, gotEnd, tt.wantName, tt.wantStart, tt.wantEnd)
			}
		})
	}
}
