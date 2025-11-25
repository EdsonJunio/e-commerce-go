CREATE TABLE IF NOT EXISTS categories
(
    id             BIGSERIAL PRIMARY KEY,
    name           TEXT   NOT NULL,
    slug           TEXT   NOT NULL UNIQUE,
    parent_id      BIGINT REFERENCES categories (id) ON DELETE SET NULL,
    is_active      BOOLEAN     DEFAULT true,
    description    TEXT,
    created_at     TIMESTAMPTZ DEFAULT now(),
    updated_at     TIMESTAMPTZ DEFAULT now(),
    deleted_at     TIMESTAMPTZ,
    deleted_reason TEXT
);
CREATE TRIGGER set_categories_updated_at
    BEFORE UPDATE
    ON categories
    FOR EACH ROW
EXECUTE FUNCTION trigger_set_updated_at();
CREATE TRIGGER soft_delete_categories
    BEFORE DELETE
    ON categories
    FOR EACH ROW
EXECUTE FUNCTION trigger_soft_delete();


CREATE TABLE IF NOT EXISTS products
(
    id              BIGSERIAL PRIMARY KEY,
    name            TEXT   NOT NULL,
    slug            TEXT   NOT NULL UNIQUE,
    description     TEXT,
    seo_title       TEXT,
    seo_description TEXT,
    category_id     BIGINT REFERENCES categories (id) ON DELETE SET NULL,
    is_active       BOOLEAN     DEFAULT true,
    created_at      TIMESTAMPTZ DEFAULT now(),
    updated_at      TIMESTAMPTZ DEFAULT now(),
    deleted_at      TIMESTAMPTZ,
    deleted_reason  TEXT
);
CREATE TRIGGER set_products_updated_at
    BEFORE UPDATE
    ON products
    FOR EACH ROW
EXECUTE FUNCTION trigger_set_updated_at();
CREATE TRIGGER soft_delete_products
    BEFORE DELETE
    ON products
    FOR EACH ROW
EXECUTE FUNCTION trigger_soft_delete();


CREATE TABLE IF NOT EXISTS product_skus
(
    id             BIGSERIAL PRIMARY KEY,
    product_id     BIGINT NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    sku_code       TEXT   NOT NULL UNIQUE,
    barcode        TEXT,
    price_cents    BIGINT NOT NULL CHECK (price_cents > 0),
    attributes     JSONB,
    is_active      BOOLEAN     DEFAULT true,
    created_at     TIMESTAMPTZ DEFAULT now(),
    updated_at     TIMESTAMPTZ DEFAULT now(),
    deleted_at     TIMESTAMPTZ,
    deleted_reason TEXT
);
CREATE TRIGGER set_product_skus_updated_at
    BEFORE UPDATE
    ON product_skus
    FOR EACH ROW
EXECUTE FUNCTION trigger_set_updated_at();
CREATE TRIGGER soft_delete_product_skus
    BEFORE DELETE
    ON product_skus
    FOR EACH ROW
EXECUTE FUNCTION trigger_soft_delete();


CREATE TABLE IF NOT EXISTS price_history
(
    id          BIGSERIAL PRIMARY KEY,
    sku_id      BIGINT NOT NULL REFERENCES product_skus (id) ON DELETE CASCADE,
    price_cents BIGINT NOT NULL CHECK (price_cents > 0),
    changed_by  BIGINT,
    changed_at  TIMESTAMPTZ DEFAULT now()
);


CREATE TABLE IF NOT EXISTS stock
(
    sku_id            BIGINT PRIMARY KEY REFERENCES product_skus (id) ON DELETE CASCADE,
    quantity          BIGINT NOT NULL CHECK (quantity >= 0),
    reserved_quantity BIGINT NOT NULL DEFAULT 0 CHECK (reserved_quantity >= 0),
    updated_at        TIMESTAMPTZ     DEFAULT now(),
    deleted_at        TIMESTAMPTZ,
    deleted_reason    TEXT
);
CREATE TRIGGER set_stock_updated_at
    BEFORE UPDATE
    ON stock
    FOR EACH ROW
EXECUTE FUNCTION trigger_set_updated_at();


CREATE TABLE IF NOT EXISTS stock_movements
(
    id           BIGSERIAL PRIMARY KEY,
    sku_id       BIGINT NOT NULL REFERENCES product_skus (id) ON DELETE CASCADE,
    change       BIGINT NOT NULL,
    reason       TEXT   NOT NULL,
    reference_id BIGINT,
    created_by   BIGINT,
    created_at   TIMESTAMPTZ DEFAULT now()
);