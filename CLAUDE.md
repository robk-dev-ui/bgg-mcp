# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

This is a BoardGameGeek MCP (Model Context Protocol) server written in Go that provides access to BGG API data through standardized MCP tools.

## Build Commands

- `make build` - Build the binary to `build/bgg-mcp`
- `make clean` - Remove build artifacts
- `make all` - Clean and build
- `go build -o build/bgg-mcp` - Direct Go build

## Architecture

The codebase implements an MCP server with 4 tools for accessing BoardGameGeek data:

1. **bgg-details** - Get detailed information about a board game by name
2. **bgg-collection** - Query a user's game collection with extensive filtering
3. **bgg-hot** - Get the current BGG hotness list
4. **bgg-user** - Get user profile information

Each tool is implemented in its own file under the `tools/` directory and registered in `main.go`.

## Key Dependencies

- `github.com/kkjdaniel/gogeek` - BoardGameGeek API client library
- `github.com/mark3labs/mcp-go` - MCP protocol implementation

## Development Notes

- The server runs via stdio and is configured as an MCP server in client applications
- All BGG API interactions are handled through the gogeek library
- Tool responses are JSON-marshaled and returned as text
- Error handling follows a consistent pattern of returning error messages as text results
- No external configuration files are needed - the server is self-contained