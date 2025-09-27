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
)

var (
	ErrInvalidCategoryID           = errors.New("invalid category ID")
	ErrCategoryNotFound            = errors.New("category not found")
	ErrCategoryNameRequired        = errors.New("category name is required")
	ErrCategorySlugRequired        = errors.New("category slug is required")
	ErrCategorySlugExists          = errors.New("category with this slug already exists")
	ErrCategoryDescriptionRequired = errors.New("slug is required")
)
