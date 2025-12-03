package plugin

import (
	"context"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
)

// HandlerWrapper wraps the instance manager to implement handler interfaces
type HandlerWrapper struct {
	im instancemgmt.InstanceManager
}

// NewHandlerWrapper creates a new handler wrapper
func NewHandlerWrapper(im instancemgmt.InstanceManager) *HandlerWrapper {
	return &HandlerWrapper{im: im}
}

// QueryData implements backend.QueryDataHandler
func (h *HandlerWrapper) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	instance, err := h.im.Get(ctx, req.PluginContext)
	if err != nil {
		return nil, err
	}
	ds := instance.(*Datasource)
	return ds.QueryData(ctx, req)
}

// CheckHealth implements backend.CheckHealthHandler
func (h *HandlerWrapper) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	instance, err := h.im.Get(ctx, req.PluginContext)
	if err != nil {
		return nil, err
	}
	ds := instance.(*Datasource)
	return ds.CheckHealth(ctx, req)
}

// CallResource implements backend.CallResourceHandler
func (h *HandlerWrapper) CallResource(ctx context.Context, req *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	instance, err := h.im.Get(ctx, req.PluginContext)
	if err != nil {
		return err
	}
	ds := instance.(*Datasource)
	return ds.CallResource(ctx, req, sender)
}

