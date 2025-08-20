# API E-commerce em Go

Esta é uma API RESTful para uma plataforma de e-commerce construída em Go, seguindo os princípios de arquitetura limpa (Clean Architecture). A aplicação utiliza uma abordagem de Domain-Driven Design (DDD) com clara separação de responsabilidades.

## 🏗️ Arquitetura

A aplicação segue uma arquitetura limpa com as seguintes camadas principais:

1. **Camada de Entrega (HTTP)**: Responsável por lidar com requisições e respostas HTTP
2. **Camada de Serviço**: Contém a lógica de negócios
3. **Camada de Repositório**: Gerencia o acesso a dados
4. **Camada de Domínio**: Contém os modelos de negócio e interfaces

### Estrutura do Projeto



## 🔄 Fluxo de Dados

### 1. Fluxo de Processamento de Requisições

1. **Requisição HTTP** → **Roteador** → **Middleware** → **Handler** → **Serviço** → **Repositório** → **Banco de Dados**

### 2. Fluxo Detalhado

1. **Requisição HTTP**
    - Endpoint: [[MÉTODO] /api/v1/products[/:id]](cci:1://file://wsl.localhost/Ubuntu/home/edsonjr/golang/e-commerce-go/cmd/api/main.go:24:0-95:1)
    - Cabeçalhos: `Content-Type: application/json`

2. **Roteador**
    - Roteia a requisição para o handler correto
    - Aplica middlewares (CORS, ID de requisição, tratamento de erros)

3. **Handler (Camada de Entrega)**
    - Valida a requisição
    - Extrai parâmetros e dados
    - Chama o serviço apropriado
    - Formata a resposta

4. **Camada de Serviço**
    - Implementa as regras de negócio
    - Gerencia transações
    - Valida regras de negócio
    - Chama os métodos do repositório

5. **Camada de Repositório**
    - Interage com o banco de dados
    - Executa consultas
    - Mapeia resultados para modelos de domínio

6. **Banco de Dados**
    - PostgreSQL
    - Gerenciado via migrações

## 🚀 Funcionalidades

- API RESTful
- Arquitetura limpa
- Migrações de banco de dados
- Suporte a Docker
- Configuração por ambiente
- Validação de requisições
- Tratamento de erros
- Logging

## 🛠️ Começando

### Pré-requisitos

- Go 1.20+
- Docker e Docker Compose
- PostgreSQL

### Configuração do Ambiente

1. Copie o arquivo `.env.example` para `.env` e atualize os valores:
   ```bash
   cp .env.example .env