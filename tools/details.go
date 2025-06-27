package tools

import (
	"context"
	"encoding/json"
	"fmt"
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
		mcp.WithBoolean("full_details",
			mcp.Description("Return the complete BGG API response instead of essential info. WARNING: This returns significantly more data and can overload AI context windows. ONLY set this to true if the user explicitly requests 'full details', 'complete data', or similar. Default behavior returns essential info which is sufficient for most use cases."),
		),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		arguments := request.GetArguments()

		var gameID int
		var err error

		if idVal, ok := arguments["id"]; ok && idVal != nil {
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
			fullDetails := false
			if fd, ok := arguments["full_details"].(bool); ok {
				fullDetails = fd
			}

			var out []byte
			var err error
			
			if fullDetails {
				out, err = json.Marshal(things.Items[0])
			} else {
				essentialInfo := extractEssentialInfo(things.Items[0])
				out, err = json.Marshal(essentialInfo)
			}
			
			if err != nil {
				return mcp.NewToolResultText(fmt.Sprintf("Error formatting results: %v", err)), nil
			}
			return mcp.NewToolResultText(string(out)), nil
		}

		return mcp.NewToolResultText("No query results found"), nil
	}

	return tool, handler
}
