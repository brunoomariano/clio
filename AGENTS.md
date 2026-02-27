# AGENTS

## Objetivo
Descrever o protocolo oficial de colaboracao entre humano e LLM na evolucao da TUI.

## Principios obrigatorios
1. Nunca modificar multiplos contratos ao mesmo tempo.
2. Alteracoes devem ser feitas via diff.
3. Layout, estado e comportamento sao arquivos distintos.
4. Sempre respeitar largura minima definida.
5. Snapshots ASCII sao o contrato visual.
6. Nunca regenerar a aplicacao inteira sem solicitacao explicita.
7. Mudancas devem ser reversiveis e rastreaveis.

## Fluxo de Trabalho
Sempre responder no formato:
1. Alteracoes propostas
2. Diff dos arquivos afetados
3. Snapshot ASCII atualizado (se aplicavel)
4. Justificativa tecnica breve
