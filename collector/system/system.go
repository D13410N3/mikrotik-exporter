package system

import (
	"context"

	"github.com/mikrotik-exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
)

// Collector implements the collector.Collector interface for system metrics
type Collector struct {
	uptimeDesc        *prometheus.Desc
	cpuLoadDesc       *prometheus.Desc
	memoryUsedDesc    *prometheus.Desc
	memoryTotalDesc   *prometheus.Desc
	diskUsedDesc      *prometheus.Desc
	diskTotalDesc     *prometheus.Desc
	temperatureDesc   *prometheus.Desc
}

// NewCollector creates a new system collector
func NewCollector() *Collector {
	return &Collector{
		uptimeDesc: prometheus.NewDesc(
			"mikrotik_system_uptime_seconds",
			"System uptime in seconds",
			[]string{"target", "identity", "version"},
			nil,
		),
		cpuLoadDesc: prometheus.NewDesc(
			"mikrotik_system_cpu_load_percent",
			"CPU load percentage",
			[]string{"target", "identity"},
			nil,
		),
		memoryUsedDesc: prometheus.NewDesc(
			"mikrotik_system_memory_used_bytes",
			"Used memory in bytes",
			[]string{"target", "identity"},
			nil,
		),
		memoryTotalDesc: prometheus.NewDesc(
			"mikrotik_system_memory_total_bytes",
			"Total memory in bytes",
			[]string{"target", "identity"},
			nil,
		),
		diskUsedDesc: prometheus.NewDesc(
			"mikrotik_system_disk_used_bytes",
			"Used disk space in bytes",
			[]string{"target", "identity"},
			nil,
		),
		diskTotalDesc: prometheus.NewDesc(
			"mikrotik_system_disk_total_bytes",
			"Total disk space in bytes",
			[]string{"target", "identity"},
			nil,
		),
		temperatureDesc: prometheus.NewDesc(
			"mikrotik_system_temperature_celsius",
			"System temperature in Celsius",
			[]string{"target", "identity"},
			nil,
		),
	}
}

// Name returns the collector name
func (c *Collector) Name() string {
	return "system"
}

// Describe sends the descriptors of each metric over to the provided channel
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.uptimeDesc
	ch <- c.cpuLoadDesc
	ch <- c.memoryUsedDesc
	ch <- c.memoryTotalDesc
	ch <- c.diskUsedDesc
	ch <- c.diskTotalDesc
	ch <- c.temperatureDesc
}

// Collect fetches the metrics from Mikrotik device and sends them to Prometheus
func (c *Collector) Collect(ctx context.Context, target string, auth collector.AuthInfo, ch chan<- prometheus.Metric) error {
	// TODO: Implement actual API call to Mikrotik device
	// This is a placeholder implementation
	
	return nil
}
