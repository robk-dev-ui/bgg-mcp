package tools

import (
	"context"
	"encoding/json"

	"github.com/kkjdaniel/gogeek/collection"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func CollectionTool() (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("bgg-collection",
		mcp.WithDescription("Find the details about a specific users board game collection on BoardGameGeek (BGG)"),
		mcp.WithString("username",
			mcp.Required(),
			mcp.Description("The username of the BoardGameGeek (BGG) user who owns the collection"),
		),
		mcp.WithString("subtype",
			mcp.Enum("boardgame", "boardgameexpansion"),
			mcp.Description("Whether to search for base games or expansions"),
		),
		mcp.WithBoolean("owned",
			mcp.Description("Filters for owned games in the collection"),
		),
		mcp.WithBoolean("wishlist",
			mcp.Description("Filterds for wishlisted games in the collection"),
		),
		mcp.WithBoolean("preordered",
			mcp.Description("Filters for preordered games in the collection"),
		),
		mcp.WithBoolean("fortrade",
			mcp.Description("Filters for games that are marked for trade in the collection"),
		),
		mcp.WithBoolean("rated",
			mcp.Description("Filters for games that are rated in the collection"),
		),
		mcp.WithBoolean("wanttoplay",
			mcp.Description("Filters for games that the user wants to play in the collection"),
		),
		mcp.WithBoolean("played",
			mcp.Description("Filters for games that have recorded plays in the collection"),
		),
		mcp.WithBoolean("wanttobuy",
			mcp.Description("Filters for games that the user wants to buy in the collection"),
		),
		mcp.WithBoolean("hasparts",
			mcp.Description("Filters for games that have spare parts or not in the collection"),
		),
		mcp.WithNumber("minrating",
			mcp.Description("Filters based on the minimum personal rating of the games in the collection"),
		),
		mcp.WithNumber("maxrating",
			mcp.Description("Filters based on the maximum personal rating of the games in the collection"),
		),
		mcp.WithNumber("minbggrating",
			mcp.Description("Filters based on the minimum global BoardGameGeek (BGG) rating of the games in the collection"),
		),
		mcp.WithNumber("maxbggrating",
			mcp.Description("Filters based on the maximum global BoardGameGeek (BGG) rating of the games in the collection"),
		),
		mcp.WithNumber("minplays",
			mcp.Description("Filters based on the minimum number of plays of the games in the collection"),
		),
		mcp.WithNumber("maxplays",
			mcp.Description("Filters based on the maximum number of plays of the games in the collection"),
		),
	)

	var handler server.ToolHandlerFunc
	handler = func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		arguments := request.GetArguments()
		username := arguments["username"].(string)
		subtype := arguments["subtype"]

		owned := arguments["owned"]
		wishlist := arguments["wishlist"]
		preordered := arguments["preordered"]
		fortrade := arguments["fortrade"]
		rated := arguments["rated"]
		wanttoplay := arguments["wanttoplay"]
		played := arguments["played"]
		wanttobuy := arguments["wanttobuy"]
		hasparts := arguments["hasparts"]

		minrating := arguments["minrating"]
		maxrating := arguments["maxrating"]
		minbggrating := arguments["minbggrating"]
		maxbggrating := arguments["maxbggrating"]
		minplays := arguments["minplays"]
		maxplays := arguments["maxplays"]

		var options []collection.CollectionOption

		if subtype != nil {
			switch subtype.(string) {
			case "boardgame":
				options = append(options, collection.WithSubtype("boardgame"))
			case "boardgameexpansion":
				options = append(options, collection.WithSubtype("boardgameexpansion"))
			}
		}

		if owned != nil {
			options = append(options, collection.WithOwned(owned.(bool)))
		}
		if wishlist != nil {
			options = append(options, collection.WithWishlist(wishlist.(bool)))
		}
		if preordered != nil {
			options = append(options, collection.WithPreordered(preordered.(bool)))
		}
		if fortrade != nil {
			options = append(options, collection.WithTrade(fortrade.(bool)))
		}
		if rated != nil {
			options = append(options, collection.WithRated(rated.(bool)))
		}
		if wanttoplay != nil {
			options = append(options, collection.WithWantToPlay(wanttoplay.(bool)))
		}
		if played != nil {
			options = append(options, collection.WithPlayed(played.(bool)))
		}
		if wanttobuy != nil {
			options = append(options, collection.WithWantToBuy(wanttobuy.(bool)))
		}
		if hasparts != nil {
			options = append(options, collection.WithHasParts(hasparts.(bool)))
		}

		if minrating != nil {
			options = append(options, collection.WithMinRating(minrating.(float64)))
		}
		if maxrating != nil {
			options = append(options, collection.WithMaxRating(maxrating.(float64)))
		}
		if minbggrating != nil {
			options = append(options, collection.WithMinBGGRating(minbggrating.(float64)))
		}
		if maxbggrating != nil {
			options = append(options, collection.WithMaxBGGRating(maxbggrating.(float64)))
		}
		if minplays != nil {
			var minPlaysValue int
			if floatVal, ok := minplays.(float64); ok {
				minPlaysValue = int(floatVal)
				options = append(options, collection.WithMinPlays(minPlaysValue))
			}
		}
		if maxplays != nil {
			var maxPlaysValue int
			if floatVal, ok := maxplays.(float64); ok {
				maxPlaysValue = int(floatVal)
				options = append(options, collection.WithMaxPlays(maxPlaysValue))
			}
		}

		result, err := collection.Query(username, options...)

		if err != nil || len(result.Items) == 0 {
			return mcp.NewToolResultText("No search result found"), nil
		}

		items := result.Items
		out, _ := json.Marshal(items)
		return mcp.NewToolResultText(string(out)), nil
	}

	return tool, handler
}
