package tools

import (
	"encoding/json"
	"html"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kkjdaniel/gogeek/collection"
	"github.com/kkjdaniel/gogeek/forum"
	"github.com/kkjdaniel/gogeek/forumlist"
	"github.com/kkjdaniel/gogeek/hot"
	"github.com/kkjdaniel/gogeek/thing"
	"github.com/kkjdaniel/gogeek/thread"
	"github.com/kkjdaniel/gogeek/user"
)

// RegisterRESTHandlers attaches REST endpoints to provided mux with CORS and basic sanitization.
// Endpoints:
// GET /health
// GET /v1/bgg/search?query=...&limit=30&type=boardgame|all
// GET /v1/bgg/details/{id}
// GET /v1/bgg/hot
// GET /v1/bgg/user?username=
// GET /v1/bgg/collection?username=...&subtype=boardgame|boardgameexpansion&owned=true...
// GET /v1/bgg/price?ids=12,844&currency=USD&destination=US
// GET /v1/bgg/recommendations?name=Azul&id=&min_votes=30
// GET /v1/bgg/trade-finder?user1=...&user2=...
// GET /v1/bgg/rules?name=Azul&id=
// GET /v1/bgg/thread/{id}
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

	mux.HandleFunc("/v1/bgg/hot", func(w http.ResponseWriter, r *http.Request) {
		res, err := hot.Query(hot.ItemTypeBoardGame)
		if err != nil {
			writeJSON(w, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, res.Items)
	})

	mux.HandleFunc("/v1/bgg/user", func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimSpace(r.URL.Query().Get("username"))
		if strings.EqualFold(name, "SELF") || name == "" {
			if env := os.Getenv("BGG_USERNAME"); env != "" {
				name = env
			}
		}
		if name == "" {
			w.WriteHeader(http.StatusBadRequest)
			writeJSON(w, map[string]string{"error": "username required"})
			return
		}
		ud, err := user.Query(name)
		if err != nil {
			writeJSON(w, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, ud)
	})

	mux.HandleFunc("/v1/bgg/collection", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		name := strings.TrimSpace(q.Get("username"))
		if strings.EqualFold(name, "SELF") || name == "" {
			if env := os.Getenv("BGG_USERNAME"); env != "" {
				name = env
			}
		}
		if name == "" {
			w.WriteHeader(http.StatusBadRequest)
			writeJSON(w, map[string]string{"error": "username required"})
			return
		}
		// Reuse buildCollectionOptions by translating query params
		args := map[string]interface{}{}
		if v := q.Get("subtype"); v != "" { args["subtype"] = v }
		for _, key := range []string{"owned","wishlist","preordered","fortrade","rated","wanttoplay","played","wanttobuy","hasparts"} {
			if v := q.Get(key); v != "" { args[key] = (strings.ToLower(v) == "true" || v == "1") }
		}
		for _, key := range []string{"minrating","maxrating","minbggrating","maxbggrating"} {
			if v := q.Get(key); v != "" { if f, err := strconv.ParseFloat(v, 64); err == nil { args[key] = f } }
		}
		if v := q.Get("minplays"); v != "" { if f, err := strconv.ParseFloat(v, 64); err == nil { args["minplays"] = f } }
		if v := q.Get("maxplays"); v != "" { if f, err := strconv.ParseFloat(v, 64); err == nil { args["maxplays"] = f } }

		opts := buildCollectionOptions(args)
		res, err := collection.Query(name, opts...)
		if err != nil {
			writeJSON(w, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, res.Items)
	})

	mux.HandleFunc("/v1/bgg/price", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		ids := strings.TrimSpace(q.Get("ids"))
		if ids == "" { writeJSON(w, map[string]string{"error":"ids required"}); return }
		currency := strings.ToUpper(strings.TrimSpace(q.Get("currency")))
		if currency == "" { currency = "USD" }
		destination := strings.ToUpper(strings.TrimSpace(q.Get("destination")))
		if destination == "" { destination = "US" }
		params := url.Values{}
		params.Add("eid", ids)
		params.Add("currency", currency)
		params.Add("destination", destination)
		params.Add("sitename", "bgg-mcp")
		resp, err := http.Get("https://boardgameprices.co.uk/api/info?" + params.Encode())
		if err != nil { writeJSON(w, map[string]string{"error": err.Error()}); return }
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		var out any
		if err := json.Unmarshal(body, &out); err != nil { writeJSON(w, map[string]string{"error": err.Error()}); return }
		writeJSON(w, out)
	})

	mux.HandleFunc("/v1/bgg/recommendations", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		name := strings.TrimSpace(q.Get("name"))
		idStr := strings.TrimSpace(q.Get("id"))
		minVotes := 30
		if v := q.Get("min_votes"); v != "" { if n, err := strconv.Atoi(v); err == nil && n > 0 { minVotes = n } }
		var gameID int
		if idStr != "" {
			if n, err := strconv.Atoi(idStr); err == nil { gameID = n }
		}
		if gameID == 0 && name != "" {
			items, err := searchAndSortGames(name, "boardgame", 1)
			if err == nil && len(items.Items) > 0 { gameID = items.Items[0].ID }
		}
		if gameID == 0 { writeJSON(w, map[string]string{"error":"name or id required"}); return }
		recURL := "https://recommend.games/api/games/" + strconv.Itoa(gameID) + "/similar.json?num_votes__gte=" + strconv.Itoa(minVotes) + "&page=1"
		resp, err := http.Get(recURL)
		if err != nil { writeJSON(w, map[string]string{"error": err.Error()}); return }
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		var parsed struct{ Results []struct{ BGGID int `json:"bgg_id"` } `json:"results"` }
		_ = json.Unmarshal(b, &parsed)
		if len(parsed.Results) == 0 { writeJSON(w, []any{}); return }
		ids := make([]int, 0, len(parsed.Results))
		for i, g := range parsed.Results { if i >= 10 { break }; ids = append(ids, g.BGGID) }
		things, err := thing.Query(ids)
		if err != nil { writeJSON(w, map[string]string{"error": err.Error()}); return }
		writeJSON(w, extractEssentialInfoList(things.Items))
	})

	mux.HandleFunc("/v1/bgg/trade-finder", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		u1 := strings.TrimSpace(q.Get("user1"))
		u2 := strings.TrimSpace(q.Get("user2"))
		if u1 == "SELF" || u1 == "" { if env := os.Getenv("BGG_USERNAME"); env != "" { u1 = env } }
		if u2 == "SELF" { if env := os.Getenv("BGG_USERNAME"); env != "" { u2 = env } }
		if u1 == "" || u2 == "" { writeJSON(w, map[string]string{"error":"user1 and user2 required"}); return }
		u1Col, err := collection.Query(u1, collection.WithOwned(true))
		if err != nil { writeJSON(w, map[string]string{"error": err.Error()}); return }
		u2Wish, err := collection.Query(u2, collection.WithWishlist(true))
		if err != nil { writeJSON(w, map[string]string{"error": err.Error()}); return }
		writeJSON(w, analyseTradeOpportunities(u1, u2, u1Col, u2Wish))
	})

	mux.HandleFunc("/v1/bgg/rules", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		name := strings.TrimSpace(q.Get("name"))
		idStr := strings.TrimSpace(q.Get("id"))
		var gameID int
		var gameName string
		if idStr != "" {
			if n, err := strconv.Atoi(idStr); err == nil { gameID = n }
		}
		if gameID == 0 && name != "" {
			best, err := findBestGameMatch(name)
			if err == nil && best != nil { gameID = best.ID; gameName = best.Name.Value }
		}
		if gameID == 0 { writeJSON(w, map[string]string{"error":"name or id required"}); return }
		forums, err := forumlist.Query(gameID, forumlist.Thing)
		if err != nil { writeJSON(w, map[string]string{"error": err.Error()}); return }
		var rulesForumID int
		var rulesForumTitle string
		for _, f := range forums.Forums {
			titleLower := strings.ToLower(f.Title)
			if strings.Contains(titleLower, "rules") { rulesForumID = f.ID; rulesForumTitle = f.Title; break }
		}
		if rulesForumID == 0 { writeJSON(w, map[string]string{"error":"no rules forum found"}); return }
		threads := []map[string]any{}
		page := 1
		for page <= 3 { // cap pages for REST
			fd, err := forum.Query(rulesForumID, forum.WithPage(page))
			if err != nil { break }
			for _, th := range fd.Threads {
				threads = append(threads, map[string]any{"id": th.ID, "subject": th.Subject, "replies": th.NumArticles - 1, "link": "https://boardgamegeek.com/thread/" + strconv.Itoa(th.ID) })
			}
			if len(fd.Threads) < 50 { break }
			page++
		}
		writeJSON(w, map[string]any{
			"game_name": gameName,
			"game_id": gameID,
			"forum_title": rulesForumTitle,
			"threads": threads,
		})
	})

	mux.HandleFunc("/v1/bgg/thread/", func(w http.ResponseWriter, r *http.Request) {
		idPart := strings.TrimPrefix(r.URL.Path, "/v1/bgg/thread/")
		id, err := strconv.Atoi(idPart)
		if err != nil { writeJSON(w, map[string]string{"error":"invalid id"}); return }
		td, err := thread.Query(id)
		if err != nil { writeJSON(w, map[string]string{"error": err.Error()}); return }
		writeJSON(w, td)
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
