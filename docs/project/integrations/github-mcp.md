# Integração do Codex com o GitHub MCP

Data da configuração: 2026-09-07

## Objetivo

Permitir que o Codex consulte e administre recursos do GitHub por meio do servidor MCP oficial. A configuração usa o servidor remoto mantido pelo GitHub com todos os toolsets habilitados para este teste:

`https://api.githubcopilot.com/mcp/x/all`

O token não é armazenado em arquivos versionados. Para o teste local autorizado, seu valor fica em `.env.github`, ignorado pelo Git, e o Codex o lê da variável de ambiente `GITHUB_PAT_TOKEN`.

## Configuração do projeto

O servidor está definido em `.codex/config.toml`:

```toml
[mcp_servers.github]
url = "https://api.githubcopilot.com/mcp/x/all"
bearer_token_env_var = "GITHUB_PAT_TOKEN"
enabled = true
required = false
startup_timeout_sec = 20
tool_timeout_sec = 60
default_tools_approval_mode = "writes"
```

Decisões de segurança:

- `required = false`: a ausência do GitHub não impede o Codex de iniciar.
- `default_tools_approval_mode = "writes"`: leituras podem ocorrer normalmente; ferramentas de escrita pedem aprovação.
- `/x/all`: habilita todos os toolsets remotos; as operações efetivas continuam limitadas pelas permissões do PAT.
- O token nunca fica no TOML versionado. O arquivo local `.env.github` deve ter permissão `600` e ser removido ou atualizado quando o token for revogado.
- As permissões do token devem seguir o menor privilégio necessário.

## Autenticação

1. Crie um Personal Access Token no GitHub.
2. Dê acesso somente ao repositório `EdsonJunio/e-commerce-go`, quando o tipo de token permitir essa restrição.
3. Habilite inicialmente apenas leitura de conteúdo/metadados e as permissões de issues e pull requests que você pretende usar. Acrescente acesso a Actions somente quando esse fluxo for necessário.
4. Para uma sessão efêmera, exporte o token no terminal que iniciará o Codex:

```bash
export GITHUB_PAT_TOKEN="seu_token"
codex
```

Neste projeto, o launcher abaixo carrega o arquivo local `.env.github` e inicia o Codex com a credencial:

```bash
./scripts/codex-with-github-mcp.sh
```

Não coloque o valor real em `.codex/config.toml`, documentação, histórico do shell ou arquivos versionados. Se optar por um arquivo local de segredos, carregue-o antes de iniciar o Codex e confirme que ele está ignorado pelo Git.

## Ativação e verificação

A configuração foi reconhecida localmente com:

```bash
codex mcp list
codex mcp get github
```

O resultado esperado é o servidor `github` habilitado, usando transporte `streamable_http` e autenticação por bearer token via `GITHUB_PAT_TOKEN`.

Em 2026-09-07, a inicialização, a listagem das ferramentas e as leituras do repositório foram validadas. Um teste de criação de issue recebeu `403 Forbidden`, pois o PAT usado no teste não tinha permissão de escrita em Issues. Nenhum recurso remoto foi criado nessa tentativa.

Depois de carregar o token, inicie uma nova sessão do Codex ou reinicie a extensão. A sessão que criou este arquivo não recebe novas ferramentas MCP dinamicamente.

Na nova sessão:

1. Abra `/mcp` e confirme que `github` apresenta ferramentas.
2. Solicite uma leitura simples, por exemplo: “liste as issues abertas de `EdsonJunio/e-commerce-go`”.
3. Faça uma escrita apenas quando necessário e confirme que o Codex solicita aprovação.

## Diagnóstico

| Sintoma | Verificação |
| --- | --- |
| Servidor não aparece | Execute `codex mcp list` dentro deste projeto e confira se ele está marcado como confiável |
| `401 Unauthorized` | Confirme se `GITHUB_PAT_TOKEN` foi exportado no mesmo ambiente que iniciou o Codex e se não expirou |
| `403 Forbidden` | Acrescente somente a permissão exigida pela operação recusada |
| Ferramentas não aparecem na sessão atual | Reinicie o cliente ou abra uma nova sessão após configurar o MCP |
| Escritas não funcionam | Confirme as permissões do token e aprove a chamada quando solicitado |

## Referências

- [Configuração MCP no Codex](https://developers.openai.com/codex/mcp)
- [Servidor MCP oficial do GitHub](https://github.com/github/github-mcp-server)
- [Guia oficial do GitHub MCP para Codex](https://github.com/github/github-mcp-server/blob/main/docs/installation-guides/install-codex.md)
