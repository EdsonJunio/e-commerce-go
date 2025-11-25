CREATE OR REPLACE FUNCTION trigger_set_updated_at()
    RETURNS TRIGGER
    LANGUAGE plpgsql AS
$$
BEGIN
    NEW.updated_at := now();
    RETURN NEW;
END;
$$;

CREATE OR REPLACE FUNCTION trigger_soft_delete()
    RETURNS TRIGGER
    LANGUAGE plpgsql AS
$$
DECLARE
    _reason TEXT := NULL;
BEGIN
    BEGIN
        _reason := current_setting('app.delete_reason', true);
    EXCEPTION
        WHEN OTHERS THEN
            _reason := NULL;
    END;

    IF _reason IS NULL AND array_length(TG_ARGV, 1) >= 1 THEN
        _reason := TG_ARGV[0];
    END IF;

    IF _reason IS NULL THEN
        _reason := 'deleted';
    END IF;

    EXECUTE format(
            'UPDATE %I SET deleted_at = now(), deleted_reason = $1 WHERE id = $2',
            TG_TABLE_NAME
            ) USING _reason, OLD.id;

    RETURN NULL;
END;
$$;

CREATE TABLE IF NOT EXISTS users
(
    id             BIGSERIAL PRIMARY KEY,
    email          TEXT NOT NULL UNIQUE,
    password_hash  TEXT NOT NULL,
    full_name      TEXT,
    phone          TEXT,
    created_at     TIMESTAMPTZ DEFAULT now(),
    updated_at     TIMESTAMPTZ DEFAULT now(),
    deleted_at     TIMESTAMPTZ,
    deleted_reason TEXT,
    deleted_by     BIGINT
);

CREATE TRIGGER set_users_updated_at
    BEFORE UPDATE
    ON users
    FOR EACH ROW
EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER soft_delete_users
    BEFORE DELETE
    ON users
    FOR EACH ROW
EXECUTE FUNCTION trigger_soft_delete();

CREATE TABLE IF NOT EXISTS addresses
(
    id             BIGSERIAL PRIMARY KEY,
    user_id        BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    address_type   TEXT   NOT NULL DEFAULT 'shipping'
        CHECK (address_type IN ('shipping', 'billing', 'other')),
    line1          TEXT   NOT NULL,
    line2          TEXT,
    city           TEXT   NOT NULL,
    state          TEXT   NOT NULL,
    zip            TEXT   NOT NULL,
    country        TEXT   NOT NULL,
    is_default     BOOLEAN         DEFAULT false,
    created_at     TIMESTAMPTZ     DEFAULT now(),
    updated_at     TIMESTAMPTZ     DEFAULT now(),
    deleted_at     TIMESTAMPTZ,
    deleted_reason TEXT
);

CREATE TRIGGER set_addresses_updated_at
    BEFORE UPDATE
    ON addresses
    FOR EACH ROW
EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER soft_delete_addresses
    BEFORE DELETE
    ON addresses
    FOR EACH ROW
EXECUTE FUNCTION trigger_soft_delete();
