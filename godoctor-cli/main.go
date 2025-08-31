package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
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
	godoctorSymbol := godocCmd.String("symbol", "", "The symbol within the package to document (optional)")

	reviewCmd := flag.NewFlagSet("review", flag.ExitOnError)
	reviewHint := reviewCmd.String("hint", "", "An optional hint for the AI reviewer")

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
		callGodoc(*godocPackage, *godoctorSymbol)
	case "review":
		reviewCmd.Parse(os.Args[2:])
		if reviewCmd.NArg() != 1 {
			fmt.Println("Error: a file path is required for the review command.")
			reviewCmd.Usage()
			os.Exit(1)
		}
		callCodeReview(reviewCmd.Arg(0), *reviewHint)
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

func callCodeReview(filePath, hint string) {
	ctx := context.Background()

	codeContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	client := mcp.NewClient(&mcp.Implementation{
		Name:    "godoctor-cli",
		Version: "v1.0.0"},
		nil)

	// Connect to the server over stdin/stdout

	transport := &mcp.CommandTransport{
		Command: exec.Command("./bin/godoctor")}
	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer session.Close()

	// Call the code_review tool on the server.
	params := map[string]any{"code_content": string(codeContent)}
	if hint != "" {
		params["hint"] = hint
	}

	res, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "code_review",
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
