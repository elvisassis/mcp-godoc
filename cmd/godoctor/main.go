package main

import (
	"context"
	"flag"
	"fmt"
	godoc "godoctor/internal/tools/godocs"
	reviewcode "godoctor/internal/tools/review_code"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var (
	version = "dev"
)

func main() {
	log.SetOutput(os.Stderr)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("godoctor", flag.ExitOnError)
	apiKeyEnvVar := fs.String("api-key-env", "GEMINI_API_KEY", "environment variable for the Gemini API key")
	fs.Parse(args)
	if err := fs.Parse(args); err != nil {
		return err
	}

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "godoctor",
		Version: version}, nil)
	addTools(server, *apiKeyEnvVar)

	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
	return nil
}

func addTools(server *mcp.Server, apiKeyEnvVar string) {
	godoc.Register(server)
	reviewcode.Register(server, apiKeyEnvVar)
}
