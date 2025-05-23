package tools

import (
	"context"
	"encoding/json"

	"github.com/kkjdaniel/gogeek/user"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func UserTool() (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("bgg-user",
		mcp.WithDescription("Find details about a specific user on BoardGameGeek (BGG)"),
		mcp.WithString("username",
			mcp.Required(),
			mcp.Description("The username of the BoardGameGeek (BGG) user"),
		),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		arguments := request.GetArguments()
		name := arguments["username"].(string)

		userDetails, err := user.Query(name)
		if err != nil {
			return mcp.NewToolResultText(err.Error()), nil
		}

		out, _ := json.Marshal(userDetails)
		return mcp.NewToolResultText(string(out)), nil

	}

	return tool, handler
}
