package domain

import "time"

// Category represents the category aggregate persisted in the database.
type Category struct {
	ID            int        `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name          string     `gorm:"column:name;not null" json:"name"`
	Slug          string     `gorm:"column:slug;unique;not null" json:"slug"`
	ParentID      *int       `gorm:"column:parent_id" json:"parent_id,omitempty"`
	IsActive      bool       `gorm:"column:is_active;default:true" json:"is_active"`
	Description   string     `gorm:"column:description" json:"description"`
	CreatedAt     time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt     *time.Time `gorm:"column:deleted_at;index" json:"deleted_at,omitempty"`
	DeletedReason string     `gorm:"column:deleted_reason" json:"deleted_reason"`
}

// TableName sets the table name for GORM.
func (Category) TableName() string { return "categories" }

// CategoryRepository defines the DB acess contract for Category
type CategoryRepository interface {
	List(limit, offset int, filters map[string]interface{}) ([]Category, int64, error)
	FindByID(id int) (*Category, error)
	FindBySlug(slug string) (*Category, error)
	Create(category *Category) error
	Update(category *Category) error
	Delete(id int) error
}

// CategoryService defines the business logic contract for Category
type CategoryService interface {
	ListCategories(p Pagination, filters map[string]interface{}) ([]Category, int64, error)
	GetCategoryByID(id int) (*Category, error)
	GetCategoryBySlug(slug string) (*Category, error)
	CreateCategory(category *Category) error
	UpdateCategory(id int, category *Category) error
	DeleteCategory(id int) error
}
