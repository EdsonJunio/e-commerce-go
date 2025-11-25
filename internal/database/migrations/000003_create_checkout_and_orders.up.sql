CREATE TABLE IF NOT EXISTS carts
(
    id             BIGSERIAL PRIMARY KEY,
    user_id        BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    created_at     TIMESTAMPTZ DEFAULT now(),
    updated_at     TIMESTAMPTZ DEFAULT now(),
    expires_at     TIMESTAMPTZ,
    deleted_at     TIMESTAMPTZ,
    deleted_reason TEXT
);
CREATE TRIGGER set_carts_updated_at
    BEFORE UPDATE
    ON carts
    FOR EACH ROW
EXECUTE FUNCTION trigger_set_updated_at();
CREATE TRIGGER soft_delete_carts
    BEFORE DELETE
    ON carts
    FOR EACH ROW
EXECUTE FUNCTION trigger_soft_delete();


CREATE TABLE IF NOT EXISTS cart_items
(
    id             BIGSERIAL PRIMARY KEY,
    cart_id        BIGINT NOT NULL REFERENCES carts (id) ON DELETE CASCADE,
    sku_id         BIGINT NOT NULL REFERENCES product_skus (id) ON DELETE RESTRICT,
    quantity       BIGINT NOT NULL CHECK (quantity >= 1),
    created_at     TIMESTAMPTZ DEFAULT now(),
    updated_at     TIMESTAMPTZ DEFAULT now(),
    deleted_at     TIMESTAMPTZ,
    deleted_reason TEXT
);
CREATE TRIGGER set_cart_items_updated_at
    BEFORE UPDATE
    ON cart_items
    FOR EACH ROW
EXECUTE FUNCTION trigger_set_updated_at();


CREATE TABLE IF NOT EXISTS wishlists
(
    id             BIGSERIAL PRIMARY KEY,
    user_id        BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    created_at     TIMESTAMPTZ DEFAULT now(),
    updated_at     TIMESTAMPTZ DEFAULT now(),
    deleted_at     TIMESTAMPTZ,
    deleted_reason TEXT
);
CREATE TRIGGER set_wishlists_updated_at
    BEFORE UPDATE
    ON wishlists
    FOR EACH ROW
EXECUTE FUNCTION trigger_set_updated_at();
CREATE TRIGGER soft_delete_wishlists
    BEFORE DELETE
    ON wishlists
    FOR EACH ROW
EXECUTE FUNCTION trigger_soft_delete();


CREATE TABLE IF NOT EXISTS wishlist_items
(
    id             BIGSERIAL PRIMARY KEY,
    wishlist_id    BIGINT NOT NULL REFERENCES wishlists (id) ON DELETE CASCADE,
    sku_id         BIGINT NOT NULL REFERENCES product_skus (id),
    created_at     TIMESTAMPTZ DEFAULT now(),
    updated_at     TIMESTAMPTZ DEFAULT now(),
    deleted_at     TIMESTAMPTZ,
    deleted_reason TEXT
);
CREATE TRIGGER set_wishlist_items_updated_at
    BEFORE UPDATE
    ON wishlist_items
    FOR EACH ROW
EXECUTE FUNCTION trigger_set_updated_at();


CREATE TABLE IF NOT EXISTS orders
(
    id                  BIGSERIAL PRIMARY KEY,
    user_id             BIGINT NOT NULL REFERENCES users (id),
    shipping_address_id BIGINT REFERENCES addresses (id),
    billing_address_id  BIGINT REFERENCES addresses (id),
    status              TEXT   NOT NULL CHECK (status IN
                                               ('pending', 'awaiting_payment', 'paid', 'failed', 'cancelled', 'expired',
                                                'refunded')),
    total_cents         BIGINT NOT NULL DEFAULT 0 CHECK (total_cents >= 0),
    currency            TEXT            DEFAULT 'BRL',
    created_at          TIMESTAMPTZ     DEFAULT now(),
    updated_at          TIMESTAMPTZ     DEFAULT now(),
    deleted_at          TIMESTAMPTZ,
    deleted_reason      TEXT
);
CREATE TRIGGER set_orders_updated_at
    BEFORE UPDATE
    ON orders
    FOR EACH ROW
EXECUTE FUNCTION trigger_set_updated_at();
CREATE TRIGGER soft_delete_orders
    BEFORE DELETE
    ON orders
    FOR EACH ROW
EXECUTE FUNCTION trigger_soft_delete();


CREATE TABLE IF NOT EXISTS order_items
(
    id             BIGSERIAL PRIMARY KEY,
    order_id       BIGINT NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    sku_id         BIGINT NOT NULL REFERENCES product_skus (id),
    quantity       BIGINT NOT NULL CHECK (quantity >= 1),
    price_cents    BIGINT NOT NULL CHECK (price_cents >= 0),
    created_at     TIMESTAMPTZ DEFAULT now(),
    updated_at     TIMESTAMPTZ DEFAULT now(),
    deleted_at     TIMESTAMPTZ,
    deleted_reason TEXT
);
CREATE TRIGGER set_order_items_updated_at
    BEFORE UPDATE
    ON order_items
    FOR EACH ROW
EXECUTE FUNCTION trigger_set_updated_at();


CREATE TABLE IF NOT EXISTS shipments
(
    id                    BIGSERIAL PRIMARY KEY,
    order_id              BIGINT NOT NULL REFERENCES orders (id),
    provider              TEXT,
    tracking_code         TEXT,
    status                TEXT   NOT NULL CHECK (status IN ('pending', 'in_transit', 'delivered', 'returned')),
    shipment_cost_cents   BIGINT,
    estimated_delivery_at TIMESTAMPTZ,
    shipped_at            TIMESTAMPTZ,
    delivered_at          TIMESTAMPTZ,
    return_reason         TEXT,
    weight_grams          BIGINT,
    height_cm             INT,
    width_cm              INT,
    length_cm             INT,
    created_at            TIMESTAMPTZ DEFAULT now(),
    updated_at            TIMESTAMPTZ DEFAULT now(),
    deleted_at            TIMESTAMPTZ,
    deleted_reason        TEXT
);
CREATE TRIGGER set_shipments_updated_at
    BEFORE UPDATE
    ON shipments
    FOR EACH ROW
EXECUTE FUNCTION trigger_set_updated_at();
CREATE TRIGGER soft_delete_shipments
    BEFORE DELETE
    ON shipments
    FOR EACH ROW
EXECUTE FUNCTION trigger_soft_delete();
