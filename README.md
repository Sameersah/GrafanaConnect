# GrafanaConnect

A unified Grafana data source plugin built in Go that integrates Prometheus-style metrics, Loki log streams, and external REST APIs into unified Grafana dashboards.

## Features

- **Multi-Source Support**: Single plugin handles Prometheus metrics, Loki logs, and REST APIs
- **Prometheus Integration**: Full PromQL support for instant and range queries
- **Loki Integration**: LogQL query support for log stream analysis
- **REST API Integration**: Generic REST API integration with configurable endpoints and JSON transformation
- **Flexible Authentication**: Support for API keys, basic auth, and bearer tokens
- **Unified Dashboards**: Combine data from multiple sources in a single dashboard

## Architecture

The plugin consists of:

- **Backend (Go)**: Handles query execution, data fetching, and transformation using the Grafana Plugin SDK for Go
- **Frontend (React/TypeScript)**: Provides configuration UI and query builder interface

## Installation

### Prerequisites

- Go 1.21 or higher
- Node.js 16+ and npm/yarn
- Grafana 9.0 or higher

### Build from Source

1. Clone the repository:
```bash
git clone https://github.com/Sameersah/GrafanaConnect.git
cd GrafanaConnect
```

2. Install dependencies:
```bash
make deps
```

3. Build the plugin:
```bash
make build
```

This will:
- Build the Go backend binary (`dist/gpx_grafana-connect`)
- Build the frontend assets

### Development Setup

1. Install dependencies:
```bash
make deps
```

2. Start development mode:
```bash
make dev
```

This will watch for changes and rebuild automatically.

## Configuration

### Data Source Setup

1. In Grafana, go to **Configuration** → **Data Sources**
2. Click **Add data source**
3. Search for **GrafanaConnect** and select it
4. Configure the following:

#### Prometheus Configuration

- **Prometheus URL**: Base URL of your Prometheus instance (e.g., `http://prometheus:9090`)

#### Loki Configuration

- **Loki URL**: Base URL of your Loki instance (e.g., `http://loki:3100`)

#### REST API Configuration

- **REST API Base URL**: Base URL for your REST API endpoints (e.g., `https://api.example.com`)

#### Authentication

Choose one of the following authentication methods:

- **API Key**: Set `API Key (Plain)` or `API Key (Secure)` for secure storage
- **Basic Auth**: Set `Basic Auth Username` and `Basic Auth Password`
- **Bearer Token**: Set `Bearer Token (Plain)` or `Bearer Token (Secure)` for secure storage

5. Click **Save & Test** to verify connectivity

## Usage

### Prometheus Queries

1. Create a new panel in Grafana
2. Select GrafanaConnect as the data source
3. Set **Query Type** to **Prometheus**
4. Enter a PromQL query in the **PromQL Query** field

**Example PromQL Queries:**
```
# CPU usage
100 - (avg(irate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)

# Memory usage
node_memory_MemTotal_bytes - node_memory_MemAvailable_bytes

# HTTP request rate
rate(http_requests_total[5m])
```

### Loki Queries

1. Create a new panel in Grafana
2. Select GrafanaConnect as the data source
3. Set **Query Type** to **Loki**
4. Enter a LogQL query in the **LogQL Query** field

**Example LogQL Queries:**
```
# All logs from a specific job
{job="varlogs"}

# Error logs
{job="varlogs"} |= "error"

# Logs with rate calculation
rate({job="varlogs"}[5m])
```

### REST API Queries

1. Create a new panel in Grafana
2. Select GrafanaConnect as the data source
3. Set **Query Type** to **REST API**
4. Configure:
   - **Endpoint**: API endpoint path (e.g., `/api/v1/metrics`)
   - **HTTP Method**: GET, POST, PUT, PATCH, or DELETE
   - **Request Body**: JSON body for POST/PUT/PATCH requests (optional)

**Example REST API Queries:**

**GET Request:**
- Endpoint: `/api/v1/metrics`
- Method: `GET`

**POST Request:**
- Endpoint: `/api/v1/data`
- Method: `POST`
- Body:
```json
{
  "start": "2024-01-01T00:00:00Z",
  "end": "2024-01-01T23:59:59Z"
}
```

### Data Format

The plugin automatically converts REST API responses to Grafana data frames:

- **Arrays of Objects**: Treated as time series if timestamp fields are detected
- **Objects**: Converted to table format
- **Primitive Values**: Wrapped in a simple time series

The plugin looks for common timestamp field names: `time`, `timestamp`, `date`, `ts`, `datetime`.

## Examples

### Example 1: Prometheus Metrics Dashboard

Create a dashboard showing CPU and memory usage from Prometheus:

1. Add a time series panel
2. Data source: GrafanaConnect
3. Query Type: Prometheus
4. PromQL: `100 - (avg(irate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)`

### Example 2: Loki Logs Dashboard

Create a dashboard showing error logs from Loki:

1. Add a logs panel
2. Data source: GrafanaConnect
3. Query Type: Loki
4. LogQL: `{job="varlogs"} |= "error"`

### Example 3: REST API Metrics

Create a dashboard from a REST API endpoint:

1. Add a time series panel
2. Data source: GrafanaConnect
3. Query Type: REST API
4. Endpoint: `/api/v1/metrics`
5. Method: `GET`

## Development

### Project Structure

```
graphana/
├── src/                    # Frontend TypeScript/React code
│   ├── datasource.ts       # Main datasource class
│   ├── ConfigEditor.tsx    # Configuration UI
│   ├── QueryEditor.tsx     # Query builder UI
│   └── types.ts            # TypeScript types
├── pkg/
│   ├── plugin/             # Backend Go code
│   │   ├── datasource.go   # Main plugin entry point
│   │   ├── prometheus.go   # Prometheus handler
│   │   ├── loki.go         # Loki handler
│   │   └── restapi.go      # REST API handler
│   └── models/
│       └── models.go       # Data models
├── plugin.json             # Plugin manifest
├── go.mod                  # Go dependencies
├── package.json            # Node.js dependencies
└── Makefile               # Build automation
```

### Running Tests

```bash
make test
```

### Building for Production

```bash
make build
```

The built plugin will be in the `dist/` directory.

## Troubleshooting

### Plugin Not Loading

- Ensure Grafana version is 9.0 or higher
- Check that the plugin binary is executable: `chmod +x dist/gpx_grafana-connect`
- Review Grafana logs: `journalctl -u grafana-server -f` (Linux) or check Grafana logs directory

### Connection Errors

- Verify data source URLs are correct and accessible
- Check authentication credentials
- Ensure firewall rules allow connections
- Test connectivity manually: `curl http://prometheus:9090/api/v1/query?query=up`

### Query Errors

- **Prometheus**: Verify PromQL syntax is correct
- **Loki**: Verify LogQL syntax is correct
- **REST API**: Check that the endpoint returns valid JSON

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## License

This project is licensed under the Apache-2.0 License - see the LICENSE file for details.

## Author

**Sameer Sah**

- GitHub: [@Sameersah](https://github.com/Sameersah)

## Acknowledgments

- Built with [Grafana Plugin SDK for Go](https://github.com/grafana/grafana-plugin-sdk-go)
- Uses [Grafana Toolkit](https://github.com/grafana/grafana-toolkit) for frontend development

