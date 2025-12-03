package models

// QueryType represents the type of data source query
type QueryType string

const (
	QueryTypePrometheus QueryType = "prometheus"
	QueryTypeLoki       QueryType = "loki"
	QueryTypeREST       QueryType = "rest"
)

// DataSourceConfig holds the configuration for the data source
type DataSourceConfig struct {
	PrometheusURL string `json:"prometheusUrl"`
	LokiURL       string `json:"lokiUrl"`
	RESTURL       string `json:"restUrl"`
	
	// Authentication
	APIKey        string `json:"apiKey"`
	BasicAuthUser string `json:"basicAuthUser"`
	BasicAuthPass string `json:"basicAuthPass"`
	BearerToken   string `json:"bearerToken"`
	
	// REST API specific
	RESTHeaders map[string]string `json:"restHeaders"`
}

// QueryModel represents a query from Grafana
type QueryModel struct {
	QueryType QueryType `json:"queryType"`
	
	// Prometheus query fields
	PromQL string `json:"promQL,omitempty"`
	
	// Loki query fields
	LogQL string `json:"logQL,omitempty"`
	
	// REST API query fields
	RESTEndpoint string            `json:"restEndpoint,omitempty"`
	RESTMethod   string            `json:"restMethod,omitempty"`
	RESTHeaders  map[string]string `json:"restHeaders,omitempty"`
	RESTBody     string            `json:"restBody,omitempty"`
	
	// Common fields
	RefID string `json:"refId"`
}

// PrometheusQueryRequest represents a Prometheus query request
type PrometheusQueryRequest struct {
	Query     string `json:"query"`
	StartTime int64  `json:"start_time,omitempty"`
	EndTime   int64  `json:"end_time,omitempty"`
	Step      string `json:"step,omitempty"`
}

// PrometheusQueryResponse represents a Prometheus query response
type PrometheusQueryResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric map[string]string `json:"metric"`
			Values [][]interface{}    `json:"values,omitempty"`
			Value  []interface{}      `json:"value,omitempty"`
		} `json:"result"`
	} `json:"data"`
}

// LokiQueryRequest represents a Loki query request
type LokiQueryRequest struct {
	Query     string `json:"query"`
	Limit     int    `json:"limit,omitempty"`
	StartTime int64  `json:"start_time,omitempty"`
	EndTime   int64  `json:"end_time,omitempty"`
}

// LokiQueryResponse represents a Loki query response
type LokiQueryResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Stream map[string]string `json:"stream"`
			Values [][]string        `json:"values"`
		} `json:"result"`
	} `json:"data"`
}

