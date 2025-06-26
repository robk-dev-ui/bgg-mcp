package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kkjdaniel/gogeek/search"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func SearchTool() (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("bgg-search",
		mcp.WithDescription("Search for board games on BoardGameGeek (BGG)"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("The search query for board games"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of results to return (default: 30)"),
		),
		mcp.WithString("type",
			mcp.Description("Filter by type (default: all, options: 'boardgame' (aka base game), 'boardgameexpansion', or 'all')"),
			mcp.Enum("all", "boardgame", "boardgameexpansion"),
		),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		arguments := request.GetArguments()
		query := arguments["query"].(string)

		limit := 30
		if l, ok := arguments["limit"].(float64); ok {
			limit = int(l)
		}

		typeFilter := "all"
		if t, ok := arguments["type"].(string); ok {
			typeFilter = t
		}

		result, err := search.Query(query, false)
		if err != nil {
			return mcp.NewToolResultText(fmt.Sprintf("Search error: %v", err)), nil
		}

		if len(result.Items) == 0 {
			return mcp.NewToolResultText("No search results found"), nil
		}

		var filteredItems []search.SearchResult
		if typeFilter == "all" {
			filteredItems = result.Items
		} else {
			for _, item := range result.Items {
				if item.Type == typeFilter {
					filteredItems = append(filteredItems, item)
				}
			}
		}

		if len(filteredItems) == 0 {
			return mcp.NewToolResultText(fmt.Sprintf("No %s results found", typeFilter)), nil
		}

		if len(filteredItems) > limit {
			filteredItems = filteredItems[:limit]
		}

		out, err := json.Marshal(filteredItems)
		if err != nil {
			return mcp.NewToolResultText(fmt.Sprintf("JSON encoding error: %v", err)), nil
		}

		return mcp.NewToolResultText(string(out)), nil
	}

	return tool, handler
}
