package reviewcode

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/genai"
)

func Register(server *mcp.Server, apiKey string) {

	if apiKey == "" {
		log.Printf("GEMINI_API_KEY environment variable not set.")
		return
	}
	name := "code_review"
	schema, err := jsonschema.For[CodeReviewParams](&jsonschema.ForOptions{})
	if err != nil {
		panic(err)
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        name,
		Description: "Invokes the 'go doc' command to display documentation for Go packages or symbols.",
		InputSchema: schema,
	}, CodeReviewHandler)
}

// CodeReviewParams defines the arguments for the code_review tool.
type CodeReviewParams struct {
	CodeContent string `json:"code_content" jsonschema:"the Go code content to review"`
	Hint        string `json:"hint,omitempty" jsonschema:"an optional hint for the AI reviewer"`
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
		Backend: genai.BackendGeminiAPI,
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
