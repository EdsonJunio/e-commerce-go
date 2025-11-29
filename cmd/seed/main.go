package main

import (
	"database/sql"
	"e-commerce-go/internal/shared/security"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Constants to control data volume
const (
	TotalUsers      = 50
	TotalCategories = 8
	TotalProducts   = 100
	TotalOrders     = 200
)

func main() {
	// Initialize random seed
	gofakeit.Seed(time.Now().UnixNano())

	// Connection string
	connStr := "postgres://postgres:1234@localhost:5432/ecommerce?sslmode=disable"
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Unable to ping database: %v\n", err)
	}

	fmt.Println("Starting full database seed...")

	// 1. Clean existing data
	cleanTables(db)

	// 2. Base Data (Users & Categories)
	userIDs := seedUsers(db)
	addressIDs := seedAddresses(db, userIDs) // Create addresses for users
	categoryIDs := seedCategories(db)

	// 3. Catalog (Products, SKUs, Stock, Prices)
	skuIDs := seedCatalog(db, categoryIDs)

	// 4. Shopping Experience (Carts, Wishlists)
	seedCarts(db, userIDs, skuIDs)
	seedWishlists(db, userIDs, skuIDs)

	// 5. Sales (Orders, Items, Shipments)
	seedOrdersAndShipments(db, userIDs, addressIDs, skuIDs)

	fmt.Println("Database seeded successfully with full dataset!")
}

func cleanTables(db *sql.DB) {
	fmt.Println("Cleaning tables...")
	tables := []string{
		"shipments", "order_items", "orders",
		"wishlist_items", "wishlists",
		"cart_items", "carts",
		"stock_movements", "stock", "price_history",
		"product_skus", "products", "categories",
		"addresses", "users",
	}

	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table))
		if err != nil {
			log.Printf("Warning: Could not truncate %s: %v", table, err)
		}
	}
}

// --- USERS & ADDRESSES ---

func seedUsers(db *sql.DB) []int {
	fmt.Printf("Seeding %d users...\n", TotalUsers)
	var ids []int

	// 1. Create fixed Admin user
	var adminID int

	// CRIPTOGRAFA A SENHA ANTES DE SALVAR
	adminPass, _ := security.HashPassword("hash123") // A senha para logar será "hash123"

	query := `INSERT INTO users (email, password_hash, full_name, phone) VALUES ($1, $2, $3, $4) RETURNING id`

	// Note que passamos 'adminPass' (o hash) e não a string pura
	err := db.QueryRow(query, "admin@gmail.com", adminPass, "Admin User", "+551199999999").Scan(&adminID)
	if err != nil {
		log.Fatalf("Failed to create admin: %v", err)
	}
	ids = append(ids, adminID)

	// 2. Create fake users
	query = `INSERT INTO users (email, password_hash, full_name, phone) VALUES ($1, $2, $3, $4) RETURNING id`
	for i := 0; i < TotalUsers; i++ {
		var id int
		email := gofakeit.Email()

		// Criptografa a senha fake também
		fakePass, _ := security.HashPassword("123456") // Todos os users fakes terão senha "123456"

		name := gofakeit.Name()
		phone := gofakeit.Phone()

		err := db.QueryRow(query, email, fakePass, name, phone).Scan(&id)
		if err != nil {
			log.Printf("Error creating user: %v", err)
			continue
		}
		ids = append(ids, id)
	}
	return ids
}

func seedAddresses(db *sql.DB, userIDs []int) map[int][]int {
	fmt.Println("Seeding addresses...")
	userAddresses := make(map[int][]int) // Map UserID -> []AddressID
	query := `
		INSERT INTO addresses (user_id, address_type, line1, city, state, zip, country, is_default)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`

	types := []string{"shipping", "billing", "other"}

	for _, userID := range userIDs {
		// Create 1 to 3 addresses per user
		count := gofakeit.Number(1, 3)
		for i := 0; i < count; i++ {
			var id int
			addrType := types[rand.Intn(len(types))]
			isDefault := i == 0 // First one is default

			err := db.QueryRow(query,
				userID, addrType, gofakeit.Street(), gofakeit.City(),
				gofakeit.StateAbr(), gofakeit.Zip(), gofakeit.Country(), isDefault,
			).Scan(&id)

			if err == nil {
				userAddresses[userID] = append(userAddresses[userID], id)
			}
		}
	}
	return userAddresses
}

// --- CATALOG ---

func seedCategories(db *sql.DB) []int {
	fmt.Printf("Seeding %d categories...\n", TotalCategories)
	var ids []int
	query := `INSERT INTO categories (name, slug, description) VALUES ($1, $2, $3) RETURNING id`

	for i := 0; i < TotalCategories; i++ {
		var id int
		name := gofakeit.Fruit() + " " + gofakeit.Adjective()
		slug := gofakeit.UUID()
		db.QueryRow(query, name, slug, gofakeit.Sentence(5)).Scan(&id)
		ids = append(ids, id)
	}
	return ids
}

func seedCatalog(db *sql.DB, categoryIDs []int) []int {
	fmt.Printf("Seeding %d products with SKUs, Stock and Price History...\n", TotalProducts)
	var allSkuIDs []int

	queryProd := `INSERT INTO products (name, slug, description, category_id) VALUES ($1, $2, $3, $4) RETURNING id`
	querySKU := `INSERT INTO product_skus (product_id, sku_code, barcode, price_cents, attributes) VALUES ($1, $2, $3, $4, $5) RETURNING id`
	queryStock := `INSERT INTO stock (sku_id, quantity, reserved_quantity) VALUES ($1, $2, 0)`
	queryStockMov := `INSERT INTO stock_movements (sku_id, change, reason, created_by) VALUES ($1, $2, 'initial_stock', 1)`
	queryPrice := `INSERT INTO price_history (sku_id, price_cents, changed_by) VALUES ($1, $2, 1)`

	for i := 0; i < TotalProducts; i++ {
		catID := categoryIDs[rand.Intn(len(categoryIDs))]
		var prodID int

		prodName := gofakeit.ProductName()
		slug := gofakeit.UUID()

		err := db.QueryRow(queryProd, prodName, slug, gofakeit.ProductDescription(), catID).Scan(&prodID)
		if err != nil {
			continue
		}

		// Create 1 to 4 SKUs per product
		skuCount := gofakeit.Number(1, 4)
		for j := 0; j < skuCount; j++ {
			var skuID int
			skuCode := gofakeit.LetterN(4) + "-" + gofakeit.DigitN(4)
			barcode := gofakeit.DigitN(12)
			price := int64(gofakeit.Price(10, 500)) * 100 // Cents

			// Attributes JSON
			attrs := map[string]string{
				"color": gofakeit.Color(),
				"size":  []string{"S", "M", "L", "XL"}[rand.Intn(4)],
			}
			attrsJSON, _ := json.Marshal(attrs)

			err := db.QueryRow(querySKU, prodID, skuCode, barcode, price, attrsJSON).Scan(&skuID)
			if err == nil {
				allSkuIDs = append(allSkuIDs, skuID)

				// Initial Stock
				qty := gofakeit.Number(0, 100)
				db.Exec(queryStock, skuID, qty)
				db.Exec(queryStockMov, skuID, qty)

				// Price History
				db.Exec(queryPrice, skuID, price)
			}
		}
	}
	return allSkuIDs
}

// --- SHOPPING ---

func seedCarts(db *sql.DB, userIDs []int, skuIDs []int) {
	fmt.Println("Seeding carts...")
	queryCart := `INSERT INTO carts (user_id) VALUES ($1) RETURNING id`
	queryItem := `INSERT INTO cart_items (cart_id, sku_id, quantity) VALUES ($1, $2, $3)`

	// Create carts for 50% of users
	for _, uid := range userIDs {
		if gofakeit.Bool() {
			var cartID int
			db.QueryRow(queryCart, uid).Scan(&cartID)

			// Add random items
			itemCount := gofakeit.Number(1, 5)
			for i := 0; i < itemCount; i++ {
				skuID := skuIDs[rand.Intn(len(skuIDs))]
				db.Exec(queryItem, cartID, skuID, gofakeit.Number(1, 3))
			}
		}
	}
}

func seedWishlists(db *sql.DB, userIDs []int, skuIDs []int) {
	fmt.Println("Seeding wishlists...")
	queryList := `INSERT INTO wishlists (user_id) VALUES ($1) RETURNING id`
	queryItem := `INSERT INTO wishlist_items (wishlist_id, sku_id) VALUES ($1, $2)`

	for _, uid := range userIDs {
		if gofakeit.Bool() {
			var listID int
			db.QueryRow(queryList, uid).Scan(&listID)

			itemCount := gofakeit.Number(1, 5)
			for i := 0; i < itemCount; i++ {
				skuID := skuIDs[rand.Intn(len(skuIDs))]
				db.Exec(queryItem, listID, skuID)
			}
		}
	}
}

// --- ORDERS & SHIPMENTS ---

func seedOrdersAndShipments(db *sql.DB, userIDs []int, userAddresses map[int][]int, skuIDs []int) {
	fmt.Printf("Seeding %d orders with history...\n", TotalOrders)

	queryOrder := `
		INSERT INTO orders (user_id, shipping_address_id, billing_address_id, status, total_cents)
		VALUES ($1, $2, $2, $3, 0) RETURNING id`

	queryItem := `
		INSERT INTO order_items (order_id, sku_id, quantity, price_cents)
		VALUES ($1, $2, $3, $4)`

	queryShipment := `
		INSERT INTO shipments (order_id, status, tracking_code, provider)
		VALUES ($1, $2, $3, $4)`

	statuses := []string{"pending", "awaiting_payment", "paid", "failed", "cancelled", "expired", "refunded"}

	for i := 0; i < TotalOrders; i++ {
		userID := userIDs[rand.Intn(len(userIDs))]

		// Get User Address
		addrs, ok := userAddresses[userID]
		if !ok || len(addrs) == 0 {
			continue
		}
		addrID := addrs[0]

		status := statuses[rand.Intn(len(statuses))]
		var orderID int

		// Create Order (Total 0 initially)
		err := db.QueryRow(queryOrder, userID, addrID, status).Scan(&orderID)
		if err != nil {
			log.Printf("Error creating order: %v", err)
			continue
		}

		// Add Items and Calc Total
		itemCount := gofakeit.Number(1, 5)
		var totalCents int64

		for j := 0; j < itemCount; j++ {
			skuID := skuIDs[rand.Intn(len(skuIDs))]
			qty := int64(gofakeit.Number(1, 3))
			price := int64(gofakeit.Price(10, 200)) * 100 // Mock price

			_, err := db.Exec(queryItem, orderID, skuID, qty, price)
			if err == nil {
				totalCents += price * qty
			}
		}

		// Update Order Total
		db.Exec("UPDATE orders SET total_cents = $1 WHERE id = $2", totalCents, orderID)

		// Create Shipment if Paid
		if status == "paid" || status == "shipped" {
			shipStatus := "pending"
			if gofakeit.Bool() {
				shipStatus = "delivered"
			}

			db.Exec(queryShipment, orderID, shipStatus, gofakeit.UUID(), "DHL")
		}
	}
}
