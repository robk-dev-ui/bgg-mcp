package tools

import (
	"context"
	"encoding/json"

	"github.com/kkjdaniel/gogeek/search"
	"github.com/kkjdaniel/gogeek/thing"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func SearchTool() (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("bgg-search",
		mcp.WithDescription("Find the details about a specific board game on BGG"),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("The name of the board game"),
		),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		arguments := request.GetArguments()
		name := arguments["name"].(string)

		result, err := search.Query(name, true)
		if err != nil || len(result.Items) == 0 {
			return mcp.NewToolResultText("No search result found"), nil
		}

		things, err := thing.Query([]int{result.Items[0].ID})
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
