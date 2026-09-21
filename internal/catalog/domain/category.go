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

type CategoryListFilters struct {
	ParentID *int
	IsActive *bool
}

type CategoryChanges struct {
	Name        *string
	Slug        *string
	Description *string
	ParentID    *int
	ClearParent bool
	IsActive    *bool
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

func (c *Category) UpdateState(changes CategoryChanges) {
	if changes.Name != nil {
		c.Name = strings.TrimSpace(*changes.Name)
	}
	if changes.Slug != nil {
		c.Slug = strings.TrimSpace(*changes.Slug)
	}
	if changes.Description != nil {
		c.Description = strings.TrimSpace(*changes.Description)
	}
	if changes.ClearParent {
		c.ParentID = nil
	} else if changes.ParentID != nil {
		c.ParentID = changes.ParentID
	}
	if changes.IsActive != nil {
		c.IsActive = *changes.IsActive
	}
}

type CategoryRepository interface {
	List(ctx context.Context, limit, offset int, filters CategoryListFilters) ([]Category, int64, error)
	FindByID(ctx context.Context, id int) (*Category, error)
	FindParentByID(ctx context.Context, id int) (*int, error)
	FindBySlug(ctx context.Context, slug string) (*Category, error)
	Create(ctx context.Context, category *Category) error
	Update(ctx context.Context, category *Category) error
	Delete(ctx context.Context, id int) error
}

type CategoryService interface {
	ListCategories(ctx context.Context, p Pagination, filters CategoryListFilters) ([]Category, int64, error)
	GetCategoryByID(ctx context.Context, id int) (*Category, error)
	GetCategoryBySlug(ctx context.Context, slug string) (*Category, error)
	CreateCategory(ctx context.Context, category *Category) error
	UpdateCategory(ctx context.Context, id int, changes CategoryChanges) error
	DeleteCategory(ctx context.Context, id int) error
}
