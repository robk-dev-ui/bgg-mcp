package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
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
		"1.4.0",
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

	rulesTool, rulesHandler := tools.RulesTool()
	s.AddTool(rulesTool, rulesHandler)

	threadDetailsTool, threadDetailsHandler := tools.ThreadDetailsTool()
	s.AddTool(threadDetailsTool, threadDetailsHandler)

	prompts.RegisterPrompts(s)

	return s
}

func main() {
	var mode string
	var port string
	
	flag.StringVar(&mode, "mode", "stdio", "Server mode: stdio or http")
	flag.StringVar(&port, "port", "8080", "Port for HTTP server (only used in http mode)")
	flag.Parse()

	if envMode := os.Getenv("MCP_MODE"); envMode != "" {
		mode = envMode
	}
	
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
	baseURL := os.Getenv("MCP_BASE_URL")
	if baseURL == "" {
		baseURL = fmt.Sprintf("http://localhost:%s", port)
	}

	mux := http.NewServeMux()

	// Register REST endpoints
	tools.RegisterRESTHandlers(mux)

	// MCP HTTP server (streamable) mounted under /mcp
	httpServer := server.NewStreamableHTTPServer(mcpServer,
		server.WithEndpointPath("/mcp"),
		server.WithStateLess(true),
		server.WithHeartbeatInterval(30*time.Second),
	)

	mux.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
		// Delegate to existing MCP server handler
		httpServer.ServeHTTP(w, r)
	})

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	serverInstance := &http.Server{Addr: ":" + port, Handler: mux}

	go func() {
		<-sigChan
		log.Println("Shutting down HTTP server...")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		if err := serverInstance.Shutdown(shutdownCtx); err != nil {
			log.Printf("Error during shutdown: %v", err)
		}
	}()

	log.Printf("Starting HTTP server on port %s", port)
	log.Printf("HTTP endpoint: %s/mcp", baseURL)
	log.Printf("REST endpoints: %s/health, %s/v1/bgg/search, %s/v1/bgg/details/{id}", baseURL, baseURL, baseURL)

	if err := serverInstance.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("HTTP server error: %v", err)
	}
}
