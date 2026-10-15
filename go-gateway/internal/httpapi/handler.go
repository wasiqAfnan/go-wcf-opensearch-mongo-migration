package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"example.com/go-wcf-learning/go-gateway/internal/soapclient"
)

type customerGetter interface {
	CustomerByID(ctx context.Context, id int) (soapclient.Customer, error)
}

type Handler struct {
	customers customerGetter
	logger    *slog.Logger
}

func NewHandler(customers customerGetter, logger *slog.Logger) http.Handler {
	h := &Handler{customers: customers, logger: logger}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", h.health)
	mux.HandleFunc("GET /api/v1/customers/", h.getCustomer)
	return h.accessLog(mux)
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) getCustomer(w http.ResponseWriter, r *http.Request) {
	idText := strings.TrimPrefix(r.URL.Path, "/api/v1/customers/")
	if idText == "" || strings.Contains(idText, "/") {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "route not found"})
		return
	}

	id, err := strconv.Atoi(idText)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "customer ID must be a positive integer"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 9*time.Second)
	defer cancel()

	customer, err := h.customers.CustomerByID(ctx, id)
	if err != nil {
		if errors.Is(err, soapclient.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "customer not found"})
			return
		}
		h.logger.Error("WCF customer request failed", "customer_id", id, "error", err)
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "upstream customer service failed"})
		return
	}

	writeJSON(w, http.StatusOK, customer)
}

func (h *Handler) accessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		h.logger.Info("HTTP request", "method", r.Method, "path", r.URL.Path, "duration_ms", time.Since(start).Milliseconds())
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
