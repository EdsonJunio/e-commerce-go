# 📦 Logger

Este pacote implementa um logger estruturado e performático baseado no [`uber-go/zap`](https://github.com/uber-go/zap), com foco em aplicações web escritas em Go.

---

## ✅ O que este logger faz:

- Inicializa o `zap.Logger` com configuração específica para `development` (console com cores) ou `production` (JSON estruturado).
- Adiciona campos padrão a todos os logs: `env`, `service`, `version`, `pid`.
- Redireciona logs da biblioteca padrão (`log`) para o `zap`.
- Fornece funções globais (`logger.L()`, `logger.With()`) para logar facilmente em qualquer lugar.
- Inclui middleware de logging para o `Gin`, que registra:
    - Método HTTP
    - Status da resposta
    - IP do cliente
    - Tempo de execução
    - User-Agent
    - Request ID (se presente)
- Inclui middleware de `recovery` para capturar panics e logar stack trace automaticamente.
- Garante o flush dos logs com `logger.Sync()` (deve ser chamado com `defer` no `main.go`).
- Permite extração e propagação do logger via `context.Context` para rastreio completo da requisição.

---

## 📌 Uso comum:

```go
logger.L().Info("Produto criado com sucesso")
logger.L().Error("Erro ao criar produto", zap.Error(err))
