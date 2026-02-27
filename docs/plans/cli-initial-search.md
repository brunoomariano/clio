# Plano: CLI com termo inicial no TUI

## Resumo
Adicionar suporte a `clio '<text>'` para abrir o TUI com o termo ja preenchido na caixa de busca. Implementar parsing de args com erro quando houver mais de um argumento e passar o termo para o modelo de UI na inicializacao. Atualizar a documentacao de uso.

## Mudancas de interface publica
- CLI: passa a aceitar 0 ou 1 argumento posicional.
- `clio` mantem comportamento atual.
- `clio "<text>"` abre o TUI com a busca ja preenchida.
- Se houver mais de 1 argumento, o app deve falhar com mensagem de uso e codigo de erro.
- UI interna: `ui.New` passa a receber um parametro `initialQuery string` ou um novo construtor que aceite esse valor.

## Estrategia de implementacao
1. Criar estrutura de planos.
2. Criar o arquivo `docs/plans/cli-initial-search.md` com este conteudo.
3. Em `cmd/clio/main.go`, adicionar `parseArgs(args []string) (string, error)`.
4. `parseArgs` deve retornar `""` se `len(args) == 0`.
5. `parseArgs` deve retornar `args[0]` se `len(args) == 1`.
6. `parseArgs` deve retornar erro se `len(args) > 1`.
7. Adicionar funcao injetavel `getArgs = func() []string { return os.Args[1:] }` para isolar testes.
8. `run()` deve usar `parseArgs(getArgs())` e retornar erro em caso de uso invalido.
9. Em `internal/ui/model.go`, atualizar `New` para receber `initialQuery string` ou criar `NewWithQuery` e manter `New` como wrapper.
10. Definir `search.SetValue(initialQuery)` antes de retornar o `Model`.
11. Garantir que `Init()` chama `runSearch()` para aplicar o termo inicial.
12. Atualizar `cmd/clio/main.go` para passar `initialQuery` na criacao do modelo.
13. Ajustar testes de UI que chamam `ui.New` para incluir o novo parametro.
14. Adicionar testes de `parseArgs` em `cmd/clio`.
15. Ajustar testes de `run()` para definir `getArgs` como `[]string{}`.
16. Atualizar `README.md` na secao Usage com o novo uso.

## Casos de teste e cenarios
- `clio` inicia sem texto na busca.
- `clio "erro"` inicia com `erro` na caixa de busca e resultados filtrados.
- `clio foo bar` retorna erro e nao inicia o TUI.
- `clio ""` inicia com busca vazia.

## Assuncoes e defaults
- Mensagem de erro: `usage: clio '<text>'`.
- O termo inicial deve ser usado exatamente como recebido.
- Nao ha flags adicionais neste momento; apenas argumento posicional.
