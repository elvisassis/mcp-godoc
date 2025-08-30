package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	log.SetOutput(os.Stderr)

	// Define subcommands
	godocCmd := flag.NewFlagSet("godoc", flag.ExitOnError)
	godocPackage := godocCmd.String("package", "", "The Go package to document (required)")
	godocSymbol := godocCmd.String("symbol", "", "The symbol within the package to document (optional)")

	// Parse the command line arguments
	if len(os.Args) < 2 {
		fmt.Println("Usage: godoctor-cli <command> [arguments]")
		fmt.Println("\nCommands:")
		fmt.Println("  godoc -package <package> [-symbol <symbol>]")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "godoc":
		godocCmd.Parse(os.Args[2:])
		if *godocPackage == "" {
			fmt.Println("Error: -package is required for godoc command.")
			godocCmd.Usage()
			os.Exit(1)
		}
		callGodoc(*godocPackage, *godocSymbol)
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		fmt.Println("Usage: godoctor-cli <command> [arguments]")
		fmt.Println("\nCommands:")
		fmt.Println("  godoc -package <package> [-symbol <symbol>]")
		os.Exit(1)
	}
}

func callGodoc(pkg, symbol string) {
	ctx := context.Background()

	client := mcp.NewClient(&mcp.Implementation{Name: "godoctor-cli", Version: "v1.0.0"}, nil)

	// Connect to the server over stdin/stdout
	transport := &mcp.CommandTransport{Command: exec.Command("./bin/godoctor")}
	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer session.Close()

	// Call the godoc tool on the server.
	params := map[string]any{"package": pkg}
	if symbol != "" {
		params["symbol"] = symbol
	}

	res, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "godoc",
		Arguments: params,
	})
	if err != nil {
		log.Fatalf("CallTool failed: %v", err)
	}

	if res.IsError {
		for _, c := range res.Content {
			if textContent, ok := c.(*mcp.TextContent); ok {
				fmt.Fprintf(os.Stderr, "Tool error: %s\n", textContent.Text)
			}
		}
		os.Exit(1)
	}

	for _, c := range res.Content {
		if textContent, ok := c.(*mcp.TextContent); ok {
			fmt.Println(strings.TrimSpace(textContent.Text))
		}
	}
}
