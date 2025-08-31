package godoc

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func Register(server *mcp.Server) {
	schema, err := jsonschema.For[GodocParams](&jsonschema.ForOptions{})
	if err != nil {
		panic(err)
	}

	name := "godoc"
	mcp.AddTool(server, &mcp.Tool{
		Name:        name,
		Description: "Invokes the 'go doc' command to display documentation for Go packages or symbols.",
		InputSchema: schema,
	}, GodocHandler)
}

// GodocParams defines the input parameters for the godoc tool.
type GodocParams struct {
	Package string `json:"package"`
	Symbol  string `json:"symbol,omitempty"`
}

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
