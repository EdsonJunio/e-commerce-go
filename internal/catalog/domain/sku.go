package domain

import (
	"time"

	"gorm.io/gorm"
)

type Sku struct {
	ID            int                    `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ProductID     int                    `gorm:"column:product_id;not null;index" json:"product_id"`
	SkuCode       string                 `gorm:"column:sku_code;unique;not null" json:"sku_code"`
	BarCode       string                 `gorm:"column:barcode" json:"barcode,omitempty"`
	PriceCents    int64                  `gorm:"column:price_cents;not null;check:price_cents > 0" json:"price_cents"`
	Stock         int                    `gorm:"column:stock;not null;default:0;check:stock >= 0" json:"stock"`
	Attributes    map[string]interface{} `gorm:"type:jsonb" json:"attributes,omitempty"`
	IsActive      bool                   `gorm:"column:is_active;not null;default:true" json:"is_active"`
	CreatedAt     time.Time              `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time              `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt     gorm.DeletedAt         `gorm:"column:deleted_at;index" json:"deleted_at,omitempty"`
	DeletedReason string                 `gorm:"column:deleted_reason" json:"deleted_reason,omitempty"`
}

type SkuRepository interface {
}

type SkuService interface {
}
