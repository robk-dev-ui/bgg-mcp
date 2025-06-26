package tools

import (
	"context"
	"fmt"

	"github.com/fbiville/markdown-table-formatter/pkg/markdown"
	"github.com/kkjdaniel/gogeek/search"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func SearchTool() (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("bgg-search",
		mcp.WithDescription("Search for board games on BoardGameGeek (BGG) and return results in a markdown table"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("The search query for board games"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of results to return (default: 10)"),
		),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		arguments := request.GetArguments()
		query := arguments["query"].(string)

		limit := 40
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

		tableData := [][]string{}

		count := 0
		for _, item := range result.Items {
			if count >= limit {
				break
			}

			yearStr := ""
			if item.YearPublished.Value != 0 {
				yearStr = fmt.Sprintf("%d", item.YearPublished.Value)
			}

			row := []string{
				fmt.Sprintf("%d", item.ID),
				item.Name.Value,
				yearStr,
				item.Type,
			}
			tableData = append(tableData, row)
			count++
		}

		formatter := markdown.NewTableFormatterBuilder().
			WithPrettyPrint().
			Build("ID", "Name", "Year Published", "Type")

		table, err := formatter.Format(tableData)
		if err != nil {
			return mcp.NewToolResultText(fmt.Sprintf("Table formatting error: %v", err)), nil
		}

		return mcp.NewToolResultText(table), nil
	}

	return tool, handler
}
