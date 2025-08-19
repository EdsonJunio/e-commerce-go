package domain

import (
	"time"
)

// Product representa um produto no sistema
type Product struct {
	ID         int        `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	CategoryID int        `gorm:"column:category_id;not null" json:"category_id"`
	Name       string     `gorm:"column:name;not null" json:"name"`
	Slug       string     `gorm:"column:slug;unique;not null" json:"slug"`
	PriceCents int64      `gorm:"column:price_cents;not null" json:"price_cents"`
	IsActive   bool       `gorm:"column:is_active;default:true" json:"is_active"`
	CreatedAt  time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt  *time.Time `gorm:"column:deleted_at;index" json:"deleted_at,omitempty"`
}

// TableName define o nome da tabela no banco de dados
func (Product) TableName() string {
	return "products"
}

// Repository define a interface para operações de banco de dados para Product
type Repository interface {
	Create(product *Product) error
	FindByID(id int) (*Product, error)
	FindBySlug(slug string) (*Product, error)
	Update(product *Product) error
	Delete(id int) error
	List(limit, offset int, filters map[string]interface{}) ([]Product, int64, error)
}

// Service define a interface para a lógica de negócios de Product
type Service interface {
	CreateProduct(product *Product) error
	GetProductByID(id int) (*Product, error)
	GetProductBySlug(slug string) (*Product, error)
	UpdateProduct(id int, product *Product) error
	DeleteProduct(id int) error
	ListProducts(limit, page int, filters map[string]interface{}) ([]Product, int64, error)
}
