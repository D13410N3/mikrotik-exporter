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
- **system**: System metrics (uptime, CPU, memory, disk)
- **wireless**: Wireless interface and client metrics
- **firewall**: Firewall rule metrics (enabled status, bytes, packets, rule info)

## Configuration

### Environment Variables

- `LISTEN_ADDR`: Listen address (default: `0.0.0.0`)
- `LISTEN_PORT`: Listen port (default: `9261`)
- `CONFIG_FILE`: Configuration file path (default: `./config.yaml`)
- `METRICS_NAMESPACE`: Metrics namespace (default: `mikrotik_exporter`)

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
      firewall: true
      
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

## Available Metrics

### Interface Metrics
| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `interface_enabled` | gauge | Interface enabled status (1=enabled, 0=disabled) | name, type |
| `interface_up` | gauge | Interface running status (1=running, 0=not running) | mac, name, type, comment |
| `interface_rx_bytes_total` | counter | Total bytes received on interface | name, type |
| `interface_rx_packets_total` | counter | Total packets received on interface | name, type |
| `interface_tx_bytes_total` | counter | Total bytes transmitted on interface | name, type |
| `interface_tx_packets_total` | counter | Total packets transmitted on interface | name, type |
| `interface_fp_rx_bytes_total` | counter | Fast path bytes received on interface | name, type |
| `interface_fp_rx_packets_total` | counter | Fast path packets received on interface | name, type |
| `interface_fp_tx_bytes_total` | counter | Fast path bytes transmitted on interface | name, type |
| `interface_fp_tx_packets_total` | counter | Fast path packets transmitted on interface | name, type |
| `interface_tx_queue_drop_total` | counter | Packets dropped from TX queue | name, type |
| `interface_mtu` | gauge | Interface MTU in bytes | name, type |
| `interface_link_downs_total` | counter | Number of link down events | name, type |
| `interface_last_link_up_time` | gauge | Last link up time (Unix timestamp) | name, type |
| `interface_last_link_down_time` | gauge | Last link down time (Unix timestamp) | name, type |

### DHCP Metrics
| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `dhcp_bound` | gauge | DHCP lease bound status (1=bound, 0=not bound) | device_ip, mac, dhcp_server, device_hostname |

### BGP Metrics
| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `bgp_session_up` | gauge | BGP session status (1=established, 0=not established) | name |
| `bgp_session_prefix_count` | gauge | Number of prefixes in BGP session | name |
| `bgp_session_remote_bytes_total` | counter | Total bytes received from remote BGP peer | name |
| `bgp_session_remote_messages_total` | counter | Total messages received from remote BGP peer | name |
| `bgp_session_local_bytes_total` | counter | Total bytes sent to remote BGP peer | name |
| `bgp_session_local_messages_total` | counter | Total messages sent to remote BGP peer | name |
| `bgp_session_uptime` | gauge | BGP session uptime in seconds | name |
| `bgp_session_info` | gauge | BGP session information (always 1) | name, remote_address, remote_id, remote_as, local_address, local_id, local_as |

### System Metrics
| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `system_info` | gauge | System information (always 1) | board_name, cpu_model, version, platform |
| `system_cpu_cores` | gauge | Number of CPU cores | - |
| `system_cpu_freq` | gauge | CPU frequency in MHz | - |
| `system_cpu_load` | gauge | CPU load percentage | - |
| `system_total_disk` | gauge | Total disk space in bytes | - |
| `system_free_disk` | gauge | Free disk space in bytes | - |
| `system_bad_blocks` | gauge | Number of bad blocks | - |
| `system_write_sect_total` | counter | Total write sectors | - |
| `system_total_memory` | gauge | Total memory in bytes | - |
| `system_free_memory` | gauge | Free memory in bytes | - |
| `system_uptime` | gauge | System uptime in seconds | - |
| `system_voltage` | gauge | System voltage in volts | - |
| `system_temperature` | gauge | System temperature in Celsius | - |

### Wireless Metrics
| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `wireless_client_info` | gauge | Wireless client information (always 1 for connected clients) | mac, interface, ssid |
| `wireless_tx_bytes_total` | counter | Total bytes transmitted to wireless client | mac |
| `wireless_tx_packets_total` | counter | Total packets transmitted to wireless client | mac |
| `wireless_rx_bytes_total` | counter | Total bytes received from wireless client | mac |
| `wireless_rx_packets_total` | counter | Total packets received from wireless client | mac |
| `wireless_rx_rate` | gauge | Wireless client RX rate in bps | mac |
| `wireless_tx_rate` | gauge | Wireless client TX rate in bps | mac |
| `wireless_uptime` | gauge | Wireless client connection uptime in seconds | mac |
| `wireless_signal` | gauge | Wireless client signal strength in dBm | mac |

### Firewall Metrics
| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `firewall_rule_enabled` | gauge | Firewall rule enabled status (1=enabled, 0=disabled) | id, table |
| `firewall_rule_bytes` | counter | Number of bytes matched by firewall rule | id, table |
| `firewall_rule_packets` | counter | Number of packets matched by firewall rule | id, table |
| `firewall_rule_info` | gauge | Firewall rule information (always 1) | id, table, chain, action, comment |

### Exporter Metrics
| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `collector_success` | gauge | Whether a collector succeeded (1) or failed (0) | collector |

## License

MIT License
