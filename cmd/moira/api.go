package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"time"

	"fmt"
	"github.com/moira-alert/moira-alert/api"
	"github.com/moira-alert/moira-alert/api/handler"
	"github.com/moira-alert/moira-alert/database/redis"
	"github.com/moira-alert/moira-alert/logging/go-logging"
)

// APIServer is a HTTP server for Moira API
type APIServer struct {
	Config         *api.Config
	DatabaseConfig *redis.Config

	LogFile  string
	LogLevel string

	http *http.Server
}

// Start Moira API HTTP server
func (apiService *APIServer) Start() error {
	logger, err := logging.ConfigureLog(apiService.LogFile, apiService.LogLevel, "api")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't configure logger for API: %v\n", err)
		os.Exit(1)
	}

	if !apiService.Config.Enabled {
		logger.Info("Moira Api Disabled")
		return nil
	}

	dataBase := redis.NewDatabase(logger, *apiService.DatabaseConfig)
	listener, err := net.Listen("tcp", apiService.Config.Listen)
	if err != nil {
		return err
	}

	httpHandler := handler.NewHandler(dataBase, logger)
	apiService.http = &http.Server{
		Handler: httpHandler,
	}

	go func() {
		apiService.http.Serve(listener)
	}()

	logger.Info("Moira Api Started")
	return nil
}

// Stop Moira API HTTP server
func (apiService *APIServer) Stop() error {
	if !apiService.Config.Enabled {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return apiService.http.Shutdown(ctx)
}
