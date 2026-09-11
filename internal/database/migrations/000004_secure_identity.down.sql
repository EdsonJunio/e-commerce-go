DROP INDEX IF EXISTS users_active_email_idx;
DROP INDEX IF EXISTS users_email_normalized_uq;

ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_status_check,
    DROP CONSTRAINT IF EXISTS users_role_check,
    DROP COLUMN IF EXISTS status,
    DROP COLUMN IF EXISTS role;
