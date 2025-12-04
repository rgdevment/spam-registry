package http

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rgdevment/spam-registry/internal/service"
)

type Handler struct {
	service service.Service
}

func NewHandler(s service.Service) *Handler {
	return &Handler{
		service: s,
	}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/v1/reports", h.CreateReport)
	r.Get("/v1/phone/{number}", h.CheckRisk)
}

func (h *Handler) CreateReport(w http.ResponseWriter, r *http.Request) {
	var req CreateReportRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	reporterRaw := r.Header.Get("X-Reporter-ID")
	if reporterRaw == "" {
		reporterRaw = "anonymous"
	}

	err := h.service.IngestReport(
		r.Context(),
		req.PhoneNumber,
		reporterRaw,
		req.Category,
		req.Comment,
	)

	if err != nil {
		log.Printf("❌ ERROR IngestReport: %v", err)

		if err.Error() == "invalid phone format: ensure it includes country code (e.g. +569...)" ||
			err.Error() == "invalid phone number: number does not exist" ||
			err.Error() == "could not detect country from phone number" {

			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"status": "received"})
}

func (h *Handler) CheckRisk(w http.ResponseWriter, r *http.Request) {
	phoneNumber := chi.URLParam(r, "number")

	if len(phoneNumber) < 5 {
		http.Error(w, "Invalid phone number", http.StatusBadRequest)
		return
	}

	score, err := h.service.CheckRisk(r.Context(), phoneNumber)
	if err != nil {
		http.Error(w, "Error retrieval failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(score)
}
