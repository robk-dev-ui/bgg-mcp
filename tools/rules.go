package tools

import (
	"context"
	"fmt"
	"html"
	"strconv"
	"strings"

	"github.com/kkjdaniel/gogeek/forum"
	"github.com/kkjdaniel/gogeek/forumlist"
	"github.com/kkjdaniel/gogeek/search"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type RulesForumResult struct {
	GameName     string         `json:"game_name"`
	GameID       int            `json:"game_id"`
	ForumID      int            `json:"forum_id"`
	ForumTitle   string         `json:"forum_title"`
	TotalThreads int            `json:"total_threads"`
	Threads      []forum.Thread `json:"threads"`
}

func RulesTool() (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("bgg-rules",
		mcp.WithDescription("Find rules clarifications and discussions for a board game from BoardGameGeek forums. Returns thread summaries that can be explored further with bgg-thread-details."),
		mcp.WithString("name",
			mcp.Description("The name of the board game"),
		),
		mcp.WithNumber("id",
			mcp.Description("The BoardGameGeek ID of the board game"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of threads to return (default: 100, max: 200)"),
		),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		arguments := request.GetArguments()

		var gameID int
		var gameName string
		var err error

		if idVal, ok := arguments["id"]; ok && idVal != nil {
			switch v := idVal.(type) {
			case float64:
				gameID = int(v)
			case string:
				gameID, err = strconv.Atoi(v)
				if err != nil {
					return mcp.NewToolResultText("Invalid game ID format"), nil
				}
			}
		} else if nameVal, ok := arguments["name"]; ok && nameVal != nil {
			gameName = nameVal.(string)
			searchResults, err := search.Query(gameName)
			if err != nil {
				return mcp.NewToolResultText(fmt.Sprintf("Failed to search for game: %v", err)), nil
			}
			if len(searchResults.Items) == 0 {
				return mcp.NewToolResultText(fmt.Sprintf("No games found matching '%s'", gameName)), nil
			}
			gameID = searchResults.Items[0].ID
			gameName = searchResults.Items[0].Name.Value
		} else {
			return mcp.NewToolResultText("Either 'name' or 'id' parameter is required"), nil
		}

		limit := 100
		if l, ok := arguments["limit"].(float64); ok {
			limit = int(l)
			if limit > 200 {
				limit = 200
			}
		}
		forums, err := forumlist.Query(gameID, forumlist.Thing)
		if err != nil {
			return mcp.NewToolResultText(fmt.Sprintf("Failed to get forum list: %v", err)), nil
		}
		var rulesForumID int
		var rulesForumTitle string
		for _, f := range forums.Forums {
			titleLower := strings.ToLower(f.Title)
			if strings.Contains(titleLower, "rules") {
				rulesForumID = f.ID
				rulesForumTitle = f.Title
				break
			}
		}

		if rulesForumID == 0 {
			return mcp.NewToolResultText(fmt.Sprintf("No rules forum found for game ID %d", gameID)), nil
		}
		result := RulesForumResult{
			GameName:   gameName,
			GameID:     gameID,
			ForumID:    rulesForumID,
			ForumTitle: rulesForumTitle,
			Threads:    []forum.Thread{},
		}

		allThreads := []forum.Thread{}
		page := 1
		maxPages := (limit / 50) + 1
		if maxPages > 4 {
			maxPages = 4
		}

		for page <= maxPages {
			var rulesForumData *forum.Forum
			var err error

			if page == 1 {
				rulesForumData, err = forum.Query(rulesForumID)
			} else {
				rulesForumData, err = forum.Query(rulesForumID, forum.WithPage(page))
			}

			if err != nil {
				return mcp.NewToolResultText(fmt.Sprintf("Failed to get rules forum threads: %v", err)), nil
			}

			if page == 1 {
				result.TotalThreads = rulesForumData.NumThreads
			}

			if len(rulesForumData.Threads) == 0 {
				break
			}

			allThreads = append(allThreads, rulesForumData.Threads...)

			if len(allThreads) >= limit || len(rulesForumData.Threads) < 50 {
				break
			}

			page++
		}

		if len(allThreads) > limit {
			result.Threads = allThreads[:limit]
		} else {
			result.Threads = allThreads
		}

		// Return structured XML response with instructions for the AI
		var response strings.Builder
		response.WriteString("<rules_forum_analysis>\n")
		response.WriteString("<instructions>\n")
		response.WriteString("Your goal is to help the user resolve their rules question or understand game mechanics.\n")
		response.WriteString("1. First, identify threads that directly address the user's specific rules query based on their titles\n")
		response.WriteString("2. Look for threads with high reply counts (indicating thorough discussions) or official-sounding titles\n")
		response.WriteString("3. Present the 1-4 most relevant threads with brief descriptions of what the titles suggest they discuss\n")
		response.WriteString("4. For the most promising thread(s), proactively use bgg-thread-details to fetch the actual content\n")
		response.WriteString("5. After reading the thread content, provide a clear answer to the user's rules question\n")
		response.WriteString("Remember: You're seeing thread titles only. Use bgg-thread-details to get actual answers.\n")
		response.WriteString("</instructions>\n\n")

		response.WriteString("<game_context>\n")
		response.WriteString(fmt.Sprintf("  <game_name>%s</game_name>\n", html.EscapeString(result.GameName)))
		response.WriteString(fmt.Sprintf("  <game_id>%d</game_id>\n", result.GameID))
		response.WriteString(fmt.Sprintf("  <forum_title>%s</forum_title>\n", html.EscapeString(result.ForumTitle)))
		response.WriteString(fmt.Sprintf("  <total_threads>%d</total_threads>\n", result.TotalThreads))
		response.WriteString(fmt.Sprintf("  <threads_retrieved>%d</threads_retrieved>\n", len(result.Threads)))
		response.WriteString("</game_context>\n\n")

		response.WriteString("<threads>\n")
		response.WriteString("<!-- Threads are sorted by most recent activity. High reply counts often indicate thorough rules discussions. -->\n")
		for _, thread := range result.Threads {
			response.WriteString("  <thread>\n")
			response.WriteString(fmt.Sprintf("    <id>%d</id>\n", thread.ID))
			response.WriteString(fmt.Sprintf("    <subject>%s</subject>\n", html.EscapeString(thread.Subject)))
			response.WriteString(fmt.Sprintf("    <author>%s</author>\n", html.EscapeString(thread.Author)))
			response.WriteString(fmt.Sprintf("    <replies>%d</replies>\n", thread.NumArticles-1))
			response.WriteString(fmt.Sprintf("    <posted>%s</posted>\n", thread.PostDate))
			response.WriteString(fmt.Sprintf("    <last_post>%s</last_post>\n", thread.LastPostDate))
			response.WriteString(fmt.Sprintf("    <link>https://boardgamegeek.com/thread/%d</link>\n", thread.ID))
			response.WriteString("  </thread>\n")
		}
		response.WriteString("</threads>\n")
		response.WriteString("</rules_forum_analysis>\n")

		return mcp.NewToolResultText(response.String()), nil
	}

	return tool, handler
}
