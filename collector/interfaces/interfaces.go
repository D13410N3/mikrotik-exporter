package interfaces

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/mikrotik-exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
)

// Collector implements the collector.Collector interface for interface metrics
type Collector struct {
	// Interface status metrics
	enabledDesc *prometheus.Desc
	upDesc      *prometheus.Desc

	// RX metrics
	rxBytesDesc     *prometheus.Desc
	rxPacketsDesc   *prometheus.Desc
	fpRxBytesDesc   *prometheus.Desc
	fpRxPacketsDesc *prometheus.Desc

	// TX metrics
	txBytesDesc     *prometheus.Desc
	txPacketsDesc   *prometheus.Desc
	fpTxBytesDesc   *prometheus.Desc
	fpTxPacketsDesc *prometheus.Desc
	txQueueDropDesc *prometheus.Desc

	// Interface properties
	mtuDesc          *prometheus.Desc
	linkDownsDesc    *prometheus.Desc
	lastLinkUpDesc   *prometheus.Desc
	lastLinkDownDesc *prometheus.Desc

	namespace string
}

// InterfaceData represents the structure returned by Mikrotik REST API
type InterfaceData struct {
	ID               string `json:".id"`
	ActualMTU        string `json:"actual-mtu"`
	Comment          string `json:"comment"`
	DefaultName      string `json:"default-name"`
	Disabled         string `json:"disabled"`
	FpRxByte         string `json:"fp-rx-byte"`
	FpRxPacket       string `json:"fp-rx-packet"`
	FpTxByte         string `json:"fp-tx-byte"`
	FpTxPacket       string `json:"fp-tx-packet"`
	L2MTU            string `json:"l2mtu"`
	LastLinkDownTime string `json:"last-link-down-time"`
	LastLinkUpTime   string `json:"last-link-up-time"`
	LinkDowns        string `json:"link-downs"`
	MacAddress       string `json:"mac-address"`
	MaxL2MTU         string `json:"max-l2mtu"`
	MTU              string `json:"mtu"`
	Name             string `json:"name"`
	Running          string `json:"running"`
	RxByte           string `json:"rx-byte"`
	RxPacket         string `json:"rx-packet"`
	TxByte           string `json:"tx-byte"`
	TxPacket         string `json:"tx-packet"`
	TxQueueDrop      string `json:"tx-queue-drop"`
	Type             string `json:"type"`
}

// NewCollector creates a new interfaces collector
func NewCollector() *Collector {
	c := &Collector{
		namespace: "mikrotik_exporter", // default namespace
	}
	c.initMetrics()
	return c
}

// initMetrics initializes the metric descriptors with the current namespace
func (c *Collector) initMetrics() {
	allLabels := []string{"mac", "name", "type", "comment"}
	basicLabels := []string{"name", "type"}

	// Interface status metrics
	c.enabledDesc = prometheus.NewDesc(
		c.namespace+"_interface_enabled",
		"Interface enabled status (1 = enabled, 0 = disabled)",
		basicLabels, nil,
	)
	c.upDesc = prometheus.NewDesc(
		c.namespace+"_interface_up",
		"Interface running status (1 = running, 0 = not running)",
		allLabels, nil,
	)

	// RX metrics
	c.rxBytesDesc = prometheus.NewDesc(
		c.namespace+"_interface_rx_bytes_total",
		"Number of bytes received on interface",
		basicLabels, nil,
	)
	c.rxPacketsDesc = prometheus.NewDesc(
		c.namespace+"_interface_rx_packets_total",
		"Number of packets received on interface",
		basicLabels, nil,
	)
	c.fpRxBytesDesc = prometheus.NewDesc(
		c.namespace+"_interface_fp_rx_bytes_total",
		"Number of fast path bytes received on interface",
		basicLabels, nil,
	)
	c.fpRxPacketsDesc = prometheus.NewDesc(
		c.namespace+"_interface_fp_rx_packets_total",
		"Number of fast path packets received on interface",
		basicLabels, nil,
	)

	// TX metrics
	c.txBytesDesc = prometheus.NewDesc(
		c.namespace+"_interface_tx_bytes_total",
		"Number of bytes transmitted on interface",
		basicLabels, nil,
	)
	c.txPacketsDesc = prometheus.NewDesc(
		c.namespace+"_interface_tx_packets_total",
		"Number of packets transmitted on interface",
		basicLabels, nil,
	)
	c.fpTxBytesDesc = prometheus.NewDesc(
		c.namespace+"_interface_fp_tx_bytes_total",
		"Number of fast path bytes transmitted on interface",
		basicLabels, nil,
	)
	c.fpTxPacketsDesc = prometheus.NewDesc(
		c.namespace+"_interface_fp_tx_packets_total",
		"Number of fast path packets transmitted on interface",
		basicLabels, nil,
	)
	c.txQueueDropDesc = prometheus.NewDesc(
		c.namespace+"_interface_tx_queue_drop_total",
		"Number of packets dropped from TX queue",
		basicLabels, nil,
	)

	// Interface properties
	c.mtuDesc = prometheus.NewDesc(
		c.namespace+"_interface_mtu",
		"Interface MTU in bytes",
		basicLabels, nil,
	)
	c.linkDownsDesc = prometheus.NewDesc(
		c.namespace+"_interface_link_downs_total",
		"Number of link down events",
		basicLabels, nil,
	)
	c.lastLinkUpDesc = prometheus.NewDesc(
		c.namespace+"_interface_last_link_up_time",
		"Last link up time (Unix timestamp)",
		basicLabels, nil,
	)
	c.lastLinkDownDesc = prometheus.NewDesc(
		c.namespace+"_interface_last_link_down_time",
		"Last link down time (Unix timestamp)",
		basicLabels, nil,
	)
}

// Name returns the collector name
func (c *Collector) Name() string {
	return "interfaces"
}

// Describe sends the descriptors of each metric over to the provided channel
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.enabledDesc
	ch <- c.upDesc
	ch <- c.rxBytesDesc
	ch <- c.rxPacketsDesc
	ch <- c.fpRxBytesDesc
	ch <- c.fpRxPacketsDesc
	ch <- c.txBytesDesc
	ch <- c.txPacketsDesc
	ch <- c.fpTxBytesDesc
	ch <- c.fpTxPacketsDesc
	ch <- c.txQueueDropDesc
	ch <- c.mtuDesc
	ch <- c.linkDownsDesc
	ch <- c.lastLinkUpDesc
	ch <- c.lastLinkDownDesc
}

// SetNamespace sets the metrics namespace prefix
func (c *Collector) SetNamespace(namespace string) {
	c.namespace = namespace
	c.initMetrics()
}

// Collect fetches the metrics from Mikrotik device and sends them to Prometheus
func (c *Collector) Collect(ctx context.Context, target string, auth collector.AuthInfo, ch chan<- prometheus.Metric) error {
	// Fetch interface data from Mikrotik REST API
	interfaces, err := c.fetchInterfaces(ctx, target, auth)
	if err != nil {
		return fmt.Errorf("failed to fetch interfaces: %w", err)
	}

	// Process each interface
	for _, iface := range interfaces {
		comment := iface.Comment
		if comment == "" {
			comment = ""
		}
		allLabels := []string{iface.MacAddress, iface.Name, iface.Type, comment}
		basicLabels := []string{iface.Name, iface.Type}

		// Interface status metrics
		enabledValue := 0.0
		if iface.Disabled != "true" {
			enabledValue = 1.0
		}
		ch <- prometheus.MustNewConstMetric(c.enabledDesc, prometheus.GaugeValue, enabledValue, basicLabels...)

		upValue := 0.0
		if iface.Running == "true" {
			upValue = 1.0
		}
		ch <- prometheus.MustNewConstMetric(c.upDesc, prometheus.GaugeValue, upValue, allLabels...)

		// RX metrics
		if rxBytes, err := parseUint64(iface.RxByte); err == nil {
			ch <- prometheus.MustNewConstMetric(c.rxBytesDesc, prometheus.CounterValue, float64(rxBytes), basicLabels...)
		}
		if rxPackets, err := parseUint64(iface.RxPacket); err == nil {
			ch <- prometheus.MustNewConstMetric(c.rxPacketsDesc, prometheus.CounterValue, float64(rxPackets), basicLabels...)
		}
		if fpRxBytes, err := parseUint64(iface.FpRxByte); err == nil {
			ch <- prometheus.MustNewConstMetric(c.fpRxBytesDesc, prometheus.CounterValue, float64(fpRxBytes), basicLabels...)
		}
		if fpRxPackets, err := parseUint64(iface.FpRxPacket); err == nil {
			ch <- prometheus.MustNewConstMetric(c.fpRxPacketsDesc, prometheus.CounterValue, float64(fpRxPackets), basicLabels...)
		}

		// TX metrics
		if txBytes, err := parseUint64(iface.TxByte); err == nil {
			ch <- prometheus.MustNewConstMetric(c.txBytesDesc, prometheus.CounterValue, float64(txBytes), basicLabels...)
		}
		if txPackets, err := parseUint64(iface.TxPacket); err == nil {
			ch <- prometheus.MustNewConstMetric(c.txPacketsDesc, prometheus.CounterValue, float64(txPackets), basicLabels...)
		}
		if fpTxBytes, err := parseUint64(iface.FpTxByte); err == nil {
			ch <- prometheus.MustNewConstMetric(c.fpTxBytesDesc, prometheus.CounterValue, float64(fpTxBytes), basicLabels...)
		}
		if fpTxPackets, err := parseUint64(iface.FpTxPacket); err == nil {
			ch <- prometheus.MustNewConstMetric(c.fpTxPacketsDesc, prometheus.CounterValue, float64(fpTxPackets), basicLabels...)
		}
		if txQueueDrop, err := parseUint64(iface.TxQueueDrop); err == nil {
			ch <- prometheus.MustNewConstMetric(c.txQueueDropDesc, prometheus.CounterValue, float64(txQueueDrop), basicLabels...)
		}

		// Interface properties
		if mtu, err := parseUint64(iface.MTU); err == nil {
			ch <- prometheus.MustNewConstMetric(c.mtuDesc, prometheus.GaugeValue, float64(mtu), basicLabels...)
		}
		if linkDowns, err := parseUint64(iface.LinkDowns); err == nil {
			ch <- prometheus.MustNewConstMetric(c.linkDownsDesc, prometheus.CounterValue, float64(linkDowns), basicLabels...)
		}

		// Last link up/down times
		if lastLinkUp := parseTimestamp(iface.LastLinkUpTime); lastLinkUp > 0 {
			ch <- prometheus.MustNewConstMetric(c.lastLinkUpDesc, prometheus.GaugeValue, float64(lastLinkUp), basicLabels...)
		}
		if lastLinkDown := parseTimestamp(iface.LastLinkDownTime); lastLinkDown > 0 {
			ch <- prometheus.MustNewConstMetric(c.lastLinkDownDesc, prometheus.GaugeValue, float64(lastLinkDown), basicLabels...)
		}
	}

	return nil
}

// fetchInterfaces fetches interface data from Mikrotik REST API
func (c *Collector) fetchInterfaces(ctx context.Context, target string, auth collector.AuthInfo) ([]InterfaceData, error) {
	url := fmt.Sprintf("http://%s/rest/interface", target)

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

	var interfaces []InterfaceData
	if err := json.NewDecoder(resp.Body).Decode(&interfaces); err != nil {
		return nil, err
	}

	return interfaces, nil
}

// parseUint64 safely parses a string to uint64
func parseUint64(s string) (uint64, error) {
	if s == "" {
		return 0, fmt.Errorf("empty string")
	}
	return strconv.ParseUint(s, 10, 64)
}

// parseTimestamp converts Mikrotik timestamp format to Unix timestamp
func parseTimestamp(timeStr string) int64 {
	if timeStr == "" {
		return 0
	}

	// Parse Mikrotik timestamp format: "2025-09-21 01:08:49"
	t, err := time.Parse("2006-01-02 15:04:05", timeStr)
	if err != nil {
		return 0
	}

	return t.Unix()
}
