package system

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

// Collector implements the collector.Collector interface for system metrics
type Collector struct {
	systemInfoDesc     *prometheus.Desc
	cpuCoresDesc       *prometheus.Desc
	cpuFreqDesc        *prometheus.Desc
	cpuLoadDesc        *prometheus.Desc
	totalDiskDesc      *prometheus.Desc
	freeDiskDesc       *prometheus.Desc
	badBlocksDesc      *prometheus.Desc
	writeSectTotalDesc *prometheus.Desc
	totalMemoryDesc    *prometheus.Desc
	freeMemoryDesc     *prometheus.Desc
	uptimeDesc         *prometheus.Desc
	namespace          string
}

// SystemResourceData represents the structure returned by Mikrotik system resource API
type SystemResourceData struct {
	ArchitectureName     string `json:"architecture-name"`
	BadBlocks            string `json:"bad-blocks"`
	BoardName            string `json:"board-name"`
	BuildTime            string `json:"build-time"`
	CPU                  string `json:"cpu"`
	CPUCount             string `json:"cpu-count"`
	CPUFrequency         string `json:"cpu-frequency"`
	CPULoad              string `json:"cpu-load"`
	FactorySoftware      string `json:"factory-software"`
	FreeHDDSpace         string `json:"free-hdd-space"`
	FreeMemory           string `json:"free-memory"`
	Platform             string `json:"platform"`
	TotalHDDSpace        string `json:"total-hdd-space"`
	TotalMemory          string `json:"total-memory"`
	Uptime               string `json:"uptime"`
	Version              string `json:"version"`
	WriteSectSinceReboot string `json:"write-sect-since-reboot"`
	WriteSectTotal       string `json:"write-sect-total"`
}

// NewCollector creates a new system collector
func NewCollector() *Collector {
	c := &Collector{
		namespace: "mikrotik_exporter", // default namespace
	}
	c.initMetrics()
	return c
}

// initMetrics initializes the metric descriptors with the current namespace
func (c *Collector) initMetrics() {
	targetLabel := []string{"target"}

	// System info metric with additional labels
	c.systemInfoDesc = prometheus.NewDesc(
		c.namespace+"_system_info",
		"System information",
		[]string{"target", "board_name", "cpu_model", "version", "platform"},
		nil,
	)

	// CPU metrics
	c.cpuCoresDesc = prometheus.NewDesc(
		c.namespace+"_system_cpu_cores",
		"Number of CPU cores",
		targetLabel, nil,
	)
	c.cpuFreqDesc = prometheus.NewDesc(
		c.namespace+"_system_cpu_freq",
		"CPU frequency in MHz",
		targetLabel, nil,
	)
	c.cpuLoadDesc = prometheus.NewDesc(
		c.namespace+"_system_cpu_load",
		"CPU load percentage",
		targetLabel, nil,
	)

	// Disk metrics
	c.totalDiskDesc = prometheus.NewDesc(
		c.namespace+"_system_total_disk",
		"Total disk space in bytes",
		targetLabel, nil,
	)
	c.freeDiskDesc = prometheus.NewDesc(
		c.namespace+"_system_free_disk",
		"Free disk space in bytes",
		targetLabel, nil,
	)
	c.badBlocksDesc = prometheus.NewDesc(
		c.namespace+"_system_bad_blocks",
		"Number of bad blocks",
		targetLabel, nil,
	)
	c.writeSectTotalDesc = prometheus.NewDesc(
		c.namespace+"_system_write_sect_total",
		"Total write sectors",
		targetLabel, nil,
	)

	// Memory metrics
	c.totalMemoryDesc = prometheus.NewDesc(
		c.namespace+"_system_total_memory",
		"Total memory in bytes",
		targetLabel, nil,
	)
	c.freeMemoryDesc = prometheus.NewDesc(
		c.namespace+"_system_free_memory",
		"Free memory in bytes",
		targetLabel, nil,
	)

	// Uptime metric
	c.uptimeDesc = prometheus.NewDesc(
		c.namespace+"_system_uptime",
		"System uptime in seconds",
		targetLabel, nil,
	)
}

// Name returns the collector name
func (c *Collector) Name() string {
	return "system"
}

// Describe sends the descriptors of each metric over to the provided channel
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.systemInfoDesc
	ch <- c.cpuCoresDesc
	ch <- c.cpuFreqDesc
	ch <- c.cpuLoadDesc
	ch <- c.totalDiskDesc
	ch <- c.freeDiskDesc
	ch <- c.badBlocksDesc
	ch <- c.writeSectTotalDesc
	ch <- c.totalMemoryDesc
	ch <- c.freeMemoryDesc
	ch <- c.uptimeDesc
}

// SetNamespace sets the metrics namespace prefix
func (c *Collector) SetNamespace(namespace string) {
	c.namespace = namespace
	c.initMetrics()
}

// Collect fetches the metrics from Mikrotik device and sends them to Prometheus
func (c *Collector) Collect(ctx context.Context, target string, auth collector.AuthInfo, ch chan<- prometheus.Metric) error {
	// Fetch system resource data from Mikrotik REST API
	resource, err := c.fetchSystemResource(ctx, target, auth)
	if err != nil {
		return fmt.Errorf("failed to fetch system resource: %w", err)
	}

	targetLabel := []string{target}

	// System info metric with labels
	ch <- prometheus.MustNewConstMetric(
		c.systemInfoDesc,
		prometheus.GaugeValue,
		1.0,
		target, resource.BoardName, resource.CPU, resource.Version, resource.Platform,
	)

	// CPU metrics
	if cpuCores, err := parseUint64(resource.CPUCount); err == nil {
		ch <- prometheus.MustNewConstMetric(c.cpuCoresDesc, prometheus.GaugeValue, float64(cpuCores), targetLabel...)
	}
	if cpuFreq, err := parseUint64(resource.CPUFrequency); err == nil {
		ch <- prometheus.MustNewConstMetric(c.cpuFreqDesc, prometheus.GaugeValue, float64(cpuFreq), targetLabel...)
	}
	if cpuLoad, err := parseUint64(resource.CPULoad); err == nil {
		ch <- prometheus.MustNewConstMetric(c.cpuLoadDesc, prometheus.GaugeValue, float64(cpuLoad), targetLabel...)
	}

	// Disk metrics
	if totalDisk, err := parseUint64(resource.TotalHDDSpace); err == nil {
		ch <- prometheus.MustNewConstMetric(c.totalDiskDesc, prometheus.GaugeValue, float64(totalDisk), targetLabel...)
	}
	if freeDisk, err := parseUint64(resource.FreeHDDSpace); err == nil {
		ch <- prometheus.MustNewConstMetric(c.freeDiskDesc, prometheus.GaugeValue, float64(freeDisk), targetLabel...)
	}
	if badBlocks, err := parseUint64(resource.BadBlocks); err == nil {
		ch <- prometheus.MustNewConstMetric(c.badBlocksDesc, prometheus.GaugeValue, float64(badBlocks), targetLabel...)
	}
	if writeSectTotal, err := parseUint64(resource.WriteSectTotal); err == nil {
		ch <- prometheus.MustNewConstMetric(c.writeSectTotalDesc, prometheus.CounterValue, float64(writeSectTotal), targetLabel...)
	}

	// Memory metrics
	if totalMemory, err := parseUint64(resource.TotalMemory); err == nil {
		ch <- prometheus.MustNewConstMetric(c.totalMemoryDesc, prometheus.GaugeValue, float64(totalMemory), targetLabel...)
	}
	if freeMemory, err := parseUint64(resource.FreeMemory); err == nil {
		ch <- prometheus.MustNewConstMetric(c.freeMemoryDesc, prometheus.GaugeValue, float64(freeMemory), targetLabel...)
	}

	// Uptime metric
	if uptime := parseUptime(resource.Uptime); uptime > 0 {
		ch <- prometheus.MustNewConstMetric(c.uptimeDesc, prometheus.GaugeValue, float64(uptime), targetLabel...)
	}

	return nil
}

// fetchSystemResource fetches system resource data from Mikrotik REST API
func (c *Collector) fetchSystemResource(ctx context.Context, target string, auth collector.AuthInfo) (*SystemResourceData, error) {
	url := fmt.Sprintf("http://%s/rest/system/resource", target)

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

	var resource SystemResourceData
	if err := json.NewDecoder(resp.Body).Decode(&resource); err != nil {
		return nil, err
	}

	return &resource, nil
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
