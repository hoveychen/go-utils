package goutils

import (
	"sync"
	"time"
)

var (
	ChinaTimezone = time.FixedZone("Asia/Shanghai", 8*60*60) // In case the deploying machine is not in +8 Timezone.
)

func GetNow() time.Time {
	return time.Now().In(ChinaTimezone)
}

// CronTick returns a channel which trigger every *interval* time.
// It's almost the same as time.Tick, except it align the start time to integer.
// For example, an interval of 10 minutes, will start in xx:00:00, xx:10:00, etc.
func CronTick(interval time.Duration) <-chan struct{} {
	if interval.Hours() > 24 {
		LogFatal("CronTick don't accept duration more than 1 day.")
	}
	ticker := make(chan struct{})
	go func() {
		now := GetNow()
		time.Sleep(interval - time.Duration(now.UnixNano())%interval)
		t := time.NewTicker(interval)
		defer t.Stop()

		for {
			ticker <- struct{}{}
			<-t.C
		}
	}()
	return ticker
}

// Monitor is a polling helper to get new updates from external source, by comparing the last modified field.
// Example Usage:
//   m := NewMonitor(time.Now(), time.Minute)
//   for {
//       since := m.Next()
//       results := dbQuery(table.updated_at, "$gt", since)
//       for _, result := range results {
//           m.Update(result.LastModified)
//           ... Remaining processing
//       }
//   }
type Monitor struct {
	since time.Time
	c     <-chan struct{}
	first bool
	sync.RWMutex
}

// Update the modified time of last read record.
func (m *Monitor) Update(t time.Time) {
	m.Lock()
	defer m.Unlock()
	if t.After(m.since) {
		m.since = t
	}
}

// Next returns the latest record time to catch up. Any records with update time
// later than the return value is regarded new unprocessed records.
func (m *Monitor) Next() time.Time {
	m.RLock()
	if m.first {
		m.RUnlock()
		m.first = false
		return m.since
	}
	m.RUnlock()
	<-m.c
	return m.since
}

func NewMonitor(since time.Time, duration time.Duration) (m *Monitor) {
	m = &Monitor{}
	m.since = since
	m.c = CronTick(duration)
	m.first = true
	return
}
