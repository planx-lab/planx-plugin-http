package main

import (
	"flag"
	"os"

	"github.com/planx-lab/planx-common/logger"
	"github.com/planx-lab/planx-plugin-http/internal/plugin"
	planxv1 "github.com/planx-lab/planx-proto/gen/go/planx/v1"
	"github.com/planx-lab/planx-sdk-go/server"
)

func main() {
	address := flag.String("address", ":50052", "gRPC server address")
	debug := flag.Bool("debug", false, "Enable debug logging")
	flag.Parse()

	// Initialize logger
	logLevel := "info"
	if *debug {
		logLevel = "debug"
	}
	logger.Init(logger.Config{
		Level:       logLevel,
		Pretty:      true,
		Output:      os.Stdout,
		ServiceName: "planx-plugin-http",
	})

	// Create server
	srv := server.New(server.Config{
		Address:          *address,
		PluginName:       "http",
		PluginType:       server.PluginTypeSink,
		EnableReflection: true,
	})

	// Register sink plugin
	sink := plugin.NewHTTPSink()
	planxv1.RegisterSinkPluginServer(srv.GRPCServer(), sink)

	logger.Info().Str("address", *address).Msg("Starting HTTP sink plugin")

	// Run server
	if err := srv.RunWithSignals(); err != nil {
		logger.Fatal().Err(err).Msg("Server error")
	}
}
