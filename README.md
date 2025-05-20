# BGG MCP

BGG MCP is an MCP (Model Context Protocol) server that enables AI tools like Claude to interact with the BoardGameGeek API (XML API2). The server is implemented in Go, using the [GoGeek](https://github.com/kkjdaniel/gogeek) library.

## Tools

- Search _(find games by name)_
- Collection _(find and filter about a users collection)_

## Example

![Example of BGG MCP in action](example.png)

## Installation

### Go

You will need to have Go installed on your system to run the binary. This can be easily [downloaded and setup here](https://go.dev/doc/install), or you can use the package manager that you prefer such as Brew.

### Using Makefile

The project includes a Makefile to simplify building and managing the binary.

```bash
# Build the application (output goes to build/bgg-mcp)
make build

# Clean build artifacts
make clean

# Both clean and build
make all
```

### VS Code (Insiders), Claude, Cursor

Download the compiled server from the latest release, or build it yourself:

```bash
# Using Go directly
go build -o build/bgg-mcp

# Or using Make
make build
```

In the `settings.json` (VS Code / Cursor) or `claude_desktop_config.json` add the following to your list of servers:

```json
"bgg": {
    "command": "path/to/build/bgg-mcp",
    "args": []
}
```

More details for configuring Claude can be [found here](https://modelcontextprotocol.io/quickstart/user).
