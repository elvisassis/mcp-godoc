Usando a ferramenta use-prompt com o prompt dev-pt sua tarefa é:
Criar um component godoctor-cli que chamará o servidor MCP usando o comando transport. Este CLI exporá todas as ferramentas usando subcomandos e nos permitirá testar a implementação do servidor MCP sem a necessidade de compilar as chamadas JSON-RPC manualmente

Use a implementação de referência em https://github.com/modelcontextprotocol/go-sdk/blob/main/README.md para construir o cliente

Teste chamando na linha de comando:

- the godoc tool with a local package
- the godoc tool with a local package and symbol
- the godoc tool with an external package
- the godoc tool with an external package and symbol

Use o resultado retornado desta ferramenta como contexto e inicie o desenvolvimento.