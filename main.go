package main

import (
	"os"

	"github.com/Sameersah/GrafanaConnect/pkg/plugin"
	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

func main() {
	log.DefaultLogger.Info("Starting GrafanaConnect datasource plugin")

	provider := plugin.NewInstanceProvider()
	im := instancemgmt.New(provider)

	handler := plugin.NewHandlerWrapper(im)

	if err := datasource.Serve(datasource.ServeOpts{
		QueryDataHandler:    handler,
		CheckHealthHandler:  handler,
		CallResourceHandler: handler,
	}); err != nil {
		log.DefaultLogger.Error("Error starting plugin", "error", err)
		os.Exit(1)
	}
}
