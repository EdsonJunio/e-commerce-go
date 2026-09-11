# Documentação do projeto

Esta pasta contém a documentação técnica escrita e revisada manualmente. Os arquivos `docs.go`, `swagger.json` e `swagger.yaml` no diretório pai são artefatos gerados pelo Swaggo e não devem ser usados para guardar decisões ou análises manuais.

## Documentos disponíveis

| Documento | Estado | Finalidade |
| --- | --- | --- |
| [Engenharia reversa do código atual](./01-engenharia-reversa-codigo-atual.md) | Concluído em 2026-09-10 | Contexto de cada struct/interface, métodos, dependências, fluxos e comportamento observado |
| [Current-scope quality gap analysis](./02-current-scope-quality-gap-analysis.md) | Active roadmap | Stabilization subtasks for the currently executable API |
| [Complete E-commerce implementation roadmap](./03-complete-ecommerce-implementation-roadmap.md) | Proposed roadmap | Incremental construction plan for this repository's E-commerce service |
| [Arquitetura proposta](./arquitetura-e-commerce-go.md) | Em evolução | Visão arquitetural, limites de serviço e direção técnica |
| [Integração do Codex com o GitHub MCP](./integrations/github-mcp.md) | Leitura validada; escrita depende das permissões do PAT | Acesso seguro do Codex a repositórios, issues e pull requests |

## Documentos previstos

Os documentos abaixo serão criados nas próximas etapas. A ordem poderá mudar após cada revisão.

1. Requisitos funcionais e não funcionais.
2. Comparação arquitetural e ADR da arquitetura escolhida.
3. Contextos delimitados, responsabilidades e regras de domínio.
4. Contratos HTTP e catálogo de erros.
5. Eventos e contratos de mensageria.
6. Modelagem de dados proposta e estratégia de persistência.
7. Fluxos e diagramas de sequência.
8. Segurança, testes, observabilidade e tratamento de falhas.
9. Execução local, implantação e CI/CD.
10. Glossário e índice de ADRs.

## Regra de manutenção

Cada etapa deve atualizar o documento afetado, registrar decisões relevantes em ADR e indicar evidências de validação. Funcionalidade presente apenas em migration, seed, comentário ou Swagger não será marcada como implementada sem um caminho executável na aplicação e testes que comprovem o comportamento principal.
