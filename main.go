package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kkjdanie/bgg-mcp/prompts"
	"github.com/kkjdanie/bgg-mcp/tools"
	"github.com/mark3labs/mcp-go/server"
)

func createMCPServer() *server.MCPServer {
	s := server.NewMCPServer(
		"BGG MCP",
		"1.2.0",
		server.WithResourceCapabilities(true, true),
		server.WithPromptCapabilities(true),
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

	searchTool, searchHandler := tools.SearchTool()
	s.AddTool(searchTool, searchHandler)

	priceTool, priceHandler := tools.PriceTool()
	s.AddTool(priceTool, priceHandler)

	tradeFinderTool, tradeFinderHandler := tools.TradeFinderTool()
	s.AddTool(tradeFinderTool, tradeFinderHandler)

	recommenderTool, recommenderHandler := tools.RecommenderTool()
	s.AddTool(recommenderTool, recommenderHandler)

	prompts.RegisterPrompts(s)

	return s
}

func main() {
	var mode string
	var port string
	
	flag.StringVar(&mode, "mode", "stdio", "Server mode: stdio or http")
	flag.StringVar(&port, "port", "8080", "Port for HTTP server (only used in http mode)")
	flag.Parse()

	// Override mode from environment variable if set
	if envMode := os.Getenv("MCP_MODE"); envMode != "" {
		mode = envMode
	}
	
	// Override port from environment variable if set
	if envPort := os.Getenv("MCP_PORT"); envPort != "" {
		port = envPort
	}

	mcpServer := createMCPServer()

	switch mode {
	case "http":
		runHTTPServer(mcpServer, port)
	case "stdio":
		runStdioServer(mcpServer)
	default:
		log.Fatalf("Invalid mode: %s. Use 'stdio' or 'http'", mode)
	}
}

func runStdioServer(mcpServer *server.MCPServer) {
	if err := server.ServeStdio(mcpServer); err != nil {
		log.Fatalf("STDIO server error: %v", err)
	}
}

func runHTTPServer(mcpServer *server.MCPServer, port string) {
	// Get base URL from environment or use default
	baseURL := os.Getenv("MCP_BASE_URL")
	if baseURL == "" {
		// Default to localhost for local development
		baseURL = fmt.Sprintf("http://localhost:%s", port)
	}
	
	// Create SSE server for HTTP transport
	sseServer := server.NewSSEServer(mcpServer,
		server.WithBaseURL(baseURL),
		server.WithStaticBasePath("/mcp"),
		server.WithKeepAlive(true),
		server.WithKeepAliveInterval(30*time.Second),
	)

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down HTTP server...")
		
		// Give the server time to shutdown gracefully
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		
		if err := sseServer.Shutdown(shutdownCtx); err != nil {
			log.Printf("Error during shutdown: %v", err)
		}
	}()

	// Start the HTTP server
	log.Printf("Starting HTTP server on port %s", port)
	log.Printf("SSE endpoint: %s/mcp/sse", baseURL)
	log.Printf("Message endpoint: %s/mcp/message", baseURL)
	
	if err := sseServer.Start(":" + port); err != nil {
		log.Fatalf("HTTP server error: %v", err)
	}
}
