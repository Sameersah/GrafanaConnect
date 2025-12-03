package plugin

import (
	"bytes"
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

// LokiHandler handles Loki log queries
type LokiHandler struct {
	config *models.DataSourceConfig
	logger log.Logger
}

// handleLokiQuery processes Loki queries
func (d *Datasource) handleLokiQuery(ctx context.Context, query backend.DataQuery, queryModel *models.QueryModel) backend.DataResponse {
	handler := &LokiHandler{
		config: d.config,
		logger: d.logger,
	}

	if d.config.LokiURL == "" {
		return backend.DataResponse{
			Error: fmt.Errorf("Loki URL not configured"),
		}
	}

	if queryModel.LogQL == "" {
		return backend.DataResponse{
			Error: fmt.Errorf("LogQL query is required"),
		}
	}

	return handler.executeQuery(ctx, query, queryModel)
}

// executeQuery executes a Loki query
func (h *LokiHandler) executeQuery(ctx context.Context, query backend.DataQuery, queryModel *models.QueryModel) backend.DataResponse {
	// Build query URL
	queryURL := fmt.Sprintf("%s/loki/api/v1/query_range", h.config.LokiURL)

	// Build query parameters
	params := url.Values{}
	params.Set("query", queryModel.LogQL)
	params.Set("start", strconv.FormatInt(query.TimeRange.From.UnixNano(), 10))
	params.Set("end", strconv.FormatInt(query.TimeRange.To.UnixNano(), 10))
	params.Set("limit", "1000") // Default limit

	// Make HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", queryURL+"?"+params.Encode(), nil)
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
			Error: fmt.Errorf("Loki API returned status %d: %s", resp.StatusCode, string(body)),
		}
	}

	// Parse response
	var lokiResp models.LokiQueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&lokiResp); err != nil {
		return backend.DataResponse{
			Error: fmt.Errorf("failed to parse response: %w", err),
		}
	}

	if lokiResp.Status != "success" {
		return backend.DataResponse{
			Error: fmt.Errorf("Loki query failed: %s", lokiResp.Status),
		}
	}

	// Convert to Grafana data frames
	frames, err := h.convertToDataFrames(&lokiResp)
	if err != nil {
		return backend.DataResponse{
			Error: fmt.Errorf("failed to convert response: %w", err),
		}
	}

	return backend.DataResponse{
		Frames: frames,
	}
}

// convertToDataFrames converts Loki response to Grafana data frames
func (h *LokiHandler) convertToDataFrames(resp *models.LokiQueryResponse) (data.Frames, error) {
	var frames data.Frames

	for _, result := range resp.Data.Result {
		// Extract labels
		labels := result.Stream

		// Parse log entries
		times := make([]time.Time, 0, len(result.Values))
		values := make([]string, 0, len(result.Values))

		for _, val := range result.Values {
			if len(val) < 2 {
				continue
			}

			// Parse timestamp (nanoseconds)
			tsNano, err := strconv.ParseInt(val[0], 10, 64)
			if err != nil {
				h.logger.Warn("Failed to parse timestamp", "error", err, "value", val[0])
				continue
			}
			timestamp := time.Unix(0, tsNano)

			// Log line
			logLine := val[1]

			times = append(times, timestamp)
			values = append(values, logLine)
		}

		if len(times) == 0 {
			continue
		}

		// Create data frame
		timeField := data.NewField("time", nil, times)
		valueField := data.NewField("value", labels, values)

		// Set field config
		valueField.Config = &data.FieldConfig{
			DisplayNameFromDS: h.buildSeriesName(labels),
		}

		frame := data.NewFrame("", timeField, valueField)
		frame.Meta = &data.FrameMeta{
			Type: data.FrameTypeLogLines,
		}

		// Add labels as frame metadata
		frame.Meta.Custom = map[string]interface{}{
			"labels": labels,
		}

		frames = append(frames, frame)
	}

	return frames, nil
}

// buildSeriesName creates a series name from log labels
func (h *LokiHandler) buildSeriesName(labels map[string]string) string {
	if job, ok := labels["job"]; ok {
		return job
	}
	if instance, ok := labels["instance"]; ok {
		return instance
	}
	if len(labels) > 0 {
		// Use first label value
		for _, v := range labels {
			return v
		}
	}
	return "logs"
}

// addAuthHeaders adds authentication headers to the request
func (h *LokiHandler) addAuthHeaders(req *http.Request) {
	if h.config.BearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+h.config.BearerToken)
	} else if h.config.APIKey != "" {
		req.Header.Set("X-API-Key", h.config.APIKey)
	} else if h.config.BasicAuthUser != "" && h.config.BasicAuthPass != "" {
		req.SetBasicAuth(h.config.BasicAuthUser, h.config.BasicAuthPass)
	}
}

// handleLokiResource handles resource calls for Loki
func (d *Datasource) handleLokiResource(ctx context.Context, req *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	// Proxy the request to Loki
	client := &http.Client{Timeout: 30 * time.Second}

	// Build URL
	targetURL := d.config.LokiURL + req.Path
	if len(req.URL) > 0 && req.URL != req.Path {
		// Parse URL to extract query string if present
		if parsedURL, err := url.Parse(req.URL); err == nil && parsedURL.RawQuery != "" {
			targetURL += "?" + parsedURL.RawQuery
		}
	}

	var bodyReader io.Reader
	if len(req.Body) > 0 {
		bodyReader = bytes.NewReader(req.Body)
	}

	proxyReq, err := http.NewRequestWithContext(ctx, req.Method, targetURL, bodyReader)
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

