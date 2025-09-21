package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/mikrotik-exporter/collector"
	"github.com/mikrotik-exporter/collector/bgp"
	"github.com/mikrotik-exporter/collector/dhcp"
	"github.com/mikrotik-exporter/collector/interfaces"
	"github.com/mikrotik-exporter/collector/system"
	"github.com/mikrotik-exporter/collector/wireless"
	"github.com/mikrotik-exporter/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	cfg               *config.Config
	collectorRegistry *collector.Registry
)

func init() {
	// Initialize collector registry - collectors will be registered in main() with namespace
	collectorRegistry = collector.NewRegistry()
}

func main() {
	// Get configuration from environment variables
	listenAddr := getEnv("LISTEN_ADDR", "0.0.0.0")
	listenPort := getEnv("LISTEN_PORT", "9261")
	configFile := getEnv("CONFIG_FILE", "./config.yaml")
	metricsNamespace := getEnv("METRICS_NAMESPACE", "mikrotik_exporter")

	// Register collectors with namespace
	interfacesCollector := interfaces.NewCollector()
	interfacesCollector.SetNamespace(metricsNamespace)
	collectorRegistry.Register(interfacesCollector)

	dhcpCollector := dhcp.NewCollector()
	dhcpCollector.SetNamespace(metricsNamespace)
	collectorRegistry.Register(dhcpCollector)

	bgpCollector := bgp.NewCollector()
	bgpCollector.SetNamespace(metricsNamespace)
	collectorRegistry.Register(bgpCollector)

	systemCollector := system.NewCollector()
	systemCollector.SetNamespace(metricsNamespace)
	collectorRegistry.Register(systemCollector)

	wirelessCollector := wireless.NewCollector()
	wirelessCollector.SetNamespace(metricsNamespace)
	collectorRegistry.Register(wirelessCollector)

	// Load configuration
	var err error
	cfg, err = config.LoadConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Setup HTTP handlers
	http.HandleFunc("/probe", probeHandler)
	http.HandleFunc("/health-check", healthCheckHandler)
	http.HandleFunc("/", indexHandler)

	// Setup metrics with default Go metrics
	registry := prometheus.NewRegistry()
	registry.MustRegister(prometheus.NewGoCollector())
	registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	// Start HTTP server
	addr := fmt.Sprintf("%s:%s", listenAddr, listenPort)
	log.Printf("Starting Mikrotik Prometheus Exporter on %s", addr)
	log.Printf("Available collectors: %v", collectorRegistry.List())

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}

func probeHandler(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("target")
	authName := r.URL.Query().Get("auth")
	moduleName := r.URL.Query().Get("module")

	// Validate required parameters
	if target == "" {
		http.Error(w, "Missing 'target' parameter", http.StatusBadRequest)
		return
	}
	if authName == "" {
		authName = "default"
	}
	if moduleName == "" {
		moduleName = "default"
	}

	// Get authentication configuration
	authConfig, err := cfg.GetAuth(authName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Auth configuration error: %v", err), http.StatusBadRequest)
		return
	}

	// Get module configuration
	moduleConfig, err := cfg.GetModule(moduleName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Module configuration error: %v", err), http.StatusBadRequest)
		return
	}

	// Create a custom registry for this probe
	registry := prometheus.NewRegistry()

	// Get enabled collectors
	enabledCollectors := collectorRegistry.GetEnabled(moduleConfig.Collectors)
	if len(enabledCollectors) == 0 {
		http.Error(w, "No collectors enabled for this module", http.StatusBadRequest)
		return
	}

	// Create a custom collector that will run all enabled collectors
	probeCollector := &ProbeCollector{
		target:     target,
		auth:       collector.AuthInfo{Username: authConfig.Username, Password: authConfig.Password},
		collectors: enabledCollectors,
	}

	registry.MustRegister(probeCollector)

	// Create a timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Store context in the collector for use during collection
	probeCollector.ctx = ctx

	// Serve metrics
	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok","service":"mikrotik-exporter"}`))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Mikrotik Prometheus Exporter</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .container { max-width: 800px; }
        .endpoint { background: #f5f5f5; padding: 10px; margin: 10px 0; border-radius: 5px; }
        code { background: #e8e8e8; padding: 2px 4px; border-radius: 3px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Mikrotik Prometheus Exporter</h1>
        <p>This exporter provides Prometheus metrics for Mikrotik devices using the REST API.</p>
        
        <h2>Usage</h2>
        <div class="endpoint">
            <strong>Probe endpoint:</strong><br>
            <code>/probe?target=&lt;ip:port&gt;&amp;auth=&lt;auth_name&gt;&amp;module=&lt;module_name&gt;</code>
        </div>
        
        <h3>Parameters:</h3>
        <ul>
            <li><strong>target</strong> (required): IP address and port of the Mikrotik device (e.g., 192.168.1.1:80)</li>
            <li><strong>auth</strong> (optional): Authentication configuration name (default: "default")</li>
            <li><strong>module</strong> (optional): Module configuration name (default: "default")</li>
        </ul>
        
        <h3>Available Collectors:</h3>
        <ul>`

	for _, collectorName := range collectorRegistry.List() {
		html += fmt.Sprintf("<li>%s</li>", collectorName)
	}

	html += `        </ul>
        
        <h3>Examples:</h3>
        <div class="endpoint">
            <code>/probe?target=192.168.1.1:80</code><br>
            <small>Probe device with default auth and module</small>
        </div>
        <div class="endpoint">
            <code>/probe?target=192.168.1.1:80&amp;auth=production&amp;module=minimal</code><br>
            <small>Probe device with custom auth and module</small>
        </div>
        
        <h3>Other Endpoints:</h3>
        <div class="endpoint">
            <code>/metrics</code> - Exporter's own metrics (includes Go runtime metrics)
        </div>
        <div class="endpoint">
            <code>/health-check</code> - Health check endpoint (returns JSON status)
        </div>
    </div>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// ProbeCollector implements prometheus.Collector for multi-target probing
type ProbeCollector struct {
	target     string
	auth       collector.AuthInfo
	collectors []collector.Collector
	ctx        context.Context
}

func (pc *ProbeCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, c := range pc.collectors {
		c.Describe(ch)
	}
}

func (pc *ProbeCollector) Collect(ch chan<- prometheus.Metric) {
	for _, c := range pc.collectors {
		if err := c.Collect(pc.ctx, pc.target, pc.auth, ch); err != nil {
			log.Printf("Error collecting metrics from %s collector: %v", c.Name(), err)
			// Continue with other collectors even if one fails
		}
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
