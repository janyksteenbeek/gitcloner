package main

import (
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

	config := mirror.Config{
		Type:        os.Getenv("DESTINATION_TYPE"),
		URL:         os.Getenv("DESTINATION_URL"),
		Token:       os.Getenv("DESTINATION_TOKEN"),
		OrgID:       os.Getenv("DESTINATION_ORG"),
		SourceToken: os.Getenv("SOURCE_TOKEN"),
	}

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
