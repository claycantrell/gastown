package web

import (
	"encoding/json"
	"html/template"
	"net/http"
)

// ConvoyFetcher defines the interface for fetching convoy data.
type ConvoyFetcher interface {
	FetchConvoys() ([]ConvoyRow, error)
	FetchMergeQueue() ([]MergeQueueRow, error)
	FetchPolecats() ([]PolecatRow, error)
}

// ConvoyHandler handles HTTP requests for the convoy dashboard.
type ConvoyHandler struct {
	fetcher  ConvoyFetcher
	template *template.Template
	mux      *http.ServeMux
}

// NewConvoyHandler creates a new convoy handler with the given fetcher.
func NewConvoyHandler(fetcher ConvoyFetcher) (*ConvoyHandler, error) {
	tmpl, err := LoadTemplates()
	if err != nil {
		return nil, err
	}

	h := &ConvoyHandler{
		fetcher:  fetcher,
		template: tmpl,
		mux:      http.NewServeMux(),
	}

	// Register routes
	h.mux.HandleFunc("/", h.handleDashboard)
	h.mux.HandleFunc("/map", h.handleMap)
	h.mux.HandleFunc("/api/polecat-positions", h.handlePolecatPositions)

	return h, nil
}

// ServeHTTP routes requests to appropriate handlers.
func (h *ConvoyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

// handleDashboard handles GET / requests and renders the convoy dashboard.
func (h *ConvoyHandler) handleDashboard(w http.ResponseWriter, r *http.Request) {
	convoys, err := h.fetcher.FetchConvoys()
	if err != nil {
		http.Error(w, "Failed to fetch convoys", http.StatusInternalServerError)
		return
	}

	mergeQueue, err := h.fetcher.FetchMergeQueue()
	if err != nil {
		// Non-fatal: show convoys even if merge queue fails
		mergeQueue = nil
	}

	polecats, err := h.fetcher.FetchPolecats()
	if err != nil {
		// Non-fatal: show convoys even if polecats fail
		polecats = nil
	}

	data := ConvoyData{
		Convoys:    convoys,
		MergeQueue: mergeQueue,
		Polecats:   polecats,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if err := h.template.ExecuteTemplate(w, "convoy.html", data); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}

// handleMap handles GET /map requests and renders the desert map view.
func (h *ConvoyHandler) handleMap(w http.ResponseWriter, r *http.Request) {
	polecats, err := h.fetcher.FetchPolecats()
	if err != nil {
		http.Error(w, "Failed to fetch polecats", http.StatusInternalServerError)
		return
	}

	data := MapData{
		Polecats: polecats,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if err := h.template.ExecuteTemplate(w, "map.html", data); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}

// handlePolecatPositions handles GET /api/polecat-positions and returns JSON.
func (h *ConvoyHandler) handlePolecatPositions(w http.ResponseWriter, r *http.Request) {
	polecats, err := h.fetcher.FetchPolecats()
	if err != nil {
		http.Error(w, "Failed to fetch polecats", http.StatusInternalServerError)
		return
	}

	// Return same structure as template data for consistency
	data := MapData{
		Polecats: polecats,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
