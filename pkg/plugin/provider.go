package plugin

import (
	"context"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
)

// InstanceProvider implements instancemgmt.InstanceProvider
type InstanceProvider struct{}

// NewInstanceProvider creates a new instance provider
func NewInstanceProvider() *InstanceProvider {
	return &InstanceProvider{}
}

// GetKey returns a key for the instance
func (p *InstanceProvider) GetKey(ctx context.Context, pluginContext backend.PluginContext) (interface{}, error) {
	return pluginContext.DataSourceInstanceSettings.UID, nil
}

// NewInstance creates a new instance
func (p *InstanceProvider) NewInstance(ctx context.Context, pluginContext backend.PluginContext) (instancemgmt.Instance, error) {
	return NewDatasource(ctx, *pluginContext.DataSourceInstanceSettings)
}

// NeedsUpdate checks if an instance needs to be updated
func (p *InstanceProvider) NeedsUpdate(ctx context.Context, pluginContext backend.PluginContext, cachedInstance instancemgmt.CachedInstance) bool {
	// Simple implementation: always return false to use cached instances
	// In production, you might want to compare settings to determine if update is needed
	return false
}

