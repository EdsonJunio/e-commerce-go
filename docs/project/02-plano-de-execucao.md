# Plano de execução priorizado

Data da proposta: 2026-09-07
Base: [diagnóstico do estado atual](./01-diagnostico-atual.md)

## Princípios de execução

- Trabalhar em etapas pequenas, demonstráveis e reversíveis.
- Não considerar migration, seed, comentário ou Swagger como funcionalidade pronta sem caminho executável e teste.
- Definir regra de negócio e contrato antes da implementação.
- Adicionar padrões distribuídos somente quando houver um problema distribuído concreto.
- Não alterar migrations antigas sem decidir antes se há dados ou ambientes a preservar.
- Encerrar cada etapa com código formatado, análise estática, testes adequados e documentação atualizada.

## Ordem recomendada

| Ordem | Prioridade | Etapa | Resultado verificável |
| --- | --- | --- | --- |
| 1 | P0 | Decisão arquitetural | Comparação objetiva, limites dos contextos e ADR da arquitetura escolhida |
| 2 | P0 | Segurança e configuração | JWT sem segredo fixo, RBAC mínimo, configuração única validada e pprof controlado |
| 3 | P0 | Correção do catálogo | Updates e filtros corretos, SKU coerente com banco, cache consistente e erros estáveis |
| 4 | P0 | Base de testes e CI | Regras críticas cobertas, migrations testadas e pipeline obrigatório |
| 5 | P1 | Persistência e ambiente local | Schema reconciliado, índices/restrições, seed seguro e Compose reproduzível |
| 6 | P1 | E-commerce essencial | Cliente, catálogo/SKU, estoque, carrinho e pedido funcionando em cortes verticais |
| 7 | P2 | Payment API simulada | Pagamento idempotente com histórico e integração confiável com pedido |
| 8 | P2 | Billing/Fiscal API simulada | Boleto e documento fiscal idempotentes com histórico |
| 9 | P2 | Comunicação assíncrona | Eventos versionados e outbox/inbox apenas para fluxos que precisarem disso |
| 10 | P3 | Operação e evolução | Tracing, SLOs, backups testados, segurança ampliada e preparação de deploy |

## Etapa 1 — decisão arquitetural

### Objetivo

Escolher uma arquitetura proporcional ao caráter educacional e capaz de evoluir para integrações reais.

### Trabalho

- Definir requisitos funcionais, não funcionais e limites explícitos do projeto.
- Identificar contextos: identidade/cliente, catálogo, inventário, carrinho, pedido, pagamento e cobrança/fiscal.
- Comparar monólito modular, três microsserviços desde o início e evolução híbrida.
- Comparar REST, gRPC, filas e eventos por fluxo, incluindo custo operacional, acoplamento, consistência e teste.
- Definir propriedade dos dados e regra de proibição de acesso direto ao banco alheio.
- Registrar decisão, alternativas rejeitadas, consequências e gatilhos de revisão em ADR.

### Arquivos previstos

- `docs/project/03-requisitos.md`
- `docs/project/04-arquitetura.md`
- `docs/project/05-contextos-e-dominio.md`
- `docs/project/adrs/ADR-001-arquitetura-inicial.md`

### Critério de conclusão

É possível responder quem é responsável por cada regra e dado, como cada limite se comunica hoje, quais falhas são aceitas e em que condição um módulo vira serviço.

## Etapa 2 — segurança e configuração

### Objetivo

Eliminar os riscos que tornam o código atual inadequado até para uma demonstração compartilhada.

### Trabalho

- Configurar e validar chave, emissor, audiência e duração do JWT.
- Modelar papel/permissão sem depender do ID do usuário.
- Exigir autorização administrativa nas mutações do catálogo.
- Unificar `APP_ENVIRONMENT` e injetar uma única configuração.
- Controlar pprof explicitamente e revisar CORS, limites de corpo e rate limit do login.
- Criar `.dockerignore` e remover credenciais fixas dos caminhos de execução.

### Testes

- Token válido, expirado, assinatura inválida, algoritmo inválido e issuer/audience incorretos.
- Usuário comum recebe 403 e administrador executa mutação.
- Startup falha com configuração de produção insegura ou inválida.

### Critério de conclusão

Nenhum segredo fica no código; autenticação e autorização são comportamentos distintos e testados; ambiente local e Docker usam a mesma convenção.

## Etapa 3 — correção do catálogo

### Objetivo

Transformar categoria, produto e SKU em uma base confiável antes de construir estoque e carrinho.

### Trabalho

- Substituir mapas de filtros por tipos explícitos.
- Corrigir atualização parcial preservando campos omitidos.
- Alinhar DTOs, nomes JSON, validação e erros.
- Reconciliar `ProductSKU` com `product_skus` e `stock`.
- Definir operações mínimas de SKU e relação com produto/preço.
- Corrigir invalidação do cache e decidir se Redis é opcional.
- Remover Wire, mocks e erros sem uso apenas depois de confirmar a estratégia escolhida.

### Testes

- Tabelas de casos para validação, filtros, atualização e transições de ativo.
- Handler tests para status, código e envelope de resposta.
- Repository tests com PostgreSQL e cache tests com Redis compatível.

### Critério de conclusão

Todos os endpoints documentados têm comportamento testado; filtros alteram a consulta; updates não mudam campos omitidos; SKU e estoque retornam dados coerentes.

## Etapa 4 — base de testes, contratos e CI

### Objetivo

Criar uma barreira automática contra regressões.

### Trabalho

- Separar testes unitários, integração e contrato por propósito.
- Subir dependências efêmeras nos testes de integração e aplicar migrations do zero.
- Validar `up` e `down`, constraints e queries críticas.
- Validar OpenAPI e impedir divergência entre contrato e resposta.
- Criar CI com `gofmt`, `go vet`, testes, cobertura informativa, build e Compose config.

### Critério de conclusão

Um clone limpo executa a pipeline sem passos manuais e uma falha nas regras críticas interrompe a entrega.

## Etapa 5 — persistência e ambiente local

### Objetivo

Estabelecer um schema coerente e uma experiência local previsível.

### Trabalho

- Confirmar política para migrations já aplicadas.
- Adicionar índices, unicidades e checks a partir dos casos de uso definidos.
- Definir snapshots de item/endereço do pedido e históricos de estado.
- Modelar reserva de estoque com expiração e idempotência.
- Proteger o seed contra ambientes não locais e tratar todos os erros.
- Fixar versões das imagens, provisionar observabilidade e documentar reset/backup/restore local.

### Critério de conclusão

Migrations sobem e descem em banco limpo, invariantes rejeitam dados inválidos e os comandos documentados reproduzem o ambiente.

## Etapa 6 — e-commerce essencial em cortes verticais

Cada corte deve atravessar contrato HTTP, caso de uso, domínio, persistência, teste e documentação.

1. Cliente e endereço.
2. Categoria, produto e SKU completos.
3. Ajuste e consulta de estoque.
4. Carrinho e itens.
5. Checkout com validação de preço e disponibilidade.
6. Reserva/liberação de estoque.
7. Criação e consulta de pedido.
8. Cancelamento de pedido antes e depois de pagamento, conforme regras definidas.

O pedido deve possuir chave idempotente desde sua primeira implementação. Transações locais devem proteger criação de pedido, itens, snapshot e reserva, conforme o limite arquitetural escolhido.

## Etapa 7 — Payment API simulada

### Objetivo

Exercitar um limite de serviço com autonomia, falhas próprias e histórico financeiro auditável.

### Trabalho

- Contrato idempotente para solicitar pagamento.
- Banco próprio e usuário de banco exclusivo.
- Estado corrente mais histórico imutável de tentativas e transições.
- Simulador de provedor com aprovação, recusa, timeout e erro técnico controláveis.
- Cancelamento e estorno com regras explícitas.
- Autenticação serviço a serviço, timeouts e retry somente em operações seguras.
- Entrega confiável do resultado ao pedido.

### Critério de conclusão

Repetir a mesma solicitação não duplica cobrança; falhas e reprocessamentos convergem para um estado conhecido; auditoria explica cada transição.

## Etapa 8 — Billing/Fiscal API simulada

### Objetivo

Modelar cobrança e documento fiscal sem depender inicialmente de provedor real.

### Trabalho

- Separar cobrança, boleto e documento fiscal como conceitos distintos.
- Implementar emissão, consulta, vencimento/cancelamento e histórico.
- Definir gatilhos a partir de pedido/pagamento conforme regras fiscais simuladas.
- Isolar adaptadores para futura integração real.
- Aplicar idempotência, inbox e reprocessamento quando eventos forem adotados.

### Critério de conclusão

O mesmo comando/evento não gera documento duplicado, integrações simuladas podem ser substituídas e todo estado possui histórico.

## Etapa 9 — eventos, consistência eventual e resiliência

Esta etapa depende da existência de pelo menos dois processos independentes. Antes disso, chamadas internas e transações locais são mais simples e observáveis.

### Trabalho condicionado à necessidade

- Publicar eventos de fatos concluídos, com nome e versão estáveis.
- Adotar outbox no produtor e inbox/deduplicação no consumidor.
- Definir retry com backoff, dead-letter, replay e política de ordenação.
- Modelar compensações do fluxo pedido–estoque–pagamento–fiscal.
- Usar saga orquestrada ou coreografada somente após comparar clareza, acoplamento e operação.
- Aplicar circuit breaker apenas em dependências remotas com falha recorrente mensurável.

### Critério de conclusão

Testes demonstram perda temporária de comunicação, entrega duplicada, evento fora de ordem e reprocessamento sem efeitos duplicados.

## Etapa 10 — operação e evolução

- OpenTelemetry com propagação de trace e correlação nos logs.
- Métricas de pedido, reserva, pagamento, cobrança e filas.
- Dashboards, alertas e objetivos de nível de serviço proporcionais ao projeto.
- Backup automatizado, teste periódico de restore e política de retenção.
- Varredura de dependências/imagens, proteção de dados e gestão de segredos.
- Estratégia de deploy e rollback por serviço.
- Testes de carga somente depois de existirem metas e cenários representativos.

## Modelo para cada entrega

Cada entrega futura deve registrar:

1. Problema e comportamento esperado.
2. Alternativas consideradas e decisão.
3. Arquivos criados ou alterados.
4. Regras, estados, contrato e persistência envolvidos.
5. Testes adicionados e comandos executados.
6. Riscos ou limitações restantes.
7. Documentação atualizada.
8. Próximo passo pequeno e verificável.
