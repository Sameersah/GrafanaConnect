package main

import (
	"os"

	"github.com/Sameersah/GrafanaConnect/pkg/plugin"
	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

func main() {
	log.DefaultLogger.Info("Starting GrafanaConnect datasource plugin")

	if err := datasource.Serve(plugin.NewDatasource); err != nil {
		log.DefaultLogger.Error("Error starting plugin", "error", err)
		os.Exit(1)
	}
}

