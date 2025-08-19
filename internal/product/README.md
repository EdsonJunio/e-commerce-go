# Módulo de Produtos

Este módulo implementa as operações CRUD para gerenciamento de produtos no sistema.

## Estrutura do Projeto

```
product/
├── domain/               # Definições de domínio e interfaces
│   └── product.go
├── repository/           # Implementação do repositório (banco de dados)
│   └── product_repository.go
├── service/              # Lógica de negócios
│   └── product_service.go
├── delivery/             # Camada de entrega (HTTP, gRPC, etc.)
│   └── http/
│       └── product_handler.go
└── wire.go               # Configuração de injeção de dependências
```

## Funcionalidades

- Criação de produtos
- Listagem de produtos com paginação e filtros
- Busca de produto por ID
- Busca de produto por slug
- Atualização de produtos
- Exclusão lógica de produtos

## Rotas da API

### Criar Produto
```
POST /products
```

**Request Body:**
```json
{
  "category_id": 1,
  "name": "Produto de Exemplo",
  "slug": "produto-de-exemplo",
  "price_cents": 9999,
  "is_active": true
}
```

### Listar Produtos
```
GET /products?page=1&limit=10&category_id=1&is_active=true
```

### Obter Produto por ID
```
GET /products/{id}
```

### Obter Produto por Slug
```
GET /products/slug/{slug}
```

### Atualizar Produto
```
PUT /products/{id}
```

**Request Body:**
```json
{
  "category_id": 1,
  "name": "Produto Atualizado",
  "slug": "produto-atualizado",
  "price_cents": 8999,
  "is_active": true
}
```

### Deletar Produto
```
DELETE /products/{id}
```

## Configuração

Para usar este módulo, você precisará configurar a injeção de dependências. Exemplo:

```go
// Inicialize o banco de dados
db, err := gorm.Open(...)

// Inicialize o handler de produtos
productHandler := InitializeProductHandler(db)

// Configure as rotas
r := chi.NewRouter()
productHandler.RegisterRoutes(r)
```

## Validações

- Nome do produto é obrigatório
- Slug é obrigatório e único
- Preço deve ser maior que zero
- Categoria deve existir (validação externa)

## Tratamento de Erros

O módulo retorna códigos de status HTTP apropriados e mensagens de erro descritivas em caso de falha.
