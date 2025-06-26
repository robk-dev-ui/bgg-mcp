package tools

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/kkjdaniel/gogeek/search"
	"github.com/kkjdaniel/gogeek/thing"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func DetailsTool() (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("bgg-details",
		mcp.WithDescription("Find the details about a specific board game on BoardGameGeek (BGG)"),
		mcp.WithString("name",
			mcp.Description("The name of the board game"),
		),
		mcp.WithNumber("id",
			mcp.Description("The BoardGameGeek ID of the board game"),
		),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		arguments := request.GetArguments()
		
		var gameID int
		var err error
		
		// Check if ID is provided
		if idVal, ok := arguments["id"]; ok && idVal != nil {
			// Handle both float64 and string types
			switch v := idVal.(type) {
			case float64:
				gameID = int(v)
			case string:
				gameID, err = strconv.Atoi(v)
				if err != nil {
					return mcp.NewToolResultText("Invalid ID format"), nil
				}
			default:
				return mcp.NewToolResultText("Invalid ID type"), nil
			}
		} else if nameVal, ok := arguments["name"]; ok && nameVal != nil {
			// Fall back to name-based search
			name := nameVal.(string)
			result, err := search.Query(name, true)
			if err != nil || len(result.Items) == 0 {
				return mcp.NewToolResultText("No search result found"), nil
			}
			gameID = result.Items[0].ID
		} else {
			return mcp.NewToolResultText("Either 'name' or 'id' parameter must be provided"), nil
		}

		things, err := thing.Query([]int{gameID})
		if err != nil {
			return mcp.NewToolResultText(err.Error()), nil
		}

		if len(things.Items) > 0 {
			thing := things.Items[0]
			out, _ := json.Marshal(thing)
			return mcp.NewToolResultText(string(out)), nil
		}

		return mcp.NewToolResultText("No query results found"), nil
	}

	return tool, handler
}
