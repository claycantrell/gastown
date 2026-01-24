package web

import (
	"html/template"
	"net/http"
	"strings"
)

// ConvoyFetcher defines the interface for fetching convoy data.
type ConvoyFetcher interface {
	FetchConvoys() ([]ConvoyRow, error)
	FetchMergeQueue() ([]MergeQueueRow, error)
	FetchPolecats() ([]PolecatRow, error)
	FetchConvoyByID(id string) (*ConvoyRow, error)
	FetchPolecatBySessionID(sessionID string) (*PolecatRow, error)
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
	h.mux.HandleFunc("/map/building/", h.handleBuildingDetail)
	h.mux.HandleFunc("/map/polecat/", h.handlePolecatDetail)

	return h, nil
}

// ServeHTTP delegates to the internal mux.
func (h *ConvoyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

// handleDashboard handles GET / requests and renders the convoy dashboard.
func (h *ConvoyHandler) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

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

// handleMap handles GET /map requests and renders the 2D map view.
func (h *ConvoyHandler) handleMap(w http.ResponseWriter, r *http.Request) {
	convoys, err := h.fetcher.FetchConvoys()
	if err != nil {
		http.Error(w, "Failed to fetch convoys", http.StatusInternalServerError)
		return
	}

	polecats, err := h.fetcher.FetchPolecats()
	if err != nil {
		// Non-fatal: show map even if polecats fail
		polecats = nil
	}

	mapData := LayoutMap(convoys, polecats)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if err := h.template.ExecuteTemplate(w, "map.html", mapData); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}

// handleBuildingDetail handles GET /map/building/:id requests.
func (h *ConvoyHandler) handleBuildingDetail(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path
	id := strings.TrimPrefix(r.URL.Path, "/map/building/")
	if id == "" {
		http.Error(w, "Building ID required", http.StatusBadRequest)
		return
	}

	convoy, err := h.fetcher.FetchConvoyByID(id)
	if err != nil {
		http.Error(w, "Convoy not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if err := h.template.ExecuteTemplate(w, "building-detail.html", convoy); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}

// handlePolecatDetail handles GET /map/polecat/:sessionID requests.
func (h *ConvoyHandler) handlePolecatDetail(w http.ResponseWriter, r *http.Request) {
	// Extract session ID from path
	sessionID := strings.TrimPrefix(r.URL.Path, "/map/polecat/")
	if sessionID == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	polecat, err := h.fetcher.FetchPolecatBySessionID(sessionID)
	if err != nil {
		http.Error(w, "Polecat not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if err := h.template.ExecuteTemplate(w, "polecat-detail.html", polecat); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}
