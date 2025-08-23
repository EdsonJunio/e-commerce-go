# Módulo de Catálogo

Gerencia **Categorias, Produtos, SKUs e Estoque** seguindo **Clean Architecture / DDD** com **Ports & Adapters** (interfaces no domínio, implementações em adapters).

## Arquitetura

- **domain/**: entidades, regras, portas (interfaces) e erros.
- **repository/**: adapters de persistência (GORM).
- **service/**: casos de uso e orquestração de regras.
- **delivery/http/**: handlers e rotas (Gin).

## Estrutura do Projeto

```
internal/catalog/
├── domain/
│   ├── category.go
│   ├── product.go
│   ├── sku.go
│   ├── stock.go
│   ├── ports.go
│   └── errors.go
├── repository/
│   └── gorm/
│       ├── models.go
│       ├── category_repo.go
│       ├── product_repo.go
│       ├── sku_repo.go
│       └── stock_repo.go
├── service/
│   └── catalog_service.go
└── delivery/
    └── http/
        ├── routes.go
        ├── category_handler.go
        ├── product_handler.go
        ├── sku_handler.go
        └── stock_handler.go
```

## Funcionalidades

- **Categorias**: CRUD, hierarquia (`parent_id`), ativação/inativação.
- **Produtos**: CRUD, busca por ID e slug, filtros e paginação.
- **SKUs**: CRUD por produto, preço em centavos, atributos dinâmicos.
- **Estoque**: leitura/atualização, **reserva/liberação atômicas** (rotas internas protegidas).

## Rotas da API

### Categorias

**Criar**
```
POST /v1/catalog/categories
```
Body:
```json
{
  "name": "Eletrônicos",
  "slug": "eletronicos",
  "parent_id": null,
  "is_active": true
}
```

**Listar**
```
GET /v1/catalog/categories?page=1&limit=10&parent_id=&is_active=true
```

**Obter por ID**
```
GET /v1/catalog/categories/{id}
```

**Atualizar**
```
PUT /v1/catalog/categories/{id}
```
Body:
```json
{
  "name": "Smartphones",
  "slug": "smartphones",
  "parent_id": null,
  "is_active": true
}
```

**Deletar**
```
DELETE /v1/catalog/categories/{id}
```

---

### Produtos

**Criar**
```
POST /v1/catalog/products
```
Body:
```json
{
  "category_id": 1,
  "name": "iPhone 14",
  "slug": "iphone-14",
  "description": "128GB",
  "is_active": true
}
```

**Listar**
```
GET /v1/catalog/products?page=1&limit=10&q=iphone&category_id=1&is_active=true&sort=created_at&order=desc
```

**Obter por ID**
```
GET /v1/catalog/products/{id}
```

**Obter por Slug**
```
GET /v1/catalog/products/slug/{slug}
```

**Atualizar**
```
PUT /v1/catalog/products/{id}
```
Body:
```json
{
  "category_id": 1,
  "name": "iPhone 14",
  "slug": "iphone-14",
  "description": "128GB",
  "is_active": true
}
```

**Deletar**
```
DELETE /v1/catalog/products/{id}
```

---

### SKUs

**Criar**
```
POST /v1/catalog/products/{product_id}/skus
```
Body:
```json
{
  "code": "IPH14-128-BLK",
  "price_cents": 459900,
  "attributes": { "color": "black", "storage": "128GB" },
  "is_active": true
}
```

**Listar por Produto**
```
GET /v1/catalog/products/{product_id}/skus
```

**Obter por ID**
```
GET /v1/catalog/skus/{id}
```

**Atualizar**
```
PUT /v1/catalog/skus/{id}
```
Body:
```json
{
  "code": "IPH14-128-BLK",
  "price_cents": 449900,
  "attributes": { "color": "black", "storage": "128GB" },
  "is_active": true
}
```

**Deletar**
```
DELETE /v1/catalog/skus/{id}
```

---

### Estoque

**Obter**
```
GET /v1/catalog/skus/{id}/stock
```

**Atualizar**
```
PUT /v1/catalog/skus/{id}/stock
```
Body:
```json
{ "quantity": 12 }
```

**Rotas internas (protegidas)**
```
POST /internal/stock/reserve
POST /internal/stock/release
```
Reserve:
```json
{ "sku_id": 10, "quantity": 2, "reason": "checkout:order#123" }
```
Release:
```json
{ "sku_id": 10, "quantity": 2, "reason": "payment_failed:order#123" }
```

## Configuração

```go
db, _ := gorm.Open(...)
repo := gormrepo.New(db)
svc := service.NewCatalogService(repo)

r := gin.Default()
cataloghttp.Register(r, svc)
r.Run()
```

## Validações

- `name` obrigatório em **categoria** e **produto**.
- `slug` obrigatório e **único por recurso**.
- `category_id` deve existir para **produto**.
- **SKU**: `code` obrigatório e **único**; `price_cents >= 0`.
- **Estoque**: `quantity >= 0`.
- `parent_id` não pode criar ciclos na hierarquia de categorias.

## Tratamento de Erros

- Códigos HTTP: **400**, **404**, **409**, **422**, **500**.
- Resposta de erro:
```json
{
  "error": {
    "code": "validation_error",
    "message": "slug already exists"
  }
}
```

## Requisitos

- Go 1.20+
- Gin (HTTP)
- GORM (ORM) e driver do banco (ex.: PostgreSQL)

## Boas Práticas

- IDs imutáveis; slugs estáveis.
- Preço sempre em **centavos** (`price_cents` como `int`).
- Padrão de paginação consistente (`page`, `limit`).
- Handlers finos; regras no `service`.
- Idempotência em reservas/liberações de estoque.
- Logs e correlação por `reason` nas rotas internas.
