package dhcp

import (
	"context"

	"github.com/mikrotik-exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
)

// Collector implements the collector.Collector interface for DHCP metrics
type Collector struct {
	leasesActiveDesc   *prometheus.Desc
	leasesTotalDesc    *prometheus.Desc
	serverStatusDesc   *prometheus.Desc
}

// NewCollector creates a new DHCP collector
func NewCollector() *Collector {
	return &Collector{
		leasesActiveDesc: prometheus.NewDesc(
			"mikrotik_dhcp_leases_active",
			"Number of active DHCP leases",
			[]string{"target", "server"},
			nil,
		),
		leasesTotalDesc: prometheus.NewDesc(
			"mikrotik_dhcp_leases_total",
			"Total number of DHCP leases",
			[]string{"target", "server"},
			nil,
		),
		serverStatusDesc: prometheus.NewDesc(
			"mikrotik_dhcp_server_running",
			"DHCP server running status (1 = running, 0 = not running)",
			[]string{"target", "server", "interface"},
			nil,
		),
	}
}

// Name returns the collector name
func (c *Collector) Name() string {
	return "dhcp"
}

// Describe sends the descriptors of each metric over to the provided channel
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.leasesActiveDesc
	ch <- c.leasesTotalDesc
	ch <- c.serverStatusDesc
}

// Collect fetches the metrics from Mikrotik device and sends them to Prometheus
func (c *Collector) Collect(ctx context.Context, target string, auth collector.AuthInfo, ch chan<- prometheus.Metric) error {
	// TODO: Implement actual API call to Mikrotik device
	// This is a placeholder implementation
	
	return nil
}
