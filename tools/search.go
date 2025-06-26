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
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		arguments := request.GetArguments()
		query := arguments["query"].(string)

		limit := 30
		if l, ok := arguments["limit"].(float64); ok {
			limit = int(l)
		}

		result, err := search.Query(query, false)
		if err != nil {
			return mcp.NewToolResultText(fmt.Sprintf("Search error: %v", err)), nil
		}

		if len(result.Items) == 0 {
			return mcp.NewToolResultText("No search results found"), nil
		}

		items := result.Items
		if len(items) > limit {
			items = items[:limit]
		}

		out, err := json.Marshal(items)
		if err != nil {
			return mcp.NewToolResultText(fmt.Sprintf("JSON encoding error: %v", err)), nil
		}

		return mcp.NewToolResultText(string(out)), nil
	}

	return tool, handler
}
