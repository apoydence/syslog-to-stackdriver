package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/logging"
	envstruct "code.cloudfoundry.org/go-envstruct"
	"github.com/apoydence/syslog-to-stackdriver/pkg/conversion"
	"github.com/apoydence/syslog-to-stackdriver/pkg/web"
)

func main() {
	log := log.New(os.Stderr, "[SYSLOG-STACKDRIVER] ", log.LstdFlags)
	cfg := LoadConfig(log)

	log.Println("Starting syslog to stackdriver...")
	defer log.Println("Closing syslog to stackdriver...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := logging.NewClient(ctx, cfg.ProjectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	logger := client.Logger(cfg.LogID)
	defer func() {
		if err := client.Close(); err != nil {
			log.Fatalf("Failed to close client: %v", err)
		}
	}()

	handler := web.NewDrain(conversion.Convert, logger)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), handler); err != nil {
		log.Fatal(err)
	}
}

type Config struct {
	Port                   int    `env:"PORT, report"`
	ProjectID              string `env:"PROJECT_ID, required, report"`
	LogID                  string `env:"LOG_ID, report"`
	GoogleApplicationCreds string `env:"GOOGLE_APPLICATION_CREDENTIALS, required"`
}

func LoadConfig(log *log.Logger) Config {
	cfg := Config{
		Port:  8080,
		LogID: "syslog",
	}

	if err := envstruct.Load(&cfg); err != nil {
		log.Fatal(err)
	}

	envstruct.WriteReport(&cfg)

	return cfg
}
