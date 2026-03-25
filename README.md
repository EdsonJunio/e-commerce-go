# E-Commerce API (Golang)

![Go Version](https://img.shields.io/badge/Go-1.25.1-00ADD8?style=flat&logo=go)
![Gin Framework](https://img.shields.io/badge/Gin-1.11.0-00ADD8?style=flat)
![GORM](https://img.shields.io/badge/GORM-1.31.0-red?style=flat)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-13-336791?style=flat&logo=postgresql)
![Redis](https://img.shields.io/badge/Redis-7-DC382D?style=flat&logo=redis)
![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat&logo=docker)

Uma API robusta, rápida e escalável para e-commerce, construída em **Go**. Embora outras tecnologias tenham construído a web e mereçam respeito, este projeto consolida a linguagem do mascote azul como a principal ferramenta de estudo e experimentação do portfólio. A arquitetura foi desenhada para proporcionar aquela sensação viciante de compilar um binário enxuto, ver a aplicação subindo em milissegundos e aguentar tráfego pesado sem suar a camisa. 💙🦫

O projeto adota princípios de **Clean Architecture** (Domain-Driven), garantindo separação clara de responsabilidades, fácil manutenção e alta performance.

## 🚀 Tecnologias e Stack

* **Linguagem:** Go 1.25+
* **Web Framework:** [Gin](https://gin-gonic.com/)
* **Banco de Dados Relacional:** PostgreSQL + [GORM](https://gorm.io/)
* **Cache:** Redis
* **Autenticação:** JWT (JSON Web Tokens)
* **Logger:** Uber Zap
* **Documentação:** Swagger (swaggo/gin-swagger)
* **Monitoramento:** Prometheus + Grafana (via Docker)
* **Infraestrutura:** Docker & Docker Compose

## 🏗 Arquitetura e Modelagem

O projeto segue uma estrutura modular dividida por domínios da regra de negócio:

* **`identity`**: Autenticação, usuários e endereços.
* **`catalog`**: Vitrine, produtos, categorias e controle de SKUs (Stock Keeping Units).
* *(Em breve)* **`inventory`**: Controle físico de estoque e movimentações.
* *(Em breve)* **`shopping`**: Carrinho de compras e lista de desejos.
* *(Em breve)* **`checkout`**: Gestão de pedidos e logística.

### Diagrama de Banco de Dados

*(Salve a imagem do fluxograma na pasta `docs/` com o nome `er-diagram.jpg` para renderizar aqui)*

![Diagrama ER](./docs/er-diagram.jpg)

## 📂 Estrutura de Diretórios

```text
.
├── cmd/                # Pontos de entrada da aplicação (main.go, seeders)
├── deploy/             # Arquivos de infra (Prometheus, Grafana)
├── docs/               # Documentação gerada pelo Swagger
├── internal/           # Código privado da aplicação (Regras de negócio)
│   ├── catalog/        # Domínio de Catálogo (Produtos, Categorias, SKUs)
│   ├── database/       # Migrations SQL
│   ├── identity/       # Domínio de Identidade (Auth, Usuários)
│   └── shared/         # Middlewares, configs, cache, responses base
├── pkg/                # Pacotes públicos e utilitários (ex: logger)
└── scripts/            # Scripts utilitários (testes, hooks de git)
⚙️ Como Executar o Projeto
Pré-requisitos
Go 1.25+ instalado

Docker e Docker Compose instalados

Make (Opcional, mas recomendado)

Passo a Passo
Clone o repositório:

Bash
git clone [https://github.com/seu-usuario/e-commerce-go.git](https://github.com/seu-usuario/e-commerce-go.git)
cd e-commerce-go
Configure as variáveis de ambiente:
Crie um arquivo .env na raiz do projeto (use as variáveis necessárias conforme o seu pacote de config).

Suba a infraestrutura (Postgres, Redis, Prometheus):

Bash
docker-compose up -d db redis prometheus grafana
Rode as migrações:

Bash
make migrate-up
Inicie a API:

Bash
go run cmd/api/main.go
A aplicação estará rodando em http://localhost:8081 (ou a porta que estiver configurada no seu .env).

📚 Documentação da API (Swagger)
A documentação interativa das rotas é gerada automaticamente pelo Swagger.
Com a aplicação rodando, acesse:

👉 http://localhost:8081/swagger/index.html

Para atualizar o Swagger após modificar os comentários nas rotas:

Bash
swag init -g cmd/api/main.go --parseDependency --parseInternal
📊 Monitoramento e Observabilidade
Este projeto possui integração nativa com o Prometheus. As métricas da aplicação (tempo de resposta, taxa de erros,
consumo de memória) são expostas na rota /metrics e podem ser visualizadas no dashboard do Grafana na porta 3000
(configurado via Docker Compose).

