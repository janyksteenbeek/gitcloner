package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/janyksteenbeek/gitcloner/pkg/mirror"
	"github.com/janyksteenbeek/gitcloner/pkg/webhook"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Parse command line flags
	importRepos := flag.String("import", "", "Platform and repository to import (e.g., 'github username/repo')")
	flag.Parse()

	config := mirror.Config{
		Type:        os.Getenv("DESTINATION_TYPE"),
		URL:         os.Getenv("DESTINATION_URL"),
		Token:       os.Getenv("DESTINATION_TOKEN"),
		OrgID:       os.Getenv("DESTINATION_ORG"),
		SourceToken: os.Getenv("SOURCE_TOKEN"),
	}

	// Handle one-time imports if specified
	if *importRepos != "" {
		if err := mirror.HandleImport(config, *importRepos); err != nil {
			log.Fatalf("Failed to import repository: %v", err)
		}
		return
	}

	// Start webhook server
	log.Printf("type: %s", config.Type)
	log.Printf("url: %s", config.URL)
	log.Printf("orgID: %s", config.OrgID)

	handler := webhook.NewHandler(config)
	http.HandleFunc("/webhook", handler.HandleWebhook)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
