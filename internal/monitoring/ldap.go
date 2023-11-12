package monitoring

import (
	"log/slog"
	"time"
)

type LDAPMonitorWatcher struct {
	syncTicker *time.Ticker

	ldap LDAPServerInterface

	monitor MonitorInterface
}

func (m *LDAPMonitorWatcher) sync() {
	// for {
	// 	select {
	// 	case tick := <-m.syncTicker.C:
	// 		m.logger.Debug().Time("value", tick).Msg("Tick")
	// 		m.storeMetrics()
	// 	default:
	// 		continue
	// 	}
	// }
}

func (m *LDAPMonitorWatcher) storeMetrics() {
	stats := m.ldap.GetStats()

	if err := m.monitor.SetLDAPMetric(map[string]string{"type": "conns"}, float64(stats.Conns)); err != nil {
		slog.Error("failed to set metric", err)
	}
	if err := m.monitor.SetLDAPMetric(map[string]string{"type": "binds"}, float64(stats.Binds)); err != nil {
		slog.Error("failed to set metric", err)
	}
	if err := m.monitor.SetLDAPMetric(map[string]string{"type": "unbinds"}, float64(stats.Unbinds)); err != nil {
		slog.Error("failed to set metric", err)
	}
	if err := m.monitor.SetLDAPMetric(map[string]string{"type": "searches"}, float64(stats.Searches)); err != nil {
		slog.Error("failed to set metric", err)
	}
}

func NewLDAPMonitorWatcher(ldap LDAPServerInterface, monitor MonitorInterface) *LDAPMonitorWatcher {
	m := new(LDAPMonitorWatcher)

	m.syncTicker = time.NewTicker(15 * time.Second)
	m.ldap = ldap
	m.monitor = monitor

	m.ldap.SetStats(true)

	go m.sync()

	return m
}
