package collector

import (
	"context"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

// Collector interface defines the contract for all collectors
type Collector interface {
	// Name returns the collector name
	Name() string

	// Describe sends the descriptors of each metric over to the provided channel
	Describe(ch chan<- *prometheus.Desc)

	// Collect fetches the metrics from Mikrotik device and sends them to Prometheus
	Collect(ctx context.Context, target string, auth AuthInfo, ch chan<- prometheus.Metric) error

	// SetNamespace sets the metrics namespace prefix
	SetNamespace(namespace string)
}

// AuthInfo contains authentication information for connecting to Mikrotik device
type AuthInfo struct {
	Username string
	Password string
}

// Registry holds all available collectors
type Registry struct {
	collectors map[string]Collector
}

// NewRegistry creates a new collector registry
func NewRegistry() *Registry {
	return &Registry{
		collectors: make(map[string]Collector),
	}
}

// Register adds a collector to the registry
func (r *Registry) Register(collector Collector) {
	r.collectors[collector.Name()] = collector
}

// Get returns a collector by name
func (r *Registry) Get(name string) (Collector, error) {
	collector, exists := r.collectors[name]
	if !exists {
		return nil, fmt.Errorf("collector '%s' not found", name)
	}
	return collector, nil
}

// GetEnabled returns all enabled collectors based on the module configuration
func (r *Registry) GetEnabled(enabledCollectors map[string]bool) []Collector {
	var enabled []Collector
	for name, isEnabled := range enabledCollectors {
		if isEnabled {
			if collector, exists := r.collectors[name]; exists {
				enabled = append(enabled, collector)
			}
		}
	}
	return enabled
}

// List returns all available collector names
func (r *Registry) List() []string {
	var names []string
	for name := range r.collectors {
		names = append(names, name)
	}
	return names
}
