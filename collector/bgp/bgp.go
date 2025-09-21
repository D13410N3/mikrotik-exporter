package bgp

import (
	"context"

	"github.com/mikrotik-exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
)

// Collector implements the collector.Collector interface for BGP metrics
type Collector struct {
	peerStatusDesc     *prometheus.Desc
	peerUptimeDesc     *prometheus.Desc
	prefixesRecvDesc   *prometheus.Desc
	prefixesSentDesc   *prometheus.Desc
}

// NewCollector creates a new BGP collector
func NewCollector() *Collector {
	return &Collector{
		peerStatusDesc: prometheus.NewDesc(
			"mikrotik_bgp_peer_established",
			"BGP peer established status (1 = established, 0 = not established)",
			[]string{"target", "peer_address", "peer_as", "remote_as"},
			nil,
		),
		peerUptimeDesc: prometheus.NewDesc(
			"mikrotik_bgp_peer_uptime_seconds",
			"BGP peer uptime in seconds",
			[]string{"target", "peer_address", "peer_as", "remote_as"},
			nil,
		),
		prefixesRecvDesc: prometheus.NewDesc(
			"mikrotik_bgp_prefixes_received_total",
			"Number of prefixes received from BGP peer",
			[]string{"target", "peer_address", "peer_as", "remote_as"},
			nil,
		),
		prefixesSentDesc: prometheus.NewDesc(
			"mikrotik_bgp_prefixes_sent_total",
			"Number of prefixes sent to BGP peer",
			[]string{"target", "peer_address", "peer_as", "remote_as"},
			nil,
		),
	}
}

// Name returns the collector name
func (c *Collector) Name() string {
	return "bgp"
}

// Describe sends the descriptors of each metric over to the provided channel
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.peerStatusDesc
	ch <- c.peerUptimeDesc
	ch <- c.prefixesRecvDesc
	ch <- c.prefixesSentDesc
}

// Collect fetches the metrics from Mikrotik device and sends them to Prometheus
func (c *Collector) Collect(ctx context.Context, target string, auth collector.AuthInfo, ch chan<- prometheus.Metric) error {
	// TODO: Implement actual API call to Mikrotik device
	// This is a placeholder implementation
	
	return nil
}
