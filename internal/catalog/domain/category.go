package domain

import (
	"context"
	"strings"
	"time"

	"gorm.io/gorm"
)

type Category struct {
	ID            int            `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name          string         `gorm:"column:name;not null" json:"name"`
	Slug          string         `gorm:"column:slug;unique;not null" json:"slug"`
	ParentID      *int           `gorm:"column:parent_id" json:"parent_id,omitempty"`
	IsActive      bool           `gorm:"column:is_active;default:true" json:"is_active"`
	Description   string         `gorm:"column:description" json:"description"`
	CreatedAt     time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at,omitempty"`
	DeletedReason string         `gorm:"column:deleted_reason" json:"deleted_reason"`
}

func (Category) TableName() string { return "categories" }

func (c *Category) Validate() error {
	c.Name = strings.TrimSpace(c.Name)
	c.Slug = strings.TrimSpace(c.Slug)
	c.Description = strings.TrimSpace(c.Description)

	if c.Name == "" {
		return ErrCategoryNameRequired
	}
	if c.Slug == "" {
		return ErrCategorySlugRequired
	}
	if c.Description == "" {
		return ErrCategoryDescriptionRequired
	}

	return nil
}

func (c *Category) UpdateState(newData *Category) {
	if newData.Name != "" {
		c.Name = strings.TrimSpace(newData.Name)
	}
	if newData.Slug != "" {
		c.Slug = strings.TrimSpace(newData.Slug)
	}
	if newData.Description != "" {
		c.Description = strings.TrimSpace(newData.Description)
	}

	if newData.ParentID != nil {
		c.ParentID = newData.ParentID
	}

	c.IsActive = newData.IsActive
}

type CategoryRepository interface {
	List(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]Category, int64, error)
	FindByID(ctx context.Context, id int) (*Category, error)
	FindBySlug(ctx context.Context, slug string) (*Category, error)
	Create(ctx context.Context, category *Category) error
	Update(ctx context.Context, category *Category) error
	Delete(ctx context.Context, id int) error
}

type CategoryService interface {
	ListCategories(ctx context.Context, p Pagination, filters map[string]interface{}) ([]Category, int64, error)
	GetCategoryByID(ctx context.Context, id int) (*Category, error)
	GetCategoryBySlug(ctx context.Context, slug string) (*Category, error)
	CreateCategory(ctx context.Context, category *Category) error
	UpdateCategory(ctx context.Context, id int, category *Category) error
	DeleteCategory(ctx context.Context, id int) error
}
