package tools

import (
	"encoding/json"
	"html"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/kkjdaniel/gogeek/thing"
)

// RegisterRESTHandlers attaches REST endpoints to provided mux with CORS and basic sanitization.
// Endpoints:
// GET /health
// GET /v1/bgg/search?query=...&limit=30&type=boardgame|all
// GET /v1/bgg/details/{id}
func RegisterRESTHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("/v1/bgg/search", func(w http.ResponseWriter, r *http.Request) {
		q := strings.TrimSpace(r.URL.Query().Get("query"))
		if len(q) < 3 {
			writeJSON(w, map[string]any{"games": []any{}, "total": 0, "warning": "query too short"})
			return
		}
		limit := 30
		if l := r.URL.Query().Get("limit"); l != "" {
			if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 100 {
				limit = n
			}
		}
		// type filter: default boardgame unless explicitly 'all'
		filterType := r.URL.Query().Get("type")
		if filterType == "" {
			filterType = "boardgame"
		}
		if filterType != "boardgame" && filterType != "all" { // guard invalid values
			filterType = "boardgame"
		}

		items, err := searchAndSortGames(q, filterType, limit)
		if err != nil {
			writeJSON(w, map[string]any{"games": []any{}, "total": 0, "error": err.Error()})
			return
		}
		es := extractEssentialInfoList(items.Items)
		writeJSON(w, map[string]any{"games": es, "total": len(es)})
	})

	mux.HandleFunc("/v1/bgg/details/", func(w http.ResponseWriter, r *http.Request) {
		idPart := strings.TrimPrefix(r.URL.Path, "/v1/bgg/details/")
		if idPart == "" {
			w.WriteHeader(http.StatusBadRequest)
			writeJSON(w, map[string]string{"error": "missing id"})
			return
		}
		id, err := strconv.Atoi(idPart)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			writeJSON(w, map[string]string{"error": "invalid id"})
			return
		}

		things, err := thing.Query([]int{id})
		if err != nil || len(things.Items) == 0 {
			w.WriteHeader(http.StatusNotFound)
			writeJSON(w, map[string]string{"error": "not found"})
			return
		}
		info := extractEssentialInfo(things.Items[0])
		// Clean description & add short form if present
		if info.Description != "" {
			clean := sanitizeDescription(info.Description)
			info.Description = clean
			if len(clean) > 400 {
				info.DescriptionShort = truncateWordSafe(clean, 400)
			} else {
				info.DescriptionShort = clean
			}
		}
		writeJSON(w, info)
	})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	_ = json.NewEncoder(w).Encode(v)
}

func sanitizeDescription(s string) string {
	// Decode HTML entities
	decoded := html.UnescapeString(s)
	// Replace common encoded newline artifacts
	decoded = strings.ReplaceAll(decoded, "\u0026#10;", "\n")
	decoded = strings.ReplaceAll(decoded, "&#10;", "\n")
	decoded = strings.ReplaceAll(decoded, "\r", "\n")
	// Collapse excessive whitespace
	lines := strings.Split(decoded, "\n")
	outLines := make([]string, 0, len(lines))
	for _, l := range lines {
		trim := strings.TrimSpace(l)
		if trim != "" {
			outLines = append(outLines, trim)
		}
	}
	res := strings.Join(outLines, "\n")
	return strings.TrimSpace(res)
}

func truncateWordSafe(s string, max int) string {
	if len(s) <= max {
		return s
	}
	cut := s[:max]
	// backtrack to last space for word safety
	if idx := strings.LastIndex(cut, " "); idx > 0 && idx > max-80 { // allow some backtracking
		cut = cut[:idx]
	}
	return strings.TrimSpace(cut) + "…"
}

// (Optional) simple ETag helper if needed later
func generateETag(id string) string {
	return "W/\"" + id + "-" + strconv.FormatInt(time.Now().Unix()/3600, 10) + "\"" // hourly weak ETag
}
