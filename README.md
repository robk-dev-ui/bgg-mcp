<p align="center">
  <img src="bgg-mcp-logo.png" width="250" alt="BGG MCP Logo">
</p>

# BGG MCP: BoardGameGeek MCP API Server

[![smithery badge](https://smithery.ai/badge/@kkjdaniel/bgg-mcp)](https://smithery.ai/server/@kkjdaniel/bgg-mcp)

> [!WARNING]  
> This project is under active developmennt, therefore expect tooling to change.

BGG MCP provides access to the BoardGameGeek API through the [Model Context Protocol](https://www.anthropic.com/news/model-context-protocol), enabling retrieval and filtering of board game data, user collections, and profiles. The server is implemented in Go, using the [GoGeek](https://github.com/kkjdaniel/gogeek) library, which helps ensure robust API interactions.

<a href="https://boardgamegeek.com/">
  <img src="powered-bgg.webp" width="160" alt="Powered by BGG">
</a>

## Example

![Example of BGG MCP in action](example.png)

## Tools

- Game Details _(find game by name, currently returns best match)_
- Collection _(find and filter about a users collection)_
- Hottness _(get the current BGG hotness)_
- User _(find details of a user by username)_

## Roadmap

- [x] Specific Game Details _(by name)_
- [x] Collection (+ filters)
- [x] Hot Games
- [x] User Details
- [ ] Broad Search
- [ ] Recommended Games

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

## Using Makefile

The project includes a Makefile to simplify building and managing the binary.

```bash
# Build the application (output goes to build/bgg-mcp)
make build

# Clean build artifacts
make clean

# Both clean and build
make all
```
