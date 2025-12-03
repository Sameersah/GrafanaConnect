package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/Sameersah/GrafanaConnect/pkg/models"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

// PrometheusHandler handles Prometheus queries
type PrometheusHandler struct {
	config *models.DataSourceConfig
	logger log.Logger
}

// handlePrometheusQuery processes Prometheus queries
func (d *Datasource) handlePrometheusQuery(ctx context.Context, query backend.DataQuery, queryModel *models.QueryModel) backend.DataResponse {
	handler := &PrometheusHandler{
		config: d.config,
		logger: d.logger,
	}

	if d.config.PrometheusURL == "" {
		return backend.DataResponse{
			Error: fmt.Errorf("Prometheus URL not configured"),
		}
	}

	if queryModel.PromQL == "" {
		return backend.DataResponse{
			Error: fmt.Errorf("PromQL query is required"),
		}
	}

	return handler.executeQuery(ctx, query, queryModel)
}

// executeQuery executes a Prometheus query
func (h *PrometheusHandler) executeQuery(ctx context.Context, query backend.DataQuery, queryModel *models.QueryModel) backend.DataResponse {
	// Determine query type (instant vs range)
	isRangeQuery := !query.TimeRange.From.Equal(query.TimeRange.To)

	var promURL string
	if isRangeQuery {
		// Range query
		promURL = fmt.Sprintf("%s/api/v1/query_range", h.config.PrometheusURL)
	} else {
		// Instant query
		promURL = fmt.Sprintf("%s/api/v1/query", h.config.PrometheusURL)
	}

	// Build query parameters
	params := url.Values{}
	params.Set("query", queryModel.PromQL)

	if isRangeQuery {
		params.Set("start", strconv.FormatInt(query.TimeRange.From.Unix(), 10))
		params.Set("end", strconv.FormatInt(query.TimeRange.To.Unix(), 10))
		
		// Calculate step (default to 15s if not specified)
		step := query.Interval
		if step == 0 {
			step = 15 * time.Second
		}
		params.Set("step", strconv.FormatInt(int64(step.Seconds()), 10)+"s")
	} else {
		params.Set("time", strconv.FormatInt(query.TimeRange.To.Unix(), 10))
	}

	// Make HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", promURL+"?"+params.Encode(), nil)
	if err != nil {
		return backend.DataResponse{
			Error: fmt.Errorf("failed to create request: %w", err),
		}
	}

	// Add authentication
	h.addAuthHeaders(req)

	// Execute request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return backend.DataResponse{
			Error: fmt.Errorf("failed to execute request: %w", err),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return backend.DataResponse{
			Error: fmt.Errorf("Prometheus API returned status %d: %s", resp.StatusCode, string(body)),
		}
	}

	// Parse response
	var promResp models.PrometheusQueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&promResp); err != nil {
		return backend.DataResponse{
			Error: fmt.Errorf("failed to parse response: %w", err),
		}
	}

	if promResp.Status != "success" {
		return backend.DataResponse{
			Error: fmt.Errorf("Prometheus query failed: %s", promResp.Status),
		}
	}

	// Convert to Grafana data frames
	frames, err := h.convertToDataFrames(&promResp, isRangeQuery)
	if err != nil {
		return backend.DataResponse{
			Error: fmt.Errorf("failed to convert response: %w", err),
		}
	}

	return backend.DataResponse{
		Frames: frames,
	}
}

// convertToDataFrames converts Prometheus response to Grafana data frames
func (h *PrometheusHandler) convertToDataFrames(resp *models.PrometheusQueryResponse, isRangeQuery bool) (data.Frames, error) {
	var frames data.Frames

	for _, result := range resp.Data.Result {
		var timeField *data.Field
		var valueField *data.Field

		if isRangeQuery {
			// Range query: multiple values
			times := make([]time.Time, len(result.Values))
			values := make([]float64, len(result.Values))

			for i, val := range result.Values {
				if len(val) < 2 {
					continue
				}

				// Parse timestamp
				ts, ok := val[0].(float64)
				if !ok {
					return nil, fmt.Errorf("invalid timestamp format")
				}
				times[i] = time.Unix(int64(ts), 0)

				// Parse value
				valStr, ok := val[1].(string)
				if !ok {
					return nil, fmt.Errorf("invalid value format")
				}
				v, err := strconv.ParseFloat(valStr, 64)
				if err != nil {
					return nil, fmt.Errorf("failed to parse value: %w", err)
				}
				values[i] = v
			}

			timeField = data.NewField("time", nil, times)
			valueField = data.NewField("value", result.Metric, values)
		} else {
			// Instant query: single value
			if len(result.Value) < 2 {
				return nil, fmt.Errorf("invalid instant query response")
			}

			ts, ok := result.Value[0].(float64)
			if !ok {
				return nil, fmt.Errorf("invalid timestamp format")
			}
			timestamp := time.Unix(int64(ts), 0)

			valStr, ok := result.Value[1].(string)
			if !ok {
				return nil, fmt.Errorf("invalid value format")
			}
			v, err := strconv.ParseFloat(valStr, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse value: %w", err)
			}

			timeField = data.NewField("time", nil, []time.Time{timestamp})
			valueField = data.NewField("value", result.Metric, []float64{v})
		}

		// Set field config
		valueField.Config = &data.FieldConfig{
			DisplayNameFromDS: h.buildSeriesName(result.Metric),
		}

		frame := data.NewFrame("", timeField, valueField)
		frame.Meta = &data.FrameMeta{
			Type: data.FrameTypeTimeSeriesMany,
		}

		frames = append(frames, frame)
	}

	return frames, nil
}

// buildSeriesName creates a series name from metric labels
func (h *PrometheusHandler) buildSeriesName(metric map[string]string) string {
	if name, ok := metric["__name__"]; ok {
		return name
	}
	if instance, ok := metric["instance"]; ok {
		return instance
	}
	return "series"
}

// addAuthHeaders adds authentication headers to the request
func (h *PrometheusHandler) addAuthHeaders(req *http.Request) {
	if h.config.BearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+h.config.BearerToken)
	} else if h.config.APIKey != "" {
		req.Header.Set("X-API-Key", h.config.APIKey)
	} else if h.config.BasicAuthUser != "" && h.config.BasicAuthPass != "" {
		req.SetBasicAuth(h.config.BasicAuthUser, h.config.BasicAuthPass)
	}
}

// checkHealth verifies Prometheus connectivity
func (h *PrometheusHandler) checkHealth(ctx context.Context) error {
	healthURL := fmt.Sprintf("%s/-/healthy", h.config.PrometheusURL)
	req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
	if err != nil {
		return err
	}

	h.addAuthHeaders(req)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check returned status %d", resp.StatusCode)
	}

	return nil
}

// handlePrometheusResource handles resource calls for Prometheus
func (d *Datasource) handlePrometheusResource(ctx context.Context, req *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	// Proxy the request to Prometheus
	client := &http.Client{Timeout: 30 * time.Second}
	
	// Build URL
	targetURL := d.config.PrometheusURL + req.Path
	if len(req.URL.RawQuery) > 0 {
		targetURL += "?" + req.URL.RawQuery
	}

	proxyReq, err := http.NewRequestWithContext(ctx, req.Method, targetURL, req.Body)
	if err != nil {
		return sender.Send(&backend.CallResourceResponse{
			Status: 500,
			Body:   []byte(fmt.Sprintf(`{"error": "Failed to create request: %v"}`, err)),
		})
	}

	// Copy headers
	for k, v := range req.Headers {
		proxyReq.Header[k] = v
	}

	// Add auth
	if d.config.BearerToken != "" {
		proxyReq.Header.Set("Authorization", "Bearer "+d.config.BearerToken)
	} else if d.config.APIKey != "" {
		proxyReq.Header.Set("X-API-Key", d.config.APIKey)
	} else if d.config.BasicAuthUser != "" && d.config.BasicAuthPass != "" {
		proxyReq.SetBasicAuth(d.config.BasicAuthUser, d.config.BasicAuthPass)
	}

	resp, err := client.Do(proxyReq)
	if err != nil {
		return sender.Send(&backend.CallResourceResponse{
			Status: 500,
			Body:   []byte(fmt.Sprintf(`{"error": "Request failed: %v"}`, err)),
		})
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return sender.Send(&backend.CallResourceResponse{
			Status: 500,
			Body:   []byte(fmt.Sprintf(`{"error": "Failed to read response: %v"}`, err)),
		})
	}

	return sender.Send(&backend.CallResourceResponse{
		Status: resp.StatusCode,
		Headers: resp.Header,
		Body:   body,
	})
}

