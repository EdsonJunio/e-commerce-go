ALTER TABLE users
    ADD COLUMN role TEXT NOT NULL DEFAULT 'customer',
    ADD COLUMN status TEXT NOT NULL DEFAULT 'active';

ALTER TABLE users
    ADD CONSTRAINT users_role_check CHECK (role IN ('customer', 'admin')),
    ADD CONSTRAINT users_status_check CHECK (status IN ('active', 'disabled'));

UPDATE users
SET role = 'admin'
WHERE id = 1;

CREATE UNIQUE INDEX users_email_normalized_uq ON users (LOWER(email));
CREATE INDEX users_active_email_idx ON users (LOWER(email)) WHERE deleted_at IS NULL AND status = 'active';
