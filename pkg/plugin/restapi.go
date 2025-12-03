package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Sameersah/GrafanaConnect/pkg/models"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

// RESTAPIHandler handles REST API queries
type RESTAPIHandler struct {
	config *models.DataSourceConfig
	logger log.Logger
}

// handleRESTQuery processes REST API queries
func (d *Datasource) handleRESTQuery(ctx context.Context, query backend.DataQuery, queryModel *models.QueryModel) backend.DataResponse {
	handler := &RESTAPIHandler{
		config: d.config,
		logger: d.logger,
	}

	if queryModel.RESTEndpoint == "" {
		return backend.DataResponse{
			Error: fmt.Errorf("REST endpoint is required"),
		}
	}

	return handler.executeQuery(ctx, query, queryModel)
}

// executeQuery executes a REST API query
func (h *RESTAPIHandler) executeQuery(ctx context.Context, query backend.DataQuery, queryModel *models.QueryModel) backend.DataResponse {
	// Build full URL
	baseURL := h.config.RESTURL
	if baseURL == "" {
		return backend.DataResponse{
			Error: fmt.Errorf("REST API base URL not configured"),
		}
	}

	// Ensure base URL doesn't end with /
	baseURL = strings.TrimSuffix(baseURL, "/")
	endpoint := strings.TrimPrefix(queryModel.RESTEndpoint, "/")
	fullURL := baseURL + "/" + endpoint

	// Determine HTTP method
	method := strings.ToUpper(queryModel.RESTMethod)
	if method == "" {
		method = "GET"
	}

	// Create request body if provided
	var bodyReader io.Reader
	if queryModel.RESTBody != "" && (method == "POST" || method == "PUT" || method == "PATCH") {
		bodyReader = bytes.NewBufferString(queryModel.RESTBody)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return backend.DataResponse{
			Error: fmt.Errorf("failed to create request: %w", err),
		}
	}

	// Add headers
	if queryModel.RESTHeaders != nil {
		for k, v := range queryModel.RESTHeaders {
			req.Header.Set(k, v)
		}
	}

	// Add default headers if not present
	if req.Header.Get("Content-Type") == "" && bodyReader != nil {
		req.Header.Set("Content-Type", "application/json")
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

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return backend.DataResponse{
			Error: fmt.Errorf("REST API returned status %d: %s", resp.StatusCode, string(body)),
		}
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return backend.DataResponse{
			Error: fmt.Errorf("failed to read response: %w", err),
		}
	}

	// Parse JSON response
	var jsonData interface{}
	if err := json.Unmarshal(body, &jsonData); err != nil {
		return backend.DataResponse{
			Error: fmt.Errorf("failed to parse JSON response: %w", err),
		}
	}

	// Convert to Grafana data frames
	frames, err := h.convertToDataFrames(jsonData, query)
	if err != nil {
		return backend.DataResponse{
			Error: fmt.Errorf("failed to convert response: %w", err),
		}
	}

	return backend.DataResponse{
		Frames: frames,
	}
}

// convertToDataFrames converts REST API JSON response to Grafana data frames
func (h *RESTAPIHandler) convertToDataFrames(jsonData interface{}, query backend.DataQuery) (data.Frames, error) {
	var frames data.Frames

	// Handle different JSON structures
	switch v := jsonData.(type) {
	case []interface{}:
		// Array of objects - treat as time series or table
		frame, err := h.arrayToDataFrame(v, query)
		if err != nil {
			return nil, err
		}
		frames = append(frames, frame)

	case map[string]interface{}:
		// Object - try to extract time series data
		frame, err := h.objectToDataFrame(v, query)
		if err != nil {
			return nil, err
		}
		frames = append(frames, frame)

	default:
		// Primitive value - create simple frame
		frame := h.primitiveToDataFrame(jsonData, query)
		frames = append(frames, frame)
	}

	return frames, nil
}

// arrayToDataFrame converts an array of objects to a data frame
func (h *RESTAPIHandler) arrayToDataFrame(arr []interface{}, query backend.DataQuery) (*data.Frame, error) {
	if len(arr) == 0 {
		return data.NewFrame("", data.NewField("value", nil, []interface{}{})), nil
	}

	// Check if first element has timestamp field
	_, ok := arr[0].(map[string]interface{})
	if !ok {
		// Not an array of objects, create simple array frame
		return data.NewFrame("", data.NewField("value", nil, arr)), nil
	}

	// Try to detect time series structure
	var timeField *data.Field
	var valueFields []*data.Field

	// Look for common timestamp fields
	var times []time.Time
	var hasTimeField bool

	for _, item := range arr {
		obj, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		// Try to find timestamp
		var timestamp time.Time
		for _, timeKey := range []string{"time", "timestamp", "date", "ts", "datetime"} {
			if tsVal, exists := obj[timeKey]; exists {
				timestamp = h.parseTimestamp(tsVal)
				hasTimeField = true
				break
			}
		}

		if !hasTimeField {
			// Use query time range if no timestamp found
			timestamp = query.TimeRange.From.Add(time.Duration(len(times)) * query.Interval)
		}

		times = append(times, timestamp)

		// Extract numeric values
		if len(valueFields) == 0 {
			// Initialize fields on first iteration
			for key, val := range obj {
				if key == "time" || key == "timestamp" || key == "date" || key == "ts" || key == "datetime" {
					continue
				}
				if h.isNumeric(val) {
					valueFields = append(valueFields, data.NewField(key, nil, []float64{}))
				}
			}
		}

		// Add values to fields
		fieldIdx := 0
		for key, val := range obj {
			if key == "time" || key == "timestamp" || key == "date" || key == "ts" || key == "datetime" {
				continue
			}
			if h.isNumeric(val) {
				if fieldIdx < len(valueFields) {
					valueFields[fieldIdx].Append(h.toFloat64(val))
					fieldIdx++
				}
			}
		}
	}

	if hasTimeField {
		timeField = data.NewField("time", nil, times)
		frame := data.NewFrame("", timeField)
		for _, f := range valueFields {
			frame.Fields = append(frame.Fields, f)
		}
		frame.Meta = &data.FrameMeta{
			Type: data.FrameTypeTimeSeriesMany,
		}
		return frame, nil
	}

	// No time field - create table frame
	frame := data.NewFrame("")
	for _, f := range valueFields {
		frame.Fields = append(frame.Fields, f)
	}
	return frame, nil
}

// objectToDataFrame converts an object to a data frame
func (h *RESTAPIHandler) objectToDataFrame(obj map[string]interface{}, query backend.DataQuery) (*data.Frame, error) {
	frame := data.NewFrame("")

	// Check if it's a time series object with data array
	if dataArr, ok := obj["data"].([]interface{}); ok {
		return h.arrayToDataFrame(dataArr, query)
	}

	// Otherwise, treat as single row table
	for key, val := range obj {
		field := data.NewField(key, nil, []interface{}{val})
		frame.Fields = append(frame.Fields, field)
	}

	return frame, nil
}

// primitiveToDataFrame creates a simple frame from a primitive value
func (h *RESTAPIHandler) primitiveToDataFrame(val interface{}, query backend.DataQuery) *data.Frame {
	now := time.Now()
	timeField := data.NewField("time", nil, []time.Time{now})
	valueField := data.NewField("value", nil, []interface{}{val})
	return data.NewFrame("", timeField, valueField)
}

// parseTimestamp attempts to parse various timestamp formats
func (h *RESTAPIHandler) parseTimestamp(val interface{}) time.Time {
	switch v := val.(type) {
	case string:
		// Try ISO 8601
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			return t
		}
		// Try Unix timestamp string
		if ts, err := strconv.ParseInt(v, 10, 64); err == nil {
			if ts > 1e12 {
				// Milliseconds
				return time.Unix(ts/1000, (ts%1000)*1e6)
			}
			return time.Unix(ts, 0)
		}
	case float64:
		// Unix timestamp (seconds or milliseconds)
		if v > 1e12 {
			// Milliseconds
			return time.Unix(int64(v)/1000, (int64(v)%1000)*1e6)
		}
		return time.Unix(int64(v), 0)
	case int64:
		if v > 1e12 {
			return time.Unix(v/1000, (v%1000)*1e6)
		}
		return time.Unix(v, 0)
	}
	return time.Now()
}

// isNumeric checks if a value is numeric
func (h *RESTAPIHandler) isNumeric(val interface{}) bool {
	switch val.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return true
	case string:
		// Try to parse as number
		_, err := strconv.ParseFloat(val.(string), 64)
		return err == nil
	}
	return false
}

// toFloat64 converts a value to float64
func (h *RESTAPIHandler) toFloat64(val interface{}) float64 {
	switch v := val.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return 0
}

// addAuthHeaders adds authentication headers to the request
func (h *RESTAPIHandler) addAuthHeaders(req *http.Request) {
	if h.config.BearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+h.config.BearerToken)
	} else if h.config.APIKey != "" {
		req.Header.Set("X-API-Key", h.config.APIKey)
	} else if h.config.BasicAuthUser != "" && h.config.BasicAuthPass != "" {
		req.SetBasicAuth(h.config.BasicAuthUser, h.config.BasicAuthPass)
	}
}

// handleRESTResource handles resource calls for REST API
func (d *Datasource) handleRESTResource(ctx context.Context, req *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	// Proxy the request to REST API
	client := &http.Client{Timeout: 30 * time.Second}

	// Build URL
	baseURL := d.config.RESTURL
	if baseURL == "" {
		return sender.Send(&backend.CallResourceResponse{
			Status: 400,
			Body:   []byte(`{"error": "REST API base URL not configured"}`),
		})
	}

	baseURL = strings.TrimSuffix(baseURL, "/")
	path := strings.TrimPrefix(req.Path, "/")
	targetURL := baseURL + "/" + path

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

