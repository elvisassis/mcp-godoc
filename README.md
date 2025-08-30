# GoDoctor: Model Context Protocol (MCP) Server and CLI Client

This project demonstrates a Model Context Protocol (MCP) server implemented in Go, exposing a `godoc` tool. It also includes a command-line interface (CLI) client that interacts with this server, allowing users to easily query Go documentation without manually constructing JSON-RPC calls.

## Features

- **MCP Server (`godoctor`):** A lightweight MCP server that listens for requests over standard input/output (stdio).
- **`godoc` Tool:** An MCP tool exposed by the server that invokes the `go doc` command to retrieve documentation for Go packages and symbols.
- **CLI Client (`godoctor-cli`):** A user-friendly command-line interface to interact with the `godoctor` MCP server. It simplifies calling the `godoc` tool.

## Prerequisites

- [Go](https://go.dev/dl/) (version 1.25.0 or higher)

## Project Structure

```
./
├── bin/
│   ├── godoctor          # Compiled MCP server executable
│   └── godoctor-cli      # Compiled CLI client executable
├── godoctor-cli/         # Source code for the CLI client
│   └── main.go
│   └── go.mod
├── go.mod                # Go module file for the MCP server
├── main.go               # Source code for the MCP server
└── README.md             # This file
```

## Building the Project (Developer Perspective)

To build the `godoctor` MCP server and the `godoctor-cli` client, follow these steps:

1.  **Navigate to the project root directory:**

    ```bash
    cd /path/to/your/mcp-godoc
    ```

2.  **Build the MCP Server (`godoctor`):**

    This command will download the necessary Go SDK dependencies and compile the server executable into the `bin/` directory.

    ```bash
    go mod tidy
    go build -o bin/godoctor .
    ```

3.  **Build the CLI Client (`godoctor-cli`):**

    This command will navigate into the `godoctor-cli` directory, download its dependencies, and compile the client executable into the `bin/` directory.

    ```bash
    cd godoctor-cli && go mod tidy
    cd godoctor-cli && go build -o ../bin/godoctor-cli .
    ```

    *Note: The `cd godoctor-cli &&` prefix is necessary because `godoctor-cli` is a separate Go module within the project.*

## Usage (User Perspective)

Once built, you can use the `godoctor-cli` to interact with the `godoctor` MCP server.

### General Usage

```bash
./bin/godoctor-cli <command> [arguments]
```

### `godoc` Command

The `godoc` command allows you to retrieve documentation for Go packages and symbols.

-   **Required Argument:**
    -   `-package <package_path>`: The import path of the Go package (e.g., `fmt`, `net/http`).
-   **Optional Argument:**
    -   `-symbol <symbol_name>`: The name of a specific function, type, variable, or constant within the package (e.g., `Println`, `Client`).

#### Examples:

1.  **Get documentation for a local package (e.g., `fmt`):**

    ```bash
    ./bin/godoctor-cli godoc -package fmt
    ```

2.  **Get documentation for a specific symbol in a local package (e.g., `fmt.Println`):**

    ```bash
    ./bin/godoctor-cli godoc -package fmt -symbol Println
    ```

3.  **Get documentation for an external package (e.g., `net/http`):**

    ```bash
    ./bin/godoctor-cli godoc -package net/http
    ```

4.  **Get documentation for a specific symbol in an external package (e.g., `net/http.Client`):**

    ```bash
    ./bin/godoctor-cli godoc -package net/http -symbol Client
    ```

## Testing (Advanced/Developer)

If you wish to test the MCP server directly without the CLI client, you can pipe JSON-RPC commands to its standard input. Note that due to buffering in the `mcp.StdioTransport`, the JSON output might not be immediately visible when piping all commands at once.

```bash
(
  echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18"}}'
  echo '{"jsonrpc":"2.0","method":"notifications/initialized","params":{}}'
  echo '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}'
  echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"godoc","arguments":{"package":"fmt"}}}'
) | ./bin/godoctor 2>&1
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
