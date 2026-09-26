package repository

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"e-commerce-go/internal/catalog/domain"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestProductSkuRepositoryListUsesMigratedSchema(t *testing.T) {
	databaseURL := os.Getenv("SKU_REPOSITORY_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("SKU_REPOSITORY_TEST_DATABASE_URL is required for the PostgreSQL repository test")
	}

	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		t.Fatalf("open PostgreSQL test database: %v", err)
	}

	tx := db.Begin()
	if tx.Error != nil {
		t.Fatalf("begin transaction: %v", tx.Error)
	}
	defer tx.Rollback()

	productID, suffix := insertProductSKUFixtures(t, tx)
	repository := NewProductSkuRepository(tx)

	skus, total, err := repository.List(context.Background(), 2, 1, map[string]interface{}{})
	if err != nil {
		t.Fatalf("list product SKUs through migrated schema: %v", err)
	}
	if total != 3 {
		t.Errorf("total = %d, want 3", total)
	}
	if len(skus) != 2 {
		t.Fatalf("SKU count = %d, want 2", len(skus))
	}

	wantCodes := []string{"inactive-middle-" + suffix, "active-oldest-" + suffix}
	for index, wantCode := range wantCodes {
		if skus[index].SkuCode != wantCode {
			t.Errorf("SKU[%d].SkuCode = %q, want %q", index, skus[index].SkuCode, wantCode)
		}
		if skus[index].ProductID != productID {
			t.Errorf("SKU[%d].ProductID = %d, want %d", index, skus[index].ProductID, productID)
		}
		if skus[index].ID == 0 || skus[index].CreatedAt.IsZero() || skus[index].UpdatedAt.IsZero() {
			t.Errorf("SKU[%d] missing migrated identity or timestamps: %+v", index, skus[index])
		}
	}

	if skus[0].BarCode != "inactive-barcode" || skus[0].PriceCents != 2200 || skus[0].IsActive {
		t.Errorf("inactive SKU scalar fields = %+v", skus[0])
	}
	if got := skus[0].Attributes["color"]; got != "blue" {
		t.Errorf("inactive SKU color = %#v, want %q", got, "blue")
	}
	if skus[1].BarCode != "active-barcode" || skus[1].PriceCents != 1100 || !skus[1].IsActive {
		t.Errorf("active SKU scalar fields = %+v", skus[1])
	}
	if got := skus[1].Attributes["size"]; got != "M" {
		t.Errorf("active SKU size = %#v, want %q", got, "M")
	}
}

func insertProductSKUFixtures(t *testing.T, tx *gorm.DB) (int, string) {
	t.Helper()
	suffix := time.Now().UTC().Format("20060102150405.000000000")
	category := domain.Category{
		Name:        "SKU mapping category",
		Slug:        "sku-mapping-category-" + suffix,
		Description: "SKU mapping category",
		IsActive:    true,
	}
	if err := tx.Create(&category).Error; err != nil {
		t.Fatalf("insert SKU category fixture: %v", err)
	}
	product := domain.Product{
		Name:       "SKU mapping product",
		Slug:       "sku-mapping-product-" + suffix,
		CategoryID: &category.ID,
		IsActive:   true,
	}
	if err := tx.Create(&product).Error; err != nil {
		t.Fatalf("insert SKU product fixture: %v", err)
	}

	fixtures := []struct {
		code, barcode, attributes string
		price                     int64
		active                    bool
	}{
		{"active-oldest", "active-barcode", `{"size":"M"}`, 1100, true},
		{"inactive-middle", "inactive-barcode", `{"color":"blue"}`, 2200, false},
		{"active-newest", "newest-barcode", `{"material":"cotton"}`, 3300, true},
	}
	for index, fixture := range fixtures {
		code := fmt.Sprintf("%s-%s", fixture.code, suffix)
		if err := tx.Exec(
			`INSERT INTO product_skus (product_id, sku_code, barcode, price_cents, attributes, is_active) VALUES (?, ?, ?, ?, ?::jsonb, ?)`,
			product.ID,
			code,
			fixture.barcode,
			fixture.price,
			fixture.attributes,
			fixture.active,
		).Error; err != nil {
			t.Fatalf("insert SKU fixture %d: %v", index, err)
		}
	}

	return product.ID, suffix
}
