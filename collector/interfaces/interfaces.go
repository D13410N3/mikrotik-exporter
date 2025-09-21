package interfaces

import (
	"context"

	"github.com/mikrotik-exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
)

// Collector implements the collector.Collector interface for interface metrics
type Collector struct {
	rxBytesDesc   *prometheus.Desc
	txBytesDesc   *prometheus.Desc
	rxPacketsDesc *prometheus.Desc
	txPacketsDesc *prometheus.Desc
	statusDesc    *prometheus.Desc
}

// NewCollector creates a new interfaces collector
func NewCollector() *Collector {
	return &Collector{
		rxBytesDesc: prometheus.NewDesc(
			"mikrotik_interface_rx_bytes_total",
			"Number of bytes received on interface",
			[]string{"target", "interface", "name"},
			nil,
		),
		txBytesDesc: prometheus.NewDesc(
			"mikrotik_interface_tx_bytes_total",
			"Number of bytes transmitted on interface",
			[]string{"target", "interface", "name"},
			nil,
		),
		rxPacketsDesc: prometheus.NewDesc(
			"mikrotik_interface_rx_packets_total",
			"Number of packets received on interface",
			[]string{"target", "interface", "name"},
			nil,
		),
		txPacketsDesc: prometheus.NewDesc(
			"mikrotik_interface_tx_packets_total",
			"Number of packets transmitted on interface",
			[]string{"target", "interface", "name"},
			nil,
		),
		statusDesc: prometheus.NewDesc(
			"mikrotik_interface_running",
			"Interface running status (1 = running, 0 = not running)",
			[]string{"target", "interface", "name"},
			nil,
		),
	}
}

// Name returns the collector name
func (c *Collector) Name() string {
	return "interfaces"
}

// Describe sends the descriptors of each metric over to the provided channel
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.rxBytesDesc
	ch <- c.txBytesDesc
	ch <- c.rxPacketsDesc
	ch <- c.txPacketsDesc
	ch <- c.statusDesc
}

// Collect fetches the metrics from Mikrotik device and sends them to Prometheus
func (c *Collector) Collect(ctx context.Context, target string, auth collector.AuthInfo, ch chan<- prometheus.Metric) error {
	// TODO: Implement actual API call to Mikrotik device
	// This is a placeholder implementation
	
	// Example of how metrics would be sent:
	// ch <- prometheus.MustNewConstMetric(
	//     c.rxBytesDesc,
	//     prometheus.CounterValue,
	//     float64(rxBytes),
	//     target, interfaceID, interfaceName,
	// )
	
	return nil
}
