package storage

import (
	"fmt"
	"log"
	"time"
)

// partitionFor returns the partition name and date range for the month containing t,
// e.g. any time in October 2026 gives "short_urls_2026_10" covering Oct 1 to Nov 1 (UTC).
func partitionFor(t time.Time) (name string, start, end time.Time) {
	t = t.UTC()
	start = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	end = start.AddDate(0, 1, 0)
	name = fmt.Sprintf("short_urls_%d_%02d", start.Year(), int(start.Month()))
	return name, start, end
}

// ensurePartition creates the monthly partition containing t if it doesn't exist yet.
// Table names can't be query parameters, so the SQL is built with Sprintf; every value comes from the clock, never user input.
func ensurePartition(t time.Time) error {
	name, start, end := partitionFor(t)
	query := fmt.Sprintf(
		"CREATE TABLE IF NOT EXISTS %s PARTITION OF short_urls FOR VALUES FROM ('%s') TO ('%s')",
		name, start.Format(time.RFC3339), end.Format(time.RFC3339),
	)
	_, err := DB.Exec(query)
	return err
}

// EnsurePartitions creates this month's and next month's partitions, so inserts never hit a missing month.
func EnsurePartitions() error {
	now := time.Now()
	_, _, nextMonth := partitionFor(now)

	if err := ensurePartition(now); err != nil {
		return err
	}
	return ensurePartition(nextMonth)
}

// StartPartitionMaintenance re-runs EnsurePartitions every interval in the background.
func StartPartitionMaintenance(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			if err := EnsurePartitions(); err != nil {
				log.Println("partition maintenance failed:", err)
			}
		}
	}()
}
