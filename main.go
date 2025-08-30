package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// GodocParams defines the arguments for the godoc tool.
type GodocParams struct {
	Package string `json:"package" jsonschema:"the Go package to document"`
	Symbol  string `json:"symbol,omitempty" jsonschema:"the symbol within the package to document (optional)"`
}

// GodocHandler handles the godoc tool call.
func GodocHandler(ctx context.Context, req *mcp.CallToolRequest, args GodocParams) (*mcp.CallToolResult, any, error) {
	cmdArgs := []string{"doc"}
	if args.Symbol != "" {
		cmdArgs = append(cmdArgs, fmt.Sprintf("%s.%s", args.Package, args.Symbol))
	} else {
		cmdArgs = append(cmdArgs, args.Package)
	}

	cmd := exec.CommandContext(ctx, "go", cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if ok {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error executing 'go doc': %s\n%s", exitErr.Error(), string(output))}},
			}, nil, nil
		}
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to execute 'go doc': %v", err)}},
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: strings.TrimSpace(string(output))}},
	}, nil, nil
}

func main() {
	log.SetOutput(os.Stderr)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "godoc-server",
		Version: "v1.0.0"},
	 nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "godoc",
		Description: "Invokes the 'go doc' command to display documentation for Go packages or symbols.",
	}, GodocHandler)

	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
		log.Printf("Server failed: %v", err)
	}
}
