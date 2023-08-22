//nolint:typecheck
package main

import (
	"context"
	"net/http"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"davensi.com/core/internal/util"
)

const (
	// crdbTimeout = 5 * time.Second
	httpTimeout = 10 * time.Second
)

// Run gRPC server
func main() {
	// Set default values
	viper.SetDefault("DEBUG", "false")
	viper.SetDefault("APP_ADDRESS_PORT", ":8080")
	viper.SetDefault("COCKROACHDB_MAX_CONN", "100")

	viper.AutomaticEnv()
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Info().Msg("config file not found, will only environment variables only")
		} else {
			log.Error().Err(err).Msg("error reading config file")
		}
	} else {
		log.Info().Msg("config file found")
	}
	debug := viper.GetBool("DEBUG")
	address := viper.GetString("APP_ADDRESS_PORT")

	// Configure logger
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	conn, err := util.PgxConn()
	if err != nil {
		log.Fatal().Err(err).Msg("Error connecting to CockroachDB")
	}

	if err := conn.Ping(context.Background()); err != nil {
		log.Panic().Err(err).Msg("Error connecting to CockroachDB")
	}

	defer conn.Close()

	mux := routes(conn)

	server := &http.Server{
		Addr:              address,
		ReadHeaderTimeout: httpTimeout,
		Handler:           h2c.NewHandler(mux, &http2.Server{}),
	}

	log.Info().Msg("Server started, listening on " + address)
	if err := server.ListenAndServe(); err != nil {
		log.Panic().Err(err).Msg("Unable to start server")
	}
}
