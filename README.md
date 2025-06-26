<p align="center">
  <img src="bgg-mcp-logo.png" width="250" alt="BGG MCP Logo">
</p>

# BGG MCP: BoardGameGeek MCP API Server

[![smithery badge](https://smithery.ai/badge/@kkjdaniel/bgg-mcp)](https://smithery.ai/server/@kkjdaniel/bgg-mcp)
[![Go Version](https://img.shields.io/github/go-mod/go-version/kkjdaniel/bgg-mcp)](https://go.dev/)
[![License](https://img.shields.io/github/license/kkjdaniel/bgg-mcp)](LICENSE)
[![MCP Protocol](https://img.shields.io/badge/MCP-Protocol-blue)](https://modelcontextprotocol.io)

BGG MCP provides access to the BoardGameGeek API through the [Model Context Protocol](https://www.anthropic.com/news/model-context-protocol), enabling retrieval and filtering of board game data, user collections, and profiles. The server is implemented in Go, using the [GoGeek](https://github.com/kkjdaniel/gogeek) library, which helps ensure robust API interactions.

<a href="https://boardgamegeek.com/">
  <img src="powered-bgg.webp" width="160" alt="Powered by BGG">
</a>

## Example

![Example of BGG MCP in action](example.png)

## Tools

- **Search** - Search for board games on BoardGameGeek
- **Game Details** - Get detailed information about a specific board game
- **Collection** - Query and filter a user's game collection
- **Hotness** - Get the current BGG hotness list
- **User** - Get user profile information
- **Price** - Get current prices from multiple retailers using BGG IDs

## Prompts

- **trade-sales-post** - Generate a formatted sales post for your BGG 'for trade' collection with discounted market prices

![Example of trade-sales-post prompt in action](prompt-example.png)

## Example Prompts

Here are some example prompts you can use to interact with the BGG MCP tools:

### 🔍 Search

```
"Search for Wingspan on BGG"
"How many expansions does Grand Austria Hotel have?"
"Search for Wingspan expansions only"
```

### 📊 Game Details

```
"Get details for Azul"
"Show me information about game ID 224517"
"What's the BGG rating for Gloomhaven?"
```

### 📚 Collection

```
"Show me ZeeGarcia's game collection"
"Show games rated 0+ in kkjdaniel's collection"
"List unplayed games in rahdo's collection"
"Find games for 6 players in kkjdaniel's collection"
```

### 🔥 Hotness

```
"Show me the current BGG hotness list"
"What's trending on BGG?"
```

### 👤 User Profile

```
"Show me details about BGG user rahdo"
"When did user ZeeGarcia join BGG?"
```

### 💰 Prices

```
"Get the best price for Wingspan in GBP"
"Show me the best UK price for Ark Nova"
"Compare prices for: Wingspan & Ark Nova"
```

## Setup

You have two options for setting up, the easiest is to use the integration of Smithery.

### A) Installing via Smithery

To install bgg-mcp for Claude Desktop automatically via [Smithery](https://smithery.ai/server/@kkjdaniel/bgg-mcp):

```bash
npx -y @smithery/cli install @kkjdaniel/bgg-mcp --client claude
```

### B) Manual Setup

#### 1. Install Go

You will need to have Go installed on your system to build binary. This can be easily [downloaded and setup here](https://go.dev/doc/install), or you can use the package manager that you prefer such as Brew.

#### 2. Build

The project includes a Makefile to simplify building and managing the binary.

```bash
# Build the application (output goes to build/bgg-mcp)
make build

# Clean build artifacts
make clean

# Both clean and build
make all
```

Or you can simply build it directly with Go...

```bash
go build -o build/bgg-mcp
```

#### 3. Add MCP Config

In the `settings.json` (VS Code / Cursor) or `claude_desktop_config.json` add the following to your list of servers, pointing it to the binary you created earlier, once you load up your AI tool you should see the tools provided by the server connected:

```json
"bgg": {
    "command": "path/to/build/bgg-mcp",
    "args": []
}
```

More details for configuring Claude can be [found here](https://modelcontextprotocol.io/quickstart/user).
