Usando a ferramenta use-prompt com o prompt dev-pt sua tarefa é:
Criar um servidor Model Context Protocol (MCP) para expor uma ferramenta "godoc". Para a implementação do MCP, você deve usar o Go SDK oficial para MCP e o transporte stdio.
A ferramenta do nosso servidor MCP chamada "godoc" invoca o comando shell "go doc". A ferramenta receberá um argumento obrigatório "package" e um argumento opcional "símbolo".

Leia estas referências para reunir informações sobre a tecnologia e a estrutura do projeto antes de escrever qualquer código:

- https://github.com/modelcontextprotocol/go-sdk
- https://modelcontextprotocol.io/specification/2025-06-18/basic/lifecycle
- https://go.dev/doc/modules/layout
- https://pkg.go.dev/golang.org/x/tools/cmd/godoc

Para testar o servidor, use comandos shell como estes:
(
  echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18"}}'
  echo '{"jsonrpc":"2.0","method":"notifications/initialized","params":{}}'
  echo '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}'
  echo '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name": "godoc", "arguments": {"package": "fmt"} } }'
  echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name": "godoc", "arguments": {"package": "fmt", "symbol": "Println"} } }'
) | ./bin/godoctor

Use o resultado retornado desta ferramenta como contexto e inicie o desenvolvimento.