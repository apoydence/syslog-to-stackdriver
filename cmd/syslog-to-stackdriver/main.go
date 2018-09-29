package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"cloud.google.com/go/logging"
	envstruct "code.cloudfoundry.org/go-envstruct"
	"github.com/apoydence/syslog-to-stackdriver/pkg/conversion"
	"github.com/apoydence/syslog-to-stackdriver/pkg/web"
	"google.golang.org/appengine"
)

func main() {
	log := log.New(os.Stderr, "[SYSLOG-STACKDRIVER] ", log.LstdFlags)
	cfg := LoadConfig(log)

	log.Println("Starting syslog to stackdriver...")
	defer log.Println("Closing syslog to stackdriver...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	handler := buildHandler(ctx, cfg, log)

	if !cfg.AppEngine {
		if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), handler); err != nil {
			log.Fatal(err)
		}
	}

	http.Handle("/", handler)
	appengine.Main()
}

func buildHandler(ctx context.Context, cfg Config, log *log.Logger) http.Handler {
	if !cfg.AppEngine {
		client, err := logging.NewClient(ctx, cfg.ProjectID)
		if err != nil {
			log.Fatalf("Failed to create client: %v", err)
		}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := client.Logger(getLogID(cfg, r.URL.Path))
			web.NewDrain(conversion.Convert, logger).ServeHTTP(w, r)
		})
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := appengine.NewContext(r)
		client, err := logging.NewClient(ctx, appengine.AppID(ctx))
		if err != nil {
			log.Fatalf("Failed to create client: %v", err)
		}

		defer func() {
			if err := client.Close(); err != nil {
				log.Fatalf("Failed to close client: %v", err)
			}
		}()

		logger := client.Logger(getLogID(cfg, r.URL.Path))

		web.NewDrain(conversion.Convert, logger).ServeHTTP(w, r)
	})
}

func getLogID(cfg Config, path string) string {
	split := strings.Split(path, "/")
	if len(split) == 0 || split[len(split)-1] == "" {
		return cfg.LogID
	}

	return split[len(split)-1]
}

type Config struct {
	Port                   int    `env:"PORT, report"`
	ProjectID              string `env:"PROJECT_ID, report"`
	LogID                  string `env:"LOG_ID, report"`
	GoogleApplicationCreds string `env:"GOOGLE_APPLICATION_CREDENTIALS"`
	AppEngine              bool   `env:"APP_ENGINE"`
}

func LoadConfig(log *log.Logger) Config {
	cfg := Config{
		Port:  8080,
		LogID: "syslog",
	}

	if err := envstruct.Load(&cfg); err != nil {
		log.Fatal(err)
	}

	if cfg.ProjectID == "" && !cfg.AppEngine {
		log.Fatal("missing PROJECT_ID")
	}

	envstruct.WriteReport(&cfg)

	return cfg
}
