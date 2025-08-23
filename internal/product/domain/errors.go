package domain

import "errors"

var (
	ErrInvalidID    = errors.New("invalid product ID")
	ErrNotFound     = errors.New("product not found")
	ErrNameReq      = errors.New("product name is required")
	ErrSlugReq      = errors.New("product slug is required")
	ErrPriceInvalid = errors.New("product price must be greater than zero")
	ErrSlugExists   = errors.New("product with this slug already exists")
	ErrSlugIsReq    = errors.New("slug is required")
)
