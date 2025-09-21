package firewall

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

// Collector implements the collector.Collector interface for firewall metrics
type Collector struct {
	ruleEnabledDesc *prometheus.Desc
	ruleBytesDesc   *prometheus.Desc
	rulePacketsDesc *prometheus.Desc
	ruleInfoDesc    *prometheus.Desc
	namespace       string
}

// FirewallRuleData represents the structure returned by Mikrotik firewall API
type FirewallRuleData struct {
	ID                 string `json:".id"`
	Action             string `json:"action"`
	Bytes              string `json:"bytes"`
	Chain              string `json:"chain"`
	Comment            string `json:"comment"`
	ConnectionNatState string `json:"connection-nat-state"`
	ConnectionState    string `json:"connection-state"`
	Disabled           string `json:"disabled"`
	Dynamic            string `json:"dynamic"`
	InInterfaceList    string `json:"in-interface-list"`
	Invalid            string `json:"invalid"`
	Packets            string `json:"packets"`
}

// NewCollector creates a new firewall collector
func NewCollector() *Collector {
	c := &Collector{
		namespace: "mikrotik_exporter", // default namespace
	}
	c.initMetrics()
	return c
}

// initMetrics initializes the metric descriptors with the current namespace
func (c *Collector) initMetrics() {
	c.ruleEnabledDesc = prometheus.NewDesc(
		c.namespace+"_firewall_rule_enabled",
		"Firewall rule enabled status (1 = enabled, 0 = disabled)",
		[]string{"id", "table"},
		nil,
	)
	c.ruleBytesDesc = prometheus.NewDesc(
		c.namespace+"_firewall_rule_bytes",
		"Number of bytes matched by firewall rule",
		[]string{"id", "table"},
		nil,
	)
	c.rulePacketsDesc = prometheus.NewDesc(
		c.namespace+"_firewall_rule_packets",
		"Number of packets matched by firewall rule",
		[]string{"id", "table"},
		nil,
	)
	c.ruleInfoDesc = prometheus.NewDesc(
		c.namespace+"_firewall_rule_info",
		"Firewall rule information",
		[]string{"id", "table", "chain", "action", "comment"},
		nil,
	)
}

// Name returns the collector name
func (c *Collector) Name() string {
	return "firewall"
}

// Describe sends the descriptors of each metric over to the provided channel
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.ruleEnabledDesc
	ch <- c.ruleBytesDesc
	ch <- c.rulePacketsDesc
	ch <- c.ruleInfoDesc
}

// SetNamespace sets the metrics namespace prefix
func (c *Collector) SetNamespace(namespace string) {
	c.namespace = namespace
	c.initMetrics()
}

// Collect fetches the metrics from Mikrotik device and sends them to Prometheus
func (c *Collector) Collect(ctx context.Context, target string, auth collector.AuthInfo, ch chan<- prometheus.Metric) error {
	// List of firewall tables to query
	tables := []string{"filter", "nat", "mangle", "raw"}

	for _, table := range tables {
		rules, err := c.fetchFirewallRules(ctx, target, auth, table)
		if err != nil {
			return fmt.Errorf("failed to fetch %s rules: %w", table, err)
		}

		// Process each firewall rule
		for _, rule := range rules {
			// Skip rules without ID
			if rule.ID == "" {
				continue
			}

			labels := []string{rule.ID, table}

			// Firewall rule enabled status
			enabled := 1.0
			if rule.Disabled == "true" {
				enabled = 0.0
			}
			ch <- prometheus.MustNewConstMetric(c.ruleEnabledDesc, prometheus.GaugeValue, enabled, labels...)

			// Rule bytes
			if bytes, err := c.parseNumericField(rule.Bytes); err == nil {
				ch <- prometheus.MustNewConstMetric(c.ruleBytesDesc, prometheus.CounterValue, bytes, labels...)
			}

			// Rule packets
			if packets, err := c.parseNumericField(rule.Packets); err == nil {
				ch <- prometheus.MustNewConstMetric(c.rulePacketsDesc, prometheus.CounterValue, packets, labels...)
			}

			// Rule info
			comment := rule.Comment
			if comment == "" {
				comment = ""
			}
			infoLabels := []string{
				rule.ID,
				table, // table name (filter/nat/mangle/raw)
				rule.Chain,
				rule.Action,
				comment,
			}
			ch <- prometheus.MustNewConstMetric(c.ruleInfoDesc, prometheus.GaugeValue, 1, infoLabels...)
		}
	}

	return nil
}

// fetchFirewallRules fetches firewall rule data from Mikrotik REST API
func (c *Collector) fetchFirewallRules(ctx context.Context, target string, auth collector.AuthInfo, table string) ([]FirewallRuleData, error) {
	url := fmt.Sprintf("http://%s/rest/ip/firewall/%s", target, table)

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

	var rules []FirewallRuleData
	if err := json.NewDecoder(resp.Body).Decode(&rules); err != nil {
		return nil, err
	}

	return rules, nil
}

// parseNumericField safely parses a string field to float64
func (c *Collector) parseNumericField(value string) (float64, error) {
	if value == "" {
		return 0, fmt.Errorf("empty value")
	}
	return strconv.ParseFloat(value, 64)
}
