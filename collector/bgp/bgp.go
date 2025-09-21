package bgp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/mikrotik-exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
)

// Collector implements the collector.Collector interface for BGP metrics
type Collector struct {
	sessionUpDesc           *prometheus.Desc
	prefixCountDesc         *prometheus.Desc
	remoteBytesTotalDesc    *prometheus.Desc
	remoteMessagesTotalDesc *prometheus.Desc
	localBytesDesc          *prometheus.Desc
	localMessagesTotalDesc  *prometheus.Desc
	uptimeDesc              *prometheus.Desc
	sessionInfoDesc         *prometheus.Desc
	namespace               string
}

// BGPSessionData represents the structure returned by Mikrotik BGP session API
type BGPSessionData struct {
	ID                 string `json:".id"`
	EBGP               string `json:"ebgp"`
	Established        string `json:"established"`
	HoldTime           string `json:"hold-time"`
	InputProcID        string `json:"input.procid"`
	KeepaliveTime      string `json:"keepalive-time"`
	LastStarted        string `json:"last-started"`
	LastStopped        string `json:"last-stopped"`
	LocalAddress       string `json:"local.address"`
	LocalAFI           string `json:"local.afi"`
	LocalAS            string `json:"local.as"`
	LocalBytes         string `json:"local.bytes"`
	LocalCapabilities  string `json:"local.capabilities"`
	LocalClusterID     string `json:"local.cluster-id"`
	LocalEOR           string `json:"local.eor"`
	LocalID            string `json:"local.id"`
	LocalMessages      string `json:"local.messages"`
	LocalRole          string `json:"local.role"`
	Multihop           string `json:"multihop"`
	Name               string `json:"name"`
	OutputProcID       string `json:"output.procid"`
	PrefixCount        string `json:"prefix-count"`
	RemoteAddress      string `json:"remote.address"`
	RemoteAFI          string `json:"remote.afi"`
	RemoteAS           string `json:"remote.as"`
	RemoteBytes        string `json:"remote.bytes"`
	RemoteCapabilities string `json:"remote.capabilities"`
	RemoteEOR          string `json:"remote.eor"`
	RemoteGRAFI        string `json:"remote.gr-afi"`
	RemoteGRAFIFWP     string `json:"remote.gr-afi-fwp"`
	RemoteGRTime       string `json:"remote.gr-time"`
	RemoteHoldTime     string `json:"remote.hold-time"`
	RemoteID           string `json:"remote.id"`
	RemoteMessages     string `json:"remote.messages"`
	Uptime             string `json:"uptime"`
}

// NewCollector creates a new BGP collector
func NewCollector() *Collector {
	c := &Collector{
		namespace: "mikrotik_exporter", // default namespace
	}
	c.initMetrics()
	return c
}

// initMetrics initializes the metric descriptors with the current namespace
func (c *Collector) initMetrics() {
	c.sessionUpDesc = prometheus.NewDesc(
		c.namespace+"_bgp_session_up",
		"BGP session status (1 = established, 0 = not established)",
		[]string{"name"},
		nil,
	)
	c.prefixCountDesc = prometheus.NewDesc(
		c.namespace+"_bgp_session_prefix_count",
		"Number of prefixes in BGP session",
		[]string{"name"},
		nil,
	)
	c.remoteBytesTotalDesc = prometheus.NewDesc(
		c.namespace+"_bgp_session_remote_bytes_total",
		"Total bytes received from remote BGP peer",
		[]string{"name"},
		nil,
	)
	c.remoteMessagesTotalDesc = prometheus.NewDesc(
		c.namespace+"_bgp_session_remote_messages_total",
		"Total messages received from remote BGP peer",
		[]string{"name"},
		nil,
	)
	c.localBytesDesc = prometheus.NewDesc(
		c.namespace+"_bgp_session_local_bytes_total",
		"Total bytes sent to remote BGP peer",
		[]string{"name"},
		nil,
	)
	c.localMessagesTotalDesc = prometheus.NewDesc(
		c.namespace+"_bgp_session_local_messages_total",
		"Total messages sent to remote BGP peer",
		[]string{"name"},
		nil,
	)
	c.uptimeDesc = prometheus.NewDesc(
		c.namespace+"_bgp_session_uptime",
		"BGP session uptime in seconds",
		[]string{"name"},
		nil,
	)
	c.sessionInfoDesc = prometheus.NewDesc(
		c.namespace+"_bgp_session_info",
		"BGP session information",
		[]string{"name", "remote_address", "remote_id", "remote_as", "local_address", "local_id", "local_as"},
		nil,
	)
}

// Name returns the collector name
func (c *Collector) Name() string {
	return "bgp"
}

// Describe sends the descriptors of each metric over to the provided channel
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.sessionUpDesc
	ch <- c.prefixCountDesc
	ch <- c.remoteBytesTotalDesc
	ch <- c.remoteMessagesTotalDesc
	ch <- c.localBytesDesc
	ch <- c.localMessagesTotalDesc
	ch <- c.uptimeDesc
	ch <- c.sessionInfoDesc
}

// SetNamespace sets the metrics namespace prefix
func (c *Collector) SetNamespace(namespace string) {
	c.namespace = namespace
	c.initMetrics()
}

// Collect fetches the metrics from Mikrotik device and sends them to Prometheus
func (c *Collector) Collect(ctx context.Context, target string, auth collector.AuthInfo, ch chan<- prometheus.Metric) error {
	// Fetch BGP session data from Mikrotik REST API
	sessions, err := c.fetchBGPSessions(ctx, target, auth)
	if err != nil {
		return fmt.Errorf("failed to fetch BGP sessions: %w", err)
	}

	// Process each BGP session
	for _, session := range sessions {
		// Skip sessions without name
		if session.Name == "" {
			continue
		}

		labels := []string{session.Name}

		// BGP session up status
		sessionUp := 0.0
		if session.Established == "true" {
			sessionUp = 1.0
		}
		ch <- prometheus.MustNewConstMetric(c.sessionUpDesc, prometheus.GaugeValue, sessionUp, labels...)

		// Prefix count
		if prefixCount, err := c.parseNumericField(session.PrefixCount); err == nil {
			ch <- prometheus.MustNewConstMetric(c.prefixCountDesc, prometheus.GaugeValue, prefixCount, labels...)
		}

		// Remote bytes and messages
		if remoteBytes, err := c.parseNumericField(session.RemoteBytes); err == nil {
			ch <- prometheus.MustNewConstMetric(c.remoteBytesTotalDesc, prometheus.CounterValue, remoteBytes, labels...)
		}
		if remoteMessages, err := c.parseNumericField(session.RemoteMessages); err == nil {
			ch <- prometheus.MustNewConstMetric(c.remoteMessagesTotalDesc, prometheus.CounterValue, remoteMessages, labels...)
		}

		// Local bytes and messages
		if localBytes, err := c.parseNumericField(session.LocalBytes); err == nil {
			ch <- prometheus.MustNewConstMetric(c.localBytesDesc, prometheus.CounterValue, localBytes, labels...)
		}
		if localMessages, err := c.parseNumericField(session.LocalMessages); err == nil {
			ch <- prometheus.MustNewConstMetric(c.localMessagesTotalDesc, prometheus.CounterValue, localMessages, labels...)
		}

		// Uptime
		if uptime, err := c.parseUptime(session.Uptime); err == nil && uptime > 0 {
			ch <- prometheus.MustNewConstMetric(c.uptimeDesc, prometheus.GaugeValue, uptime, labels...)
		}

		// Session info
		infoLabels := []string{
			session.Name,
			session.RemoteAddress,
			session.RemoteID,
			session.RemoteAS,
			session.LocalAddress,
			session.LocalID,
			session.LocalAS,
		}
		ch <- prometheus.MustNewConstMetric(c.sessionInfoDesc, prometheus.GaugeValue, 1, infoLabels...)
	}

	return nil
}

// fetchBGPSessions fetches BGP session data from Mikrotik REST API
func (c *Collector) fetchBGPSessions(ctx context.Context, target string, auth collector.AuthInfo) ([]BGPSessionData, error) {
	url := fmt.Sprintf("http://%s/rest/routing/bgp/session", target)

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

	var sessions []BGPSessionData
	if err := json.NewDecoder(resp.Body).Decode(&sessions); err != nil {
		return nil, err
	}

	return sessions, nil
}

// parseNumericField safely parses a string field to float64
func (c *Collector) parseNumericField(value string) (float64, error) {
	if value == "" {
		return 0, fmt.Errorf("empty value")
	}
	return strconv.ParseFloat(value, 64)
}

// parseUptime parses Mikrotik uptime format to seconds
func (c *Collector) parseUptime(uptime string) (float64, error) {
	if uptime == "" {
		return 0, fmt.Errorf("empty uptime")
	}

	// Remove milliseconds part if present (e.g., "2w4d36m5s950ms" -> "2w4d36m5s")
	uptime = regexp.MustCompile(`\d+ms$`).ReplaceAllString(uptime, "")

	// Parse uptime format: 2w4d1h12m27s
	re := regexp.MustCompile(`(?:(\d+)w)?(?:(\d+)d)?(?:(\d+)h)?(?:(\d+)m)?(?:(\d+)s)?`)
	matches := re.FindStringSubmatch(uptime)

	if len(matches) == 0 {
		return 0, fmt.Errorf("invalid uptime format: %s", uptime)
	}

	var totalSeconds float64

	// Parse weeks
	if matches[1] != "" {
		if weeks, err := strconv.Atoi(matches[1]); err == nil {
			totalSeconds += float64(weeks * 7 * 24 * 3600)
		}
	}

	// Parse days
	if matches[2] != "" {
		if days, err := strconv.Atoi(matches[2]); err == nil {
			totalSeconds += float64(days * 24 * 3600)
		}
	}

	// Parse hours
	if matches[3] != "" {
		if hours, err := strconv.Atoi(matches[3]); err == nil {
			totalSeconds += float64(hours * 3600)
		}
	}

	// Parse minutes
	if matches[4] != "" {
		if minutes, err := strconv.Atoi(matches[4]); err == nil {
			totalSeconds += float64(minutes * 60)
		}
	}

	// Parse seconds
	if matches[5] != "" {
		if seconds, err := strconv.Atoi(matches[5]); err == nil {
			totalSeconds += float64(seconds)
		}
	}

	return totalSeconds, nil
}
