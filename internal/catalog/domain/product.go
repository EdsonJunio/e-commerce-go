package domain

import "time"

// Product represents the product aggregate persisted in the database.
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

// TableName sets the table name for GORM.
func (Product) TableName() string {
	return "products"
}

// ProductRepository defines the DB access contract for Product.
type ProductRepository interface {
	List(limit, offset int, filters map[string]interface{}) ([]Product, int64, error)
	FindByID(id int) (*Product, error)
	FindBySlug(slug string) (*Product, error)
	Create(product *Product) error
	Update(product *Product) error
	Delete(id int) error
}

// ProductService defines the business logic contract for Product.
type ProductService interface {
	ListProducts(p Pagination, filters map[string]interface{}) ([]Product, int64, error)
	GetProductByID(id int) (*Product, error)
	GetProductBySlug(slug string) (*Product, error)
	CreateProduct(product *Product) error
	UpdateProduct(id int, product *Product) error
	DeleteProduct(id int) error
}
