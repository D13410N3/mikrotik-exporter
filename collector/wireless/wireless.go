package wireless

import (
	"context"

	"github.com/mikrotik-exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
)

// Collector implements the collector.Collector interface for wireless metrics
type Collector struct {
	clientsConnectedDesc *prometheus.Desc
	signalStrengthDesc   *prometheus.Desc
	txRateDesc           *prometheus.Desc
	rxRateDesc           *prometheus.Desc
	interfaceStatusDesc  *prometheus.Desc
}

// NewCollector creates a new wireless collector
func NewCollector() *Collector {
	return &Collector{
		clientsConnectedDesc: prometheus.NewDesc(
			"mikrotik_wireless_clients_connected",
			"Number of connected wireless clients",
			[]string{"target", "interface", "ssid"},
			nil,
		),
		signalStrengthDesc: prometheus.NewDesc(
			"mikrotik_wireless_signal_strength_dbm",
			"Wireless signal strength in dBm",
			[]string{"target", "interface", "client_mac"},
			nil,
		),
		txRateDesc: prometheus.NewDesc(
			"mikrotik_wireless_tx_rate_mbps",
			"Wireless TX rate in Mbps",
			[]string{"target", "interface", "client_mac"},
			nil,
		),
		rxRateDesc: prometheus.NewDesc(
			"mikrotik_wireless_rx_rate_mbps",
			"Wireless RX rate in Mbps",
			[]string{"target", "interface", "client_mac"},
			nil,
		),
		interfaceStatusDesc: prometheus.NewDesc(
			"mikrotik_wireless_interface_running",
			"Wireless interface running status (1 = running, 0 = not running)",
			[]string{"target", "interface", "ssid"},
			nil,
		),
	}
}

// Name returns the collector name
func (c *Collector) Name() string {
	return "wireless"
}

// Describe sends the descriptors of each metric over to the provided channel
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.clientsConnectedDesc
	ch <- c.signalStrengthDesc
	ch <- c.txRateDesc
	ch <- c.rxRateDesc
	ch <- c.interfaceStatusDesc
}

// Collect fetches the metrics from Mikrotik device and sends them to Prometheus
func (c *Collector) Collect(ctx context.Context, target string, auth collector.AuthInfo, ch chan<- prometheus.Metric) error {
	// TODO: Implement actual API call to Mikrotik device
	// This is a placeholder implementation
	
	return nil
}
