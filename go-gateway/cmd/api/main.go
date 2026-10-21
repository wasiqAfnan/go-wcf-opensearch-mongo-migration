package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"example.com/go-wcf-learning/go-gateway/internal/httpapi"
	"example.com/go-wcf-learning/go-gateway/internal/mongostore"
	"example.com/go-wcf-learning/go-gateway/internal/soapclient"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	wcfURL := os.Getenv("WCF_SOAP_URL")
	if wcfURL == "" {
		wcfURL = "http://localhost:8081"
	}

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://127.0.0.1:27017"
	}

	mongoDB := os.Getenv("MONGO_DB")
	if mongoDB == "" {
		mongoDB = "learningstack"
	}

	mongoCollection := os.Getenv("MONGO_COLLECTION")
	if mongoCollection == "" {
		mongoCollection = "customer_orders_migrated"
	}

	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	// 1. Initialize SOAP client
	soap := soapclient.New(wcfURL, 15*time.Second)

	// 2. Initialize MongoDB store
	mongoInitCtx, mongoCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer mongoCancel()

	store, err := mongostore.New(mongoInitCtx, mongoURI, mongoDB, mongoCollection)
	if err != nil {
		logger.Error("failed to connect to MongoDB", "uri", mongoURI, "error", err)
		os.Exit(1)
	}

	// 3. Initialize HTTP Handler
	handler := httpapi.NewHandler(soap, soap, store, logger)

	server := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// 4. Handle graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		logger.Info("starting Go API Gateway",
			"address", addr,
			"wcf_endpoint", wcfURL,
			"mongo_uri", mongoURI,
			"mongo_collection", mongoCollection,
		)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-stop
	logger.Info("shutting down Go API Gateway...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown error", "error", err)
	}

	if err := store.Close(shutdownCtx); err != nil {
		logger.Error("mongo disconnect error", "error", err)
	}

	logger.Info("server gracefully stopped")
}
