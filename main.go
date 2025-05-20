package main

import (
	"fmt"

	"github.com/kkjdanie/bgg-mcp/tools"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	s := server.NewMCPServer(
		"BGG MCP",
		"1.0.0",
		server.WithResourceCapabilities(true, true),
		server.WithLogging(),
		server.WithRecovery(),
	)

	searchTool, searchHandler := tools.SearchTool()
	s.AddTool(searchTool, searchHandler)

	collectionTool, collectionHandler := tools.CollectionTool()
	s.AddTool(collectionTool, collectionHandler)

	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
