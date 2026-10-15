package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"example.com/go-wcf-learning/go-gateway/internal/httpapi"
	"example.com/go-wcf-learning/go-gateway/internal/soapclient"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	wcfURL := os.Getenv("WCF_SOAP_URL")
	if wcfURL == "" {
		wcfURL = "http://localhost:8081/CustomerService"
	}

	soap := soapclient.New(wcfURL, 8*time.Second)
	handler := httpapi.NewHandler(soap, logger)

	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	server := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	logger.Info("starting Go API", "address", addr, "wcf_endpoint", wcfURL)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
