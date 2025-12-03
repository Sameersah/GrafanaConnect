package plugin

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Sameersah/GrafanaConnect/pkg/models"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

// Make sure Datasource implements required interfaces
var (
	_ backend.QueryDataHandler      = (*Datasource)(nil)
	_ backend.CheckHealthHandler    = (*Datasource)(nil)
	_ backend.CallResourceHandler   = (*Datasource)(nil)
	_ instancemgmt.InstanceDisposer = (*Datasource)(nil)
)

// Datasource is the main plugin struct
type Datasource struct {
	settings *backend.DataSourceInstanceSettings
	config   *models.DataSourceConfig
	logger   log.Logger
}

// NewDatasource creates a new instance of the datasource
func NewDatasource(ctx context.Context, settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	ds := &Datasource{
		settings: &settings,
		logger:   log.New(),
	}

	// Parse configuration
	config := &models.DataSourceConfig{}
	if err := json.Unmarshal(settings.JSONData, config); err != nil {
		ds.logger.Warn("Failed to parse JSON data, using defaults", "error", err)
	}

	// Load secure settings
	if val, ok := settings.DecryptedSecureJSONData["apiKey"]; ok {
		config.APIKey = val
	}
	if val, ok := settings.DecryptedSecureJSONData["basicAuthPass"]; ok {
		config.BasicAuthPass = val
	}
	if val, ok := settings.DecryptedSecureJSONData["bearerToken"]; ok {
		config.BearerToken = val
	}

	ds.config = config
	ds.logger.Info("Datasource initialized", "prometheusUrl", config.PrometheusURL, "lokiUrl", config.LokiURL)

	return ds, nil
}

// Dispose cleans up resources
func (d *Datasource) Dispose() {
	d.logger.Info("Disposing datasource")
}

// QueryData handles data queries
func (d *Datasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	response := backend.NewQueryDataResponse()

	for _, q := range req.Queries {
		res := d.handleQuery(ctx, q)
		response.Responses[q.RefID] = res
	}

	return response, nil
}

// handleQuery routes queries to appropriate handlers
func (d *Datasource) handleQuery(ctx context.Context, query backend.DataQuery) backend.DataResponse {
	var queryModel models.QueryModel
	if err := json.Unmarshal(query.JSON, &queryModel); err != nil {
		return backend.DataResponse{
			Error: fmt.Errorf("failed to parse query: %w", err),
		}
	}

	queryModel.RefID = query.RefID

	d.logger.Debug("Handling query", "type", queryModel.QueryType, "refId", query.RefID)

	switch queryModel.QueryType {
	case models.QueryTypePrometheus:
		return d.handlePrometheusQuery(ctx, query, &queryModel)
	case models.QueryTypeLoki:
		return d.handleLokiQuery(ctx, query, &queryModel)
	case models.QueryTypeREST:
		return d.handleRESTQuery(ctx, query, &queryModel)
	default:
		return backend.DataResponse{
			Error: fmt.Errorf("unknown query type: %s", queryModel.QueryType),
		}
	}
}

// CheckHealth checks the health of the datasource
func (d *Datasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	var status backend.HealthStatus
	var message string

	// Check if at least one data source is configured
	if d.config.PrometheusURL == "" && d.config.LokiURL == "" && d.config.RESTURL == "" {
		status = backend.HealthStatusError
		message = "No data source URLs configured. Please configure at least one data source."
	} else {
		status = backend.HealthStatusOk
		message = "Data source is ready"
		
		// Try to verify connectivity
		if d.config.PrometheusURL != "" {
			if err := d.checkPrometheusHealth(ctx); err != nil {
				status = backend.HealthStatusError
				message = fmt.Sprintf("Prometheus connection issue: %v", err)
			}
		}
	}

	return &backend.CheckHealthResult{
		Status:  status,
		Message: message,
	}, nil
}

// CallResource handles resource calls
func (d *Datasource) CallResource(ctx context.Context, req *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	d.logger.Debug("Resource call", "path", req.Path, "method", req.Method)
	
	// Handle resource calls for proxying requests
	switch req.Path {
	case "prometheus":
		return d.handlePrometheusResource(ctx, req, sender)
	case "loki":
		return d.handleLokiResource(ctx, req, sender)
	case "rest":
		return d.handleRESTResource(ctx, req, sender)
	default:
		return sender.Send(&backend.CallResourceResponse{
			Status: 404,
			Body:   []byte(`{"error": "Unknown resource path"}`),
		})
	}
}

// checkPrometheusHealth verifies Prometheus connectivity
func (d *Datasource) checkPrometheusHealth(ctx context.Context) error {
	promHandler := &PrometheusHandler{
		config: d.config,
		logger: d.logger,
	}
	return promHandler.checkHealth(ctx)
}

