package dhcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mikrotik-exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
)

// Collector implements the collector.Collector interface for DHCP metrics
type Collector struct {
	boundDesc *prometheus.Desc
	namespace string
}

// DHCPLeaseData represents the structure returned by Mikrotik DHCP lease API
type DHCPLeaseData struct {
	ID               string `json:".id"`
	Address          string `json:"address"`
	ActiveAddress    string `json:"active-address"`
	ActiveMacAddress string `json:"active-mac-address"`
	ActiveServer     string `json:"active-server"`
	AddressLists     string `json:"address-lists"`
	Blocked          string `json:"blocked"`
	DHCPOption       string `json:"dhcp-option"`
	Disabled         string `json:"disabled"`
	Dynamic          string `json:"dynamic"`
	ExpiresAfter     string `json:"expires-after"`
	HostName         string `json:"host-name"`
	LastSeen         string `json:"last-seen"`
	MacAddress       string `json:"mac-address"`
	Radius           string `json:"radius"`
	Server           string `json:"server"`
	Status           string `json:"status"`
}

// NewCollector creates a new DHCP collector
func NewCollector() *Collector {
	c := &Collector{
		namespace: "mikrotik_exporter", // default namespace
	}
	c.initMetrics()
	return c
}

// initMetrics initializes the metric descriptors with the current namespace
func (c *Collector) initMetrics() {
	c.boundDesc = prometheus.NewDesc(
		c.namespace+"_dhcp_bound",
		"DHCP lease bound status (1 = bound, 0 = not bound)",
		[]string{"device_ip", "mac", "dhcp_server", "device_hostname"},
		nil,
	)
}

// Name returns the collector name
func (c *Collector) Name() string {
	return "dhcp"
}

// Describe sends the descriptors of each metric over to the provided channel
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.boundDesc
}

// SetNamespace sets the metrics namespace prefix
func (c *Collector) SetNamespace(namespace string) {
	c.namespace = namespace
	c.initMetrics()
}

// Collect fetches the metrics from Mikrotik device and sends them to Prometheus
func (c *Collector) Collect(ctx context.Context, target string, auth collector.AuthInfo, ch chan<- prometheus.Metric) error {
	// Fetch DHCP lease data from Mikrotik REST API
	leases, err := c.fetchDHCPLeases(ctx, target, auth)
	if err != nil {
		return fmt.Errorf("failed to fetch DHCP leases: %w", err)
	}

	// Process each DHCP lease
	for _, lease := range leases {
		// Use active fields if available, fallback to regular fields
		ip := lease.ActiveAddress
		if ip == "" {
			ip = lease.Address
		}

		mac := lease.ActiveMacAddress
		if mac == "" {
			mac = lease.MacAddress
		}

		dhcpServer := lease.ActiveServer
		if dhcpServer == "" {
			dhcpServer = lease.Server
		}

		hostname := lease.HostName
		if hostname == "" {
			hostname = "unknown"
		}

		// Skip entries without essential information
		if ip == "" || mac == "" || dhcpServer == "" {
			continue
		}

		// Create labels for this lease
		labels := []string{
			ip,
			mac,
			dhcpServer,
			hostname,
		}

		// DHCP bound status
		boundValue := 0.0
		if lease.Status == "bound" {
			boundValue = 1.0
		}
		ch <- prometheus.MustNewConstMetric(c.boundDesc, prometheus.GaugeValue, boundValue, labels...)
	}

	return nil
}

// fetchDHCPLeases fetches DHCP lease data from Mikrotik REST API
func (c *Collector) fetchDHCPLeases(ctx context.Context, target string, auth collector.AuthInfo) ([]DHCPLeaseData, error) {
	url := fmt.Sprintf("http://%s/rest/ip/dhcp-server/lease", target)

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

	var leases []DHCPLeaseData
	if err := json.NewDecoder(resp.Body).Decode(&leases); err != nil {
		return nil, err
	}

	return leases, nil
}
