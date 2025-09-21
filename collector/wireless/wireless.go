package wireless

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mikrotik-exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
)

// Collector implements the collector.Collector interface for wireless metrics
type Collector struct {
	clientInfoDesc *prometheus.Desc
	txBytesDesc    *prometheus.Desc
	txPacketsDesc  *prometheus.Desc
	rxBytesDesc    *prometheus.Desc
	rxPacketsDesc  *prometheus.Desc
	rxRateDesc     *prometheus.Desc
	txRateDesc     *prometheus.Desc
	uptimeDesc     *prometheus.Desc
	signalDesc     *prometheus.Desc
	namespace      string
}

// WirelessRegistrationData represents the structure returned by Mikrotik WiFi registration table API
type WirelessRegistrationData struct {
	ID           string `json:".id"`
	Authorized   string `json:"authorized"`
	Bytes        string `json:"bytes"`
	Interface    string `json:"interface"`
	MacAddress   string `json:"mac-address"`
	Packets      string `json:"packets"`
	RxBitsPerSec string `json:"rx-bits-per-second"`
	RxRate       string `json:"rx-rate"`
	Signal       string `json:"signal"`
	SSID         string `json:"ssid"`
	TxBitsPerSec string `json:"tx-bits-per-second"`
	TxRate       string `json:"tx-rate"`
	Uptime       string `json:"uptime"`
}

// NewCollector creates a new wireless collector
func NewCollector() *Collector {
	c := &Collector{
		namespace: "mikrotik_exporter", // default namespace
	}
	c.initMetrics()
	return c
}

// initMetrics initializes the metric descriptors with the current namespace
func (c *Collector) initMetrics() {
	clientInfoLabels := []string{"mac", "interface", "ssid"}
	macLabel := []string{"mac"}

	c.clientInfoDesc = prometheus.NewDesc(
		c.namespace+"_wireless_client_info",
		"Wireless client information (always 1 for connected clients)",
		clientInfoLabels, nil,
	)
	c.txBytesDesc = prometheus.NewDesc(
		c.namespace+"_wireless_tx_bytes_total",
		"Number of bytes transmitted by wireless client",
		macLabel, nil,
	)
	c.txPacketsDesc = prometheus.NewDesc(
		c.namespace+"_wireless_tx_packets_total",
		"Number of packets transmitted by wireless client",
		macLabel, nil,
	)
	c.rxBytesDesc = prometheus.NewDesc(
		c.namespace+"_wireless_rx_bytes_total",
		"Number of bytes received by wireless client",
		macLabel, nil,
	)
	c.rxPacketsDesc = prometheus.NewDesc(
		c.namespace+"_wireless_rx_packets_total",
		"Number of packets received by wireless client",
		macLabel, nil,
	)
	c.rxRateDesc = prometheus.NewDesc(
		c.namespace+"_wireless_rx_rate",
		"Wireless RX rate in bits per second",
		macLabel, nil,
	)
	c.txRateDesc = prometheus.NewDesc(
		c.namespace+"_wireless_tx_rate",
		"Wireless TX rate in bits per second",
		macLabel, nil,
	)
	c.uptimeDesc = prometheus.NewDesc(
		c.namespace+"_wireless_uptime",
		"Wireless client uptime in seconds",
		macLabel, nil,
	)
	c.signalDesc = prometheus.NewDesc(
		c.namespace+"_wireless_signal",
		"Wireless client signal strength in dBm",
		macLabel, nil,
	)
}

// Name returns the collector name
func (c *Collector) Name() string {
	return "wireless"
}

// Describe sends the descriptors of each metric over to the provided channel
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.clientInfoDesc
	ch <- c.txBytesDesc
	ch <- c.txPacketsDesc
	ch <- c.rxBytesDesc
	ch <- c.rxPacketsDesc
	ch <- c.rxRateDesc
	ch <- c.txRateDesc
	ch <- c.uptimeDesc
	ch <- c.signalDesc
}

// SetNamespace sets the metrics namespace prefix
func (c *Collector) SetNamespace(namespace string) {
	c.namespace = namespace
	c.initMetrics()
}

// Collect fetches the metrics from Mikrotik device and sends them to Prometheus
func (c *Collector) Collect(ctx context.Context, target string, auth collector.AuthInfo, ch chan<- prometheus.Metric) error {
	// Fetch wireless registration data from Mikrotik REST API
	registrations, err := c.fetchWirelessRegistrations(ctx, target, auth)
	if err != nil {
		return fmt.Errorf("failed to fetch wireless registrations: %w", err)
	}

	// Process each wireless client
	for _, reg := range registrations {
		clientInfoLabels := []string{reg.MacAddress, reg.Interface, reg.SSID}
		macLabels := []string{reg.MacAddress}

		// Client info (always 1 for entries in registration table)
		ch <- prometheus.MustNewConstMetric(c.clientInfoDesc, prometheus.GaugeValue, 1.0, clientInfoLabels...)

		// Parse bytes (format: "tx_bytes,rx_bytes")
		if txBytes, rxBytes, err := parseCommaSeparatedPair(reg.Bytes); err == nil {
			ch <- prometheus.MustNewConstMetric(c.txBytesDesc, prometheus.CounterValue, float64(txBytes), macLabels...)
			ch <- prometheus.MustNewConstMetric(c.rxBytesDesc, prometheus.CounterValue, float64(rxBytes), macLabels...)
		}

		// Parse packets (format: "tx_packets,rx_packets")
		if txPackets, rxPackets, err := parseCommaSeparatedPair(reg.Packets); err == nil {
			ch <- prometheus.MustNewConstMetric(c.txPacketsDesc, prometheus.CounterValue, float64(txPackets), macLabels...)
			ch <- prometheus.MustNewConstMetric(c.rxPacketsDesc, prometheus.CounterValue, float64(rxPackets), macLabels...)
		}

		// RX/TX rates
		if rxRate, err := parseUint64(reg.RxRate); err == nil {
			ch <- prometheus.MustNewConstMetric(c.rxRateDesc, prometheus.GaugeValue, float64(rxRate), macLabels...)
		}
		if txRate, err := parseUint64(reg.TxRate); err == nil {
			ch <- prometheus.MustNewConstMetric(c.txRateDesc, prometheus.GaugeValue, float64(txRate), macLabels...)
		}

		// Uptime
		if uptime := parseUptime(reg.Uptime); uptime > 0 {
			ch <- prometheus.MustNewConstMetric(c.uptimeDesc, prometheus.GaugeValue, float64(uptime), macLabels...)
		}

		// Signal strength
		if signal, err := strconv.ParseFloat(reg.Signal, 64); err == nil {
			ch <- prometheus.MustNewConstMetric(c.signalDesc, prometheus.GaugeValue, signal, macLabels...)
		}
	}

	return nil
}

// fetchWirelessRegistrations fetches wireless registration data from Mikrotik REST API
func (c *Collector) fetchWirelessRegistrations(ctx context.Context, target string, auth collector.AuthInfo) ([]WirelessRegistrationData, error) {
	url := fmt.Sprintf("http://%s/rest/interface/wifi/registration-table", target)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(auth.Username, auth.Password)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	var registrations []WirelessRegistrationData
	if err := json.NewDecoder(resp.Body).Decode(&registrations); err != nil {
		return nil, err
	}

	return registrations, nil
}

// parseCommaSeparatedPair parses "value1,value2" format and returns both values
func parseCommaSeparatedPair(s string) (uint64, uint64, error) {
	if s == "" {
		return 0, 0, fmt.Errorf("empty string")
	}

	parts := strings.Split(s, ",")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("expected format 'value1,value2', got: %s", s)
	}

	val1, err := strconv.ParseUint(strings.TrimSpace(parts[0]), 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse first value: %w", err)
	}

	val2, err := strconv.ParseUint(strings.TrimSpace(parts[1]), 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse second value: %w", err)
	}

	return val1, val2, nil
}

// parseUint64 safely parses a string to uint64
func parseUint64(s string) (uint64, error) {
	if s == "" {
		return 0, fmt.Errorf("empty string")
	}
	return strconv.ParseUint(s, 10, 64)
}

// parseUptime converts Mikrotik uptime format to seconds
// Format examples: "2w4d1h12m27s", "1h30m", "45s"
func parseUptime(uptimeStr string) int64 {
	if uptimeStr == "" {
		return 0
	}

	// Regular expression to match Mikrotik uptime format
	re := regexp.MustCompile(`(?:(\d+)w)?(?:(\d+)d)?(?:(\d+)h)?(?:(\d+)m)?(?:(\d+)s)?`)
	matches := re.FindStringSubmatch(uptimeStr)

	if len(matches) == 0 {
		return 0
	}

	var totalSeconds int64

	// Parse weeks
	if matches[1] != "" {
		if weeks, err := strconv.ParseInt(matches[1], 10, 64); err == nil {
			totalSeconds += weeks * 7 * 24 * 3600
		}
	}

	// Parse days
	if matches[2] != "" {
		if days, err := strconv.ParseInt(matches[2], 10, 64); err == nil {
			totalSeconds += days * 24 * 3600
		}
	}

	// Parse hours
	if matches[3] != "" {
		if hours, err := strconv.ParseInt(matches[3], 10, 64); err == nil {
			totalSeconds += hours * 3600
		}
	}

	// Parse minutes
	if matches[4] != "" {
		if minutes, err := strconv.ParseInt(matches[4], 10, 64); err == nil {
			totalSeconds += minutes * 60
		}
	}

	// Parse seconds
	if matches[5] != "" {
		if seconds, err := strconv.ParseInt(matches[5], 10, 64); err == nil {
			totalSeconds += seconds
		}
	}

	return totalSeconds
}
