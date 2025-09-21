# Mikrotik Prometheus Exporter

A Prometheus exporter for Mikrotik devices using the REST API (RouterOS 7.x+). This exporter follows a multi-target architecture, allowing you to scrape multiple devices from a single exporter instance.

## Features

- **Multi-target architecture**: Scrape multiple devices from one exporter
- **Modular collectors**: Enable/disable specific metric collectors per device
- **Flexible authentication**: Support for multiple authentication configurations
- **REST API based**: Uses Mikrotik's modern REST API instead of the legacy API protocol
- **Docker support**: Ready-to-use Docker image with multi-stage build

## Available Collectors

- **interfaces**: Network interface metrics (RX/TX bytes, packets, status)
- **dhcp**: DHCP server and lease metrics
- **bgp**: BGP peer status and prefix information
- **system**: System metrics (uptime, CPU, memory, disk, temperature)
- **wireless**: Wireless interface and client metrics

## Configuration

### Environment Variables

- `LISTEN_ADDR`: Listen address (default: `0.0.0.0`)
- `LISTEN_PORT`: Listen port (default: `9261`)
- `CONFIG_FILE`: Configuration file path (default: `./config.yaml`)

### Configuration File

Copy `config.dist.yaml` to `config.yaml` and modify with your authentication and module configurations:

```yaml
# Authentication configurations
auths:
  default:
    username: admin
    password: admin
  
  production:
    username: monitoring
    password: secure_password

# Module configurations
modules:
  default:
    collectors:
      interfaces: true
      dhcp: true
      bgp: true
      system: true
      wireless: true
      
  minimal:
    collectors:
      interfaces: true
      system: true
```

## Usage

### Running with Go

```bash
# Copy and configure the config file
cp config.dist.yaml config.yaml
# Edit config.yaml with your device credentials

go mod tidy
go run main.go
```

### Running with Docker

```bash
# Build the image
docker build -t mikrotik-exporter .

# Run the container
docker run -p 9261:9261 -v $(pwd)/config.yaml:/app/config.yaml mikrotik-exporter
```

### Prometheus Configuration

Add the following to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'mikrotik'
    static_configs:
      - targets:
        - 192.168.1.1:80    # Mikrotik device IP:port
        - 192.168.1.2:80    # Another device
    metrics_path: /probe
    params:
      auth: [default]       # Auth configuration name
      module: [default]     # Module configuration name
    relabel_configs:
      - source_labels: [__address__]
        target_label: __param_target
      - source_labels: [__param_target]
        target_label: instance
      - target_label: __address__
        replacement: localhost:9261  # Exporter address
```

## API Endpoints

### Probe Endpoint

```
GET /probe?target=<ip:port>&auth=<auth_name>&module=<module_name>
```

**Parameters:**
- `target` (required): IP address and port of the Mikrotik device
- `auth` (optional): Authentication configuration name (default: "default")
- `module` (optional): Module configuration name (default: "default")

**Examples:**
- `/probe?target=192.168.1.1:80`
- `/probe?target=192.168.1.1:80&auth=production&module=minimal`

### Other Endpoints

- `/`: Web interface with usage information
- `/metrics`: Exporter's own metrics

## Development

### Project Structure

```
.
├── main.go                 # Main application entry point
├── config/
│   └── config.go          # Configuration parsing
├── collector/
│   ├── collector.go       # Collector interface and registry
│   ├── interfaces/        # Interface metrics collector
│   ├── dhcp/             # DHCP metrics collector
│   ├── bgp/              # BGP metrics collector
│   ├── system/           # System metrics collector
│   └── wireless/         # Wireless metrics collector
├── config.yaml           # Default configuration
├── Dockerfile            # Docker build configuration
├── go.mod               # Go module definition
└── README.md            # This file
```

### Adding New Collectors

1. Create a new directory under `collector/`
2. Implement the `collector.Collector` interface
3. Register the collector in `main.go`
4. Add the collector to your module configuration

### Collector Interface

```go
type Collector interface {
    Name() string
    Describe(ch chan<- *prometheus.Desc)
    Collect(ctx context.Context, target string, auth AuthInfo, ch chan<- prometheus.Metric) error
}
```

## Requirements

- Go 1.25+
- Mikrotik RouterOS 7.x+ (with REST API support)
- Prometheus

## License

MIT License

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## TODO

- [ ] Implement actual REST API calls to Mikrotik devices
- [ ] Add comprehensive error handling and logging
- [ ] Add unit tests for collectors
- [ ] Add integration tests
- [ ] Add more collectors (firewall, queues, etc.)
- [ ] Add TLS support for secure connections
- [ ] Add caching mechanism for better performance
