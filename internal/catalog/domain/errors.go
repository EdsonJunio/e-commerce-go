package domain

import "errors"

var (
	ErrInvalidProductID           = errors.New("invalid product ID")
	ErrProductNotFound            = errors.New("product not found")
	ErrProductNameRequired        = errors.New("product name is required")
	ErrProductSlugRequired        = errors.New("product slug is required")
	ErrInvalidProductPrice        = errors.New("product description is required")
	ErrProductSlugExists          = errors.New("product with this slug already exists")
	ErrProductDescriptionRequired = errors.New("slug is required")
	ErrInvalidCategoryReference   = errors.New("invalid category reference")
)

var (
	ErrInvalidCategoryID           = errors.New("invalid category ID")
	ErrCategoryNotFound            = errors.New("category not found")
	ErrCategoryNameRequired        = errors.New("category name is required")
	ErrCategorySlugRequired        = errors.New("category slug is required")
	ErrCategorySlugExists          = errors.New("category with this slug already exists")
	ErrCategoryDescriptionRequired = errors.New("category description is required")
	ErrParentCategoryRequired      = errors.New("parent category ID is required")
	ErrParentCategoryNotFound      = errors.New("specified parent category was not found")
)
