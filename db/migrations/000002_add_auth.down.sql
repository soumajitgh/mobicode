DROP INDEX IF EXISTS idx_refresh_tokens_user_id;
DROP TABLE refresh_tokens;
ALTER TABLE users DROP COLUMN password_hash;
