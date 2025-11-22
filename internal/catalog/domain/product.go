package domain

import (
	"context"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Product represents the product aggregate persisted in the database.
type Product struct {
	ID             int            `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name           string         `gorm:"column:name;not null" json:"name"`
	Slug           string         `gorm:"column:slug;unique;not null" json:"slug"`
	Description    string         `gorm:"column:description" json:"description"`
	SeoTitle       string         `gorm:"column:seo_title" json:"seo_title"`
	SeoDescription string         `gorm:"column:seo_description" json:"seo_description"`
	CategoryID     *int           `gorm:"column:category_id" json:"category_id"`
	IsActive       bool           `gorm:"column:is_active;default:true" json:"is_active"`
	CreatedAt      time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at,omitempty"`
	DeletedReason  string         `gorm:"column:deleted_reason" json:"deleted_reason"`
}

// TableName sets the table name for GORM.
func (Product) TableName() string {
	return "products"
}

func (p *Product) Validate() error {
	p.Name = strings.TrimSpace(p.Name)
	p.Slug = strings.TrimSpace(p.Slug)
	p.Description = strings.TrimSpace(p.Description)
	p.SeoTitle = strings.TrimSpace(p.SeoTitle)
	p.SeoDescription = strings.TrimSpace(p.SeoDescription)

	if p.Name == "" {
		return ErrProductNameRequired
	}
	if p.Slug == "" {
		return ErrProductSlugRequired
	}
	if p.Description == "" {
		return ErrProductDescriptionRequired
	}
	if p.SeoTitle == "" {
		return ErrSeoTitle
	}
	if p.SeoDescription == "" {
		return ErrSeoDescription
	}

	if p.CategoryID == nil || *p.CategoryID <= 0 {
		return ErrInvalidCategoryReference
	}

	return nil
}

func (p *Product) UpdateState(newData *Product) {
	if newData.Name != "" {
		p.Name = strings.TrimSpace(newData.Name)
	}
	if newData.Slug != "" {
		p.Slug = strings.TrimSpace(newData.Slug)
	}
	if newData.Description != "" {
		p.Description = strings.TrimSpace(newData.Description)
	}
	if newData.SeoTitle != "" {
		p.SeoTitle = strings.TrimSpace(newData.SeoTitle)
	}
	if newData.SeoDescription != "" {
		p.SeoDescription = strings.TrimSpace(newData.SeoDescription)
	}

	if newData.CategoryID != nil && *newData.CategoryID > 0 {
		p.CategoryID = newData.CategoryID
	}

	p.IsActive = newData.IsActive
}

type ProductRepository interface {
	List(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]Product, int64, error)
	FindByID(ctx context.Context, id int) (*Product, error)
	FindBySlug(ctx context.Context, slug string) (*Product, error)
	Create(ctx context.Context, product *Product) error
	Update(ctx context.Context, product *Product) error
	Delete(ctx context.Context, id int) error
}

// ProductService defines the business logic contract for Product.
type ProductService interface {
	ListProducts(ctx context.Context, p Pagination, filters map[string]interface{}) ([]Product, int64, error)
	GetProductByID(ctx context.Context, id int) (*Product, error)
	GetProductBySlug(ctx context.Context, slug string) (*Product, error)
	CreateProduct(ctx context.Context, product *Product) error
	UpdateProduct(ctx context.Context, id int, product *Product) error
	DeleteProduct(ctx context.Context, id int) error
}
