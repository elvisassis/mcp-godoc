package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/genai"
)

// GodocParams defines the arguments for the godoc tool.
type GodocParams struct {
	Package string `json:"package"`
	Symbol  string `json:"symbol,omitempty"`
}

// CodeReviewParams defines the arguments for the code_review tool.
type CodeReviewParams struct {
	CodeContent string `json:"code_content" jsonschema:"the Go code content to review"`
	Hint        string `json:"hint,omitempty" jsonschema:"an optional hint for the AI reviewer"`
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

// CodeReviewHandler handles the code_review tool call.
func CodeReviewHandler(ctx context.Context, req *mcp.CallToolRequest, args CodeReviewParams) (*mcp.CallToolResult, any, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{&mcp.TextContent{Text: "GEMINI_API_KEY environment variable not set."}},
		}, nil, nil
	}

	// Criar cliente Gemini
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI, // backend correto
	})
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to create Gemini client: %v", err)}},
		}, nil, nil
	}

	// Montar prompt
	prompt := fmt.Sprintf(
		"Analyze the following Go code and provide a list of improvements in JSON format. "+
			"The JSON should be an array of objects, where each object has 'line', 'column', and 'description' fields. "+
			"Follow Go best practices. Hint: %s\n\nCode:\n```go\n%s\n```",
		args.Hint,
		args.CodeContent,
	)

	// Chamada para o modelo Gemini
	resp, err := client.Models.GenerateContent(
		ctx,
		"gemini-2.5-flash", // modelo mais novo/rápido
		genai.Text(prompt),
		nil, // opções extras (pode ser nil)
	)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to generate content from Gemini: %v", err)}},
		}, nil, nil
	}

	// Extrair o texto gerado
	content := resp.Text()

	// Remove markdown code block fences if present
	cleanedContent := strings.TrimSpace(content)
	if strings.HasPrefix(cleanedContent, "```json") && strings.HasSuffix(cleanedContent, "```") {
		cleanedContent = strings.TrimPrefix(cleanedContent, "```json")
		cleanedContent = strings.TrimSuffix(cleanedContent, "```")
		cleanedContent = strings.TrimSpace(cleanedContent)
	}

	// Validar se o retorno é JSON válido
	var js json.RawMessage
	if err := json.Unmarshal([]byte(cleanedContent), &js); err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Gemini response is not valid JSON: %v\nResponse: %s", err, cleanedContent)}}, 
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: cleanedContent}},
	}, nil, nil
}

func main() {
	log.SetOutput(os.Stderr)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	server := mcp.NewServer(&mcp.Implementation{Name: "godoc-server", Version: "v1.0.0"}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "godoc",
		Description: "Invokes the 'go doc' command to display documentation for Go packages or symbols.",
	}, GodocHandler)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "code_review",
		Description: "Analyzes Go code using the Gemini API and provides a list of improvements in JSON format.",
	}, CodeReviewHandler)

	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
