package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"time"

	"github.com/TV4/graceful"
	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const serviceName = "http-proxy"

var (
	buildTag = "dev"

	httpPort    = flag.Int("port", 8080, "HTTP port to listen")
	forwardPort = flag.Int("forward_port", 8080, "HTTP port to forward")

	debug   = flag.Bool("debug", false, "enable debug mode")
	console = flag.Bool("console_log", false, "change logs format from JSON to console output")
)

func main() {
	flag.Parse()

	logger := initLogger(*console, *debug, serviceName, buildTag)

	r := chi.NewRouter()
	r.HandleFunc("/health", healthHandler)
	r.HandleFunc("/v{version}/{service}/*", proxyHandler(logger))

	graceful.LogListenAndServe(&http.Server{
		Addr:    fmt.Sprintf(":%d", *httpPort),
		Handler: r,
	}, httpLoggerWrapper(logger))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func proxyHandler(log zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		service := chi.URLParam(r, "service")
		forwardURL := r.URL
		forwardURL.Scheme = "http"
		forwardURL.Host = fmt.Sprintf("%s:%d", service, *forwardPort)
		log.Debug().Interface("target", forwardURL).Interface("request", r.URL).Msg("proxy request")
		httputil.NewSingleHostReverseProxy(forwardURL).ServeHTTP(w, r)
	}
}

func initLogger(console, debug bool, serviceName, buildTag string) (logger zerolog.Logger) {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	if console {
		output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
		output.FormatLevel = func(i interface{}) string {
			return strings.ToUpper(fmt.Sprintf("| %-6s|", i))
		}
		logger = zerolog.New(output).With().Timestamp().Logger()
	} else {
		logger = log.Logger
	}

	logger = logger.With().Str("service", serviceName).Str("build", buildTag).Logger()
	logger.Debug().Msg("debug mode enabled")

	return logger
}

type loggerWrapper struct {
	log zerolog.Logger
}

func (l loggerWrapper) Printf(format string, v ...interface{}) {
	l.log.Info().Msgf(format, v...)
}

func (l loggerWrapper) Fatal(v ...interface{}) {
	for _, err := range v {
		l.log.Fatal().Interface("error", err)
	}
}

func httpLoggerWrapper(log zerolog.Logger) loggerWrapper {
	return loggerWrapper{log: log}
}
