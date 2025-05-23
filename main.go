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

	detailsTool, detailsHandler := tools.DetailsTool()
	s.AddTool(detailsTool, detailsHandler)

	collectionTool, collectionHandler := tools.CollectionTool()
	s.AddTool(collectionTool, collectionHandler)

	hotnessTool, hotnessHandler := tools.HotnessTool()
	s.AddTool(hotnessTool, hotnessHandler)

	userTool, userHandler := tools.UserTool()
	s.AddTool(userTool, userHandler)

	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
